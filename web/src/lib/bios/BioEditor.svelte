<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Earlier, Section, Version } from '#lib/api/types/bios.ts';
	import { realTime } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { bioDependency, createBio, getHistory, kinds, restoreBio, saveBio, sectionLabel } from './api.ts';

	let {
		save,
		template,
		name,
		versions
	}: {
		save: string;
		template: string;
		name: string; // the character's, as shown
		versions: Version[]; // the dynamic bio, the baseline, the default: whichever exist
	} = $props();

	let readOnly = $derived(live.gameRunning);
	// Starts on the bio NPCs use; follows the versions when they reload (a
	// save, a created bio).
	let kind = $state(untrack(() => versions[0]?.kind ?? 'dynamic'));
	$effect(() => {
		if (!versions.some((v) => v.kind === kind)) kind = versions[0]?.kind ?? 'dynamic';
	});
	let version = $derived(versions.find((v) => v.kind === kind));
	let editable = $derived(!!version && version.kind !== 'default' && !readOnly);
	let hasBaseline = $derived(versions.some((v) => v.kind === 'baseline'));

	// What the pane is doing with the selected section, or the whole file.
	let mode = $state<'view' | 'edit' | 'raw' | 'history'>('view');
	let draft = $state('');
	let saving = $state(false);
	let notice = $state('');
	let error = $state('');

	const selection = new Selection(() => version?.sections ?? [], (s) => s.name);
	let section = $derived(selection.current);

	let history = $state<Earlier[]>([]);
	const chosen = new Selection(() => history, (e) => e.id);
	let earlier = $derived(chosen.current);

	async function run(action: () => Promise<string>) {
		error = '';
		notice = '';
		saving = true;
		try {
			notice = await action();
			mode = 'view';
			await invalidate(bioDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function startEdit() {
		if (!editable || !section) return;
		draft = section.text;
		mode = 'edit';
	}

	function startRaw() {
		if (!editable || !version) return;
		draft = version.raw;
		mode = 'raw';
	}

	async function startHistory() {
		if (!version || version.kind === 'default') return;
		error = '';
		try {
			history = (await getHistory(save, template, version.kind)).earlier;
			chosen.key = null;
			mode = 'history';
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	function submit() {
		if (!version) return;
		const v = version;
		const edit = mode === 'raw' ? { modified: v.modified, raw: draft } : { modified: v.modified, sections: { [section!.name]: draft } };
		run(async () => {
			const r = await saveBio(save, template, v.kind, edit);
			return [`${kinds[v.kind]?.label ?? v.kind} bio saved.`, r.warning].filter(Boolean).join(' ');
		});
	}

	function create() {
		run(async () => {
			await createBio(save, template);
			kind = 'baseline';
			return `${name} now has a baseline bio in this playthrough, copied from the default.`;
		});
	}

	function restore() {
		if (!version || !earlier) return;
		const v = version;
		const e = earlier;
		run(async () => {
			await restoreBio(save, template, v.kind, { id: e.id, modified: v.modified });
			return `Restored the version from ${realTime(e.replaced)}. The one it replaced is kept in the history.`;
		});
	}

	let kindOptions = $derived(versions.map((v) => ({ value: v.kind, label: kinds[v.kind]?.label ?? v.kind, title: kinds[v.kind]?.what })));

	useShortcuts(() => {
		const viewing = mode === 'view';
		const editing = mode === 'edit' || mode === 'raw';
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing && mode !== 'history', run: () => (mode === 'history' ? chosen.move(1) : selection.move(1)) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing && mode !== 'history', run: () => (mode === 'history' ? chosen.move(-1) : selection.move(-1)) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !editable || !section, run: startEdit },
			{ key: 'r', shown: 'R', label: 'Raw file', hidden: !viewing, disabled: !editable, run: startRaw },
			{ key: 'h', shown: 'H', label: 'History', hidden: !viewing, disabled: !version || version.kind === 'default', run: startHistory },
			{ key: 'c', shown: 'C', label: 'Create save bio', hidden: !viewing || version?.kind !== 'default' || hasBaseline, disabled: readOnly, run: create },
			{ key: 'Enter', label: 'Save', command: true, hidden: !editing, disabled: saving, run: submit },
			{ key: 'Enter', label: 'Restore', hidden: mode !== 'history', disabled: readOnly || !earlier || earlier.current, run: restore },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: mode === 'history' ? 'Back' : 'Cancel', run: () => (mode = 'view') }])
		];
	});
</script>

<div class="tools">
	{#if versions.length > 1}
		<Segmented bind:value={kind} options={kindOptions} label="Which bio" />
	{:else if version}
		<span class="caps muted">{kinds[version.kind]?.label ?? version.kind} bio</span>
	{/if}
	{#if version}
		<span class="muted">{kinds[version.kind]?.what}{#if version.source}{' '}From {version.source}{#if version.shared}, shared as {version.shared}{/if}.{/if}</span>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !version}
	<p class="muted">{name} has no bio: nothing in this playthrough, and nothing in SkyrimNet’s content for the template <span class="id">{template}</span>.</p>
{:else}
	<div class="split">
		<div class="list">
			{#if mode === 'history'}
				<table>
					<thead>
						<tr><th class="mark" aria-label="Selected"></th><th class="caps">Replaced</th><th class="caps">By</th><th class="caps">Playthrough</th></tr>
					</thead>
					<tbody>
						{#if !history.length}
							<tr><td colspan="4" class="muted">No earlier versions of this bio were found.</td></tr>
						{/if}
						{#each history as e (e.id)}
							<tr aria-selected={chosen.is(e)} onclick={() => chosen.select(e)}>
								<td class="mark">{#if chosen.is(e)}<Diamond />{/if}</td>
								<td>{realTime(e.replaced)}{#if e.current}<span class="muted"> · same as now</span>{/if}</td>
								<td>{e.source === 'skyrimnet' ? 'SkyrimNet' : 'Hermaeus'}</td>
								<td class="muted">{e.playthrough === 'this' ? 'this one' : e.playthrough === 'other' ? e.save : 'unclear'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{:else}
				<table>
					<thead>
						<tr><th class="mark" aria-label="Selected"></th><th class="caps">Section</th><th class="caps text">Text</th></tr>
					</thead>
					<tbody>
						{#each version.sections as s (s.name)}
							<tr aria-selected={selection.is(s)} onclick={() => mode === 'view' && selection.select(s)}>
								<td class="mark">{#if selection.is(s)}<Diamond />{/if}</td>
								<td class="name caps">{sectionLabel(s.name)}</td>
								<td class="text"><span class="line">{s.text}</span></td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>

		<aside class="pane" use:fitToViewport aria-label="Bio">
			{#if mode === 'raw'}
				<h2 class="caps">Whole file</h2>
				<p class="muted small">Every section, with SkyrimNet’s own markup. Broken markup is refused; a removed section only warns.</p>
				<textarea bind:value={draft} spellcheck="false" rows="24"></textarea>
				<div class="buttons">
					<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
					<button type="button" class="button" onclick={() => (mode = 'view')} disabled={saving}>Cancel</button>
				</div>
			{:else if mode === 'history'}
				{#if earlier}
					{#key earlier.id}
						<div in:fly={{ y: 6, duration: 180 }}>
							<h2 class="caps">{realTime(earlier.replaced)}</h2>
							<p class="muted small">Kept by {earlier.source === 'skyrimnet' ? 'SkyrimNet when it replaced the bio' : 'Hermaeus before a save'}.</p>
							<pre class="raw">{earlier.raw}</pre>
							{#if !earlier.current && !readOnly}
								<div class="buttons">
									<button type="button" class="button primary" onclick={restore} disabled={saving}>Restore this version</button>
									<button type="button" class="button" onclick={() => (mode = 'view')}>Back</button>
								</div>
							{/if}
						</div>
					{/key}
				{:else}
					<p class="muted">Choose an earlier version to read it.</p>
				{/if}
			{:else if section}
				{#key `${section.name}:${mode}`}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{sectionLabel(section.name)}</h2>
						{#if mode === 'edit'}
							<textarea bind:value={draft} rows="16"></textarea>
							<div class="buttons">
								<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
								<button type="button" class="button" onclick={() => (mode = 'view')} disabled={saving}>Cancel</button>
							</div>
						{:else}
							<p class="content">{section.text}</p>
							{#if editable}
								<div class="buttons">
									<button type="button" class="button" onclick={startEdit}>Edit</button>
									<button type="button" class="button" onclick={startRaw}>Raw file</button>
									<button type="button" class="button" onclick={startHistory}>History</button>
								</div>
							{:else if version.kind === 'default' && !hasBaseline && !readOnly}
								<div class="buttons">
									<button type="button" class="button primary" onclick={create} disabled={saving}>Create a save bio from this</button>
								</div>
								<p class="muted small">A save bio belongs to this playthrough alone, and you can edit it.</p>
							{/if}
						{/if}
					</div>
				{/key}
			{:else}
				<p class="muted">This bio has no sections.</p>
			{/if}
			{#if readOnly}<p class="notice warn">Skyrim is running, so bios can’t be changed until you close it.</p>{/if}
		</aside>
	</div>
{/if}

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-3);
	}
	.id {
		font-variant-numeric: tabular-nums;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(22rem, 34vw, 36rem);
		gap: var(--space-5);
		align-items: start;
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
		vertical-align: top;
		white-space: nowrap;
	}
	tbody tr {
		cursor: default;
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	td.name {
		font-size: var(--step--1);
		color: var(--muted);
		padding-top: 0.6rem;
	}
	.text {
		width: 100%;
		max-width: 0;
		white-space: normal;
	}
	.line {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.pane {
		position: sticky;
		top: 0;
		align-self: start;
		overflow: auto; /* fitToViewport sets its height */
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.content {
		margin: var(--space-4) 0;
		line-height: 1.6;
		white-space: pre-line;
		max-width: 65ch;
	}
	.raw {
		margin: var(--space-3) 0;
		font: inherit;
		white-space: pre-wrap;
		line-height: 1.55;
	}
	textarea {
		width: 100%;
		margin-top: var(--space-3);
		resize: vertical;
		line-height: 1.55;
	}
	.small {
		font-size: var(--step--1);
	}
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		padding: var(--space-3) 0 var(--space-5);
		background: var(--panel);
		border-top: 1px solid var(--rule);
		margin-top: var(--space-3);
	}
	.pane .notice {
		margin-bottom: var(--space-4);
	}
	@media (max-width: 70rem) {
		.split {
			grid-template-columns: minmax(0, 1fr) minmax(18rem, 24rem);
		}
	}
	@media (max-width: 60rem) {
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
