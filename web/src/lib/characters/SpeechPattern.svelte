<!-- An NPC's Universal Translator speech pattern, read and edited on their
     Overview. It lives in UniversalTranslator.yaml; an empty pattern
     removes the NPC's rule, and the global pattern applies to everyone. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { SpeechRule } from '#lib/api/types/settings.ts';
	import { m } from '#lib/copy/messages.ts';
	import { setSpeech } from '#lib/settings/api.ts';
	import { translatorHref } from '#lib/translator/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';

	let { save: playthrough, template, rule: loaded, loadError = '' }: { save: string; template: string; rule: SpeechRule | null; loadError?: string } = $props();

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
			notice = pattern.trim() ? m.characters.speech.saved : m.characters.speech.removed;
			leave('mode');
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	useShortcuts(() => [
		{ key: 'e', shown: 'E', label: m.characters.speech.edit, hidden: editing, disabled: !editable, run: startEdit },
		{ key: 'Enter', label: m.common.save, command: true, hidden: !editing, disabled: saving, run: () => save(draft) },
		...(editing ? [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => leave('mode') }] : [])
	]);
</script>

<section>
	<h2 class="caps">{m.characters.speech.title}</h2>
	{#if loadError}
		<p class="notice error" role="alert">{loadError}</p>
	{:else if !rule?.found}
		<p class="muted">{m.characters.speech.notWritten}</p>
	{:else if editing}
		<form onsubmit={(e) => (e.preventDefault(), save(draft))}>
			<textarea bind:this={field} bind:value={draft} rows="4" aria-label={m.characters.speech.title} placeholder={m.characters.speech.placeholder}></textarea>
			<div class="buttons">
				<button type="submit" class="button primary" disabled={saving}>{m.common.save}</button>
				{#if rule.pattern}
					<button type="button" class="button" onclick={() => save('')} disabled={saving}>{m.common.remove}</button>
				{/if}
				<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>{m.common.cancel}</button>
			</div>
		</form>
	{:else}
		{#if rule.pattern}
			<p class="pattern">{rule.pattern}</p>
		{:else}
			<p class="muted">{rule.global ? m.characters.speech.noneGlobal : m.characters.speech.none}</p>
		{/if}
		{#if rule.global}
			<p class="muted small">{m.characters.speech.global({ pattern: rule.global })}</p>
		{/if}
		{#if !rule.enabled}
			<p class="notice warn">{m.characters.speech.off}</p>
		{/if}
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
		{#if notice}<p class="notice" role="status">{notice}</p>{/if}
		<div class="buttons">
			<button type="button" class="button" disabled={!editable} onclick={startEdit}>{rule.pattern ? m.common.edit : m.characters.speech.set}</button>
			<a class="button" href={translatorHref(playthrough)}>{m.characters.speech.allRules}</a>
		</div>
		{#if readOnly}<p class="notice warn">{m.characters.speech.closeToChange}</p>{/if}
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
	.buttons a {
		text-decoration: none;
	}
	.notice {
		margin-top: var(--space-2);
	}
</style>
