<!-- SkyrimNet's settings files, with how many settings each has and how
     many the player changed. -->
<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import type { FileValues } from '#lib/api/types/settings.ts';
	import { count } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { changedPage, fileLabel, setFilePinned, settingsDependency, settingsHref } from './api.ts';

	let { save, files }: { save: string; files: FileValues[] } = $props(); // pinned files first, in the order they were pinned, then by name

	let query = $state('');
	let sort = $state<Sort>({ key: 'group', descending: false });
	let error = $state('');

	let changed = $derived(files.reduce((n, f) => n + f.file.changed, 0));
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = files.filter((f) => !q || fileLabel(f.file).toLowerCase().includes(q) || f.file.name.toLowerCase().includes(q));
		return sorted(list, sort, {
			group: (f) => (f.file.pinned ? 0 : 1), // ties keep their order: pins as pinned, the rest by name
			name: (f) => fileLabel(f.file),
			settings: (f) => f.file.settings,
			changed: (f) => f.file.changed
		});
	});
	// In their natural order, files sit under Pinned or Everything else.
	let grouped = $derived(sort.key === 'group' && files.some((f) => f.file.pinned));
	const groupName = (f: FileValues) => (f.file.pinned ? 'Pinned' : 'Everything else');

	async function togglePin(f: FileValues) {
		error = '';
		try {
			await setFilePinned(f.file.name, !f.file.pinned);
			await invalidate(settingsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	const selection = new Selection(() => shown, (f) => f.file.name);
	let current = $derived(selection.current);
	const href = (f: FileValues) => settingsHref(save, f.file.name);

	useShortcuts(() => {
		const f = selection.current;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: 'Open', disabled: !f, run: () => f && goto(href(f)) },
			{ key: 'p', shown: 'P', label: f?.file.pinned ? 'Unpin' : 'Pin', disabled: !f, run: () => f && togglePin(f) },
			{ key: 'c', shown: 'C', label: 'Changed settings', disabled: !changed, run: () => goto(settingsHref(save, changedPage)) }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Settings</h1>
	{#if changed}
		<a class="changed n" href={settingsHref(save, changedPage)}>{count(changed)} changed from the defaults</a>
	{:else}
		<span class="muted n">Nothing differs from the defaults</span>
	{/if}
	<input type="search" bind:value={query} placeholder="Filter by file" aria-label="Filter settings files" />
</div>
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !files.length}
	<p class="muted">No settings files were found in SkyrimNet’s config folder.</p>
{:else if !shown.length}
	<p class="muted">Nothing matches “{query}”.</p>
{:else}
	<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<SortHeader label="File" key="group" bind:sort />
					<SortHeader label="Settings" key="settings" bind:sort descendingFirst align="end" />
					<SortHeader label="Changed" key="changed" bind:sort descendingFirst align="end" />
				</tr>
			</thead>
			<tbody>
				{#each shown as f, i (f.file.name)}
					{#if grouped && (i === 0 || f.file.pinned !== shown[i - 1]!.file.pinned)}
						<tr class="group"><th colspan="4" class="caps" scope="rowgroup">{groupName(f)}</th></tr>
					{/if}
					<tr aria-selected={selection.is(f)} onclick={() => selection.select(f)} ondblclick={() => goto(href(f))}>
						<td class="mark">{#if selection.is(f)}<Diamond />{/if}</td>
						<td class="name">
							<a href={href(f)}>{fileLabel(f.file)}</a>
							{#if !f.file.exists}<span class="muted"> · not created yet</span>{/if}
						</td>
						<td class="n">{count(f.file.settings)}</td>
						<td class="n">{#if f.file.changed}<span class="changed">{count(f.file.changed)}</span>{/if}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label="Settings file">
			{#if current}
				<h2 class="caps">{fileLabel(current.file)}</h2>
				<p class="path">{current.file.path}</p>
				<dl>
					<dt class="caps">Settings</dt><dd>{count(current.file.settings)}</dd>
					<dt class="caps">Changed</dt><dd>{current.file.changed ? count(current.file.changed) : 'None'}</dd>
					<dt class="caps">Defaults</dt><dd>{current.file.hasDefaults ? 'Recorded by SkyrimNet' : 'None recorded'}</dd>
				</dl>
				{#if !current.file.exists}
					<p class="notice">SkyrimNet hasn’t created this file yet; it will the next time it starts. Until then its defaults can be read but not changed.</p>
				{/if}
				<div class="actions">
					<a class="button primary" href={href(current)}>Open</a>
					<button type="button" class="button" aria-pressed={current.file.pinned} onclick={() => togglePin(current!)}>{current.file.pinned ? 'Unpin' : 'Pin'}</button>
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
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	.bar .n {
		white-space: nowrap;
	}
	.bar input {
		margin-left: auto;
		width: 18rem;
	}
	.changed {
		color: var(--changed);
	}
	a.changed {
		text-decoration-color: var(--changed);
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
	}
	.pane .notice {
		margin-bottom: var(--space-4);
	}
	.actions {
		display: flex;
		gap: var(--space-2);
	}
	.actions a {
		text-decoration: none;
	}
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		text-align: left;
		font-size: var(--step--1);
		color: var(--accent);
		border-bottom: 1px solid var(--rule-strong);
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
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
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
	td.n {
		text-align: right;
		width: 7rem;
	}
</style>
