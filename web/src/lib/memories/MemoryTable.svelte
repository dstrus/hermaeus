<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { SvelteSet } from 'svelte/reactivity';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Memory, MemoryChanges } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import { m } from '#lib/copy/messages.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import {
		addMemory,
		deleteMemories,
		memoriesDependency,
		moveToCurrentIdentity,
		savedNote,
		updateMemories,
		updateMemory
	} from './api.ts';
	import { dayLabel, timeLabel } from './dates.ts';
	import MemoryForm from './MemoryForm.svelte';

	import type { Character } from '#lib/api/types/characters.ts';
	import { characterName } from '#lib/characters/api.ts';

	let {
		save,
		uuid,
		memories,
		characters = []
	}: {
		save: string;
		uuid?: string; // whose memories these are; everyone's when left out
		memories: Memory[]; // newest first
		characters?: Character[]; // for narrowing everyone's to one character's
	} = $props();
	let actor = $state('');
	const owner = (m: Memory) => (m.actorUuid ? `/${encodeURIComponent(save)}/npc/${m.actorUuid}/memories` : undefined);

	let readOnly = $derived(live.gameRunning);
	let query = $state('');
	let type = $state('');
	let mood = $state('');
	let sort = $state<Sort>({ key: 'when', descending: true });

	let types = $derived([...new Set(memories.map((m) => m.type).filter(Boolean))].sort());
	let moods = $derived([...new Set(memories.map((m) => m.emotion).filter(Boolean))].sort());

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = memories.filter(
			(m) =>
				(!type || m.type === type) &&
				(!mood || m.emotion === mood) &&
				(!actor || m.actorUuid === actor) &&
				(!q || [m.content, m.title, m.location, m.emotion, ...m.tags].some((v) => v.toLowerCase().includes(q)))
		);
		return sorted(list, sort, {
			when: (m) => m.id,
			type: (m) => m.type,
			who: (m) => m.actor || '￿',
			mood: (m) => m.emotion || '￿',
			importance: (m) => m.importance
		});
	});
	let byDay = $derived(sort.key === 'when');

	const selection = new Selection(() => shown, (m) => String(m.id), { param: 'item' });
	let current = $derived(selection.current);

	let marked = new SvelteSet<number>();
	let notice = $state('');
	let error = $state('');

	let chosen = $derived(marked.size ? memories.filter((m) => marked.has(m.id)) : current ? [current] : []);
	let anyRetired = $derived(chosen.some((m) => m.retired !== undefined));

	// What the pane is doing: showing the selected memory, editing it, adding
	// a new one, or confirming a delete. It's in the address, so each has its
	// own URL, and Back closes a form.
	type Mode = 'view' | 'edit' | 'new' | 'delete';
	let mode = $derived.by((): Mode => {
		const m = param('mode');
		if (m === 'edit' && current) return 'edit';
		if (m === 'new' && uuid) return 'new';
		if (m === 'delete' && chosen.length) return 'delete';
		return 'view';
	});
	const open = (m: Mode) => setParams({ mode: m });
	// close ends a form: back to where it was opened from, or on to a memory just added.
	const close = (added?: string) => (added ? setParams({ item: added, mode: null }, 'replace') : leave('mode'));
	// A mode the pane can't be in (a form for nothing) is dropped from the address.
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	type Done = string | { note: string; added: string };
	async function run(action: () => Promise<Done>, countsChange = false) {
		error = '';
		notice = '';
		try {
			const done = await action();
			notice = typeof done === 'string' ? done : done.note;
			// The address settles first: a change to it cancels a reload in flight.
			await close(typeof done === 'string' ? undefined : done.added);
			await invalidate(memoriesDependency);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveForm(changes: MemoryChanges) {
		const editing = mode === 'edit' ? current : undefined;
		await run(async () => {
			if (!editing && !uuid) throw new Error(m.memories.memories.newFromPage);
			const r = editing ? await updateMemory(save, editing.id, changes) : await addMemory(save, uuid!, changes);
			const note = [savedNote(r, editing ? m.memories.memories.savedWhat : m.memories.memories.newWhat), r.warning].filter(Boolean).join(' ');
			return editing ? note : { note, added: String(r.id) };
		}, !editing);
	}

	async function act(what: 'activate' | 'deactivate' | 'move' | 'delete') {
		const ids = chosen.map((m) => m.id);
		await run(async () => {
			let note: string;
			if (what === 'move') note = m.memories.memories.moved({ count: (await moveToCurrentIdentity(save, ids)).n });
			else if (what === 'delete') note = m.memories.memories.deleted({ count: (await deleteMemories(save, ids)).n });
			else {
				const r = await updateMemories(save, { ids, changes: { active: what === 'activate' } });
				note = (what === 'activate' ? m.memories.memories.madeActive : m.memories.memories.madeInactive)({ count: r.n });
			}
			marked.clear();
			return note;
		}, what === 'move' || what === 'delete');
	}

	function toggleMark(m: Memory) {
		if (marked.has(m.id)) marked.delete(m.id);
		else marked.add(m.id);
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const can = !readOnly && viewing;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: m.common.edit, hidden: !viewing, disabled: !can || !current, run: () => open('edit') },
			{ key: 'n', shown: 'N', label: m.memories.new, hidden: !viewing || !uuid, disabled: !can, run: () => open('new') },
			{ key: 'x', shown: 'X', label: m.memories.memories.mark, hidden: !viewing, disabled: !can || !current, run: () => current && toggleMark(current) },
			...deleteShortcuts({ label: m.common.delete, hidden: !viewing, disabled: !can || !chosen.length, run: () => open('delete') }),
			{
				key: 'Enter',
				label: m.common.save,
				command: true,
				hidden: mode !== 'edit' && mode !== 'new',
				run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit()
			},
			// Esc cancels when there's something to cancel; otherwise it's free for the page.
			...(viewing && !marked.size
				? []
				: [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => (mode === 'view' ? marked.clear() : close()) }])
		];
	});
</script>

<div class="tools">
	<input type="search" bind:value={query} placeholder={m.memories.memories.search} aria-label={m.memories.memories.search} />
	<select bind:value={type} aria-label={m.memories.type}>
		<option value="">{m.memories.memories.allTypes}</option>
		{#each types as t (t)}<option value={t}>{t}</option>{/each}
	</select>
	<select bind:value={mood} aria-label={m.memories.mood}>
		<option value="">{m.memories.memories.allMoods}</option>
		{#each moods as e (e)}<option value={e}>{e}</option>{/each}
	</select>
	{#if !uuid}
		<select bind:value={actor} aria-label={m.characters.character}>
			<option value="">{m.characters.everyone}</option>
			{#each characters as c (c.uuid)}
				<option value={c.uuid}>{characterName(c)}</option>
			{/each}
		</select>
	{/if}
	{#if marked.size}
		<span class="marked">{m.memories.memories.marked({ count: marked.size })}</span>
		<button type="button" class="textbutton" onclick={() => marked.clear()}>{m.memories.memories.clear}</button>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split" class:everyone={!uuid}>
	<div class="list">
		{#if !memories.length}
			<p class="muted">{m.memories.memories.none}</p>
		{:else if !shown.length}
			<p class="muted">{m.memories.memories.noMatch}</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label={m.common.selected}></th>
						<SortHeader label={m.memories.when} key="when" bind:sort descendingFirst />
						{#if !uuid}<SortHeader label={m.memories.who} key="who" bind:sort />{/if}
						<th class="caps text">{m.memories.memories.memory}</th>
						<SortHeader label={m.memories.type} key="type" bind:sort class="narrow-hide" />
						<SortHeader label={m.memories.mood} key="mood" bind:sort class="wide" />
						<SortHeader label={m.memories.imp} key="importance" bind:sort descendingFirst align="end" class="wide" />
					</tr>
				</thead>
				<tbody>
					{#each shown as mem, i (mem.id)}
						{#if byDay && dayLabel(mem) !== (shown[i - 1] ? dayLabel(shown[i - 1]!) : '')}
							<tr class="day"><th colspan={uuid ? 6 : 7} scope="rowgroup" class="caps sticky-day">{dayLabel(mem)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(mem)} class:inactive={!mem.active} onclick={() => mode === 'view' && selection.select(mem)}>
							<td class="mark">
								{#if marked.has(mem.id)}<span class="tick" aria-label={m.memories.memories.markedLabel}></span>{:else if selection.is(mem)}<Diamond />{/if}
							</td>
							<td class="when">{timeLabel(mem)}{#if !byDay}<span class="muted"> · {dayLabel(mem)}</span>{/if}</td>
							{#if !uuid}
								<td class="who">{#if owner(mem)}<a href={owner(mem)}>{mem.actor || m.characters.noName}</a>{:else}{mem.actor || m.characters.noName}{/if}</td>
							{/if}
							<td class="text">
								<span class="line">{#if mem.title}<strong>{mem.title}: </strong>{/if}{mem.content}</span>
								{#if mem.retired !== undefined}<span class="flag warn">{m.memories.under({ name: mem.retired || m.memories.anEarlierIdentity })}</span>{/if}
								{#if !mem.active}<span class="flag">{m.memories.inactive}</span>{/if}
								{#if !mem.embedded}<span class="flag">{m.memories.vectorPending}</span>{/if}
							</td>
							<td class="type caps narrow-hide">{mem.type}</td>
							<td class="wide">{mem.emotion}</td>
							<td class="n wide">{mem.importance.toFixed(2)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label={m.memories.memories.memory}>
		{#if mode === 'new'}
			<h2 class="caps">{m.memories.memories.newTitle}</h2>
			<MemoryForm onsave={saveForm} oncancel={() => close()} />
		{:else if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						<h2 class="caps">{m.memories.memories.editTitle({ id: current.id })}</h2>
						<MemoryForm memory={current} onsave={saveForm} oncancel={() => close()} />
					{:else}
						<h2 class="caps">{current.title || m.memories.memories.untitled({ id: current.id })}</h2>
						<p class="muted">{current.date ? m.memories.dateTime(current.date) : dayLabel(current)}</p>
						<p class="content">{current.content}</p>
						<dl>
							<dt class="caps">{m.memories.type}</dt><dd>{current.type}</dd>
							<dt class="caps">{m.memories.mood}</dt><dd>{current.emotion || '—'}</dd>
							<dt class="caps">{m.memories.importance}</dt><dd>{current.importance.toFixed(2)}</dd>
							{#if current.location}<dt class="caps">{m.memories.place}</dt><dd>{current.location}</dd>{/if}
							{#if current.tags.length}<dt class="caps">{m.memories.tags}</dt><dd>{current.tags.join(', ')}</dd>{/if}
							{#if current.condition}<dt class="caps">{m.memories.onlyWhile}</dt><dd><code>{current.condition}</code></dd>{/if}
							{#if current.diaryEntry !== undefined}<dt class="caps">{m.memories.memories.copyOf}</dt><dd>{m.memories.memories.diaryEntry({ id: current.diaryEntry })}</dd>{/if}
							<dt class="caps">{m.memories.search}</dt><dd>{current.embedded ? m.memories.vectorBuilt : m.memories.vectorPendingLong}</dd>
							{#if current.alwaysInject}<dt class="caps">{m.memories.recall}</dt><dd>{m.memories.alwaysIncluded}</dd>{/if}
							{#if !current.active}<dt class="caps">{m.memories.state}</dt><dd>{m.memories.memories.notRecalled}</dd>{/if}
						</dl>
						{#if current.retired !== undefined}
							<p class="notice warn">{m.memories.storedUnder({ name: current.retired || m.characters.noNameLower })}</p>
						{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>{m.memories.memories.confirmDelete({ count: chosen.length })}</p>
								<button type="button" class="button primary" onclick={() => act('delete')}>{m.common.delete}</button>
								<button type="button" class="button" onclick={() => close()}>{m.memories.keep}</button>
							</div>
						{:else if !readOnly}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('edit')}>{m.common.edit}</button>
								<button type="button" class="button" onclick={() => act(current!.active ? 'deactivate' : 'activate')}>
									{current.active ? m.memories.memories.makeInactive : m.memories.memories.makeActive}
								</button>
								{#if anyRetired}<button type="button" class="button" onclick={() => act('move')}>{m.memories.memories.move}</button>{/if}
								<button type="button" class="button" onclick={() => open('delete')}>{m.memories.deleteMore}</button>
							</div>
							{#if marked.size}<p class="muted small">{m.memories.memories.appliesTo({ count: marked.size })}</p>{/if}
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">{m.memories.memories.gameRunning}</p>{/if}
	</aside>
</div>

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-3);
	}
	.tools input {
		width: 16rem;
		max-width: 100%;
	}
	.marked {
		margin-left: var(--space-3);
		color: var(--accent);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(20rem, 26rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		vertical-align: top;
		white-space: nowrap;
	}
	tbody tr {
		cursor: default;
	}
	/* A row you can select tints under the pointer, as a bio's sections do. */
	tbody tr[aria-selected]:not([aria-selected='true']):hover td {
		background: var(--glass);
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	tr.inactive td {
		color: var(--muted);
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .mark {
		border-bottom-color: var(--rule);
	}
	.tick {
		display: inline-block;
		width: 9px;
		height: 9px;
		border: 1.5px solid var(--accent);
		background: var(--accent);
		border-radius: 1px;
	}
	.when {
		color: var(--muted);
	}
	.who {
		max-width: 9rem;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	/* A mood list never widens the table; the pane shows it whole. */
	td.wide {
		max-width: 7rem;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	/* Everyone's records carry a Who column, so the side columns give way sooner. */
	@media (max-width: 90rem) {
		.split.everyone :global(.wide) {
			display: none;
		}
	}
	.who a {
		text-decoration: none;
	}
	.who a:hover {
		text-decoration: underline;
	}
	.text {
		width: 100%;
		max-width: 0;
		white-space: normal;
		text-align: left;
		font-size: var(--step--1);
		padding: var(--space-2) var(--space-3);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
		font-weight: inherit;
	}
	tbody .text {
		font-size: var(--step-0);
		color: inherit;
		border-bottom-color: var(--rule);
	}
	.line {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.flag {
		font-size: var(--step--1);
		color: var(--muted);
		margin-right: var(--space-3);
	}
	.flag.warn {
		color: var(--warn);
	}
	.n {
		text-align: right;
	}
	td.type {
		font-size: var(--step--1);
		color: var(--muted);
		padding-top: 0.6rem;
	}
	.day th {
		text-align: left;
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
		font-size: var(--step--1);
		border-bottom: 1px solid var(--rule-strong);
	}
	/* The item card: a full-height zone of the menu beside the list. */
	.pane {
		--pane: var(--panel);
		position: sticky;
		top: 0;
		align-self: start;
		overflow: auto; /* fitToViewport sets its height */
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.content {
		margin: var(--space-4) 0;
		line-height: 1.6;
		max-width: 65ch;
	}
	dl {
		display: grid;
		grid-template-columns: 7rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: 0 0 var(--space-4);
	}
	dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	dd {
		margin: 0;
	}
	.actions,
	.confirm {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}
	.confirm p {
		flex-basis: 100%;
	}
	.small {
		font-size: var(--step--1);
		margin-top: var(--space-2);
	}
	.pane .notice {
		margin-top: var(--space-4);
	}
	/* Narrower desktop windows keep the pane beside the list, with fewer columns. */
	@media (max-width: 70rem) {
		.split {
			grid-template-columns: minmax(0, 1fr) minmax(18rem, 22rem);
		}
		.split :global(.wide) {
			display: none;
		}
	}
	@media (max-width: 60rem) {
		.split {
			grid-template-columns: 1fr;
		}
		.pane {
			position: static;
			min-height: 0;
			max-height: none;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
		.split :global(.narrow-hide) {
			display: none;
		}
	}
</style>
