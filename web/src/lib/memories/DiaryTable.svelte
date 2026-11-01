<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { DiaryChanges, DiaryEntry } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import { m } from '#lib/copy/messages.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { addDiaryEntry, deleteDiaryEntries, diaryDependency, memoriesDependency, savedNote, updateDiary } from './api.ts';
	import { dayLabel, timeLabel } from './dates.ts';
	import DiaryForm from './DiaryForm.svelte';

	import type { Character } from '#lib/api/types/characters.ts';
	import { characterName } from '#lib/characters/api.ts';

	let {
		save,
		uuid,
		name = '',
		entries,
		characters = []
	}: {
		save: string;
		uuid?: string; // whose diary this is; everyone's when left out
		name?: string;
		entries: DiaryEntry[]; // newest first
		characters?: Character[]; // for narrowing everyone's to one character's
	} = $props();
	let actor = $state('');
	const owner = (d: DiaryEntry) => (d.actorUuid ? `/${encodeURIComponent(save)}/npc/${d.actorUuid}/diary` : undefined);

	let readOnly = $derived(live.gameRunning);
	let query = $state('');
	let sort = $state<Sort>({ key: 'when', descending: true });

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = entries.filter(
			(d) => (!actor || d.actorUuid === actor) && (!q || [d.content, d.location, d.emotion, ...d.tags].some((v) => v.toLowerCase().includes(q)))
		);
		return sorted(list, sort, { when: (d) => d.id, who: (d) => d.actor || '￿', mood: (d) => d.emotion || '￿', importance: (d) => d.importance });
	});
	let byDay = $derived(sort.key === 'when');

	const selection = new Selection(() => shown, (d) => String(d.id), { param: 'item' });
	let current = $derived(selection.current);

	let keepMemory = $state(false);
	let notice = $state('');
	let error = $state('');

	// What the pane is doing, kept in the address: each state has its own
	// URL, and Back closes a form.
	type Mode = 'view' | 'edit' | 'new' | 'delete';
	let mode = $derived.by((): Mode => {
		const m = param('mode');
		if (m === 'edit' && current) return 'edit';
		if (m === 'new' && uuid) return 'new';
		if (m === 'delete' && current) return 'delete';
		return 'view';
	});
	const open = (m: Mode) => setParams({ mode: m });
	// close ends a form: back to where it was opened from, or on to an entry just added.
	const close = (added?: string) => (added ? setParams({ item: added, mode: null }, 'replace') : leave('mode'));
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	// run does a change, then reloads the diary, the memories (copies change
	// with entries) and, when entries come or go, the character list's counts.
	type Done = string | { note: string; added: string };
	async function run(action: () => Promise<Done>, countsChange: boolean) {
		error = '';
		notice = '';
		try {
			const done = await action();
			notice = typeof done === 'string' ? done : done.note;
			await close(typeof done === 'string' ? undefined : done.added); // settles before the reload, which it would cancel
			await Promise.all([invalidate(diaryDependency), invalidate(memoriesDependency)]);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveForm(changes: DiaryChanges) {
		const editing = mode === 'edit' ? current : undefined;
		await run(async () => {
			if (!editing && !uuid) throw new Error(m.memories.diary.newFromPage);
			const r = editing ? await updateDiary(save, editing.id, changes) : await addDiaryEntry(save, uuid!, changes);
			const note = savedNote(r, editing ? m.memories.diary.savedWhat : m.memories.diary.newWhat);
			return editing ? note : { note, added: String(r.id) };
		}, !editing);
	}

	async function remove(d: DiaryEntry) {
		await run(async () => {
			const r = await deleteDiaryEntries(save, { ids: [d.id], keepMemory });
			keepMemory = false;
			return r.memories ? m.memories.diary.deletedWithCopy : m.memories.diary.deleted;
		}, true);
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const can = !readOnly && viewing;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: m.common.edit, hidden: !viewing, disabled: !can || !current, run: () => open('edit') },
			{ key: 'n', shown: 'N', label: m.memories.new, hidden: !viewing || !uuid, disabled: !can, run: () => open('new') },
			...deleteShortcuts({ label: m.common.delete, hidden: !viewing, disabled: !can || !current, run: () => ((keepMemory = false), open('delete')) }),
			{
				key: 'Enter',
				label: m.common.save,
				command: true,
				hidden: mode !== 'edit' && mode !== 'new',
				run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit()
			},
			// Esc cancels when there's something to cancel; otherwise it's free for the page.
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => close() }])
		];
	});
</script>

<div class="tools">
	<input type="search" bind:value={query} placeholder={m.memories.diary.search} aria-label={m.memories.diary.search} />
	{#if !uuid}
		<select bind:value={actor} aria-label={m.characters.character}>
			<option value="">{m.characters.everyone}</option>
			{#each characters as c (c.uuid)}
				<option value={c.uuid}>{characterName(c)}</option>
			{/each}
		</select>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split" class:everyone={!uuid}>
	<div class="list">
		{#if !entries.length}
			<p class="muted">{m.memories.diary.none}</p>
		{:else if !shown.length}
			<p class="muted">{m.memories.diary.noMatch}</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label={m.common.selected}></th>
						<SortHeader label={m.memories.when} key="when" bind:sort descendingFirst />
						{#if !uuid}<SortHeader label={m.memories.who} key="who" bind:sort />{/if}
						<th class="caps text">{m.memories.diary.entry}</th>
						<SortHeader label={m.memories.mood} key="mood" bind:sort class="wide" />
						<SortHeader label={m.memories.imp} key="importance" bind:sort descendingFirst align="end" class="wide" />
					</tr>
				</thead>
				<tbody>
					{#each shown as d, i (d.id)}
						{#if byDay && dayLabel(d) !== (shown[i - 1] ? dayLabel(shown[i - 1]!) : '')}
							<tr class="day"><th colspan={uuid ? 5 : 6} scope="rowgroup" class="caps sticky-day">{dayLabel(d)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(d)} onclick={() => mode === 'view' && selection.select(d)}>
							<td class="mark">{#if selection.is(d)}<Diamond />{/if}</td>
							<td class="when">{timeLabel(d)}{#if !byDay}<span class="muted"> · {dayLabel(d)}</span>{/if}</td>
							{#if !uuid}
								<td class="who">{#if owner(d)}<a href={owner(d)}>{d.actor || m.characters.noName}</a>{:else}{d.actor || m.characters.noName}{/if}</td>
							{/if}
							<td class="text">
								<span class="line">{d.content}</span>
								{#if d.retired !== undefined}<span class="flag warn">{m.memories.under({ name: d.retired || m.memories.anEarlierIdentity })}</span>{/if}
								{#if d.memory === undefined}<span class="flag warn">{m.memories.diary.noCopy}</span>{/if}
							</td>
							<td class="wide">{d.emotion}</td>
							<td class="n wide">{d.importance.toFixed(2)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label={m.memories.diary.pane}>
		{#if mode === 'new'}
			<h2 class="caps">{m.memories.diary.newTitle}</h2>
			<DiaryForm onsave={saveForm} oncancel={() => close()} />
		{:else if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						<h2 class="caps">{m.memories.diary.editTitle({ id: current.id })}</h2>
						<DiaryForm entry={current} onsave={saveForm} oncancel={() => close()} />
					{:else}
						<h2 class="caps">{m.memories.diary.title({ id: current.id })}</h2>
						<p class="muted">
							{current.date ? m.memories.dateTime(current.date) : dayLabel(current)}{#if current.emotion}{' · '}{current.emotion}{/if}{#if current.location}{' · '}{current.location}{/if}
						</p>
						<p class="entry">{current.content}</p>
						{#if current.memory === undefined}
							<p class="notice warn">{m.memories.diary.missingCopy}</p>
						{/if}
						{#if current.retired !== undefined}
							<p class="notice warn">{m.memories.storedUnder({ name: current.retired || m.characters.noNameLower })}</p>
						{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>{m.memories.diary.confirmDelete}</p>
								{#if current.memory !== undefined}
									<label class="check"><input type="checkbox" bind:checked={keepMemory} /> {m.memories.diary.keepCopy({ name: name || m.memories.diary.theCharacter })}</label>
								{/if}
								<button type="button" class="button primary" onclick={() => remove(current!)}>{m.common.delete}</button>
								<button type="button" class="button" onclick={() => close()}>{m.memories.keep}</button>
							</div>
						{:else if !readOnly}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('edit')}>{m.common.edit}</button>
								<button type="button" class="button" onclick={() => ((keepMemory = false), open('delete'))}>{m.memories.deleteMore}</button>
							</div>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">{m.memories.diary.gameRunning}</p>{/if}
	</aside>
</div>

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-3);
		margin-bottom: var(--space-3);
	}
	.tools input {
		width: 16rem;
		max-width: 100%;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(22rem, 32rem);
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
	.mark {
		width: 1.5rem;
		padding-right: 0;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .mark {
		border-bottom-color: var(--rule);
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
		margin-right: var(--space-3);
	}
	.flag.warn {
		color: var(--warn);
	}
	.n {
		text-align: right;
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
	.entry {
		margin: var(--space-4) 0;
		line-height: 1.65;
		white-space: pre-line;
		max-width: 65ch;
	}
	.actions,
	.confirm {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}
	.confirm p,
	.confirm .check {
		flex-basis: 100%;
	}
	.check {
		display: flex;
		gap: var(--space-2);
		align-items: center;
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
