<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Earlier, Version } from '#lib/api/types/bios.ts';
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

	// The bio reads as one document. One section is selected (for the keys);
	// editing turns that section into a field where it is. Raw and history
	// replace the document.
	let mode = $state<'read' | 'edit' | 'raw' | 'history'>('read');
	let draft = $state('');
	let saving = $state(false);
	let notice = $state('');
	let error = $state('');

	const selection = new Selection(() => version?.sections ?? [], (s) => s.name, '[data-selected="true"]');
	let section = $derived(selection.current);

	// Which section is in view, for the index, from where the page has scrolled to.
	let inView = $state<string | null>(null);
	let doc = $state<HTMLElement>();
	$effect(() => {
		if (!doc || mode === 'raw' || mode === 'history') return;
		const seen = new Map<string, number>();
		const watch = new IntersectionObserver(
			(entries) => {
				for (const e of entries) seen.set((e.target as HTMLElement).dataset.section!, e.intersectionRatio);
				let best: string | null = null;
				let most = 0;
				for (const [name, ratio] of seen) if (ratio > most) ((most = ratio), (best = name));
				inView = best;
			},
			{ rootMargin: '-10% 0px -60% 0px', threshold: [0, 0.25, 0.5, 1] }
		);
		for (const el of doc.querySelectorAll<HTMLElement>('[data-section]')) watch.observe(el);
		return () => watch.disconnect();
	});

	let history = $state<Earlier[]>([]);
	const chosen = new Selection(() => history, (e) => e.id);
	let earlier = $derived(chosen.current);

	async function run(action: () => Promise<string>) {
		error = '';
		notice = '';
		saving = true;
		try {
			notice = await action();
			mode = 'read';
			await invalidate(bioDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function startEdit(name?: string) {
		if (!editable) return;
		if (name) selection.key = name;
		if (!selection.current) return;
		draft = selection.current.text;
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

	function jump(name: string) {
		selection.key = name;
		doc?.querySelector(`[data-section="${name}"]`)?.scrollIntoView({ block: 'start', behavior: 'smooth' });
	}

	let kindOptions = $derived(versions.map((v) => ({ value: v.kind, label: kinds[v.kind]?.label ?? v.kind, title: kinds[v.kind]?.what })));

	useShortcuts(() => {
		const reading = mode === 'read';
		const editing = mode === 'edit' || mode === 'raw';
		const listing = reading || mode === 'history';
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !listing, run: () => (mode === 'history' ? chosen.move(1) : selection.move(1)) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !listing, run: () => (mode === 'history' ? chosen.move(-1) : selection.move(-1)) },
			{ key: 'e', shown: 'E', label: 'Edit section', hidden: !reading, disabled: !editable || !section, run: () => startEdit() },
			{ key: 'r', shown: 'R', label: 'Raw file', hidden: !reading, disabled: !editable, run: startRaw },
			{ key: 'h', shown: 'H', label: 'History', hidden: !reading, disabled: !version || version.kind === 'default', run: startHistory },
			{ key: 'c', shown: 'C', label: 'Create save bio', hidden: !reading || version?.kind !== 'default' || hasBaseline, disabled: readOnly, run: create },
			{ key: 'Enter', label: 'Save', command: true, hidden: !editing, disabled: saving, run: submit },
			{ key: 'Enter', label: 'Restore', hidden: mode !== 'history', disabled: readOnly || !earlier || earlier.current, run: restore },
			...(reading ? [] : [{ key: 'Escape', shown: 'Esc', label: mode === 'history' ? 'Back' : 'Cancel', run: () => (mode = 'read') }])
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
	{#if version?.kind === 'default' && !hasBaseline && !readOnly && mode === 'read'}
		<button type="button" class="button primary" onclick={create} disabled={saving}>Create a save bio from this</button>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !version}
	<p class="muted">{name} has no bio: nothing in this playthrough, and nothing in SkyrimNet’s content for the template <span class="id">{template}</span>.</p>
{:else if mode === 'raw'}
	<div class="raw">
		<p class="muted small">The whole file, with SkyrimNet’s own markup. Broken markup is refused; a removed section only warns.</p>
		<textarea bind:value={draft} spellcheck="false" use:fitToViewport></textarea>
		<div class="buttons">
			<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
			<button type="button" class="button" onclick={() => (mode = 'read')} disabled={saving}>Cancel</button>
		</div>
	</div>
{:else if mode === 'history'}
	<div class="split">
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
		<aside class="pane" use:fitToViewport aria-label="Earlier version">
			{#if earlier}
				{#key earlier.id}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{realTime(earlier.replaced)}</h2>
						<p class="muted small">Kept by {earlier.source === 'skyrimnet' ? 'SkyrimNet when it replaced the bio' : 'Hermaeus before a save'}.</p>
						<pre class="file">{earlier.raw}</pre>
						<div class="buttons">
							{#if !earlier.current && !readOnly}
								<button type="button" class="button primary" onclick={restore} disabled={saving}>Restore this version</button>
							{/if}
							<button type="button" class="button" onclick={() => (mode = 'read')}>Back</button>
						</div>
					</div>
				{/key}
			{:else}
				<p class="muted">Choose an earlier version to read it.</p>
			{/if}
		</aside>
	</div>
{:else}
	<div class="reading">
		<article class="document" bind:this={doc}>
			{#each version.sections as s (s.name)}
				{@const selected = selection.is(s)}
				{@const editing = mode === 'edit' && selected}
				<!-- The label's button is the accessible control; clicking anywhere in
				     the section is a larger target for the same thing, and a double
				     click opens the editor. -->
				<!-- svelte-ignore a11y_no_static_element_interactions, a11y_click_events_have_key_events -->
				<section
					data-section={s.name}
					data-selected={selected}
					class:editing
					onclick={() => mode === 'read' && selection.select(s)}
					ondblclick={() => mode === 'read' && startEdit(s.name)}
				>
					<div class="label">
						<h2 class="caps">
							<button type="button" onclick={() => mode === 'read' && selection.select(s)} disabled={mode !== 'read'}>
								<span class="mark">{#if selected}<Diamond />{/if}</span>{sectionLabel(s.name)}
							</button>
						</h2>
						{#if selected && editable && mode === 'read'}
							<button type="button" class="textbutton small" onclick={() => startEdit(s.name)}>Edit</button>
						{/if}
					</div>
					{#if editing}
						<div class="field">
							<textarea bind:value={draft} rows={Math.max(6, Math.min(28, Math.ceil(draft.length / 80) + draft.split('\n').length + 1))}></textarea>
							<div class="buttons">
								<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
								<button type="button" class="button" onclick={() => (mode = 'read')} disabled={saving}>Cancel</button>
							</div>
						</div>
					{:else}
						<p class="text">{s.text}</p>
					{/if}
				</section>
			{:else}
				<p class="muted">This bio has no sections.</p>
			{/each}
			{#if readOnly}<p class="notice warn">Skyrim is running, so bios can’t be changed until you close it.</p>{/if}
		</article>
		<nav class="index" aria-label="Sections">
			<ul>
				{#each version.sections as s (s.name)}
					<li>
						<button type="button" class="caps" class:here={inView === s.name} aria-current={selection.is(s) ? 'true' : undefined} onclick={() => jump(s.name)}>
							{sectionLabel(s.name)}
						</button>
					</li>
				{/each}
			</ul>
		</nav>
	</div>
{/if}

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.id {
		font-variant-numeric: tabular-nums;
	}
	.small {
		font-size: var(--step--1);
	}

	/* Reading: the document, with the index alongside. */
	.reading {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 11rem;
		gap: var(--space-7);
		align-items: start;
	}
	.document {
		max-width: 52rem;
	}
	section {
		display: grid;
		grid-template-columns: 10rem minmax(0, 1fr);
		gap: var(--space-2) var(--space-4);
		padding: var(--space-3) 0 var(--space-3) var(--space-2);
		border-top: 1px solid var(--rule);
		scroll-margin-top: var(--space-4);
	}
	section:first-child {
		border-top: 0;
	}
	section[data-selected='true'] {
		background: var(--select);
	}
	section:not(.editing):hover {
		background: var(--glass);
	}
	section[data-selected='true']:not(.editing):hover {
		background: var(--select);
	}
	.label {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-1);
	}
	section h2 {
		font-size: var(--step--1);
		padding-top: 0.35em;
	}
	section h2 button {
		display: flex;
		gap: var(--space-2);
		align-items: baseline;
		background: none;
		border: 0;
		padding: 0;
		color: var(--muted);
		font: inherit;
		letter-spacing: inherit;
		text-transform: inherit;
		text-align: left;
		cursor: pointer;
	}
	section h2 button:hover:not(:disabled) {
		color: var(--ink);
	}
	section h2 .mark {
		width: 0.75rem;
		flex: none;
	}
	.label .textbutton {
		margin-left: calc(0.75rem + var(--space-2));
	}
	.text {
		white-space: pre-line;
		line-height: 1.6;
		max-width: 65ch;
		margin-right: var(--space-4);
	}
	.field {
		margin-right: var(--space-4);
	}
	.field textarea,
	.raw textarea {
		width: 100%;
		resize: vertical;
		line-height: 1.55;
	}
	.index {
		position: sticky;
		top: var(--space-4);
	}
	.index ul {
		list-style: none;
		margin: 0;
		padding: 0;
		border-left: 1px solid var(--rule-strong);
	}
	.index button {
		display: block;
		width: 100%;
		background: none;
		border: 0;
		padding: var(--space-1) var(--space-3);
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		cursor: pointer;
		margin-left: -1px;
		border-left: 1px solid transparent;
	}
	.index button:hover {
		color: var(--ink);
	}
	.index button.here {
		color: var(--ink);
		border-left-color: var(--accent);
	}
	.index button[aria-current='true'] {
		color: var(--accent);
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}

	/* The whole file. */
	.raw {
		max-width: 64rem;
	}
	.raw textarea {
		display: block;
		margin-top: var(--space-3);
	}

	/* Earlier versions: the item table and the pane. */
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
		white-space: nowrap;
	}
	tbody tr {
		cursor: default;
	}
	/* A row you can select tints under the pointer, as a bio's sections do. */
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
	.pane {
		position: sticky;
		top: 0;
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
	.file {
		margin: var(--space-3) 0;
		font: inherit;
		white-space: pre-wrap;
		line-height: 1.55;
	}
	.pane .buttons {
		position: sticky;
		bottom: 0;
		padding: var(--space-3) 0 var(--space-5);
		margin-top: var(--space-3);
		background: var(--panel);
		border-top: 1px solid var(--rule);
	}

	@media (max-width: 70rem) {
		.reading {
			grid-template-columns: 1fr;
		}
		.index {
			display: none;
		}
		.split {
			grid-template-columns: minmax(0, 1fr) minmax(18rem, 24rem);
		}
	}
	@media (max-width: 60rem) {
		section {
			grid-template-columns: 1fr;
			gap: var(--space-1);
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
