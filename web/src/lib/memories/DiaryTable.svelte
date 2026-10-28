<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { DiaryChanges, DiaryEntry } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { addDiaryEntry, deleteDiaryEntries, diaryDependency, memoriesDependency, savedNote, updateDiary } from './api.ts';
	import { dayLabel, timeLabel } from './dates.ts';
	import DiaryForm from './DiaryForm.svelte';

	let { save, uuid, name, entries }: { save: string; uuid: string; name: string; entries: DiaryEntry[] } = $props(); // newest first

	let readOnly = $derived(live.gameRunning);
	let query = $state('');
	let sort = $state<Sort>({ key: 'when', descending: true });

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = entries.filter((d) => !q || [d.content, d.location, d.emotion, ...d.tags].some((v) => v.toLowerCase().includes(q)));
		return sorted(list, sort, { when: (d) => d.id, mood: (d) => d.emotion || '￿', importance: (d) => d.importance });
	});
	let byDay = $derived(sort.key === 'when');

	const selection = new Selection(() => shown, (d) => String(d.id));
	let current = $derived(selection.current);

	let mode = $state<'view' | 'edit' | 'new' | 'delete'>('view');
	let keepMemory = $state(false);
	let notice = $state('');
	let error = $state('');

	// run does a change, then reloads the diary, the memories (copies change
	// with entries) and, when entries come or go, the character list's counts.
	async function run(action: () => Promise<string>, countsChange: boolean) {
		error = '';
		notice = '';
		try {
			notice = await action();
			mode = 'view';
			await Promise.all([invalidate(diaryDependency), invalidate(memoriesDependency)]);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveForm(changes: DiaryChanges) {
		const editing = mode === 'edit' ? current : undefined;
		await run(async () => {
			const r = editing ? await updateDiary(save, editing.id, changes) : await addDiaryEntry(save, uuid, changes);
			if (!editing) selection.key = String(r.id);
			return savedNote(r, editing ? 'Diary entry' : 'New diary entry');
		}, !editing);
	}

	async function remove(d: DiaryEntry) {
		await run(async () => {
			const r = await deleteDiaryEntries(save, { ids: [d.id], keepMemory });
			keepMemory = false;
			return r.memories ? 'Diary entry and its memory copy deleted.' : 'Diary entry deleted.';
		}, true);
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const can = !readOnly && viewing;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !can || !current, run: () => (mode = 'edit') },
			{ key: 'n', shown: 'N', label: 'New', hidden: !viewing, disabled: !can, run: () => (mode = 'new') },
			...deleteShortcuts({ label: 'Delete', hidden: !viewing, disabled: !can || !current, run: () => ((keepMemory = false), (mode = 'delete')) }),
			{
				key: 'Enter',
				label: 'Save',
				command: true,
				hidden: mode !== 'edit' && mode !== 'new',
				run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit()
			},
			// Esc cancels when there's something to cancel; otherwise it's free for the page.
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => (mode = 'view') }])
		];
	});
</script>

<div class="tools">
	<input type="search" bind:value={query} placeholder="Search the diary" aria-label="Search the diary" />
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		{#if !entries.length}
			<p class="muted">No diary entries yet.</p>
		{:else if !shown.length}
			<p class="muted">No diary entries match.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<SortHeader label="When" key="when" bind:sort descendingFirst />
						<th class="caps text">Entry</th>
						<SortHeader label="Mood" key="mood" bind:sort class="wide" />
						<SortHeader label="Imp." key="importance" bind:sort descendingFirst align="end" class="wide" />
					</tr>
				</thead>
				<tbody>
					{#each shown as d, i (d.id)}
						{#if byDay && dayLabel(d) !== (shown[i - 1] ? dayLabel(shown[i - 1]!) : '')}
							<tr class="day"><th colspan="5" scope="rowgroup" class="caps">{dayLabel(d)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(d)} onclick={() => mode === 'view' && selection.select(d)}>
							<td class="mark">{#if selection.is(d)}<Diamond />{/if}</td>
							<td class="when">{timeLabel(d)}{#if !byDay}<span class="muted"> · {dayLabel(d)}</span>{/if}</td>
							<td class="text">
								<span class="line">{d.content}</span>
								{#if d.retired !== undefined}<span class="flag warn">under “{d.retired || 'an earlier identity'}”</span>{/if}
								{#if d.memory === undefined}<span class="flag warn">no memory copy</span>{/if}
							</td>
							<td class="wide">{d.emotion}</td>
							<td class="n wide">{d.importance.toFixed(2)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Diary entry">
		{#if mode === 'new'}
			<h2 class="caps">New diary entry</h2>
			<DiaryForm onsave={saveForm} oncancel={() => (mode = 'view')} />
		{:else if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						<h2 class="caps">Edit diary entry {current.id}</h2>
						<DiaryForm entry={current} onsave={saveForm} oncancel={() => (mode = 'view')} />
					{:else}
						<h2 class="caps">Diary entry {current.id}</h2>
						<p class="muted">
							{current.date ? `${current.date.time}, ${current.date.day}` : dayLabel(current)}{#if current.emotion}{' · '}{current.emotion}{/if}{#if current.location}{' · '}{current.location}{/if}
						</p>
						<p class="entry">{current.content}</p>
						{#if current.memory === undefined}
							<p class="notice warn">SkyrimNet keeps a copy of each entry as a memory, and none was found for this one.</p>
						{/if}
						{#if current.retired !== undefined}
							<p class="notice warn">Stored under the earlier identity “{current.retired || 'unnamed'}”, which SkyrimNet doesn’t appear to recall.</p>
						{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>Delete this diary entry? Hermaeus backs up the database before its first change each session.</p>
								{#if current.memory !== undefined}
									<label class="check"><input type="checkbox" bind:checked={keepMemory} /> Keep its memory copy, so {name} still remembers it</label>
								{/if}
								<button type="button" class="button primary" onclick={() => remove(current!)}>Delete</button>
								<button type="button" class="button" onclick={() => (mode = 'view')}>Keep</button>
							</div>
						{:else if !readOnly}
							<div class="actions">
								<button type="button" class="button" onclick={() => (mode = 'edit')}>Edit</button>
								<button type="button" class="button" onclick={() => ((keepMemory = false), (mode = 'delete'))}>Delete…</button>
							</div>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">Skyrim is running, so the diary can’t be changed until you close it.</p>{/if}
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
		background: var(--pane);
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
