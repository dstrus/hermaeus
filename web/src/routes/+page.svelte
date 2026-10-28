<script lang="ts">
	import { goto } from '$app/navigation';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { count, realTime } from '#lib/format.ts';
	import type { Playthrough } from '#lib/api/types/store.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let playthroughs = $derived(data.status.playthroughs); // most recently played first

	const selection = new Selection(() => playthroughs, (p) => p.id);
	const open = (p: Playthrough) => goto(`/${encodeURIComponent(p.id)}`);

	useShortcuts(() => [
		{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
		{ key: 'Enter', label: 'Open', disabled: !selection.current, run: () => selection.current && open(selection.current) }
	]);
</script>

<h1 class="caps">Playthroughs</h1>
<p class="muted">In <span class="path">{data.status.skyrimNetDir}</span></p>
<p class="muted defaults">
	{#if data.status.modDir}
		SkyrimNet’s own files {data.status.modDirDetected ? 'were found' : 'are'} in <span class="path">{data.status.modDir}</span>.
	{:else if !data.status.defaultBios}
		SkyrimNet’s default bios weren’t found, so characters without a save bio show none. Under Mod Organizer 2, start Hermaeus with
		<code>-mod-dir</code> pointing at SkyrimNet’s mod folder.
	{/if}
</p>

{#if !playthroughs.length}
	<p class="empty">There are no SkyrimNet databases here yet. SkyrimNet makes one the first time you load a save with it.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th class="mark" aria-label="Selected"></th>
				<th class="caps">Player</th>
				<th class="caps">Save ID</th>
				<th class="caps">Last written</th>
				<th class="caps n">Size</th>
			</tr>
		</thead>
		<tbody>
			{#each playthroughs as p (p.id)}
				<tr aria-selected={selection.is(p)} onclick={() => selection.select(p)} ondblclick={() => open(p)}>
					<td class="mark">{#if selection.is(p)}<Diamond />{/if}</td>
					<td><a href="/{encodeURIComponent(p.id)}">{p.player || 'No player yet'}</a></td>
					<td class="muted">{p.id}</td>
					<td>{realTime(p.modified)}</td>
					<td class="n muted">{count(Math.round(p.size / 1024))} KB</td>
				</tr>
			{/each}
		</tbody>
	</table>
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
