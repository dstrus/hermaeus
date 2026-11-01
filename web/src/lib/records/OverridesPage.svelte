<!-- Every file a plugin or the player's own edits put in place of
     SkyrimNet's, or of another plugin's, with what each changed: the copy in
     use read against the one it replaces. -->
<script lang="ts">
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import type { Override, OverrideList, Version } from '#lib/api/types/records.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { realTime } from '#lib/format.ts';
	import { editHref, pluginHref } from '#lib/plugins/links.ts';
	import { param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { diffLines, hunks, markWords } from '#lib/ui/linediff.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { getVersions, layerName } from './api.ts';

	let { list }: { list: OverrideList } = $props();
	const msg = m.records.overrides;

	let save = $derived(page.params.save ?? '');
	const layerAt = (i: number | undefined) => (i === undefined ? undefined : list.layers[i]);
	const winner = (o: Override) => layerAt(o.providers[0]?.layer);
	const fileName = (path: string) => path.slice(path.lastIndexOf('/') + 1);
	const folder = (path: string) => path.slice(0, Math.max(0, path.lastIndexOf('/')));

	// Narrowed by a search, and to the flagged ones with ?show=flagged.
	let query = $state('');
	let flaggedOnly = $derived(param('show') === 'flagged');
	let flagged = $derived(list.overrides.filter((o) => o.stale || o.conflict).length);
	// Grouped by the layer whose copy is used, in the Plugins page's order.
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.overrides
			.filter((o) => (!flaggedOnly || o.stale || o.conflict) && (!q || [o.path, ...o.providers.map((p) => layerName(layerAt(p.layer)))].some((v) => v.toLowerCase().includes(q))))
			.sort((a, b) => (a.providers[0]?.layer ?? 0) - (b.providers[0]?.layer ?? 0) || a.path.localeCompare(b.path));
	});

	const selection = new Selection(() => shown, (o) => o.path, { param: 'item' });
	let current = $derived(selection.current);

	// The selected file's copies, read when it's selected.
	let versions = $state<Version[] | null>(null);
	let loadError = $state('');
	$effect(() => {
		const path = current?.path;
		versions = null;
		loadError = '';
		untrack(() => (against = ''));
		if (!path) return;
		let gone = false;
		getVersions(path)
			.then((v) => {
				if (!gone) versions = v;
			})
			.catch((e) => {
				if (!gone) loadError = e instanceof ApiError ? e.message : String(e);
			});
		return () => {
			gone = true;
		};
	});
	// The copy in use, read against one it replaces: the next one down
	// unless another is chosen.
	let against = $state('');
	let using = $derived(versions?.[0]);
	let lower = $derived(versions?.slice(1) ?? []);
	let base = $derived(lower.find((v) => String(v.layer) === against) ?? lower[0]);
	let choices = $derived(lower.map((v) => ({ value: String(v.layer), label: layerName(layerAt(v.layer)), title: v.source })));
	let diff = $derived(using?.text !== undefined && base?.text !== undefined ? diffLines(base.text, using.text) : null);
	let changes = $derived(diff ? hunks(diff) : []);
	let words = $derived(diff ? markWords(diff) : new Map());
	let counts = $derived(diff ? { added: diff.filter((l) => l.kind === 'added').length, removed: diff.filter((l) => l.kind === 'removed').length } : null);

	useShortcuts(() => [
		{ key: 'ArrowDown', label: m.common.next, hidden: true, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: m.common.previous, hidden: true, run: () => selection.move(-1) }
	]);
</script>

<div class="bar">
	<h1 class="caps">{msg.title}</h1>
	<span class="muted n" use:clipped>{msg.count({ count: list.overrides.length })}</span>
	{#if flagged}
		<button type="button" class="textbutton filter" aria-pressed={flaggedOnly} onclick={() => setParams({ show: flaggedOnly ? null : 'flagged' })}>{flaggedOnly ? m.common.showAll : msg.showFlagged({ count: flagged })}</button>
	{/if}
	<input type="search" bind:value={query} placeholder={msg.search} aria-label={msg.search} />
</div>
<p class="muted lead">{msg.lead}</p>

<div class="split">
	<div class="list">
		{#if !list.overrides.length}
			<p class="muted">{msg.none}</p>
		{:else if !shown.length}
			<p class="muted">{query ? m.records.noMatch({ query }) : msg.nothingToCheck}</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label={m.common.selected}></th>
						<th class="caps">{msg.file}</th>
						<th class="caps instead">{msg.inPlaceOf}</th>
					</tr>
				</thead>
				<tbody>
					{#each shown as o, i (o.path)}
						{#if i === 0 || shown[i - 1]!.providers[0]?.layer !== o.providers[0]?.layer}
							<tr class="group"><th colspan="3" class="caps" scope="rowgroup" class:mine={winner(o)?.kind === 'overlay'}>{layerName(winner(o))}</th></tr>
						{/if}
						<tr aria-selected={selection.is(o)} onclick={() => selection.select(o)}>
							<td class="mark">{#if selection.is(o)}<Diamond />{/if}</td>
							<td use:clipped title={o.path}><span class="file">{fileName(o.path)}</span>{#if o.stale}<span class="flag warn">{msg.older}</span>{/if}{#if o.conflict}<span class="flag warn">{msg.conflict}</span>{/if}<span class="folder">{folder(o.path)}</span></td>
							<td class="instead muted" use:clipped>
								{#if o.providers.length > 1}{o.providers.slice(1).map((p) => layerName(layerAt(p.layer))).join(m.records.listJoin)}{:else}{msg.added}{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label={msg.file}>
		{#if current}
			{@const o = current}
			{@const w = winner(o)}
			{@const edit = editHref(save, o.path)}
			{#key o.path}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2>{fileName(o.path)}</h2>
					<p class="muted small id">{o.path}</p>
					<p class="small">
						<Fill
							text={(o.providers.length > 1 ? msg.usedInPlaceOf : msg.usedAdding).text}
							layers={o.providers
								.slice(1)
								.map((p) => layerName(layerAt(p.layer)))
								.join(m.records.thenJoin)}
							>{#snippet used()}{#if w?.kind === 'pack'}<Fill text={msg.pluginUsed.text}>{#snippet plugin()}<a href={pluginHref(save, w.id)}>{w.source}</a>{/snippet}</Fill>{:else}<Fill text={msg.ownUsed.text}>{#snippet own()}<span class="changed">{msg.ownCopy}</span>{/snippet}</Fill>{/if}{/snippet}</Fill
						>
						{#if edit}<a href={edit.href}>{msg.openOn({ page: edit.page })}</a>{/if}
					</p>
					{#if o.stale}<p class="notice warn">{msg.stale}</p>{/if}
					{#if o.conflict}<p class="notice warn">
							{msg.conflictNote({
								winner: layerName(w),
								others: o.providers
									.slice(1)
									.filter((p) => layerAt(p.layer)?.kind === 'pack')
									.map((p) => layerName(layerAt(p.layer)))
									.join(msg.andJoin)
							})}
						</p>{/if}

					{#if loadError}
						<p class="notice error">{loadError}</p>
					{:else if using && base}
						<div class="against">
							{#if choices.length > 1}
								<Segmented bind:value={() => String(base?.layer ?? ''), (v) => (against = v)} options={choices} label={msg.comparedWith} />
							{:else}
								<span class="caps muted small">{msg.against({ layer: layerName(layerAt(base.layer)) })}</span>
							{/if}
						</div>
						{#if !diff}
							<p class="muted small">{msg.notComparable}</p>
						{:else if counts && !counts.added && !counts.removed}
							<p class="muted small">{using.crlf !== base.crlf ? msg.sameButEndings : msg.same}</p>
						{:else if counts}
							<p class="muted small key">
								<Fill text={msg.changes.text} layer={layerName(layerAt(base.layer))} time={realTime(base.modified)}
									>{#snippet added()}<ins>{msg.linesAdded({ count: counts.added })}</ins>{/snippet}{#snippet removed()}<del>{msg.linesRemoved({ count: counts.removed })}</del>{/snippet}</Fill
								>
							</p>
							<div class="diff" role="table" aria-label={msg.changesLabel}>
								{#each changes as h, i (i)}
									{#if 'skipped' in h}
										<div class="skipped muted">{msg.linesUnchanged({ count: h.skipped })}</div>
									{:else}
										{#each h.lines as l, j (j)}
											<div class="line {l.kind}" role="row"><span class="sign" aria-label={l.kind === 'same' ? msg.lineSame : l.kind === 'added' ? msg.lineAdded : msg.lineRemoved} role="cell">{l.kind === 'added' ? '+' : l.kind === 'removed' ? '−' : ''}</span><span role="cell">{#if words.get(l)}{#each words.get(l) as p, k (k)}{#if p.changed}<mark>{p.text}</mark>{:else}{p.text}{/if}{/each}{:else}{l.text || ' '}{/if}</span></div>
										{/each}
									{/if}
								{/each}
							</div>
						{/if}
					{:else if using}
						{#if using.text !== undefined}<pre class="whole">{using.text}</pre>{:else}<p class="muted small">{m.records.notText}</p>{/if}
					{/if}
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
		margin-right: calc(var(--pane) + var(--space-5));
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
	.bar .filter {
		margin-left: auto;
		white-space: nowrap;
	}
	.bar input {
		flex: 0 1 14rem;
		min-width: 9rem;
	}
	.bar .n + input {
		margin-left: auto;
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
	}
	.changed {
		color: var(--changed);
	}
	/* The pane holds a comparison, so it takes the larger share. */
	.bar,
	.split {
		--pane: clamp(22rem, 42vw, 46rem);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) var(--pane);
		gap: var(--space-5);
		align-items: start;
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
	th.instead {
		width: 42%;
	}
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	.group th.mine {
		color: var(--changed);
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
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
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	/* The file's name leads; its folder follows, quieter. */
	.folder {
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
		font-family: var(--font-mono);
	}
	/* A flag is a word of metadata, never a button. */
	.flag {
		margin-left: var(--space-2);
		font-size: var(--step--1);
	}
	.flag::before {
		content: '· ';
		color: var(--muted);
	}
	.flag.warn {
		color: var(--warn);
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
	.pane .small {
		margin-bottom: var(--space-2);
		overflow-wrap: anywhere;
	}
	.pane .small a:last-child {
		margin-left: var(--space-1);
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.against {
		margin: var(--space-4) 0 var(--space-3);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.key {
		margin-bottom: var(--space-3);
	}
	/* Added lines carry a wash of the accent; removed ones are struck, in a
	   wash of the error color, as in a bio's comparison. */
	ins {
		text-decoration: none;
		background: color-mix(in srgb, var(--accent) 22%, transparent);
	}
	del {
		text-decoration-color: var(--danger);
		background: color-mix(in srgb, var(--danger) 12%, transparent);
	}
	.diff,
	.whole {
		margin: 0 0 var(--space-5);
		font: 400 var(--step--1) / 1.55 var(--font-mono);
	}
	.whole {
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	.line {
		display: grid;
		grid-template-columns: 1.25rem minmax(0, 1fr);
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		tab-size: 4;
	}
	.line.same {
		color: var(--muted);
	}
	.line.added {
		background: color-mix(in srgb, var(--accent) 16%, transparent);
	}
	.line.removed {
		background: color-mix(in srgb, var(--danger) 10%, transparent);
		color: var(--muted);
	}
	/* Within a paired line, the words that differ carry a stronger wash. */
	.line mark {
		color: inherit;
		border-radius: 1px;
	}
	.line.added mark {
		background: color-mix(in srgb, var(--accent) 34%, transparent);
	}
	.line.removed mark {
		background: color-mix(in srgb, var(--danger) 24%, transparent);
		color: var(--ink);
	}
	.sign {
		text-align: center;
		user-select: none;
	}
	.line.added .sign {
		color: var(--accent);
	}
	.line.removed .sign {
		color: var(--danger);
	}
	.skipped {
		padding: var(--space-1) 0 var(--space-1) 1.25rem;
		font-family: var(--font-text);
		font-style: italic;
		border-block: 1px dashed var(--rule);
		margin: var(--space-1) 0;
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
