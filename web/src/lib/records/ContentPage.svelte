<!-- Every record file SkyrimNet reads (actions, triggers, filters, voice
     effects, and the rest), grouped by kind, with the selected one's file in
     the pane. A record can be switched, or edited as the player's own copy,
     which replaces the plugin's or SkyrimNet's in every playthrough. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { tick, untrack } from 'svelte';
	import RecordEditor from './RecordEditor.svelte';
	import { formKinds } from './forms.ts';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { ContentList, Record, Version } from '#lib/api/types/records.ts';
	import { plural, realTime } from '#lib/format.ts';
	import { pluginHref } from '#lib/plugins/links.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import CodeArea from '#lib/ui/CodeArea.svelte';
	import LimitsNotice from './LimitsNotice.svelte';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { getVersions, layerName, recordsDependency, removeRecordCopy, saveRecord, switchRecord } from './api.ts';

	let { list }: { list: ContentList } = $props();

	let save = $derived(page.params.save ?? '');
	let readOnly = $derived(live.gameRunning);
	const layerOf = (r: Record) => list.layers[r.providers[0]?.layer ?? -1];
	const kindLabel = (kind: string) => list.kinds.find((k) => k.root === kind)?.label ?? kind;

	// Narrowed to one kind with ?kind=, and by a search.
	let query = $state('');
	let kind = $derived(param('kind') ?? '');
	let kinds = $derived(list.kinds.map((k) => ({ ...k, n: list.records.filter((r) => r.kind === k.root).length })).filter((k) => k.n));
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.records.filter((r) => (!kind || r.kind === kind) && (!q || [r.name, r.path, layerName(layerOf(r))].some((v) => v.toLowerCase().includes(q))));
	});
	let off = $derived(list.records.filter((r) => !r.enabled).length);
	let yours = $derived(list.records.filter((r) => layerOf(r)?.kind === 'overlay').length);

	const selection = new Selection(() => shown, (r) => r.path, { param: 'item' });
	let current = $derived(selection.current);

	// The selected record's copies, highest priority first, read when it's
	// selected and again when the list says its file changed.
	let versions = $state<Version[] | null>(null);
	let loadError = $state('');
	let loadedPath = '';
	$effect(() => {
		const path = current?.path;
		void current?.providers[0]?.modified;
		if (path !== loadedPath) {
			versions = null;
			loadedPath = path ?? '';
			untrack(() => {
				error = '';
				notice = '';
			});
		}
		loadError = '';
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
	let using = $derived(versions?.[0]);
	let mine = $derived(versions?.find((v) => v.kind === 'overlay'));
	let below = $derived(versions?.filter((v) => v.kind !== 'overlay') ?? []);

	let saving = $state(false);
	let error = $state('');
	let notice = $state('');

	// Editing is a state of the address, so Back cancels it: ?mode=edit opens
	// a kind's form, or its YAML when it has none; ?mode=yaml its YAML; and
	// ?mode=new a new record's form.
	let mode = $derived.by<'read' | 'yaml' | 'form' | 'new'>(() => {
		const m = param('mode');
		if (m === 'new') return 'new';
		if (!current) return 'read';
		if (m === 'edit' && formKinds[current.kind]) return 'form';
		if ((m === 'edit' || m === 'yaml') && using?.text !== undefined) return 'yaml';
		return 'read';
	});
	$effect(() => {
		if (mode === 'read' && param('mode') && versions) setParams({ mode: null }, 'replace');
	});
	let draft = $state('');
	$effect(() => {
		if (mode !== 'yaml') return;
		untrack(() => {
			draft = using?.text ?? '';
			error = '';
			notice = '';
		});
	});

	// act runs a change, then shows the record's copies as they are after it.
	async function act(change: () => Promise<Version[]>, done: string) {
		if (saving || readOnly) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const v = await change();
			versions = v;
			notice = done;
			await invalidate(recordsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function toggle() {
		const r = current;
		if (!r?.switch) return;
		const made = mine ? '' : ' Hermaeus made your own copy of it, with only its switch changed.';
		act(() => switchRecord(r.path, !r.enabled), r.enabled ? `Switched off.${made} SkyrimNet ignores it from the next time it starts.` : `Switched on.${made} SkyrimNet uses it from the next time it starts.`);
	}

	async function submit() {
		const r = current;
		if (!r) return;
		const had = !!mine;
		await act(() => saveRecord(r.path, { modified: mine?.modified ?? 0, text: draft }), had ? 'Saved your copy. SkyrimNet reads it from the next time it starts.' : 'Saved as your own copy. SkyrimNet reads it in place of the one below from the next time it starts.');
		if (!error) await leave('mode');
	}

	function remove() {
		const r = current;
		if (!r || !mine) return;
		const next = below[0] ? `the copy from ${below[0].kind === 'base' ? 'SkyrimNet’s defaults' : below[0].source} applies again` : 'nothing replaces it, since only you have it';
		if (!confirm(`Remove your copy of ${r.name}? A backup is kept, and ${next}.`)) return;
		act(() => removeRecordCopy(r.path, { modified: mine!.modified }), below[0] ? 'Removed your copy. The one below applies again.' : 'Removed your copy.');
	}

	function startEdit() {
		if (current && using?.text !== undefined && !readOnly) setParams({ item: current.path, mode: 'edit' });
	}

	// A form saved: show the record, and say so.
	async function saved(path: string, done: string) {
		await setParams({ item: path, mode: null });
		await invalidate(recordsDependency);
		await tick();
		loadedPath = path;
		notice = done;
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !reading, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !reading, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: mine ? 'Edit' : 'Edit as your copy', hidden: !reading || !current, disabled: readOnly || using?.text === undefined, run: startEdit },
			{ key: 'o', shown: 'O', label: current?.enabled === false ? 'Switch on' : 'Switch off', hidden: !reading || !current?.switch, disabled: readOnly || saving || !versions, run: toggle },
			...deleteShortcuts({ label: 'Remove your copy', hidden: !reading || !mine, disabled: readOnly || saving, run: remove }),
			{ key: 'n', shown: 'N', label: 'New record', hidden: !reading, disabled: readOnly, run: () => setParams({ mode: 'new' }) },
			{ key: 'Enter', label: 'Save', command: true, hidden: mode !== 'yaml', disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: 'Cancel', hidden: mode !== 'yaml', run: () => leave('mode') }
		];
	});
</script>

{#if mode === 'form' && current}
	{#key current.path}
		<RecordEditor {list} path={current.path} {readOnly} onSaved={saved} onCancel={() => leave('mode')} onYaml={() => setParams({ mode: 'yaml' }, 'replace')} />
	{/key}
{:else if mode === 'new'}
	<RecordEditor {list} {readOnly} onSaved={saved} onCancel={() => leave('mode')} onYaml={() => {}} />
{:else if mode === 'yaml' && current}
	<header class="editing">
		<h1 class="caps">{current.name}</h1>
		<p class="path">{current.path}</p>
		<p class="muted">
			{#if mine}
				Editing <b class="changed">your own copy</b>, which SkyrimNet uses in every playthrough{below[0] ? ` in place of the one from ${layerName(list.layers[below[0].layer])}` : ''}.
			{:else}
				Starting from the copy from {layerName(layerOf(current))}. Saving makes your own copy, which replaces it in every playthrough.
			{/if}
			{#if formKinds[current.kind]}<button type="button" class="textbutton" onclick={() => setParams({ mode: 'edit' }, 'replace')}>Use the form</button>{/if}
		</p>
		{#if current.kind === 'voice_effects'}
			<LimitsNotice title="Limits Hermaeus can’t check">SkyrimNet limits what each stage’s settings and a voice sample can be, in ways that aren’t published, so Hermaeus can’t check them here. An effect outside them may not work in the game.</LimitsNotice>
		{/if}
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	</header>
	<div class="raw" use:fitToViewport>
		<CodeArea bind:value={draft} label="{current.name} as YAML" />
		<div class="buttons">
			<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
			<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
		</div>
	</div>
{:else}
	<div class="bar">
		<h1 class="caps">Content</h1>
		<span class="muted n" use:clipped>{plural(list.records.length, 'record', 'records')}{yours ? ` · ${yours} yours` : ''}{off ? ` · ${off} off` : ''}</span>
		<button type="button" class="button" onclick={() => setParams({ mode: 'new' })} disabled={readOnly}>New record</button>
		<select aria-label="Kind" value={kind} onchange={(e) => setParams({ kind: e.currentTarget.value || null })}>
			<option value="">All kinds</option>
			{#each kinds as k (k.root)}<option value={k.root}>{k.label} · {k.n}</option>{/each}
		</select>
		<input type="search" bind:value={query} placeholder="Search records" aria-label="Search records" />
	</div>
	<p class="muted lead">The record files SkyrimNet reads, from your plugins and your own edits. Where two layers have the same file, the higher one’s copy is used: yours, then the plugins in the Plugins page’s order, then SkyrimNet’s defaults.</p>
	{#if readOnly}<p class="notice warn">Skyrim is running, so records can only be read.</p>{/if}

	<div class="split">
		<div class="list">
			{#if !list.records.length}
				<p class="muted">No records in SkyrimNet’s defaults, your plugins, or your own edits.</p>
			{:else if !shown.length}
				<p class="muted">Nothing matches “{query}”.</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th class="mark" aria-label="Selected"></th>
							<th class="caps">Record</th>
							<th class="caps source">From</th>
						</tr>
					</thead>
					<tbody>
						{#each shown as r, i (r.path)}
							{#if i === 0 || shown[i - 1]!.kind !== r.kind}
								<tr class="group"><th colspan="3" class="caps" scope="rowgroup">{kindLabel(r.kind)}</th></tr>
							{/if}
							<tr aria-selected={selection.is(r)} class:off={!r.enabled} onclick={() => selection.select(r)}>
								<td class="mark">{#if selection.is(r)}<Diamond />{/if}</td>
								<td class="name" use:clipped>{r.name}{#if !r.enabled}<span class="flag">off</span>{/if}{#if r.error}<span class="flag bad">can’t be read</span>{/if}</td>
								<td class="source muted" use:clipped>
									{#if layerOf(r)?.kind === 'overlay'}<span class="changed">Your edits</span>{:else}{layerName(layerOf(r))}{/if}{#if r.providers.length > 1}<span class="flag">replaces {layerName(list.layers[r.providers[1]!.layer])}’s</span>{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>

		<aside class="pane" use:fitToViewport aria-label="Record">
			{#if current}
				{@const r = current}
				{@const from = layerOf(r)}
				{#key r.path}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{r.name}</h2>
						<p class="muted small">{kindLabel(r.kind)} · <span class="id">{r.path}</span></p>
						<p class="small">
							{#if from?.kind === 'pack'}From <a href={pluginHref(save, from.id)}>{from.source}</a>{:else if from?.kind === 'overlay'}<span class="changed">Your own copy</span>{:else}From SkyrimNet’s defaults{/if}{#if r.providers.length > 1}, in place of {r.providers
									.slice(1)
									.map((p) => layerName(list.layers[p.layer]))
									.join(', then ')}{/if}.
						</p>
						{#if r.error}<p class="notice error">SkyrimNet can’t read this file: {r.error}</p>{/if}
						{#if !r.enabled}<p class="notice">Off: SkyrimNet doesn’t use it.</p>{/if}
						{#if error}<p class="notice error" role="alert">{error}</p>{/if}
						{#if notice}<p class="notice" role="status">{notice}</p>{/if}
						<div class="buttons">
							{#if r.switch}<button type="button" class="button" onclick={toggle} disabled={readOnly || saving || !versions}>{r.enabled ? 'Switch off' : 'Switch on'}</button>{/if}
							<button type="button" class="button" onclick={startEdit} disabled={readOnly || using?.text === undefined}>{mine ? 'Edit' : 'Edit as your copy'}</button>
							{#if mine}<button type="button" class="button" onclick={remove} disabled={readOnly || saving}>Remove your copy</button>{/if}
						</div>
						{#if loadError}
							<p class="notice error">{loadError}</p>
						{:else if using}
							<p class="muted small when">{layerName(list.layers[using.layer])} · {realTime(using.modified)}{using.crlf ? ' · Windows line endings' : ''}</p>
							{#if using.text !== undefined}
								<pre class="file">{using.text}</pre>
							{:else}
								<p class="muted small">Not text, so it can’t be shown.</p>
							{/if}
						{/if}
					</div>
				{/key}
			{/if}
		</aside>
	</div>
{/if}

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
	.bar .button {
		margin-left: auto;
		flex: none;
		white-space: nowrap;
	}
	.bar select {
		flex: none;
	}
	.bar input {
		flex: 0 1 14rem;
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
	.id,
	.path {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
	.changed {
		font-weight: 400;
		color: var(--changed);
	}
	/* The pane is wider than on Actions or Triggers: it shows a file. */
	.bar,
	.split {
		--pane: clamp(20rem, 34vw, 34rem);
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
	th.source {
		width: 45%;
	}
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
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
	tr.off td.name {
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
	.flag.bad {
		color: var(--danger);
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
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin: var(--space-4) 0;
	}
	.when {
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.file {
		margin: 0 0 var(--space-5);
		font: 400 var(--step--1) / 1.55 var(--font-mono);
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		tab-size: 4;
	}
	/* Editing fills the window down to the hint strip, with Save in view
	   beneath it; the text scrolls inside, not the page. */
	.editing {
		margin-bottom: var(--space-3);
	}
	.editing h1 {
		top: 0;
		margin-bottom: var(--space-2);
	}
	.editing .path {
		color: var(--muted);
		font-size: var(--step--1);
		margin-bottom: var(--space-2);
	}
	.editing .muted {
		max-width: 64rem;
	}
	.raw {
		position: sticky;
		top: 0;
		display: flex;
		flex-direction: column;
		max-width: 64rem; /* fitToViewport sets its height */
	}
	.raw :global(.code) {
		flex: 1;
	}
	.raw .buttons {
		margin: var(--space-3) 0 0;
		padding-bottom: var(--space-4);
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
