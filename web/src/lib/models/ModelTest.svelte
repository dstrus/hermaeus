<!-- A test of a task's model: one prompt, sent on request through the
     provider SkyrimNet would use, with the reply, the time it took, and what
     it cost. Nothing is sent until the player asks. -->
<script lang="ts">
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import type { ModelTestResult, Provider, Task } from '#lib/api/types/settings.ts';
	import { providerName, testModel } from './api.ts';

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
	// action's name for action evaluation, sections for a bio. They're
	// stand-ins for SkyrimNet's real prompts, which are long templates filled
	// in from the game.
	const starts: Record<string, [system: string, prompt: string]> = {
		default: ['You are an innkeeper in Skyrim. Stay in character.', 'A traveler walks in out of the snow. Greet them in a sentence or two.'],
		combat: ['You are a bandit in the middle of a fight. Stay in character.', 'The traveler just cut your arm with a war axe. Say one line, then choose: flee, surrender, or fight on.'],
		CharacterProfileGeneration: ['Write a character bio with three short sections: Summary, Personality, and Speech.', 'Write a bio for the innkeeper of an inn in Whiterun, in under 150 words.'],
		DiaryGeneration: ['Write a first-person diary entry for the character described.', 'You are a housecarl. Today you fought a dragon at a watchtower beside your thane. Write tonight’s entry in about 120 words.'],
		memory: ['Turn the events into one first-person memory, then rate how much it matters, from 0 to 1.', 'The traveler paid for a room. They asked about the Companions. They left a large tip.'],
		gamemaster_evaluation: ['You direct a scene. Say who should speak next and why, or that no one should.', 'In a tavern, a bard has just finished a song. A guard is drinking alone, and two farmers are arguing about the harvest.'],
		action_evaluation: ['Choose the best action from the list, or none. Reply with its name only.', 'The innkeeper hears: “I’d like a room for the night.” Actions: RentRoom, OpenTrade, Follow, none.'],
		meta: ['Answer in one word.', 'Who is this line meant for, the innkeeper or the bard? “Play something cheerful.”'],
		UniversalTranslator: ['Rewrite the line in the speech pattern given, keeping its meaning.', 'Pattern: an old sea captain. Line: “I don’t know where the key is, but the jarl might.”'],
		AgentDefault: ['You help a player set up SkyrimNet.', 'How do I make characters talk to each other less often?'],
		vision: ['Describe the place in two sentences, as for someone who can’t see it.', 'A snowy courtyard with a broken well and a shuttered forge.'],
		vision_quality: ['Describe the place in two sentences, as for someone who can’t see it.', 'A snowy courtyard with a broken well and a shuttered forge.']
	};
	// A plugin's task, whose job Hermaeus doesn't know, just checks for a reply.
	const [startSystem, startPrompt] = untrack(() => starts[task.id] ?? ['', 'Reply with one short sentence to confirm you’re working.']);
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
		if (n === 0) return 'free';
		if (n < 0.01) return `$${Number(n.toPrecision(2))}`;
		return `$${n.toFixed(4)}`;
	}
</script>

<section class="test">
	<h3 class="caps">Try it</h3>
	{#if slot < 0}
		<p class="muted small">
			{#if onFateless}
				Its models are on Fateless, which SkyrimNet reaches through its own sign-in. Hermaeus doesn’t have that sign-in, so SkyrimNet’s dashboard is where to test them.
			{:else}
				{task.slots.length ? 'None of its models is on a chat provider, so there’s nothing a test can send to.' : 'It has no model of its own to try.'}
			{/if}
		</p>
	{:else}
		<form onsubmit={(e) => (e.preventDefault(), send())}>
			{#if slots.length > 1}
				<label>
					<span class="caps">Model</span>
					<select bind:value={slot}>
						{#each slots as { s, i } (i)}<option value={i}>{s.model} · {providerName(providers, s.provider)}{s.enabled ? '' : ' (off)'}</option>{/each}
					</select>
				</label>
			{/if}
			<label>
				<span class="caps">Instructions</span>
				<textarea bind:value={system} rows="2"></textarea>
			</label>
			<label>
				<span class="caps">Prompt</span>
				<textarea bind:value={prompt} rows="3"></textarea>
			</label>
			{#if vision}<p class="notice warn small">A test sends text only, so the model gets a description rather than a picture. It shows how the model writes, not how it sees.</p>{/if}
			<p class="muted small">Sends one request to {providerName(providers, chosen?.provider ?? '')}, as SkyrimNet would, with the task’s sampling settings and up to {cap.toLocaleString()} tokens of reply. A paid provider bills it like any other request.</p>
			<div class="buttons">
				<button type="submit" class="button" disabled={sending || !prompt.trim()}>{sending ? 'Waiting for a reply…' : 'Send'}</button>
			</div>
		</form>
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
		{#if result}
			<div class="result" role="status">
				<p class="reply">{result.reply || '(An empty reply.)'}</p>
				{#if result.finishReason === 'length'}<p class="notice warn small">Cut off at {result.maxTokens} tokens.</p>{/if}
				<dl class="small">
					<dt class="caps">Answered by</dt>
					<dd><span class="id">{result.answered || result.model}</span>{#if result.upstream}<span class="muted"> · through {result.upstream}</span>{/if}</dd>
					<dt class="caps">Took</dt>
					<dd>{(result.millis / 1000).toFixed(1)} s</dd>
					<dt class="caps">Tokens</dt>
					<dd>{result.promptTokens.toLocaleString()} in, {result.completionTokens.toLocaleString()} out</dd>
					<dt class="caps">Cost</dt>
					<dd>
						{#if result.cost !== undefined}{dollars(result.cost)}<span class="muted"> · {result.costFrom === 'provider' ? 'as billed' : 'by its listed prices'}</span>{:else}<span class="muted">Not reported. Fetch this provider’s models to price it by their list.</span>{/if}
					</dd>
				</dl>
				{#if result.reasoning}
					<details class="small">
						<summary>Its reasoning</summary>
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
