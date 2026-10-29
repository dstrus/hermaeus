<!-- A list of SkyrimNet records (actions, triggers), grouped by where each
     comes from, in the Plugins page's order, with the selected one in the
     pane. Read-only: which records apply is set by the plugins they come in. -->
<script lang="ts" generics="T extends Origin & { name: string; description: string; enabled: boolean }">
	import type { Snippet } from 'svelte';
	import { fly } from 'svelte/transition';
	import type { Origin } from '#lib/api/types/records.ts';
	import { plural } from '#lib/format.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sourceLabel } from './api.ts';

	let {
		title,
		noun,
		lead,
		records,
		columns,
		cells,
		pane,
		note
	}: {
		title: string;
		noun: [string, string]; // one, many
		lead: string;
		records: T[];
		columns: { label: string; class?: string }[]; // after the name
		cells: Snippet<[T]>;
		pane: Snippet<[T]>;
		note?: Snippet;
	} = $props();

	let query = $state('');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return q ? records.filter((r) => [r.name, r.description, r.source, r.id].some((v) => v.toLowerCase().includes(q))) : records;
	});
	let off = $derived(records.filter((r) => !r.enabled).length);
	let sources = $derived(new Set(records.map((r) => r.layer)).size);

	const selection = new Selection(() => shown, (r) => `${r.layer}/${r.id}`, { param: 'item' });
	let current = $derived(selection.current);

	useShortcuts(() => [
		{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) }
	]);
</script>

<div class="bar">
	<h1 class="caps">{title}</h1>
	<span class="muted n" use:clipped>{plural(records.length, noun[0], noun[1])} from {plural(sources, 'source', 'sources')}{off ? ` · ${off} off` : ''}</span>
	<input type="search" bind:value={query} placeholder="Search {noun[1]}" aria-label="Search {noun[1]}" />
</div>
<p class="muted lead">{lead}</p>

<div class="split">
	<div class="list">
		{#if !records.length}
			<p class="muted">No {noun[1]} in your plugins or your own edits.</p>
		{:else if !shown.length}
			<p class="muted">Nothing matches “{query}”.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<th class="caps">{noun[0].charAt(0).toUpperCase() + noun[0].slice(1)}</th>
						{#each columns as c (c.label)}<th class="caps {c.class ?? ''}">{c.label}</th>{/each}
					</tr>
				</thead>
				<tbody>
					{#each shown as r, i (`${r.layer}/${r.id}`)}
						{#if i === 0 || shown[i - 1]!.layer !== r.layer}
							<tr class="group"><th colspan={columns.length + 2} class="caps" scope="rowgroup">{sourceLabel(r)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(r)} class:off={!r.enabled} onclick={() => selection.select(r)}>
							<td class="mark">{#if selection.is(r)}<Diamond />{/if}</td>
							<td class="name" use:clipped>{r.name}{#if !r.enabled}<span class="flag">off</span>{/if}{#if r.replaces}<span class="flag">replaces {r.replaces}’s</span>{/if}</td>
							{@render cells(r)}
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
		{#if note}<div class="note muted small">{@render note()}</div>{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label={noun[0]}>
		{#if current}
			{@const r = current}
			{#key `${r.layer}/${r.id}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{r.name}</h2>
					<p class="muted small">{sourceLabel(r)} · <span class="id">{r.path}</span></p>
					{#if !r.enabled}<p class="notice">Off: SkyrimNet doesn’t use it.</p>{/if}
					{#if r.replaces}<p class="muted small">Replaces the one from {r.replaces}.</p>{/if}
					{@render pane(r)}
				</div>
			{/key}
		{/if}
	</aside>
</div>

<style>
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
		margin-right: calc(clamp(18rem, 26vw, 24rem) + var(--space-5));
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		position: relative;
		top: 0.21em;
	}
	.bar .n {
		white-space: nowrap;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.bar input {
		margin-left: auto;
		flex: 0 1 16rem;
		min-width: 9rem;
	}
	.lead {
		max-width: 65ch;
		margin-bottom: var(--space-4);
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	.note {
		margin-top: var(--space-4);
	}
	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
	}
	th {
		padding: var(--space-2) var(--space-3);
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
	}
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	tbody :global(td) {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	tbody tr {
		cursor: default;
	}
	tbody tr[aria-selected]:not([aria-selected='true']):hover :global(td) {
		background: var(--glass);
	}
	tr[aria-selected='true'] :global(td) {
		background: var(--select);
	}
	tr.off :global(td) {
		color: var(--muted);
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	/* A flag is a word of metadata, never a button. */
	.flag {
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.flag::before {
		content: '· ';
	}
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
		overflow-wrap: anywhere;
	}
	.pane > div > .small {
		margin-bottom: var(--space-3);
		overflow-wrap: anywhere;
	}
	.pane .notice {
		margin: var(--space-3) 0;
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
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
	}
</style>
