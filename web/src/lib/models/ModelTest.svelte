<!-- A test of a task's model: one prompt, sent on request through the
     provider SkyrimNet would use, with the reply, the time it took, and what
     it cost. Nothing is sent until the player asks. -->
<script lang="ts">
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import type { ModelTestResult, Provider, Task } from '#lib/api/types/settings.ts';
	import { m } from '#lib/copy/messages.ts';
	import { providerName, sample, testModel } from './api.ts';

	const t = m.models.test;

	let { task, providers }: { task: Task; providers: Provider[] } = $props();

	// Chat providers only: the others answer typed questions, not prompts.
	// Fateless is reached through SkyrimNet's own sign-in, which Hermaeus
	// doesn't have.
	const testable = (id: string) => {
		const p = providers.find((x) => x.id === id);
		return (p?.style ?? 'chat') === 'chat' && !p?.signIn;
	};
	let onFateless = $derived(task.slots.some((s) => providers.find((p) => p.id === s.provider)?.signIn));
	let slots = $derived(task.slots.map((s, i) => ({ s, i })).filter(({ s }) => s.model && testable(s.provider)));
	let slot = $state(-1);
	$effect(() => {
		if (!slots.some((x) => x.i === slot)) slot = (slots.find((x) => x.s.enabled) ?? slots[0])?.i ?? -1;
	});
	let chosen = $derived(task.slots[slot]);
	// The reply's cap: the task's own, or the test's, whichever is lower.
	const testCap = 1024;
	let cap = $derived(task.params.maxTokens && task.params.maxTokens > 0 ? Math.min(task.params.maxTokens, testCap) : testCap);

	// Each task starts from a prompt shaped like its job: brevity for Meta, an
	// action's name for action evaluation, sections for a bio. A plugin's
	// task, whose job Hermaeus doesn't know, just checks for a reply.
	const [startSystem, startPrompt] = untrack(() => sample(task.id));
	const vision = untrack(() => task.id === 'vision' || task.id === 'vision_quality');

	let system = $state(startSystem);
	let prompt = $state(startPrompt);
	let sending = $state(false);
	let error = $state('');
	let result = $state<ModelTestResult | null>(null);

	async function send() {
		if (sending || slot < 0) return;
		sending = true;
		error = '';
		result = null;
		try {
			result = await testModel(task.id, { slot, system, prompt });
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			sending = false;
		}
	}

	// A cost as providers quote it: tiny sums to a few significant figures.
	function dollars(n: number): string {
		if (n === 0) return t.free;
		if (n < 0.01) return `$${Number(n.toPrecision(2))}`;
		return `$${n.toFixed(4)}`;
	}
</script>

<section class="test">
	<h3 class="caps">{t.title}</h3>
	{#if slot < 0}
		<p class="muted small">
			{#if onFateless}
				{t.onFateless}
			{:else}
				{task.slots.length ? t.noChat : t.noModel}
			{/if}
		</p>
	{:else}
		<form onsubmit={(e) => (e.preventDefault(), send())}>
			{#if slots.length > 1}
				<label>
					<span class="caps">{t.model}</span>
					<select bind:value={slot}>
						{#each slots as { s, i } (i)}<option value={i}>{s.model} · {providerName(providers, s.provider)}{s.enabled ? '' : ` ${t.slotOff}`}</option>{/each}
					</select>
				</label>
			{/if}
			<label>
				<span class="caps">{t.instructions}</span>
				<textarea bind:value={system} rows="2"></textarea>
			</label>
			<label>
				<span class="caps">{t.prompt}</span>
				<textarea bind:value={prompt} rows="3"></textarea>
			</label>
			{#if vision}<p class="notice warn small">{t.vision}</p>{/if}
			<p class="muted small">{t.sends({ provider: providerName(providers, chosen?.provider ?? ''), tokens: cap.toLocaleString() })}</p>
			<div class="buttons">
				<button type="submit" class="button" disabled={sending || !prompt.trim()}>{sending ? t.waiting : t.send}</button>
			</div>
		</form>
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
		{#if result}
			<div class="result" role="status">
				<p class="reply">{result.reply || t.emptyReply}</p>
				{#if result.finishReason === 'length'}<p class="notice warn small">{t.cutOff({ tokens: result.maxTokens })}</p>{/if}
				<dl class="small">
					<dt class="caps">{t.answeredBy}</dt>
					<dd><span class="id">{result.answered || result.model}</span>{#if result.upstream}<span class="muted"> · {t.through({ upstream: result.upstream })}</span>{/if}</dd>
					<dt class="caps">{t.took}</dt>
					<dd>{t.seconds({ seconds: (result.millis / 1000).toFixed(1) })}</dd>
					<dt class="caps">{t.tokens}</dt>
					<dd>{t.tokensInOut({ input: result.promptTokens.toLocaleString(), output: result.completionTokens.toLocaleString() })}</dd>
					<dt class="caps">{t.cost}</dt>
					<dd>
						{#if result.cost !== undefined}{dollars(result.cost)}<span class="muted"> · {result.costFrom === 'provider' ? t.asBilled : t.byListedPrices}</span>{:else}<span class="muted">{t.notReported}</span>{/if}
					</dd>
				</dl>
				{#if result.reasoning}
					<details class="small">
						<summary>{t.reasoning}</summary>
						<p>{result.reasoning}</p>
					</details>
				{/if}
			</div>
		{/if}
	{/if}
</section>

<style>
	.test {
		margin-top: var(--space-5);
		padding: var(--space-3) 0 var(--space-5);
		border-top: 1px solid var(--rule);
	}
	/* A section of the pane: a step up from the field labels below it, in
	   ink, so it reads as a heading rather than another label. */
	h3 {
		font-size: var(--step-1);
		font-weight: 600;
		color: var(--ink);
		margin-bottom: var(--space-3);
	}
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	label .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	textarea,
	select {
		width: 100%;
	}
	textarea {
		resize: vertical;
		line-height: 1.5;
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.buttons {
		display: flex;
		gap: var(--space-2);
	}
	.result {
		margin-top: var(--space-3);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.reply {
		white-space: pre-wrap;
		line-height: 1.6;
		margin-bottom: var(--space-3);
	}
	dl {
		display: grid;
		grid-template-columns: 6.5rem minmax(0, 1fr);
		gap: var(--space-1) var(--space-3);
	}
	dt {
		color: var(--muted);
	}
	dd {
		overflow-wrap: anywhere;
	}
	.id {
		font-family: var(--font-mono);
	}
	details {
		margin-top: var(--space-3);
	}
	details p {
		white-space: pre-wrap;
		margin-top: var(--space-2);
	}
	.notice {
		margin: var(--space-2) 0;
	}
</style>
