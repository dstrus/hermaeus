<!-- Every character SkyrimNet or a plugin ships a bio for, met or not, with
     where the bio comes from and what this playthrough has of its own. -->
<script lang="ts">
	import { clipped } from '#lib/ui/clipped.ts';
	import { goto } from '$app/navigation';
	import type { Catalog, CatalogEntry } from '#lib/api/types/bios.ts';
	import type { Character } from '#lib/api/types/characters.ts';
	import { characterName } from '#lib/characters/api.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { count } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { catalogHref, kinds, templateName } from './api.ts';
	import { BioPreviews, type Preview } from './preview.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { pluginHref } from '#lib/plugins/links.ts';
	import { param, setParams } from '#lib/ui/address.svelte.ts';

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
	// Narrowed to the bios one plugin brings, from its pane on the Plugins page.
	let from = $derived(param('from'));
	let fromLayer = $derived(from ? catalog.layers.findIndex((l) => l.id === from) : -1);
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
		for (const k of own.get(e.template.toLowerCase()) ?? []) parts.push(m.bios.kindBio({ kind: (kinds[k]?.label ?? k).toLowerCase() }));
		return parts.join(' · ');
	};
	// A met character's bio opens on their own Bio tab.
	const href = (e: CatalogEntry) => {
		const c = met(e);
		return c ? `/${encodeURIComponent(save)}/npc/${c.uuid}/bio` : catalogHref(save, e.template);
	};

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = catalog.bios.filter((e) => (fromLayer < 0 || e.layer === fromLayer) && (!q || e.template.toLowerCase().includes(q) || templateName(e.template).toLowerCase().includes(q)));
		return sorted(list, sort, {
			name: (e) => templateName(e.template),
			template: (e) => e.template,
			source: (e) => source(e),
			here: (e) => here(e) || '￿'
		});
	});

	const selection = new Selection(() => shown, (e) => e.template, { param: 'item' });
	let current = $derived(selection.current);

	// The selected character's bio, as space in the pane permits.
	const previews = new BioPreviews(() => save);
	let preview = $state<Preview | null>(null);
	$effect(() => {
		const template = current?.template;
		preview = null;
		if (!template) return;
		previews.of(template).then((p) => {
			if (current?.template === template) preview = p;
		});
	});

	useShortcuts(() => {
		const e = selection.current;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: m.common.open, disabled: !e, run: () => e && goto(href(e)) },
			{ key: 'Escape', shown: 'Esc', label: m.characters.which.here, run: () => goto(`/${encodeURIComponent(save)}/characters`) }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">{m.characters.title}</h1>
	<Segmented
		value="all"
		options={[
			{ value: 'here', label: m.characters.which.here, title: m.characters.which.hereTitle },
			{ value: 'all', label: m.characters.which.all, title: m.characters.which.allTitle }
		]}
		label={m.characters.which.label}
		onchange={(v) => v === 'here' && goto(`/${encodeURIComponent(save)}/characters`)}
	/>
	{#if fromLayer >= 0}
		<span class="muted n" use:clipped><Fill text={m.bios.catalog.fromPlugin.text} count={count(shown.length)}>{#snippet plugin()}<a href={pluginHref(save, from!)}>{catalog.layers[fromLayer]?.source}</a>{/snippet}</Fill> · <button type="button" class="textbutton" onclick={() => setParams({ from: null })}>{m.common.showAll}</button></span>
	{:else}
		<span class="muted n" use:clipped>{m.bios.catalog.withBio({ bios: count(catalog.bios.length), count: catalog.layers.length })}</span>
	{/if}
	<input type="search" bind:value={query} placeholder={m.bios.catalog.filter} aria-label={m.bios.catalog.filterLabel} />
</div>

{#if !catalog.bios.length}
	<p class="muted">{m.bios.catalog.none}</p>
{:else if !shown.length}
	<p class="muted">{m.characters.noMatch({ query })}</p>
{:else}
	<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label={m.common.selected}></th>
					<SortHeader label={m.bios.catalog.character} key="name" bind:sort />
					<SortHeader label={m.bios.catalog.template} key="template" bind:sort class="template" />
					<SortHeader label={m.bios.catalog.from} key="source" bind:sort class="from" />
					<SortHeader label={m.bios.catalog.here} key="here" bind:sort class="here" />
				</tr>
			</thead>
			<tbody>
				{#each shown as e (e.template)}
					<tr aria-selected={selection.is(e)} onclick={() => selection.select(e)} ondblclick={() => goto(href(e))}>
						<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
						<td class="name" use:clipped><a href={href(e)}>{templateName(e.template)}</a></td>
						<td class="template muted" use:clipped>{e.template}</td>
						<td class="from muted" use:clipped>{source(e)}</td>
						<td class="here" use:clipped>{here(e)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label={m.characters.character} use:fitToViewport>
			{#if current}
				<h2 class="caps">{templateName(current.template)}</h2>
				<p class="path">{current.template}</p>
				<dl>
					<dt class="caps">{m.bios.catalog.bioFrom}</dt><dd>{source(current)}</dd>
					<dt class="caps">{m.bios.catalog.hereLabel}</dt><dd>{here(current) || m.bios.catalog.notHere}</dd>
				</dl>
				<div class="actions">
					<a class="button primary" href={href(current)}>{met(current) ? m.bios.catalog.openCharacter : m.bios.catalog.openBio}</a>
				</div>
				{#if preview}
					<section class="preview">
						<h3 class="caps">{m.bios.previewHeading({ kind: preview.kind })}</h3>
						<p>{preview.text}</p>
					</section>
				{/if}
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
		display: flex;
		flex-direction: column;
		min-height: calc(100vh - 7rem); /* fitToViewport sets the height */
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: hidden;
	}
	/* The bio preview takes what height is left, fading out at the bottom. */
	.preview {
		flex: 1;
		min-height: 4rem;
		margin-top: var(--space-4);
		overflow: hidden;
		mask-image: linear-gradient(to bottom, #000 calc(100% - 3rem), transparent);
	}
	.preview h3 {
		font-size: var(--step--1);
		color: var(--muted);
		margin-bottom: var(--space-2);
	}
	.preview p {
		font-size: var(--step--1);
		line-height: 1.6;
		white-space: pre-wrap;
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
