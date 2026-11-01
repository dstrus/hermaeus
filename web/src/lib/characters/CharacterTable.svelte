<script lang="ts">
	import { clipped } from '#lib/ui/clipped.ts';
	import { goto, invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import { BusyMemories, type Character, type Group } from '#lib/api/types/characters.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import PinButton from '#lib/ui/PinButton.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { catalogHref } from '#lib/bios/api.ts';
	import { BioPreviews, type Preview } from '#lib/bios/preview.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { m } from '#lib/copy/messages.ts';
	import { count } from '#lib/format.ts';
	import { characterName, charactersDependency, setPinned } from './api.ts';

	let { save, characters }: { save: string; characters: Character[] } = $props(); // characters in group order

	let query = $state('');
	let sort = $state<Sort>({ key: 'group', descending: false });
	let error = $state('');

	const groupNames: Record<Group, string> = { ...m.characters.groups, busy: m.characters.groups.busy({ n: BusyMemories }) };

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = characters.filter((c) => !q || characterName(c).toLowerCase().includes(q) || c.bio.toLowerCase().includes(q));
		return sorted(list, sort, {
			name: (c) => (c.name ? c.name : `￿${c.uuid}`),
			memories: (c) => c.memories,
			diaries: (c) => c.diaries
		});
	});
	// In their natural order, characters sit under their group's heading.
	let grouped = $derived(sort.key === 'group');

	const selection = new Selection(() => shown, (c) => c.uuid, { param: 'item' });
	let current = $derived(selection.current);

	// The selected character's bio, as space in the pane permits.
	const previews = new BioPreviews(() => save);
	let preview = $state<Preview | null>(null);
	$effect(() => {
		const template = current?.bio;
		preview = null;
		if (!template) return;
		previews.of(template).then((p) => {
			if (current?.bio === template) preview = p;
		});
	});
	const href = (c: Character) => `/${encodeURIComponent(save)}/npc/${c.uuid}`;

	async function togglePin(c: Character) {
		error = '';
		try {
			await setPinned(save, c.uuid, !c.pinned);
			await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const c = selection.current;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: m.common.open, disabled: !c, run: () => c && goto(href(c)) },
			{
				key: 'p',
				shown: 'P',
				label: c?.pinned ? m.characters.unpin : m.characters.pin,
				disabled: !c || c.group === 'player' || live.gameRunning,
				run: () => c && togglePin(c)
			}
		];
	});
</script>

<div class="bar">
	<h1 class="caps">{m.characters.title}</h1>
	<Segmented
		value="here"
		options={[
			{ value: 'here', label: m.characters.which.here, title: m.characters.which.hereTitle },
			{ value: 'all', label: m.characters.which.all, title: m.characters.which.allTitle }
		]}
		label={m.characters.which.label}
		onchange={(v) => v === 'all' && goto(catalogHref(save))}
	/>
	<span class="muted n" use:clipped>{m.characters.table.stored({ count: characters.length })}</span>
	<input type="search" bind:value={query} placeholder={m.characters.table.filter} aria-label={m.characters.table.filterLabel} />
</div>
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !characters.length}
	<p class="muted">{m.characters.table.none}</p>
{:else if !shown.length}
	<p class="muted">{m.characters.noMatch({ query })}</p>
{:else}
	<div class="split">
	<table>
		<thead>
			<tr>
				<th class="mark" aria-label={m.common.selected}></th>
				<SortHeader label={m.characters.table.name} key="name" bind:sort />
				<SortHeader label={m.characters.table.memories} key="memories" bind:sort descendingFirst align="end" />
				<SortHeader label={m.characters.table.diary} key="diaries" bind:sort descendingFirst align="end" />
				<th class="plugin caps">{m.characters.plugin}</th>
				<th class="pin" aria-label={m.characters.groups.pinned}></th>
			</tr>
		</thead>
		<tbody>
			{#each shown as c, i (c.uuid)}
				{#if grouped && c.group !== shown[i - 1]?.group}
					<tr class="group"><th colspan="6" class="caps" scope="rowgroup">{groupNames[c.group]}</th></tr>
				{/if}
				<tr aria-selected={selection.is(c)} onclick={() => selection.select(c)} ondblclick={() => goto(href(c))}>
					<td class="mark">{#if selection.is(c)}<Diamond />{/if}</td>
					<td class="name">
						<a href={href(c)}>{characterName(c)}</a>
						{#if c.formerNames.length}<span class="muted"> · {m.characters.table.was({ names: c.formerNames.map((n) => n || m.characters.noNameLower).join(', ') })}</span>{/if}
					</td>
					<td class="n">{count(c.memories)}</td>
					<td class="n">{c.diaries ? count(c.diaries) : ''}</td>
					<td class="plugin muted">{c.modFile}</td>
					<td class="pin">
						{#if c.group !== 'player'}
							<PinButton pinned={c.pinned} name={characterName(c)} disabled={live.gameRunning} onclick={() => togglePin(c)} />
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
	<aside class="pane" aria-label={m.characters.character} use:fitToViewport>
		{#if current}
			<div class="title">
				<h2 class="caps">{characterName(current)}</h2>
				{#if current.group !== 'player'}
					<PinButton always size={16} pinned={current.pinned} name={characterName(current)} disabled={live.gameRunning} reason={m.characters.closeToPin} onclick={() => togglePin(current!)} />
				{/if}
			</div>
			<p class="muted">{groupNames[current.group]}{#if current.formerNames.length}{' · '}{m.characters.table.previously({ names: current.formerNames.map((n) => n || m.characters.noNameLower).join(', ') })}{/if}</p>
			<dl>
				<dt class="caps">{m.characters.table.memories}</dt><dd>{count(current.memories)}</dd>
				<dt class="caps">{m.characters.table.diary}</dt><dd>{m.characters.table.entries({ count: current.diaries })}</dd>
				{#if current.modFile}<dt class="caps">{m.characters.plugin}</dt><dd>{current.modFile}</dd>{/if}
				{#if current.bio}<dt class="caps">{m.characters.bioTemplate}</dt><dd>{current.bio}</dd>{/if}
				<dt class="caps">{m.characters.skyrimNetId}</dt><dd>{current.uuid}</dd>
			</dl>
			<div class="actions">
				<a class="button primary" href={href(current)}>{m.common.open}</a>
			</div>
			{#if live.gameRunning && current.group !== 'player'}<p class="notice warn">{m.characters.closeToPinNotice}</p>{/if}
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
	table {
		width: 100%;
		border-collapse: collapse;
	}
	/* The item card: a full-height zone of the menu beside the list. */
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
	/* The name, with the pin beside it as in the item's own header. */
	.pane .title {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-1);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
	}
	.pane dl {
		display: grid;
		grid-template-columns: 7.5rem 1fr;
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
	.pane .notice {
		margin-top: var(--space-3);
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
		.split :global(.plugin) {
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
	.pin {
		width: 2.5rem;
		padding: 0 var(--space-2);
		text-align: right;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .pin {
		border-bottom-color: var(--rule);
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
	td.n {
		text-align: right;
		width: 7rem;
	}
	.plugin {
		font-size: var(--step--1);
		text-align: left;
		padding: var(--space-2) var(--space-3);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
		font-weight: inherit;
	}
	tbody .plugin {
		border-bottom-color: var(--rule);
	}
	.group th {
		text-align: left;
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
		font-size: var(--step--1);
		border-bottom: 1px solid var(--rule-strong);
	}
</style>
