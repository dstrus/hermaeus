<!-- An NPC's Universal Translator speech pattern, read and edited on their
     Overview. It lives in UniversalTranslator.yaml; an empty pattern
     removes the NPC's rule, and the global pattern applies to everyone. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { SpeechRule } from '#lib/api/types/settings.ts';
	import { setSpeech } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';

	let { template, rule: loaded, loadError = '' }: { template: string; rule: SpeechRule | null; loadError?: string } = $props();

	let rule = $derived(loaded);
	let readOnly = $derived(live.gameRunning);
	let draft = $state('');
	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let field = $state<HTMLTextAreaElement>();

	let editable = $derived(!!rule?.found && !readOnly);
	// Editing is a state of the Overview's address (?mode=speech), so it has
	// its own URL and Back cancels it.
	let editing = $derived(param('mode') === 'speech' && editable);
	$effect(() => {
		if (!editing && param('mode') === 'speech') setParams({ mode: null }, 'replace');
	});
	$effect(() => {
		if (editing) untrack(prepare);
	});

	function startEdit() {
		if (editable) setParams({ mode: 'speech' });
	}
	async function prepare() {
		draft = rule?.pattern ?? '';
		error = '';
		notice = '';
		await tick();
		field?.focus();
	}

	async function save(pattern: string) {
		if (!rule || saving) return;
		saving = true;
		error = '';
		try {
			rule = await setSpeech(template, { modified: rule.modified, pattern });
			notice = pattern.trim() ? 'Saved. SkyrimNet uses it from the next time it starts.' : 'Removed. They speak as written again.';
			leave('mode');
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	useShortcuts(() => [
		{ key: 'e', shown: 'E', label: 'Edit speech pattern', hidden: editing, disabled: !editable, run: startEdit },
		{ key: 'Enter', label: 'Save', command: true, hidden: !editing, disabled: saving, run: () => save(draft) },
		...(editing ? [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => leave('mode') }] : [])
	]);
</script>

<section>
	<h2 class="caps">Speech pattern</h2>
	{#if loadError}
		<p class="notice error" role="alert">{loadError}</p>
	{:else if !rule?.found}
		<p class="muted">Universal Translator hasn’t written its settings yet; SkyrimNet creates them the next time it starts. A pattern can be set then.</p>
	{:else if editing}
		<form onsubmit={(e) => (e.preventDefault(), save(draft))}>
			<textarea bind:this={field} bind:value={draft} rows="4" aria-label="Speech pattern" placeholder="How they speak: tone, rhythm, turns of phrase."></textarea>
			<div class="buttons">
				<button type="submit" class="button primary" disabled={saving}>Save</button>
				{#if rule.pattern}
					<button type="button" class="button" onclick={() => save('')} disabled={saving}>Remove</button>
				{/if}
				<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
			</div>
		</form>
	{:else}
		{#if rule.pattern}
			<p class="pattern">{rule.pattern}</p>
		{:else}
			<p class="muted">None: they speak as written{rule.global ? ', in the pattern set for every NPC' : ''}.</p>
		{/if}
		{#if rule.global}
			<p class="muted small">For every NPC: {rule.global}</p>
		{/if}
		{#if !rule.enabled}
			<p class="notice warn">Universal Translator is switched off, so patterns aren’t applied.</p>
		{/if}
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
		{#if notice}<p class="notice" role="status">{notice}</p>{/if}
		<div class="buttons">
			<button type="button" class="button" disabled={!editable} onclick={startEdit}>{rule.pattern ? 'Edit' : 'Set a pattern'}</button>
		</div>
		{#if readOnly}<p class="notice warn">Close Skyrim to change it.</p>{/if}
	{/if}
</section>

<style>
	h2 {
		font-size: var(--step-0);
		color: var(--accent);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-3);
	}
	.pattern {
		white-space: pre-wrap;
		line-height: 1.6;
		max-width: 65ch;
	}
	.small {
		font-size: var(--step--1);
		margin-top: var(--space-2);
	}
	textarea {
		display: block;
		width: 100%;
		max-width: 65ch;
		line-height: 1.55;
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}
	.notice {
		margin-top: var(--space-2);
	}
</style>
