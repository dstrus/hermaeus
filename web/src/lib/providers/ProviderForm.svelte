<!-- A provider as a form: where it is, how SkyrimNet talks to it, and its
     settings. The key has a form of its own, since it's written and never
     read back. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import type { ProviderEdit } from '#lib/api/types/settings.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { apiStyles, sortings } from './api.ts';

	let {
		edit: start,
		isNew = false,
		isOpenRouter = false,
		saving = false,
		onsave,
		oncancel
	}: {
		edit: ProviderEdit;
		isNew?: boolean;
		isOpenRouter?: boolean; // the sorting setting is OpenRouter's own
		saving?: boolean;
		onsave: (edit: ProviderEdit) => void;
		oncancel: () => void;
	} = $props();

	const first = untrack(() => start);
	let name = $state(first.name);
	let endpoint = $state(first.endpoint);
	let style = $state(first.style);
	let streaming = $state(first.streaming);
	let structuredOutputs = $state(first.structuredOutputs);
	let reasoning = $state(first.reasoning);
	let settings = $state(first.settings ? JSON.stringify(JSON.parse(first.settings), null, 2) : '');
	let sorting = $state(first.sorting || 'default');
	let cacheTTL = $state(first.cacheTTL || '5m');
	let more = $state(!!first.settings || first.cacheTTL === '1h');
	let error = $state('');
	let nameField = $state<HTMLInputElement>();

	$effect(() => {
		tick().then(() => nameField?.focus());
	});

	function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		let compact = '';
		if (settings.trim()) {
			try {
				compact = JSON.stringify(JSON.parse(settings));
			} catch {
				error = 'The provider settings must be valid JSON, an object of keys and values.';
				return;
			}
		}
		onsave({ modified: start.modified, name, endpoint, style, streaming, structuredOutputs, reasoning, settings: compact, sorting: isOpenRouter ? sorting : first.sorting, cacheTTL });
	}
</script>

<form onsubmit={submit}>
	<label class="field">
		<span class="caps label">Name</span>
		<input bind:this={nameField} type="text" bind:value={name} placeholder="Groq" autocomplete="off" />
		{#if isNew}<span class="muted small">Its id, for the tasks that use it, is made from the name.</span>{/if}
	</label>
	<label class="field">
		<span class="caps label">Endpoint</span>
		<input type="url" bind:value={endpoint} placeholder="https://api.example.com/v1/chat/completions" autocomplete="off" spellcheck="false" />
		<span class="muted small">The full URL SkyrimNet sends requests to.</span>
	</label>
	<div class="field">
		<span class="caps label">API style</span>
		<Segmented bind:value={style} options={apiStyles} label="API style" />
		<span class="muted small">{apiStyles.find((s) => s.value === style)?.what}</span>
	</div>
	<div class="checks">
		<label class="check"><input type="checkbox" bind:checked={streaming} /> <span>Stream the answer</span><span class="muted small">Server-sent events, so lines arrive as they're written.</span></label>
		<label class="check"><input type="checkbox" bind:checked={structuredOutputs} /> <span>Structured outputs</span><span class="muted small">Ask for JSON in the provider's own format, where it supports that.</span></label>
		<label class="check"><input type="checkbox" bind:checked={reasoning} /> <span>Reasoning</span><span class="muted small">Let models think before answering, where the provider offers it.</span></label>
	</div>
	<button type="button" class="textbutton small more" onclick={() => (more = !more)}>{more ? 'Fewer settings' : 'More settings'}</button>
	{#if more}
		{#if isOpenRouter}
			<label class="field">
				<span class="caps label">Upstream sorting</span>
				<select bind:value={sorting}>
					{#each sortings as s (s.value)}<option value={s.value}>{s.label}</option>{/each}
				</select>
				<span class="muted small">How OpenRouter chooses among the hosts that serve a model.</span>
			</label>
		{/if}
		<label class="field">
			<span class="caps label">Prompt cache</span>
			<select bind:value={cacheTTL}>
				<option value="5m">5 minutes</option>
				<option value="1h">An hour</option>
			</select>
			<span class="muted small">How long the provider keeps the cached start of a prompt. An hour lasts through longer pauses, but costs more with providers that charge for caching.</span>
		</label>
		<label class="field">
			<span class="caps label">Provider settings, as JSON</span>
			<textarea bind:value={settings} rows="6" spellcheck="false" placeholder={'{\n  "route": "fallback"\n}'}></textarea>
			<span class="muted small">Sent with every request, as the provider's API takes them. Empty sends none.</span>
		</label>
	{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<div class="buttons">
		<button type="submit" class="button primary" disabled={saving}>{isNew ? 'Add' : 'Save'}</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>Cancel</button>
	</div>
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	input,
	select,
	textarea {
		width: 100%;
	}
	textarea {
		resize: vertical;
		font-family: var(--font-mono);
		font-size: var(--step--1);
		line-height: 1.5;
	}
	.checks {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.check {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0 var(--space-2);
		align-items: center;
	}
	.check .small {
		grid-column: 2;
	}
	.more {
		align-self: flex-start;
	}
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0;
		background: var(--panel);
	}
</style>
