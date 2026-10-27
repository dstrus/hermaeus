<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import type { DiaryChanges, DiaryEntry } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import { addDiaryEntry, deleteDiaryEntries, diaryDependency, memoriesDependency, savedNote, updateDiary } from './api.ts';
	import DiaryForm from './DiaryForm.svelte';
	import { groupByDay, timeLabel } from './dates.ts';

	let {
		save,
		uuid,
		name,
		entries,
		readOnly
	}: {
		save: string;
		uuid: string; // the character's
		name: string; // the character's, as shown
		entries: DiaryEntry[]; // newest first
		readOnly: boolean; // Skyrim is running, so nothing can be saved
	} = $props();

	let query = $state('');
	let oldestFirst = $state(false);

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = entries.filter(
			(d) => !q || [d.content, d.location, d.emotion, ...d.tags].some((v) => v.toLowerCase().includes(q))
		);
		return oldestFirst ? list.toReversed() : list;
	});
	let groups = $derived(groupByDay(shown));

	let editing = $state<number | 'new' | null>(null);
	let deleting = $state<number | null>(null); // the entry being confirmed for deletion
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
			await Promise.all([invalidate(diaryDependency), invalidate(memoriesDependency)]);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveEdit(d: DiaryEntry | undefined, changes: DiaryChanges) {
		await run(async () => {
			const r = d ? await updateDiary(save, d.id, changes) : await addDiaryEntry(save, uuid, changes);
			editing = null;
			return savedNote(r, d ? 'Diary entry' : 'New diary entry');
		}, !d);
	}

	async function remove(d: DiaryEntry) {
		await run(async () => {
			const r = await deleteDiaryEntries(save, { ids: [d.id], keepMemory });
			deleting = null;
			keepMemory = false;
			return r.memories ? 'Diary entry and its memory copy deleted.' : 'Diary entry deleted.';
		}, true);
	}
</script>

<div role="search">
	<input type="search" bind:value={query} placeholder="Search the diary" aria-label="Search the diary" />
	<label><input type="checkbox" bind:checked={oldestFirst} /> Oldest first</label>
</div>

<div class="actions">
	{#if readOnly}
		<p>Skyrim is running, so the diary can’t be changed until you close it.</p>
	{:else}
		<button onclick={() => (editing = 'new')} disabled={editing === 'new'}>New entry</button>
	{/if}
</div>

{#if notice}<p role="status">{notice}</p>{/if}
{#if error}<p role="alert">{error}</p>{/if}

{#if editing === 'new'}
	<DiaryForm onsave={(c) => saveEdit(undefined, c)} oncancel={() => (editing = null)} />
{/if}

{#if !entries.length}
	<p>No diary entries yet.</p>
{:else if !shown.length}
	<p>No diary entries match.</p>
{/if}

{#each groups as group (group.items[0]?.id)}
	<h2>{group.day}</h2>
	{#each group.items as d (d.id)}
		<article>
			<p>
				<span>{timeLabel(d)}</span>
				{#if d.emotion}· {d.emotion}{/if}
				· importance {d.importance.toFixed(2)}
				{#if !readOnly && editing !== d.id}
					<button onclick={() => (editing = d.id)}>Edit</button>
					<button onclick={() => ((deleting = d.id), (keepMemory = false))}>Delete…</button>
				{/if}
			</p>
			{#if deleting === d.id}
				<div role="alert">
					<p>Delete this diary entry? Hermaeus backs up the database before its first change each session.</p>
					{#if d.memory !== undefined}
						<label>
							<input type="checkbox" bind:checked={keepMemory} />
							Keep its memory copy, so {name} still remembers it
						</label>
					{/if}
					<button onclick={() => remove(d)}>Delete</button>
					<button onclick={() => (deleting = null)}>Keep it</button>
				</div>
			{/if}
			<ul class="flags">
				{#if d.retired !== undefined}
					<li title="SkyrimNet doesn’t appear to recall entries stored under an earlier identity.">
						stored under “{d.retired || 'an earlier identity'}”
					</li>
				{/if}
				{#if d.memory === undefined}<li title="SkyrimNet keeps a copy of each entry as a memory, and none was found for this one">no memory copy</li>{/if}
			</ul>
			{#if editing === d.id}
				<DiaryForm entry={d} onsave={(c) => saveEdit(d, c)} oncancel={() => (editing = null)} />
			{:else}
				<p class="entry">{d.content}</p>
				{#if d.location}<p><small>{d.location}</small></p>{/if}
			{/if}
		</article>
	{/each}
{/each}

<style>
	.flags,
	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		list-style: none;
		padding: 0;
	}
	.entry {
		white-space: pre-line;
	}
</style>
