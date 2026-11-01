<!-- One prompt as a document: the version SkyrimNet uses, the other layers'
     versions behind a switch, and editing that saves the player's own copy. -->
<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { href } from '#lib/nav.ts';
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Layer } from '#lib/api/types/layers.ts';
	import type { Document } from '#lib/api/types/prompts.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { realTime } from '#lib/format.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import Up from '#lib/ui/Up.svelte';
	import CodeArea from '#lib/ui/CodeArea.svelte';
	import Inja from '#lib/ui/Inja.svelte';
	import { layerLabel, promptDependency, promptName, promptsDependency, removeCopy, savePrompt } from './api.ts';

	const t = m.prompts;

	let { save, path, doc: loaded }: { save: string; path: string; doc: Document } = $props();

	// The document as last loaded or saved.
	let doc = $derived(loaded);
	let readOnly = $derived(live.gameRunning);
	const layerAt = (i: number): Layer => doc.layers[i] ?? { id: '', kind: '', source: t.unknownLayer };
	// Which layer's version is shown; starts on the one in use.
	let shown = $state(untrack(() => String(loaded.versions[0]?.layer ?? 0)));
	let version = $derived(doc.versions.find((v) => String(v.layer) === shown) ?? doc.versions[0]);
	let layer = $derived(version && layerAt(version.layer));
	let mine = $derived(doc.versions.find((v) => layerAt(v.layer).kind === 'overlay'));
	let below = $derived(doc.versions.filter((v) => layerAt(v.layer).kind !== 'overlay'));
	let choices = $derived(doc.versions.map((v) => ({ value: String(v.layer), label: layerLabel(layerAt(v.layer)), title: layerAt(v.layer).source })));

	// Editing is a state of the address (?mode=edit), so it has its own URL
	// and Back cancels it.
	let mode = $derived<'read' | 'edit'>(param('mode') === 'edit' && version ? 'edit' : 'read');
	$effect(() => {
		if (mode === 'read' && param('mode')) setParams({ mode: null }, 'replace');
	});
	let draft = $state('');
	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let warnings = $state<string[]>([]);

	let list = $derived(href(save, 'prompts'));

	function startEdit() {
		if (version && !readOnly) setParams({ mode: 'edit' });
	}
	// The draft starts from the shown version whenever editing starts, however it starts.
	$effect(() => {
		if (mode !== 'edit') return;
		untrack(() => {
			draft = version?.text ?? '';
			error = '';
			notice = '';
			warnings = [];
		});
	});

	async function submit() {
		if (saving) return;
		saving = true;
		error = '';
		try {
			const saved = await savePrompt(path, { modified: mine?.modified ?? 0, text: draft });
			doc = saved.document;
			warnings = saved.warnings ?? [];
			shown = String(doc.versions[0]?.layer ?? 0);
			notice = mine ? t.savedCopy : t.savedAsCopy;
			await leave('mode'); // settles before the reload, which it would cancel
			await invalidate(promptsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	async function remove() {
		if (!mine || readOnly) return;
		const next = below[0] ? t.nextVersion({ source: layerAt(below[0].layer).source }) : t.nextNothing;
		if (!confirm(t.confirmRemove({ name: promptName(path), next }))) return;
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
			notice = t.removed;
			await invalidate(promptDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		return [
			{ key: 'e', shown: 'E', label: mine && version === mine ? m.common.edit : t.editCopy, hidden: !reading, disabled: readOnly || !version, run: startEdit },
			...deleteShortcuts({ label: t.removeCopy, hidden: !reading || !mine, disabled: readOnly, run: remove }),
			{ key: 'Enter', label: m.common.save, command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: reading ? t.title : m.common.cancel, run: () => (reading ? goto(list) : leave('mode')) }
		];
	});
</script>

<header>
	<Up href={list} label={t.title} />
	<h1 class="caps">{promptName(path)}</h1>
	<p class="path">{path}</p>
	<p class="muted">
		{#if mine}
			<Fill text={(below[0] ? t.mineInPlaceOf : t.mine).text} source={below[0] ? layerAt(below[0].layer).source : ''}>{#snippet yourCopy()}<b class="changed">{t.yourOwnCopy}</b>{/snippet}</Fill>
		{:else if layer}
			{t.usesVersion({ source: layer.source })}
		{/if}
	</p>
	{#if readOnly}<p class="notice warn">{t.readOnly}</p>{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	{#if notice}<p class="notice" role="status">{notice}</p>{/if}
	{#each warnings as w (w)}<p class="notice warn">{t.checkThis({ warning: w })}</p>{/each}
</header>

{#if mode === 'edit'}
	<div class="raw" use:fitToViewport>
		{#if version !== mine}
			<p class="muted small">{t.startingFrom({ source: layer?.source ?? '' })}</p>
		{/if}
		<CodeArea bind:value={draft} label={t.promptText} />
		<div class="buttons">
			<button type="button" class="button primary" onclick={submit} disabled={saving}>{m.common.save}</button>
			<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>{m.common.cancel}</button>
		</div>
	</div>
{:else if version}
	<div class="which">
		{#if choices.length > 1}
			<Segmented bind:value={shown} options={choices} label={t.version} />
		{/if}
		<span class="muted small">{layer?.source} · {realTime(version.modified)}</span>
	</div>
	<pre class="file"><Inja text={version.text} /></pre>
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
	.raw :global(.code) {
		flex: 1;
		margin-top: var(--space-3);
	}
	.buttons {
		display: flex;
		gap: var(--space-2);
		margin-top: var(--space-3);
		padding-bottom: var(--space-4);
	}
</style>
