<!-- Every character SkyrimNet or a plugin ships a bio for, met or not, with
     where the bio comes from and what this playthrough has of its own. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Catalog, CatalogEntry } from '#lib/api/types/bios.ts';
	import type { Character } from '#lib/api/types/characters.ts';
	import { characterName } from '#lib/characters/api.ts';
	import { count } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { catalogHref, kinds, templateName } from './api.ts';

	let {
		save,
		catalog,
		saveBios,
		characters
	}: {
		save: string;
		catalog: Catalog;
		saveBios: Record<string, string[]>; // the kinds of save bio each template has here
		characters: Character[]; // met in this playthrough
	} = $props();

	let query = $state('');
	let sort = $state<Sort>({ key: 'name', descending: false });

	// Templates match the way SkyrimNet names files: ignoring case.
	let known = $derived(new Map(characters.filter((c) => c.bio).map((c) => [c.bio.toLowerCase(), c])));
	let own = $derived(new Map(Object.entries(saveBios).map(([t, k]) => [t.toLowerCase(), k])));

	const source = (e: CatalogEntry) => catalog.layers[e.layer]?.source ?? '';
	const met = (e: CatalogEntry) => known.get(e.template.toLowerCase());
	// here is what this playthrough has for the template: the character, and any save bios.
	const here = (e: CatalogEntry) => {
		const c = met(e);
		const parts = c ? [characterName(c)] : [];
		for (const k of own.get(e.template.toLowerCase()) ?? []) parts.push((kinds[k]?.label ?? k).toLowerCase() + ' bio');
		return parts.join(' · ');
	};
	// A met character's bio opens on their own Bio tab.
	const href = (e: CatalogEntry) => {
		const c = met(e);
		return c ? `/${encodeURIComponent(save)}/npc/${c.uuid}/bio` : catalogHref(save, e.template);
	};

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = catalog.bios.filter((e) => !q || e.template.toLowerCase().includes(q) || templateName(e.template).toLowerCase().includes(q));
		return sorted(list, sort, {
			name: (e) => templateName(e.template),
			template: (e) => e.template,
			source: (e) => source(e),
			here: (e) => here(e) || '￿'
		});
	});

	const selection = new Selection(() => shown, (e) => e.template);
	let current = $derived(selection.current);

	useShortcuts(() => {
		const e = selection.current;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: 'Open', disabled: !e, run: () => e && goto(href(e)) },
			{ key: 'Escape', shown: 'Esc', label: 'This playthrough', run: () => goto(`/${encodeURIComponent(save)}`) }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Characters</h1>
	<Segmented
		value="all"
		options={[
			{ value: 'here', label: 'This playthrough', title: 'Everyone with something stored in this playthrough' },
			{ value: 'all', label: 'Every bio', title: 'Everyone SkyrimNet or a plugin ships a bio for, met or not' }
		]}
		label="Which characters"
		onchange={(v) => v === 'here' && goto(`/${encodeURIComponent(save)}`)}
	/>
	<span class="muted n">{count(catalog.bios.length)} with a bio, from {catalog.layers.length} {catalog.layers.length === 1 ? 'layer' : 'layers'}</span>
	<input type="search" bind:value={query} placeholder="Filter by name" aria-label="Filter bios" />
</div>

{#if !catalog.bios.length}
	<p class="muted">No bios were found. Check that SkyrimNet’s folder, and under Mod Organizer 2 its mod folder, are set on the Playthroughs page.</p>
{:else if !shown.length}
	<p class="muted">No one matches “{query}”.</p>
{:else}
	<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<SortHeader label="Character" key="name" bind:sort />
					<SortHeader label="Template" key="template" bind:sort class="template" />
					<SortHeader label="From" key="source" bind:sort class="from" />
					<SortHeader label="In this playthrough" key="here" bind:sort class="here" />
				</tr>
			</thead>
			<tbody>
				{#each shown as e (e.template)}
					<tr aria-selected={selection.is(e)} onclick={() => selection.select(e)} ondblclick={() => goto(href(e))}>
						<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
						<td class="name"><a href={href(e)}>{templateName(e.template)}</a></td>
						<td class="template muted">{e.template}</td>
						<td class="from muted">{source(e)}</td>
						<td class="here">{here(e)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label="Character">
			{#if current}
				<h2 class="caps">{templateName(current.template)}</h2>
				<p class="path">{current.template}</p>
				<dl>
					<dt class="caps">Bio from</dt><dd>{source(current)}</dd>
					<dt class="caps">Here</dt><dd>{here(current) || 'Not met, and no bio of this playthrough’s own'}</dd>
				</dl>
				<div class="actions">
					<a class="button primary" href={href(current)}>{met(current) ? 'Open character' : 'Open bio'}</a>
				</div>
			{/if}
		</aside>
	</div>
{/if}

<style>
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
	.bar input {
		margin-left: auto;
		width: 16rem;
		flex: 0 1 16rem;
		min-width: 9rem;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	/* Fixed columns, so the table never grows past its column into the pane. */
	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
	}
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
		.split :global(.from) {
			display: none;
		}
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
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
	.name a {
		text-decoration: none;
	}
	.name a:hover {
		text-decoration: underline;
	}
	td.template {
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	/* The template column gives way first, since the pane shows it. */
	.split :global(.template) {
		width: 11rem;
	}
	.split :global(.from),
	.split :global(.here) {
		width: 11rem;
	}
	@media (max-width: 90rem) {
		.split :global(.template) {
			display: none;
		}
	}
</style>
