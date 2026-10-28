<!-- One prompt as a document: the version SkyrimNet uses, the other layers'
     versions behind a switch, and editing that saves the player's own copy. -->
<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Layer } from '#lib/api/types/layers.ts';
	import type { Document } from '#lib/api/types/prompts.ts';
	import { realTime } from '#lib/format.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import Up from '#lib/ui/Up.svelte';
	import { layerLabel, promptDependency, promptName, promptsDependency, removeCopy, savePrompt } from './api.ts';

	let { save, path, doc: loaded }: { save: string; path: string; doc: Document } = $props();

	// The document as last loaded or saved.
	let doc = $derived(loaded);
	let readOnly = $derived(live.gameRunning);
	const layerAt = (i: number): Layer => doc.layers[i] ?? { id: '', kind: '', source: 'an unknown layer' };
	// Which layer's version is shown; starts on the one in use.
	let shown = $state(untrack(() => String(loaded.versions[0]?.layer ?? 0)));
	let version = $derived(doc.versions.find((v) => String(v.layer) === shown) ?? doc.versions[0]);
	let layer = $derived(version && layerAt(version.layer));
	let mine = $derived(doc.versions.find((v) => layerAt(v.layer).kind === 'overlay'));
	let below = $derived(doc.versions.filter((v) => layerAt(v.layer).kind !== 'overlay'));
	let choices = $derived(doc.versions.map((v) => ({ value: String(v.layer), label: layerLabel(layerAt(v.layer)), title: layerAt(v.layer).source })));

	let mode = $state<'read' | 'edit'>('read');
	let draft = $state('');
	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let warnings = $state<string[]>([]);

	let list = $derived(`/${encodeURIComponent(save)}/prompts`);

	function startEdit() {
		if (!version || readOnly) return;
		draft = version.text;
		error = '';
		notice = '';
		warnings = [];
		mode = 'edit';
	}

	async function submit() {
		if (saving) return;
		saving = true;
		error = '';
		try {
			const saved = await savePrompt(path, { modified: mine?.modified ?? 0, text: draft });
			doc = saved.document;
			warnings = saved.warnings ?? [];
			shown = String(doc.versions[0]?.layer ?? 0);
			notice = mine ? 'Saved your copy.' : 'Saved as your own copy. SkyrimNet uses it from the next time it starts.';
			mode = 'read';
			await invalidate(promptsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	async function remove() {
		if (!mine || readOnly) return;
		const next = below[0] ? `the version from ${layerAt(below[0].layer).source}` : 'nothing: only you have this prompt';
		if (!confirm(`Remove your copy of ${promptName(path)}? A backup is kept, and ${next} applies again.`)) return;
		error = '';
		try {
			const left = await removeCopy(path, { modified: mine.modified });
			await invalidate(promptsDependency);
			if (!left.versions.length) {
				await goto(list);
				return;
			}
			doc = left;
			shown = String(doc.versions[0]?.layer ?? 0);
			notice = 'Removed your copy. The version below applies again.';
			await invalidate(promptDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		return [
			{ key: 'e', shown: 'E', label: mine && version === mine ? 'Edit' : 'Edit as your copy', hidden: !reading, disabled: readOnly || !version, run: startEdit },
			...deleteShortcuts({ label: 'Remove your copy', hidden: !reading || !mine, disabled: readOnly, run: remove }),
			{ key: 'Enter', label: 'Save', command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: reading ? 'Prompts' : 'Cancel', run: () => (reading ? goto(list) : (mode = 'read')) }
		];
	});
</script>

<header>
	<Up href={list} label="Prompts" />
	<h1 class="caps">{promptName(path)}</h1>
	<p class="path">{path}</p>
	<p class="muted">
		{#if mine}
			You have <b class="changed">your own copy</b>, and SkyrimNet uses it in every playthrough{below[0] ? `, in place of the version from ${layerAt(below[0].layer).source}` : ''}.
		{:else if layer}
			SkyrimNet uses the version from {layer.source}. Editing it makes your own copy, which replaces it in every playthrough; SkyrimNet’s dashboard edits the same copy.
		{/if}
	</p>
	{#if readOnly}<p class="notice warn">Skyrim is running, so prompts can only be read.</p>{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	{#if notice}<p class="notice" role="status">{notice}</p>{/if}
	{#each warnings as w (w)}<p class="notice warn">Check this: {w}.</p>{/each}
</header>

{#if mode === 'edit'}
	<div class="raw" use:fitToViewport>
		{#if version !== mine}
			<p class="muted small">Starting from the version from {layer?.source}. Saving makes your own copy.</p>
		{/if}
		<textarea bind:value={draft} spellcheck="false" aria-label="Prompt text"></textarea>
		<div class="buttons">
			<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
			<button type="button" class="button" onclick={() => (mode = 'read')} disabled={saving}>Cancel</button>
		</div>
	</div>
{:else if version}
	<div class="which">
		{#if choices.length > 1}
			<Segmented bind:value={shown} options={choices} label="Version" />
		{/if}
		<span class="muted small">{layer?.source} · {realTime(version.modified)}</span>
	</div>
	<pre class="file">{version.text}</pre>
{/if}

<style>
	header {
		margin-bottom: var(--space-4);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		margin-bottom: var(--space-2);
	}
	.path {
		font: 400 var(--step--1) / 1.5 var(--font-mono);
		color: var(--muted);
		margin-bottom: var(--space-2);
	}
	header .muted:not(.path) {
		max-width: 64rem;
	}
	.notice {
		margin-top: var(--space-2);
	}
	.changed {
		font-weight: 400;
		color: var(--changed);
	}
	.small {
		font-size: var(--step--1);
	}
	.which {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-3);
	}
	.file {
		margin: 0;
		max-width: 64rem;
		font: 400 var(--step--1) / 1.55 var(--font-mono);
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	/* Editing fills the window down to the hint strip, with Save in view
	   beneath it; the text scrolls inside, not the page. */
	.raw {
		position: sticky;
		top: 0;
		display: flex;
		flex-direction: column;
		max-width: 64rem; /* fitToViewport sets its height */
	}
	.raw textarea {
		display: block;
		flex: 1;
		min-height: 8rem;
		margin-top: var(--space-3);
		width: 100%;
		resize: none;
		font-family: var(--font-mono);
		font-size: var(--step--1);
		line-height: 1.55;
	}
	.buttons {
		display: flex;
		gap: var(--space-2);
		margin-top: var(--space-3);
		padding-bottom: var(--space-4);
	}
</style>
