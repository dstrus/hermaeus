<!-- Every prompt SkyrimNet can use, with the layer each comes from. -->
<script lang="ts">
	import { clipped } from '#lib/ui/clipped.ts';
	import { goto, invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import type { Prompt, PromptList } from '#lib/api/types/prompts.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import PinButton from '#lib/ui/PinButton.svelte';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { count, realTime } from '#lib/format.ts';
	import { promptFolder, promptHref, promptName, promptsDependency, searchPrompts, setPromptPinned } from './api.ts';

	let { save, list }: { save: string; list: PromptList } = $props(); // pinned prompts first, in the order they were pinned, then by path

	let query = $state('');
	let inside = $state(false); // search the prompts' text too, not only their names
	let matches = $state<Set<string> | null>(null); // paths whose text matches, when searching inside
	let searching = $state(false);
	let sort = $state<Sort>({ key: 'folder', descending: false });
	let error = $state('');

	const source = (p: Prompt) => list.layers[p.layer]?.source ?? '';
	// The player's own copy is what they changed; it gets the changed mark.
	const mine = (p: Prompt) => list.layers[p.layer]?.kind === 'overlay';
	const folderName = (p: Prompt) => promptFolder(p.path) || 'prompts';

	// Searching inside reads every prompt on the server, so it waits for a
	// pause in typing.
	let timer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		const q = query.trim();
		clearTimeout(timer);
		if (!inside || !q) {
			matches = null;
			searching = false;
			return;
		}
		searching = true;
		timer = setTimeout(async () => {
			try {
				const found = (await searchPrompts(q)).paths;
				matches = new Set(found.map((p) => p.toLowerCase()));
				error = '';
			} catch (e) {
				error = e instanceof ApiError ? e.message : String(e);
			} finally {
				searching = false;
			}
		}, 250);
	});

	let grouped = $derived(sort.key === 'folder');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const found = list.prompts.filter((p) => !q || p.path.toLowerCase().includes(q) || matches?.has(p.path.toLowerCase()));
		const byKey = sorted(found, sort, {
			name: (p) => promptName(p.path),
			// Top-level prompts first, then each folder's; a space keeps a
			// folder's own files ahead of its subfolders' when collated.
			folder: (p) => promptFolder(p.path) + ' ' + promptName(p.path),
			source: (p) => source(p)
		});
		// In their natural order, pinned prompts sit first, as pinned.
		return grouped ? [...found.filter((p) => p.pinned), ...byKey.filter((p) => !p.pinned)] : byKey;
	});
	// The heading a row sits under when grouped: Pinned, or its folder.
	const groupOf = (p: Prompt) => (p.pinned ? 'Pinned' : folderName(p));

	async function togglePin(p: Prompt) {
		error = '';
		try {
			await setPromptPinned(p.path, !p.pinned);
			await invalidate(promptsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	const selection = new Selection(() => shown, (p) => p.path);
	let current = $derived(selection.current);
	const href = (p: Prompt) => promptHref(save, p.path);

	useShortcuts(() => {
		const p = selection.current;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: 'Open', disabled: !p, run: () => p && goto(href(p)) },
			{ key: 'p', shown: 'P', label: p?.pinned ? 'Unpin' : 'Pin', disabled: !p, run: () => p && togglePin(p) }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Prompts</h1>
	<span class="muted n" use:clipped>{count(list.prompts.length)} from {list.layers.length} {list.layers.length === 1 ? 'layer' : 'layers'}</span>
	<label class="inside"><input type="checkbox" bind:checked={inside} /> Search inside</label>
	<input type="search" bind:value={query} placeholder={inside ? 'Search names and text' : 'Filter by name'} aria-label="Filter prompts" />
</div>
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !list.prompts.length}
	<p class="muted">No prompts were found. Check that SkyrimNet’s folder, and under Mod Organizer 2 its mod folder, are set on the Playthroughs page.</p>
{:else if !shown.length}
	<p class="muted">{searching ? 'Searching…' : `Nothing matches “${query}”.`}</p>
{:else}
	<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<SortHeader label="Prompt" key="name" bind:sort />
					<SortHeader label="Folder" key="folder" bind:sort class="folder" />
					<SortHeader label="From" key="source" bind:sort class="from" />
					<th class="pin" aria-label="Pinned"></th>
				</tr>
			</thead>
			<tbody>
				{#each shown as p, i (p.path)}
					{#if grouped && (i === 0 || groupOf(p) !== groupOf(shown[i - 1]!))}
						<tr class="group"><th colspan="5" class="caps" scope="rowgroup">{groupOf(p)}</th></tr>
					{/if}
					<tr aria-selected={selection.is(p)} onclick={() => selection.select(p)} ondblclick={() => goto(href(p))}>
						<td class="mark">{#if selection.is(p)}<Diamond />{/if}</td>
						<td class="name"><a href={href(p)}>{promptName(p.path)}</a></td>
						<td class="folder muted">{folderName(p)}</td>
						<td class="from muted">{#if mine(p)}<span class="changed">Your copy</span>{:else}{source(p)}{/if}</td>
						<td class="pin"><PinButton pinned={p.pinned} name={promptName(p.path)} onclick={() => togglePin(p)} /></td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label="Prompt">
			{#if current}
				<h2 class="caps">{promptName(current.path)}</h2>
				<p class="path">{current.path}</p>
				<dl>
					<dt class="caps">In use</dt><dd>{#if mine(current)}<span class="changed">Your copy</span>{:else}{source(current)}{/if}</dd>
					{#if current.layers.length > 1}
						<dt class="caps">Also in</dt><dd>{current.layers.slice(1).map((i) => list.layers[i]?.source ?? '').join(', ')}</dd>
					{/if}
					<dt class="caps">Changed</dt><dd>{realTime(current.modified)}</dd>
				</dl>
				<div class="actions">
					<a class="button primary" href={href(current)}>Open</a>
					<button type="button" class="button" aria-pressed={current.pinned} onclick={() => togglePin(current!)}>{current.pinned ? 'Unpin' : 'Pin'}</button>
				</div>
			{/if}
		</aside>
	</div>
{/if}

<style>
	/* Keeps the filter over the table, not the pane. */
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
		margin-right: calc(clamp(18rem, 26vw, 24rem) + var(--space-5));
	}
	/* The bar sits on one baseline. The title's capitals have nothing
	   below that baseline, while a control's box does (its text's descent
	   and padding), so the title is set down a little to center its letters
	   on the controls, which puts its baseline near their bottom edge. */
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		position: relative;
		top: 0.21em;
	}
	.inside {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
		white-space: nowrap;
	}
	.bar input[type='search'] {
		width: 16rem;
		flex: 0 1 16rem;
		min-width: 9rem;
	}
	/* Narrower than 90rem, the bar needs the pane's width too. */
	@media (max-width: 90rem) {
		.bar {
			margin-right: 0;
		}
	}
	.bar .n {
		white-space: nowrap;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	/* The name takes what the folder and source leave; neither wraps. */
	.split :global(.folder) {
		width: 11rem;
		white-space: nowrap;
	}
	.split :global(.from) {
		width: 13rem;
		white-space: nowrap;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	/* The item card: a full-height zone of the menu beside the list. */
	.pane {
		position: sticky;
		top: 0;
		min-height: calc(100vh - 7rem);
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.path {
		font: 400 var(--step--1) / 1.5 var(--font-mono);
		color: var(--muted);
		overflow-wrap: anywhere;
	}
	.pane dl {
		display: grid;
		grid-template-columns: 5.5rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: var(--space-4) 0;
	}
	.pane dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	.pane dd {
		margin: 0;
		overflow-wrap: anywhere;
	}
	.actions {
		display: flex;
		gap: var(--space-2);
	}
	.actions a {
		text-decoration: none;
	}
	@media (max-width: 60rem) {
		.bar {
			margin-right: 0;
			flex-wrap: wrap;
		}
		.split {
			grid-template-columns: 1fr;
		}
		.pane {
			position: static;
			min-height: 0;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
		.split :global(.folder) {
			display: none;
		}
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
	}
	/* A row you can select tints under the pointer, as a bio's sections do. */
	tbody tr[aria-selected]:not([aria-selected='true']):hover td {
		background: var(--glass);
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	tbody tr {
		cursor: default;
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .mark {
		border-bottom-color: var(--rule);
	}
	.pin {
		width: 2.5rem;
		padding: 0 var(--space-2);
		text-align: right;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .pin {
		border-bottom-color: var(--rule);
	}
	.name a {
		text-decoration: none;
	}
	.name a:hover {
		text-decoration: underline;
	}
	.changed {
		color: var(--changed);
	}
	td .changed::before {
		content: '';
		display: inline-block;
		width: 0.45rem;
		height: 0.45rem;
		margin-right: var(--space-2);
		background: var(--changed);
		border-radius: 1px;
		vertical-align: 1px;
	}
	.group th {
		padding: var(--space-4) var(--space-3) var(--space-1);
		text-align: left;
		font-size: var(--step--1);
		color: var(--accent);
		border-bottom: 1px solid var(--rule-strong);
	}
</style>
