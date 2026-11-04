<script lang="ts">
	import { goto } from '$app/navigation';
	import FolderForm from '#lib/folders/FolderForm.svelte';
	import Preview from '#lib/summary/Preview.svelte';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { count, realTime } from '#lib/format.ts';
	import type { Playthrough } from '#lib/api/types/store.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let playthroughs = $derived(data.status.playthroughs); // most recently played first

	const selection = new Selection(() => playthroughs, (p) => p.id, { param: 'item' });
	const open = (p: Playthrough) => goto(`/${encodeURIComponent(p.id)}`);

	// Changing the SkyrimNet folder has its own address, so Back ends it.
	let changing = $derived(param('mode') === 'folder');
	const closeFolder = () => leave('mode'); // saved or not

	useShortcuts(() =>
		changing
			? [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: closeFolder }]
			: [
					{ key: 'ArrowDown', label: m.common.next, hidden: true, run: () => selection.move(1) },
					{ key: 'ArrowUp', label: m.common.previous, hidden: true, run: () => selection.move(-1) },
					{ key: 'Enter', label: m.common.open, disabled: !selection.current, run: () => selection.current && open(selection.current) }
				]
	);
</script>

{#if changing}
	<FolderForm current={{ path: data.status.skyrimNetDir, modPath: data.status.modDirDetected ? '' : data.status.modDir }} close={closeFolder} />
{:else}
	<h1 class="caps">{m.playthroughs.title}</h1>
	<p class="muted">
		<Fill text={m.playthroughs.in.text}>{#snippet path()}<span class="path">{data.status.skyrimNetDir}</span>{/snippet}</Fill>
		· <button type="button" class="textbutton" onclick={() => setParams({ mode: 'folder' })}>{m.playthroughs.changeFolder}</button>
	</p>
	<p class="muted defaults">
		{#if data.status.modDir}
			<Fill text={(data.status.modDirDetected ? m.playthroughs.modDirFound : m.playthroughs.modDirSet).text}>{#snippet path()}<span class="path">{data.status.modDir}</span>{/snippet}</Fill>
		{:else if !data.status.defaultBios}
			<Fill text={m.playthroughs.noDefaults.text}>{#snippet flag()}<code>-mod-dir</code>{/snippet}</Fill>
		{/if}
	</p>

	{#if !playthroughs.length}
		<p class="empty">{m.playthroughs.none}</p>
	{:else}
		<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label={m.common.selected}></th>
					<th class="caps">{m.playthroughs.player}</th>
					<th class="caps">{m.playthroughs.saveId}</th>
					<th class="caps">{m.playthroughs.lastWritten}</th>
					<th class="caps n">{m.playthroughs.size}</th>
				</tr>
			</thead>
			<tbody>
				{#each playthroughs as p (p.id)}
					<tr aria-selected={selection.is(p)} onclick={() => selection.select(p)} ondblclick={() => open(p)}>
						<td class="mark">{#if selection.is(p)}<Diamond />{/if}</td>
						<td><a href="/{encodeURIComponent(p.id)}">{p.player || m.overview.noPlayer}</a></td>
						<td class="muted">{p.id}</td>
						<td>{realTime(p.modified)}</td>
						<td class="n muted">{m.playthroughs.kb({ size: count(Math.round(p.size / 1024)) })}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label={m.playthroughs.pane} use:fitToViewport>
			{#if selection.current}<Preview playthrough={selection.current} />{/if}
		</aside>
		</div>
	{/if}
{/if}

<style>
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	h1 + p {
		margin: var(--space-2) 0 0;
	}
	.defaults {
		margin: var(--space-1) 0 var(--space-5);
		max-width: 70ch;
	}
	.path {
		color: var(--ink);
		word-break: break-all;
	}
	.empty {
		max-width: 60ch;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	/* The selected playthrough at a glance, as other lists show their item. */
	.pane {
		position: sticky;
		top: 0;
		display: flex;
		flex-direction: column;
		min-height: calc(100vh - 7rem); /* fitToViewport sets the height */
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: hidden;
	}
	@media (max-width: 60rem) {
		.split {
			grid-template-columns: 1fr;
		}
		.pane {
			position: static;
			min-height: 0;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule-strong);
		font-weight: inherit;
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
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	td a {
		font-size: var(--step-1);
		text-decoration: none;
	}
	td a:hover {
		text-decoration: underline;
	}
	.n {
		text-align: right;
	}
</style>
