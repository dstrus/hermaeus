<!-- World knowledge: the packs, each switchable, and every entry any NPC can
     draw on, with the selected entry read or edited in the pane. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { KnowledgeResponse, Memory, MemoryChanges, Pack } from '#lib/api/types/memories.ts';
	import { count, plural } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { clipped } from '#lib/ui/clipped.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { addKnowledge, deleteMemories, knowledgeDependency, savedNote, setPackEnabled, updateMemory } from './api.ts';
	import MemoryForm from './MemoryForm.svelte';

	let { save, knowledge }: { save: string; knowledge: KnowledgeResponse } = $props();

	let readOnly = $derived(live.gameRunning);
	let query = $state('');
	let pack = $state(''); // '' for every entry, a pack's id, or 'none' for entries in no pack
	let sort = $state<Sort>({ key: 'title', descending: false });
	let notice = $state('');
	let error = $state('');

	let packs = $derived(new Map(knowledge.packs.map((p) => [p.id, p])));
	const packOf = (e: Memory): Pack | undefined => (e.pack === undefined ? undefined : packs.get(e.pack));

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = knowledge.entries.filter(
			(e) =>
				(!pack || (pack === 'none' ? e.pack === undefined : String(e.pack) === pack)) &&
				(!q || [e.title, e.content, e.condition, ...e.tags].some((v) => v.toLowerCase().includes(q)))
		);
		return sorted(list, sort, {
			title: (e) => e.title || '￿',
			pack: (e) => packOf(e)?.name ?? '￿',
			importance: (e) => e.importance
		});
	});

	const selection = new Selection(() => shown, (e) => String(e.id), { param: 'item' });
	let current = $derived(selection.current);

	// What the pane is doing, kept in the address: each state has its own
	// URL, and Back closes a form.
	type Mode = 'view' | 'edit' | 'new' | 'delete';
	let mode = $derived.by((): Mode => {
		const m = param('mode');
		if (m === 'edit' && current) return 'edit';
		if (m === 'new') return 'new';
		if (m === 'delete' && current) return 'delete';
		return 'view';
	});
	const open = (m: Mode) => setParams({ mode: m });
	// close ends a form: back to where it was opened from, or on to an entry just added.
	const close = (added?: string) => (added ? setParams({ item: added, mode: null }, 'replace') : leave('mode'));
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	type Done = string | { note: string; added: string };
	async function run(action: () => Promise<Done>) {
		error = '';
		notice = '';
		try {
			const done = await action();
			notice = typeof done === 'string' ? done : done.note;
			await close(typeof done === 'string' ? undefined : done.added); // settles before the reload, which it would cancel
			await invalidate(knowledgeDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	async function saveForm(changes: MemoryChanges) {
		const editing = mode === 'edit' ? current : undefined;
		await run(async () => {
			if (editing) return savedNote(await updateMemory(save, editing.id, changes), 'Entry');
			// A new entry goes into the pack the list is narrowed to, if any.
			const into = pack && pack !== 'none' ? Number(pack) : undefined;
			const r = await addKnowledge(save, { pack: into, changes });
			return { note: savedNote(r, 'New entry'), added: String(r.id) };
		});
	}

	async function remove() {
		if (!current) return;
		const id = current.id;
		await run(async () => `${plural((await deleteMemories(save, [id])).n, 'entry', 'entries')} deleted.`);
	}

	async function togglePack(p: Pack) {
		error = '';
		try {
			await setPackEnabled(save, p.id, !p.enabled);
			notice = `${p.name} switched ${p.enabled ? 'off' : 'on'}. SkyrimNet reads packs when it starts.`;
			await invalidate(knowledgeDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const can = !readOnly && viewing;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !can || !current, run: () => open('edit') },
			{ key: 'n', shown: 'N', label: 'New', hidden: !viewing, disabled: !can, run: () => open('new') },
			...deleteShortcuts({ label: 'Delete', hidden: !viewing, disabled: !can || !current, run: () => open('delete') }),
			{
				key: 'Enter',
				label: 'Save',
				command: true,
				hidden: mode !== 'edit' && mode !== 'new',
				run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit()
			},
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => close() }])
		];
	});
</script>

{#if knowledge.packs.length}
	<section class="packs">
		<h2 class="caps">Packs</h2>
		<ul>
			{#each knowledge.packs as p (p.id)}
				<li>
					<label class="pack" class:off={!p.enabled}>
						<input type="checkbox" checked={p.enabled} disabled={readOnly} onchange={() => togglePack(p)} />
						<span class="name">{p.name}</span>
						<span class="muted small">{p.version} · {plural(p.entries, 'entry', 'entries')}{#if p.description} · {p.description}{/if}</span>
					</label>
				</li>
			{/each}
		</ul>
	</section>
{/if}

<div class="tools">
	<input type="search" bind:value={query} placeholder="Search entries" aria-label="Search entries" />
	<select bind:value={pack} aria-label="Pack">
		<option value="">Every pack</option>
		{#each knowledge.packs as p (p.id)}<option value={String(p.id)}>{p.name}</option>{/each}
		<option value="none">In no pack</option>
	</select>
	<span class="muted">{count(shown.length)} of {plural(knowledge.entries.length, 'entry', 'entries')}</span>
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		{#if !knowledge.entries.length}
			<p class="muted">No world knowledge yet: no packs are installed, and no entries have been added.</p>
		{:else if !shown.length}
			<p class="muted">No entries match.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<SortHeader label="Entry" key="title" bind:sort />
						<SortHeader label="Pack" key="pack" bind:sort class="pack-col" />
						<SortHeader label="Imp." key="importance" bind:sort descendingFirst align="end" class="wide" />
					</tr>
				</thead>
				<tbody>
					{#each shown as e (e.id)}
						<tr aria-selected={selection.is(e)} class:inactive={!e.active || packOf(e)?.enabled === false} onclick={() => mode === 'view' && selection.select(e)}>
							<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
							<td class="text">
								<span class="line">{#if e.title}<strong>{e.title}: </strong>{/if}{e.content}</span>
								{#if e.alwaysInject}<span class="flag">always included</span>{/if}
								{#if !e.active}<span class="flag">inactive</span>{/if}
								{#if packOf(e)?.enabled === false}<span class="flag">pack off</span>{/if}
								{#if !e.embedded}<span class="flag">vector pending</span>{/if}
							</td>
							<td class="pack-col muted" use:clipped>{packOf(e)?.name ?? ''}</td>
							<td class="n wide">{e.importance.toFixed(2)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Entry">
		{#if mode === 'new'}
			<h2 class="caps">New entry</h2>
			<p class="muted small">{pack && pack !== 'none' ? `Into ${packs.get(Number(pack))?.name ?? 'the pack'}.` : 'In no pack.'}</p>
			<MemoryForm onsave={saveForm} oncancel={() => close()} />
		{:else if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						<h2 class="caps">Edit entry {current.id}</h2>
						<MemoryForm memory={current} onsave={saveForm} oncancel={() => close()} />
					{:else}
						<h2 class="caps">{current.title || `Entry ${current.id}`}</h2>
						{#if packOf(current)}<p class="muted">{packOf(current)?.name}</p>{/if}
						<p class="content">{current.content}</p>
						<dl>
							<dt class="caps">Importance</dt><dd>{current.importance.toFixed(2)}</dd>
							{#if current.tags.length}<dt class="caps">Tags</dt><dd>{current.tags.join(', ')}</dd>{/if}
							{#if current.condition}<dt class="caps">Only while</dt><dd><code>{current.condition}</code></dd>{/if}
							<dt class="caps">Recall</dt><dd>{current.alwaysInject ? 'always included, not only when relevant' : 'when relevant'}</dd>
							<dt class="caps">Search</dt><dd>{current.embedded ? 'vector built' : 'vector pending; SkyrimNet builds it when it starts'}</dd>
							{#if !current.active}<dt class="caps">State</dt><dd>inactive: SkyrimNet doesn’t use it</dd>{/if}
						</dl>
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>Delete this entry? Hermaeus backs up the database before its first change each session.</p>
								<button type="button" class="button primary" onclick={remove}>Delete</button>
								<button type="button" class="button" onclick={() => close()}>Keep</button>
							</div>
						{:else if !readOnly}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('edit')}>Edit</button>
								<button type="button" class="button" onclick={() => open('delete')}>Delete…</button>
							</div>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">Skyrim is running, so knowledge can’t be changed until you close it.</p>{/if}
	</aside>
</div>

<style>
	.packs {
		margin-bottom: var(--space-4);
	}
	.packs h2 {
		font-size: var(--step--1);
		color: var(--muted);
		margin-bottom: var(--space-2);
	}
	.packs ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: var(--space-1);
	}
	.pack {
		display: flex;
		align-items: baseline;
		gap: var(--space-2);
	}
	.pack input {
		align-self: center;
		margin: 0;
	}
	.pack.off .name {
		color: var(--muted);
	}
	.small {
		font-size: var(--step--1);
	}
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
	.text {
		width: 100%;
		max-width: 0;
		white-space: normal;
	}
	.line {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.flag {
		display: inline-block;
		margin-top: var(--space-1);
		margin-right: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.split :global(.pack-col) {
		max-width: 12rem;
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: var(--step--1);
	}
	.n {
		text-align: right;
		width: 4rem;
	}
	/* No bottom padding: the form's sticky button row carries it, and a
	   sticky element stops at the scroller's padding, not its edge. */
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.content {
		margin: var(--space-3) 0;
		white-space: pre-wrap;
		line-height: 1.6;
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
		overflow-wrap: anywhere;
	}
	code {
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	.actions,
	.confirm {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}
	.confirm p {
		flex-basis: 100%;
	}
	.pane .notice {
		margin-top: var(--space-3);
	}
	.pane > div,
	.pane > .notice {
		padding-bottom: var(--space-5);
	}
	@media (max-width: 90rem) {
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
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
	}
</style>
