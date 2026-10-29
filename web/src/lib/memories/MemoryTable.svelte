<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { SvelteSet } from 'svelte/reactivity';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Memory, MemoryChanges } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import { plural } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
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

	const selection = new Selection(() => shown, (m) => String(m.id));
	let current = $derived(selection.current);

	// What the pane is doing: showing the selected memory, editing it, adding
	// a new one, or confirming a delete.
	let mode = $state<'view' | 'edit' | 'new' | 'delete'>('view');
	let marked = new SvelteSet<number>();
	let notice = $state('');
	let error = $state('');

	let chosen = $derived(marked.size ? memories.filter((m) => marked.has(m.id)) : current ? [current] : []);
	let anyRetired = $derived(chosen.some((m) => m.retired !== undefined));

	async function run(action: () => Promise<string>, countsChange = false) {
		error = '';
		notice = '';
		try {
			notice = await action();
			mode = 'view';
			await invalidate(memoriesDependency);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveForm(changes: MemoryChanges) {
		const editing = mode === 'edit' ? current : undefined;
		await run(async () => {
			if (!editing && !uuid) throw new Error('A new memory is added from its character’s page.');
			const r = editing ? await updateMemory(save, editing.id, changes) : await addMemory(save, uuid!, changes);
			if (!editing) selection.key = String(r.id);
			return [savedNote(r, editing ? 'Memory' : 'New memory'), r.warning].filter(Boolean).join(' ');
		}, !editing);
	}

	async function act(what: 'activate' | 'deactivate' | 'move' | 'delete') {
		const ids = chosen.map((m) => m.id);
		await run(async () => {
			let note: string;
			if (what === 'move') note = `${plural((await moveToCurrentIdentity(save, ids)).n, 'memory', 'memories')} moved to the current identity.`;
			else if (what === 'delete') note = `${plural((await deleteMemories(save, ids)).n, 'memory', 'memories')} deleted.`;
			else {
				const r = await updateMemories(save, { ids, changes: { active: what === 'activate' } });
				note = `${plural(r.n, 'memory', 'memories')} made ${what === 'activate' ? 'active' : 'inactive'}.`;
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
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !can || !current, run: () => (mode = 'edit') },
			{ key: 'n', shown: 'N', label: 'New', hidden: !viewing || !uuid, disabled: !can, run: () => (mode = 'new') },
			{ key: 'x', shown: 'X', label: 'Mark', hidden: !viewing, disabled: !can || !current, run: () => current && toggleMark(current) },
			...deleteShortcuts({ label: 'Delete', hidden: !viewing, disabled: !can || !chosen.length, run: () => (mode = 'delete') }),
			{
				key: 'Enter',
				label: 'Save',
				command: true,
				hidden: mode !== 'edit' && mode !== 'new',
				run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit()
			},
			// Esc cancels when there's something to cancel; otherwise it's free for the page.
			...(viewing && !marked.size
				? []
				: [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => (mode === 'view' ? marked.clear() : (mode = 'view')) }])
		];
	});
</script>

<div class="tools">
	<input type="search" bind:value={query} placeholder="Search memories" aria-label="Search memories" />
	<select bind:value={type} aria-label="Type">
		<option value="">All types</option>
		{#each types as t (t)}<option value={t}>{t}</option>{/each}
	</select>
	<select bind:value={mood} aria-label="Mood">
		<option value="">All moods</option>
		{#each moods as e (e)}<option value={e}>{e}</option>{/each}
	</select>
	{#if !uuid}
		<select bind:value={actor} aria-label="Character">
			<option value="">Everyone</option>
			{#each characters as c (c.uuid)}
				<option value={c.uuid}>{characterName(c)}</option>
			{/each}
		</select>
	{/if}
	{#if marked.size}
		<span class="marked">{plural(marked.size, 'memory', 'memories')} marked</span>
		<button type="button" class="textbutton" onclick={() => marked.clear()}>Clear</button>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split" class:everyone={!uuid}>
	<div class="list">
		{#if !memories.length}
			<p class="muted">No memories yet.</p>
		{:else if !shown.length}
			<p class="muted">No memories match.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<SortHeader label="When" key="when" bind:sort descendingFirst />
						{#if !uuid}<SortHeader label="Who" key="who" bind:sort />{/if}
						<th class="caps text">Memory</th>
						<SortHeader label="Type" key="type" bind:sort class="narrow-hide" />
						<SortHeader label="Mood" key="mood" bind:sort class="wide" />
						<SortHeader label="Imp." key="importance" bind:sort descendingFirst align="end" class="wide" />
					</tr>
				</thead>
				<tbody>
					{#each shown as m, i (m.id)}
						{#if byDay && dayLabel(m) !== (shown[i - 1] ? dayLabel(shown[i - 1]!) : '')}
							<tr class="day"><th colspan={uuid ? 6 : 7} scope="rowgroup" class="caps sticky-day">{dayLabel(m)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(m)} class:inactive={!m.active} onclick={() => mode === 'view' && selection.select(m)}>
							<td class="mark">
								{#if marked.has(m.id)}<span class="tick" aria-label="Marked"></span>{:else if selection.is(m)}<Diamond />{/if}
							</td>
							<td class="when">{timeLabel(m)}{#if !byDay}<span class="muted"> · {dayLabel(m)}</span>{/if}</td>
							{#if !uuid}
								<td class="who">{#if owner(m)}<a href={owner(m)}>{m.actor || 'Unnamed'}</a>{:else}{m.actor || 'Unnamed'}{/if}</td>
							{/if}
							<td class="text">
								<span class="line">{#if m.title}<strong>{m.title}: </strong>{/if}{m.content}</span>
								{#if m.retired !== undefined}<span class="flag warn">under “{m.retired || 'an earlier identity'}”</span>{/if}
								{#if !m.active}<span class="flag">inactive</span>{/if}
								{#if !m.embedded}<span class="flag">vector pending</span>{/if}
							</td>
							<td class="type caps narrow-hide">{m.type}</td>
							<td class="wide">{m.emotion}</td>
							<td class="n wide">{m.importance.toFixed(2)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Memory">
		{#if mode === 'new'}
			<h2 class="caps">New memory</h2>
			<MemoryForm onsave={saveForm} oncancel={() => (mode = 'view')} />
		{:else if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						<h2 class="caps">Edit memory {current.id}</h2>
						<MemoryForm memory={current} onsave={saveForm} oncancel={() => (mode = 'view')} />
					{:else}
						<h2 class="caps">{current.title || `Memory ${current.id}`}</h2>
						<p class="muted">{current.date ? `${current.date.time}, ${current.date.day}` : dayLabel(current)}</p>
						<p class="content">{current.content}</p>
						<dl>
							<dt class="caps">Type</dt><dd>{current.type}</dd>
							<dt class="caps">Mood</dt><dd>{current.emotion || '—'}</dd>
							<dt class="caps">Importance</dt><dd>{current.importance.toFixed(2)}</dd>
							{#if current.location}<dt class="caps">Place</dt><dd>{current.location}</dd>{/if}
							{#if current.tags.length}<dt class="caps">Tags</dt><dd>{current.tags.join(', ')}</dd>{/if}
							{#if current.condition}<dt class="caps">Only while</dt><dd><code>{current.condition}</code></dd>{/if}
							{#if current.diaryEntry !== undefined}<dt class="caps">Copy of</dt><dd>diary entry {current.diaryEntry}</dd>{/if}
							<dt class="caps">Search</dt><dd>{current.embedded ? 'vector built' : 'vector pending; SkyrimNet builds it when it starts'}</dd>
							{#if current.alwaysInject}<dt class="caps">Recall</dt><dd>always included, not only when relevant</dd>{/if}
							{#if !current.active}<dt class="caps">State</dt><dd>inactive: SkyrimNet doesn’t recall it</dd>{/if}
						</dl>
						{#if current.retired !== undefined}
							<p class="notice warn">Stored under the earlier identity “{current.retired || 'unnamed'}”, which SkyrimNet doesn’t appear to recall.</p>
						{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>Delete {chosen.length === 1 ? 'this memory' : plural(chosen.length, 'marked memory', 'marked memories')}? Hermaeus backs up the database before its first change each session.</p>
								<button type="button" class="button primary" onclick={() => act('delete')}>Delete</button>
								<button type="button" class="button" onclick={() => (mode = 'view')}>Keep</button>
							</div>
						{:else if !readOnly}
							<div class="actions">
								<button type="button" class="button" onclick={() => (mode = 'edit')}>Edit</button>
								<button type="button" class="button" onclick={() => act(current!.active ? 'deactivate' : 'activate')}>
									{current.active ? 'Make inactive' : 'Make active'}
								</button>
								{#if anyRetired}<button type="button" class="button" onclick={() => act('move')}>Move to current identity</button>{/if}
								<button type="button" class="button" onclick={() => (mode = 'delete')}>Delete…</button>
							</div>
							{#if marked.size}<p class="muted small">Actions apply to the {plural(marked.size, 'marked memory', 'marked memories')}.</p>{/if}
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">Skyrim is running, so memories can’t be changed until you close it.</p>{/if}
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
