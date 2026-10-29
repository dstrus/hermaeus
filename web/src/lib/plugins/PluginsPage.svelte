<!-- Plugins: what SkyrimNet loads, in order. Lower in the list wins, as in
     SkyrimNet's dashboard and in mod managers: its defaults first, your own
     edits last. A plugin can be switched off or moved; each change writes
     content-registry.json with a backup. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { List, Plugin } from '#lib/api/types/plugins.ts';
	import { plural } from '#lib/format.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { contentsLine, pluginsDependency, priorityOf, setEnabled, setOrder } from './api.ts';

	let { list }: { list: List } = $props();

	let readOnly = $derived(live.gameRunning);
	let shown = $derived(list.plugins);
	let installed = $derived(shown.filter((p) => p.kind === 'plugin'));
	let off = $derived(installed.filter((p) => !p.enabled).length);

	const selection = new Selection(() => shown, (p) => p.id, { param: 'item' });
	let current = $derived(selection.current);
	// Where an installed plugin sits among the installed ones, lowest first.
	let at = $derived(current?.kind === 'plugin' ? installed.indexOf(current) : -1);

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	async function run(action: () => Promise<unknown>, done: string) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await action();
			notice = done;
			await invalidate(pluginsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	// move takes the selected plugin a place up (earlier, so it loses more)
	// or down (later, so it wins more).
	function move(by: number) {
		const p = current;
		if (!p || p.kind !== 'plugin' || readOnly) return;
		const to = at + by;
		if (to < 0 || to >= installed.length) return;
		const order = [...installed];
		[order[at], order[to]] = [order[to]!, order[at]!];
		const other = installed[to]!;
		run(() => setOrder({ modified: list.modified, priority: priorityOf(order) }), by > 0 ? `${p.title} now wins over ${other.title}.` : `${other.title} now wins over ${p.title}.`);
	}

	function toggle() {
		const p = current;
		if (!p || p.kind !== 'plugin' || readOnly) return;
		run(() => setEnabled(p.id, { modified: list.modified, enabled: !p.enabled }), p.enabled ? `Switched off ${p.title}. SkyrimNet ignores it from the next time it starts.` : `Switched on ${p.title}. SkyrimNet loads it from the next time it starts.`);
	}

	const kindLabel = (p: Plugin) => (p.kind === 'base' ? 'SkyrimNet' : p.kind === 'overlay' ? 'Yours' : p.source === 'external' ? 'From a mod' : p.source === 'hub' ? 'From the Hub' : 'Local');

	useShortcuts(() => {
		const p = current;
		const movable = !!p && p.kind === 'plugin' && !readOnly && !saving;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
			{ key: 'o', shown: 'O', label: p?.enabled === false ? 'Switch on' : 'Switch off', disabled: !movable, run: toggle },
			{ key: 'u', shown: 'U', label: 'Move up', disabled: !movable || at <= 0, run: () => move(-1) },
			{ key: 'd', shown: 'D', label: 'Move down', disabled: !movable || at >= installed.length - 1, run: () => move(1) }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Plugins</h1>
	<span class="muted n" use:clipped>{plural(installed.length, 'plugin', 'plugins')} installed{off ? ` · ${off} off` : ''}</span>
</div>
<p class="muted lead">What SkyrimNet loads, in order. Lower in the list wins: a plugin’s prompt replaces the same prompt in any plugin above it, and your own edits, at the bottom, win over them all.</p>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<th class="caps at">Order</th>
					<th class="caps">Plugin</th>
					<th class="caps version">Version</th>
					<th class="caps holds">Holds</th>
				</tr>
			</thead>
			<tbody>
				{#each shown as p, i (p.id)}
					<tr aria-selected={selection.is(p)} class:off={!p.enabled} class:fixed={p.kind !== 'plugin'} onclick={() => selection.select(p)}>
						<td class="mark">{#if selection.is(p)}<Diamond />{/if}</td>
						<td class="at n">{p.kind === 'overlay' ? '∞' : i}</td>
						<td class="name" use:clipped>
							{p.title}
							{#if !p.enabled}<span class="flag">off</span>{/if}
							{#if p.kind === 'plugin' && p.replaced.length}<span class="flag">some replaced</span>{/if}
						</td>
						<td class="version muted">{p.version}</td>
						<td class="holds muted" use:clipped>{p.found ? contentsLine(p) || 'Nothing' : 'In another mod’s folder'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">Installing, updating, and removing plugins happens in SkyrimNet’s dashboard, through its Plugin Hub.</p>
	</div>

	<aside class="pane" use:fitToViewport aria-label="Plugin">
		{#if current}
			{@const p = current}
			{#key p.id}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{p.title}</h2>
					{#if p.tagline}<p class="tagline">{p.tagline}</p>{/if}
					<p class="muted small">{kindLabel(p)}{p.author ? ` · by ${p.author}` : ''}{p.version ? ` · ${p.version}` : ''}{p.kind === 'plugin' ? ` · ` : ''}{#if p.kind === 'plugin'}<span class="id">{p.id}</span>{/if}</p>
					{#if p.kind === 'plugin'}
						<div class="actions">
							<button type="button" class="button" onclick={toggle} disabled={readOnly || saving}>{p.enabled ? 'Switch off' : 'Switch on'}</button>
							<button type="button" class="button" onclick={() => move(-1)} disabled={readOnly || saving || at <= 0}>Move up</button>
							<button type="button" class="button" onclick={() => move(1)} disabled={readOnly || saving || at >= installed.length - 1}>Move down</button>
						</div>
					{/if}
					{#if !p.enabled}
						<p class="notice">Switched off: SkyrimNet doesn’t load it, so nothing in it applies.</p>
					{/if}
					{#if p.description}<p class="description small">{p.description}</p>{/if}
					<dl>
						{#if !p.found}
							<dt class="caps">Files</dt>
							<dd>In another mod’s folder, which Hermaeus can’t see under Mod Organizer 2. Its order and switch still work here.</dd>
						{:else}
							<dt class="caps">Holds</dt>
							<dd>{contentsLine(p) || 'Nothing'}</dd>
							<dt class="caps">Folder</dt>
							<dd class="id folder">{p.folder}</dd>
						{/if}
						{#if p.tags.length}<dt class="caps">Tags</dt><dd>{p.tags.join(', ')}</dd>{/if}
					</dl>
					{#if p.replaces.length}
						<details>
							<summary class="caps">Replaces {plural(p.replaces.length, 'prompt', 'prompts')} above it</summary>
							<ul class="paths">{#each p.replaces as path (path)}<li class="id">{path}</li>{/each}</ul>
						</details>
					{/if}
					{#if p.replaced.length}
						<details>
							<summary class="caps">{plural(p.replaced.length, 'prompt', 'prompts')} replaced below it</summary>
							<p class="muted small">A plugin lower in the list, or your own edits, has these too, so theirs apply.</p>
							<ul class="paths">{#each p.replaced as path (path)}<li class="id">{path}</li>{/each}</ul>
						</details>
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">Skyrim is running, so nothing can be changed until you close it.</p>{/if}
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
	.notice {
		margin-bottom: var(--space-4);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	.list > .small {
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
	tr.off td {
		color: var(--muted);
	}
	/* SkyrimNet's defaults and your edits sit apart from the plugins between them. */
	tr.fixed + tr:not(.fixed) td,
	tr:not(.fixed) + tr.fixed td {
		border-top: 1px solid var(--rule-strong);
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	.at {
		width: 4.5rem;
	}
	td.at {
		text-align: right;
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}
	.version {
		width: 5.5rem;
	}
	.holds {
		width: 30%;
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
	}
	.tagline {
		margin-bottom: var(--space-1);
	}
	.description {
		white-space: pre-wrap;
		margin: var(--space-3) 0;
		line-height: 1.5;
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin: var(--space-4) 0;
	}
	.pane dl {
		display: grid;
		grid-template-columns: 4.5rem 1fr;
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
		min-width: 0;
	}
	.folder {
		overflow-wrap: anywhere;
	}
	details {
		margin: var(--space-3) 0;
	}
	summary {
		font-size: var(--step--1);
		color: var(--accent);
		cursor: pointer;
	}
	.paths {
		list-style: none;
		padding: 0;
		margin: var(--space-2) 0 0;
		font-size: var(--step--1);
		overflow-wrap: anywhere;
	}
	@media (max-width: 60rem) {
		.bar {
			margin-right: 0;
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
