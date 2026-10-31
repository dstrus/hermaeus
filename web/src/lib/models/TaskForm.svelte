<!-- One task's models as a form: the rotation, slot by slot, and the
     settings every slot falls back on. A field left empty means "not set",
     so the slot takes the task's setting, and the task the provider's. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import type { Provider, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { blankSlot, reasoningEfforts } from './api.ts';
	import { describe, modelLists } from './lists.svelte.ts';

	let {
		task,
		modified,
		providers,
		choices = [],
		saving = false,
		onsave,
		oncancel
	}: {
		task: Task;
		modified: number;
		providers: Provider[];
		choices?: string[]; // models named elsewhere in the file
		saving?: boolean;
		onsave: (edit: TaskEdit) => void;
		oncancel: () => void;
	} = $props();

	// Numbers are edited as text, so a field can be empty: not set.
	const text = (n: number | undefined) => (n === undefined || n === null ? '' : String(n));
	interface Draft {
		origin?: number;
		model: string;
		provider: string;
		enabled: boolean;
		role: string;
		temperature: string;
		weight: string;
		requestTimeout: string;
		connectTimeout: string;
		maxRetries: string;
		reasoningEffort: string;
		structuredOutputs: string; // '', 'true', or 'false'
		more: boolean; // the slot's further settings are shown
	}
	const draftOf = (s: Slot): Draft => ({
		origin: s.origin,
		model: s.model,
		provider: s.provider,
		enabled: s.enabled,
		role: s.role ?? '',
		temperature: text(s.temperature),
		weight: text(s.weight),
		requestTimeout: text(s.requestTimeout),
		connectTimeout: text(s.connectTimeout),
		maxRetries: text(s.maxRetries),
		reasoningEffort: s.reasoningEffort ?? '',
		structuredOutputs: s.structuredOutputs === undefined ? '' : String(s.structuredOutputs),
		more: [s.weight, s.requestTimeout, s.connectTimeout, s.maxRetries, s.reasoningEffort, s.structuredOutputs].some((v) => v !== undefined && v !== '')
	});

	const start = untrack(() => task);
	let slots = $state<Draft[]>(start.slots.map(draftOf));
	let temperature = $state(text(start.params.temperature));
	let maxTokens = $state(text(start.params.maxTokens));
	let topP = $state(text(start.params.topP));
	let topK = $state(text(start.params.topK));
	let frequencyPenalty = $state(text(start.params.frequencyPenalty));
	let presencePenalty = $state(text(start.params.presencePenalty));
	let stopSequences = $state(start.params.stopSequences.join('\n'));
	let timeout = $state(text(start.timeout));
	let rotationMode = $state(start.rotationMode ?? '');
	let error = $state('');
	let first = $state<HTMLInputElement>();

	let isDialogue = $derived(task.id === 'default');
	let sampling = $derived(task.id !== 'decisions'); // a decision model takes no sampling settings
	let orders = $derived([
		...(isDialogue ? [] : [{ value: '', label: 'As Dialogue' }]),
		{ value: 'sequential', label: 'In turn' },
		{ value: 'random', label: 'At random' }
	]);
	let usableProviders = $derived(providers.length ? providers : [{ id: 'openrouter', name: 'OpenRouter', style: 'chat', hasKey: false, builtin: true }]);
	// The providers the slots use, for fetching each one's model list.
	let used = $derived([...new Set(slots.map((d) => d.provider))].map((id) => usableProviders.find((p) => p.id === id)).filter((p) => p !== undefined));
	const listOf = (provider: string) => modelLists.lists[provider];
	// What a provider says about a slot's model, or that it doesn't list it.
	function hint(d: Draft): string {
		const list = listOf(d.provider);
		const model = d.model.trim();
		if (!list || !model) return '';
		const m = modelLists.find(d.provider, model);
		if (!m) return `Not in ${usableProviders.find((p) => p.id === d.provider)?.name ?? d.provider}’s list.`;
		return describe(m);
	}

	$effect(() => {
		tick().then(() => first?.focus());
	});

	function add() {
		slots = [...slots, { ...draftOf(blankSlot(slots.at(-1)?.provider ?? usableProviders[0]!.id)), more: false }];
		tick().then(() => document.querySelector<HTMLInputElement>('.slot:last-of-type input[list]')?.focus());
	}
	function remove(i: number) {
		slots = slots.filter((_, j) => j !== i);
	}
	function move(i: number, by: number) {
		const to = i + by;
		if (to < 0 || to >= slots.length) return;
		const next = [...slots];
		[next[i], next[to]] = [next[to]!, next[i]!];
		slots = next;
	}

	// num reads a number field; empty is "not set".
	function num(s: string, label: string, whole = false): number | undefined {
		const t = s.trim();
		if (t === '') return undefined;
		const n = Number(t);
		if (Number.isNaN(n)) throw new Error(`${label} must be a number.`);
		if (whole && !Number.isInteger(n)) throw new Error(`${label} must be a whole number.`);
		return n;
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			if (!slots.length) throw new Error('Add at least one model.');
			const edit: TaskEdit = {
				modified,
				slots: slots.map((d, i) => {
					const at = `Model ${i + 1}`;
					if (!d.model.trim()) throw new Error(`${at} needs a name.`);
					const so = d.structuredOutputs === '' ? undefined : d.structuredOutputs === 'true';
					return {
						origin: d.origin,
						model: d.model.trim(),
						provider: d.provider,
						enabled: d.enabled,
						role: d.role || undefined,
						temperature: num(d.temperature, `${at}’s temperature`),
						weight: num(d.weight, `${at}’s weight`),
						requestTimeout: num(d.requestTimeout, `${at}’s request timeout`, true),
						connectTimeout: num(d.connectTimeout, `${at}’s connect timeout`, true),
						maxRetries: num(d.maxRetries, `${at}’s retries`, true),
						reasoningEffort: d.reasoningEffort || undefined,
						structuredOutputs: so
					};
				}),
				params: {
					temperature: num(temperature, 'Temperature'),
					maxTokens: num(maxTokens, 'Max tokens', true),
					topP: num(topP, 'Top-p'),
					topK: num(topK, 'Top-k', true),
					frequencyPenalty: num(frequencyPenalty, 'Frequency penalty'),
					presencePenalty: num(presencePenalty, 'Presence penalty'),
					stopSequences: stopSequences.split('\n').filter((l) => l !== '')
				},
				timeout: num(timeout, 'Timeout', true),
				rotationMode: rotationMode || undefined
			};
			onsave(edit);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		}
	}
</script>

<form onsubmit={submit}>
	<fieldset class="rotation">
		<legend class="caps label">Rotation</legend>
		<p class="muted small">SkyrimNet tries the models in turn or at random; a backup only answers when no primary does.</p>
		<ol class="slots">
			{#each slots as s, i (i)}
				<li class="slot" class:off={!s.enabled}>
					<div class="head">
						<span class="caps muted n">{i + 1}</span>
						<div class="tools">
							<button type="button" class="textbutton small" onclick={() => move(i, -1)} disabled={i === 0}>Up</button>
							<button type="button" class="textbutton small" onclick={() => move(i, 1)} disabled={i === slots.length - 1}>Down</button>
							<button type="button" class="textbutton small" onclick={() => remove(i)}>Remove</button>
						</div>
					</div>
					<label class="field">
						<span class="caps label">Model</span>
						{#if i === 0}
							<input bind:this={first} type="text" bind:value={s.model} list="models-{listOf(s.provider) ? s.provider : 'known'}" placeholder="vendor/model" autocomplete="off" spellcheck="false" />
						{:else}
							<input type="text" bind:value={s.model} list="models-{listOf(s.provider) ? s.provider : 'known'}" placeholder="vendor/model" autocomplete="off" spellcheck="false" />
						{/if}
						{#if hint(s)}<span class="muted small">{hint(s)}</span>{/if}
					</label>
					<div class="row">
						<label class="field">
							<span class="caps label">Provider</span>
							<select bind:value={s.provider}>
								{#each usableProviders as p (p.id)}<option value={p.id}>{p.name}{p.hasKey ? '' : ' (no key)'}</option>{/each}
							</select>
						</label>
						<label class="field">
							<span class="caps label">Role</span>
							<select bind:value={s.role}>
								<option value="">Primary</option>
								<option value="backup">Backup</option>
							</select>
						</label>
						{#if sampling}
							<label class="field short">
								<span class="caps label">Temp.</span>
								<input type="text" inputmode="decimal" bind:value={s.temperature} placeholder={temperature || '—'} />
							</label>
						{/if}
					</div>
					<div class="row">
						<label class="check"><input type="checkbox" bind:checked={s.enabled} /> <span>In rotation</span></label>
						<button type="button" class="textbutton small" onclick={() => (s.more = !s.more)}>{s.more ? 'Fewer settings' : 'More settings'}</button>
					</div>
					{#if s.more}
						<div class="more">
							<label class="field short">
								<span class="caps label">Weight</span>
								<input type="text" inputmode="decimal" bind:value={s.weight} placeholder="1" />
							</label>
							<label class="field short">
								<span class="caps label">Request timeout</span>
								<input type="text" inputmode="numeric" bind:value={s.requestTimeout} placeholder={timeout || '—'} />
							</label>
							<label class="field short">
								<span class="caps label">Connect timeout</span>
								<input type="text" inputmode="numeric" bind:value={s.connectTimeout} placeholder="—" />
							</label>
							<label class="field short">
								<span class="caps label">Retries</span>
								<input type="text" inputmode="numeric" bind:value={s.maxRetries} placeholder="—" />
							</label>
							<label class="field">
								<span class="caps label">Reasoning</span>
								<select bind:value={s.reasoningEffort}>
									{#each reasoningEfforts as r (r.value)}<option value={r.value}>{r.label}</option>{/each}
								</select>
							</label>
							<label class="field">
								<span class="caps label">Structured outputs</span>
								<select bind:value={s.structuredOutputs}>
									<option value="">Provider’s default</option>
									<option value="true">On</option>
									<option value="false">Off</option>
								</select>
							</label>
						</div>
					{/if}
				</li>
			{/each}
		</ol>
		<datalist id="models-known">
			{#each choices as m (m)}<option value={m}></option>{/each}
		</datalist>
		{#each used as p (p.id)}
			{@const list = listOf(p.id)}
			<datalist id="models-{p.id}">
				{#each list?.models ?? [] as m (m.id)}<option value={m.id}>{m.name && m.name !== m.id ? m.name : ''}</option>{/each}
			</datalist>
		{/each}
		<div class="row">
			<button type="button" class="button" onclick={add}>Add a model</button>
		</div>
		<!-- Each provider's list is fetched only when asked: the call goes to the provider, with the key when it has one. -->
		<div class="lists">
			{#each used as p (p.id)}
				{@const list = listOf(p.id)}
				<div class="list">
					<button type="button" class="textbutton" onclick={() => modelLists.fetch(p.id, !!list)} disabled={modelLists.loading[p.id] || p.style !== 'chat' || p.signIn}>
						{modelLists.loading[p.id] ? `Asking ${p.name}…` : list ? `Fetch ${p.name}’s models again` : `Fetch ${p.name}’s models`}
					</button>
					<span class="muted small">
						{#if modelLists.errors[p.id]}
							<span class="error">{modelLists.errors[p.id]}</span>
						{:else if list}
							{list.models.length} models, offered as you type.
						{:else if p.signIn}
							{p.name} is reached through SkyrimNet’s own sign-in, so Hermaeus can’t fetch its list.
						{:else if p.style !== 'chat'}
							{p.name} has no list to fetch.
						{:else}
							Asks {p.name} over the Internet{p.hasKey ? ', with your key' : ''}.
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</fieldset>

	<fieldset>
		<legend class="caps label">Order</legend>
		<Segmented bind:value={rotationMode} options={orders} label="Rotation order" />
	</fieldset>

	{#if sampling}
		<fieldset class="params">
			<legend class="caps label">For every model here</legend>
			<p class="muted small">Unless a slot above says otherwise. Empty leaves it to the provider.</p>
			<div class="grid">
				<label class="field short"><span class="caps label">Temperature</span><input type="text" inputmode="decimal" bind:value={temperature} /></label>
				<label class="field short"><span class="caps label">Max tokens</span><input type="text" inputmode="numeric" bind:value={maxTokens} /></label>
				<label class="field short"><span class="caps label">Timeout (s)</span><input type="text" inputmode="numeric" bind:value={timeout} placeholder={isDialogue ? '' : 'As Dialogue'} /></label>
				<label class="field short"><span class="caps label">Top-p</span><input type="text" inputmode="decimal" bind:value={topP} /></label>
				<label class="field short"><span class="caps label">Top-k</span><input type="text" inputmode="numeric" bind:value={topK} /></label>
				<label class="field short"><span class="caps label">Frequency penalty</span><input type="text" inputmode="decimal" bind:value={frequencyPenalty} /></label>
				<label class="field short"><span class="caps label">Presence penalty</span><input type="text" inputmode="decimal" bind:value={presencePenalty} /></label>
			</div>
			<label class="field">
				<span class="caps label">Stop sequences, one per line</span>
				<textarea bind:value={stopSequences} rows="2" spellcheck="false"></textarea>
			</label>
		</fieldset>
	{:else}
		<fieldset>
			<legend class="caps label">Timeout</legend>
			<label class="field short"><span class="caps label">Seconds</span><input type="text" inputmode="numeric" bind:value={timeout} placeholder="As Dialogue" /></label>
		</fieldset>
	{/if}

	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<div class="buttons">
		<button type="submit" class="button primary" disabled={saving}>Save</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>Cancel</button>
	</div>
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-5);
	}
	fieldset {
		border: 0;
		padding: 0;
		margin: 0;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	legend {
		padding: 0;
		margin-bottom: var(--space-2);
		color: var(--accent);
		font-size: var(--step--1);
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.slots {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.slot {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		padding: var(--space-3);
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
	}
	.slot.off {
		border-style: dashed;
	}
	.head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
	}
	.head .n {
		font-variant-numeric: tabular-nums;
	}
	.tools {
		display: flex;
		gap: var(--space-3);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		min-width: 0;
	}
	.field input,
	.field select,
	.field textarea {
		width: 100%;
	}
	.row {
		display: flex;
		flex-wrap: wrap;
		align-items: end;
		gap: var(--space-3);
	}
	.row .field {
		flex: 1 1 8rem;
	}
	.row .field.short {
		flex: 0 1 5rem;
	}
	.row .check {
		flex: 1;
	}
	.more,
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
		gap: var(--space-3);
	}
	.check {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}
	textarea {
		resize: vertical;
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	.lists {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.list {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-1) var(--space-3);
	}
	.list .error {
		color: var(--warn);
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
