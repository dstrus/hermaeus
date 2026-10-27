<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { SvelteSet } from 'svelte/reactivity';
	import { ApiError } from '#lib/api/client.ts';
	import type { Memory, MemoryChanges } from '#lib/api/types/memories.ts';
	import { charactersDependency } from '#lib/characters/api.ts';
	import { plural } from '#lib/format.ts';
	import {
		addMemory,
		deleteMemories,
		memoriesDependency,
		moveToCurrentIdentity,
		savedNote,
		updateMemories,
		updateMemory
	} from './api.ts';
	import { dayLabel, groupByDay, timeLabel } from './dates.ts';
	import MemoryForm from './MemoryForm.svelte';

	let {
		save,
		uuid,
		memories,
		readOnly
	}: {
		save: string;
		uuid: string; // the character's
		memories: Memory[]; // newest first
		readOnly: boolean; // Skyrim is running, so nothing can be saved
	} = $props();

	let query = $state('');
	let type = $state('');
	let emotion = $state('');
	let sort = $state<'newest' | 'oldest' | 'importance'>('newest');

	let types = $derived([...new Set(memories.map((m) => m.type).filter(Boolean))].sort());
	let emotions = $derived([...new Set(memories.map((m) => m.emotion).filter(Boolean))].sort());

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = memories.filter(
			(m) =>
				(!type || m.type === type) &&
				(!emotion || m.emotion === emotion) &&
				(!q || [m.content, m.title, m.location, m.emotion, ...m.tags].some((v) => v.toLowerCase().includes(q)))
		);
		if (sort === 'oldest') return list.toReversed();
		if (sort === 'importance') return list.toSorted((a, b) => b.importance - a.importance || b.id - a.id);
		return list;
	});
	// Sorted by time, memories go under day headings; by importance, each shows its day.
	let groups = $derived(sort === 'importance' ? [{ day: '', items: shown }] : groupByDay(shown));

	// Editing: one memory at a time ('new' for a new one), or a selection of several.
	let editing = $state<number | 'new' | null>(null);
	let selected = new SvelteSet<number>();
	let confirmingDelete = $state(false);
	let notice = $state('');
	let error = $state('');

	let chosen = $derived(memories.filter((m) => selected.has(m.id)));
	let anyRetired = $derived(chosen.some((m) => m.retired !== undefined));
	let allShownSelected = $derived(shown.length > 0 && shown.every((m) => selected.has(m.id)));

	// run does a change, then reloads what it affects. Counts in the character
	// list change too when memories are added, deleted or moved.
	async function run(action: () => Promise<string>, countsChange = false) {
		error = '';
		notice = '';
		try {
			notice = await action();
			await invalidate(memoriesDependency);
			if (countsChange) await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
			throw e;
		}
	}

	async function saveEdit(m: Memory | undefined, changes: MemoryChanges) {
		await run(async () => {
			const r = m ? await updateMemory(save, m.id, changes) : await addMemory(save, uuid, changes);
			editing = null;
			return [savedNote(r, m ? 'Memory' : 'New memory'), r.warning].filter(Boolean).join(' ');
		}, !m).catch(() => {}); // the error is shown on the page
	}

	async function bulk(what: 'activate' | 'deactivate' | 'move' | 'delete') {
		const ids = [...selected];
		await run(async () => {
			let note: string;
			if (what === 'move') {
				const r = await moveToCurrentIdentity(save, ids);
				note = `${plural(r.n, 'memory', 'memories')} moved to the current identity.`;
			} else if (what === 'delete') {
				const r = await deleteMemories(save, ids);
				note = `${plural(r.n, 'memory', 'memories')} deleted.`;
			} else {
				const r = await updateMemories(save, { ids, changes: { active: what === 'activate' } });
				note = `${plural(r.n, 'memory', 'memories')} made ${what === 'activate' ? 'active' : 'inactive'}.`;
			}
			selected.clear();
			confirmingDelete = false;
			return note;
		}, what === 'move' || what === 'delete').catch(() => {});
	}

	function toggleAll() {
		if (allShownSelected) shown.forEach((m) => selected.delete(m.id));
		else shown.forEach((m) => selected.add(m.id));
	}
</script>

<div role="search">
	<input type="search" bind:value={query} placeholder="Search memories" aria-label="Search memories" />
	<select bind:value={type} aria-label="Type">
		<option value="">All types</option>
		{#each types as t (t)}<option value={t}>{t}</option>{/each}
	</select>
	<select bind:value={emotion} aria-label="Emotion">
		<option value="">All emotions</option>
		{#each emotions as e (e)}<option value={e}>{e}</option>{/each}
	</select>
	<select bind:value={sort} aria-label="Order">
		<option value="newest">Newest first</option>
		<option value="oldest">Oldest first</option>
		<option value="importance">Most important first</option>
	</select>
</div>

<div class="actions">
	{#if readOnly}
		<p>Skyrim is running, so memories can’t be changed until you close it.</p>
	{:else}
		<button onclick={() => (editing = 'new')} disabled={editing === 'new'}>New memory</button>
		{#if shown.length}
			<label><input type="checkbox" checked={allShownSelected} onchange={toggleAll} /> Select all shown</label>
		{/if}
		{#if selected.size}
			<span>{plural(selected.size, 'memory', 'memories')} selected:</span>
			<button onclick={() => bulk('activate')}>Make active</button>
			<button onclick={() => bulk('deactivate')}>Make inactive</button>
			{#if anyRetired}
				<button onclick={() => bulk('move')} title="So SkyrimNet recalls them again">Move to current identity</button>
			{/if}
			{#if confirmingDelete}
				<span role="alert">Delete {plural(selected.size, 'memory', 'memories')}? Hermaeus backs up the database before its first change each session.</span>
				<button onclick={() => bulk('delete')}>Delete</button>
				<button onclick={() => (confirmingDelete = false)}>Keep them</button>
			{:else}
				<button onclick={() => (confirmingDelete = true)}>Delete…</button>
			{/if}
			<button onclick={() => selected.clear()}>Clear selection</button>
		{/if}
	{/if}
</div>

{#if notice}<p role="status">{notice}</p>{/if}
{#if error}<p role="alert">{error}</p>{/if}

{#if editing === 'new'}
	<MemoryForm onsave={(c) => saveEdit(undefined, c)} oncancel={() => (editing = null)} />
{/if}

{#if !memories.length}
	<p>No memories yet.</p>
{:else if !shown.length}
	<p>No memories match.</p>
{/if}

{#each groups as group (group.items[0]?.id)}
	{#if group.day}<h2>{group.day}</h2>{/if}
	{#each group.items as m (m.id)}
		<article class:inactive={!m.active}>
			<p>
				{#if !readOnly}
					<input
						type="checkbox"
						checked={selected.has(m.id)}
						onchange={() => (selected.has(m.id) ? selected.delete(m.id) : selected.add(m.id))}
						aria-label="Select this memory"
					/>
				{/if}
				<span>{timeLabel(m)}{#if !group.day}, {dayLabel(m)}{/if}</span>
				· {m.type}
				{#if m.emotion}· {m.emotion}{/if}
				· importance {m.importance.toFixed(2)}
				{#if !readOnly && editing !== m.id}
					<button onclick={() => (editing = m.id)}>Edit</button>
				{/if}
			</p>
			<ul class="flags">
				{#if m.retired !== undefined}
					<li title="SkyrimNet doesn’t appear to recall memories stored under an earlier identity.">
						stored under “{m.retired || 'an earlier identity'}”
					</li>
				{/if}
				{#if !m.active}<li>inactive</li>{/if}
				{#if m.alwaysInject}<li>always included</li>{/if}
				{#if m.condition}<li title="Only used while this is true: {m.condition}">conditional</li>{/if}
				{#if m.diaryEntry !== undefined}<li>copy of diary entry {m.diaryEntry}</li>{/if}
				{#if !m.embedded}<li title="SkyrimNet rebuilds missing search vectors when it starts">search vector pending</li>{/if}
			</ul>
			{#if editing === m.id}
				<MemoryForm memory={m} onsave={(c) => saveEdit(m, c)} oncancel={() => (editing = null)} />
			{:else}
				{#if m.title}<h3>{m.title}</h3>{/if}
				<p>{m.content}</p>
				{#if m.location}<p><small>{m.location}</small></p>{/if}
			{/if}
		</article>
	{/each}
{/each}

<style>
	.inactive {
		opacity: 0.6;
	}
	.flags,
	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		list-style: none;
		padding: 0;
	}
</style>
