<!-- One task's models as a form: the rotation, slot by slot, and the
     settings every slot falls back on. A field left empty means "not set",
     so the slot takes the task's setting, and the task the provider's. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import type { Provider, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';
	import { m } from '#lib/copy/messages.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { blankSlot, reasoningEfforts } from './api.ts';
	import { describe, modelLists } from './lists.svelte.ts';

	const t = m.models.form;

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
		...(isDialogue ? [] : [{ value: '', label: t.asDialogue }]),
		{ value: 'sequential', label: m.models.inTurn },
		{ value: 'random', label: m.models.atRandom }
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
		const info = modelLists.find(d.provider, model);
		if (!info) return t.notInList({ provider: usableProviders.find((p) => p.id === d.provider)?.name ?? d.provider });
		return describe(info);
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
		const trimmed = s.trim();
		if (trimmed === '') return undefined;
		const n = Number(trimmed);
		if (Number.isNaN(n)) throw new Error(t.notNumber({ label }));
		if (whole && !Number.isInteger(n)) throw new Error(t.notWhole({ label }));
		return n;
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			if (!slots.length) throw new Error(t.addOne);
			const edit: TaskEdit = {
				modified,
				slots: slots.map((d, i) => {
					const at = t.slot({ n: i + 1 });
					if (!d.model.trim()) throw new Error(t.needsName({ slot: at }));
					const so = d.structuredOutputs === '' ? undefined : d.structuredOutputs === 'true';
					return {
						origin: d.origin,
						model: d.model.trim(),
						provider: d.provider,
						enabled: d.enabled,
						role: d.role || undefined,
						temperature: num(d.temperature, t.slotTemperature({ slot: at })),
						weight: num(d.weight, t.slotWeight({ slot: at })),
						requestTimeout: num(d.requestTimeout, t.slotRequestTimeout({ slot: at }), true),
						connectTimeout: num(d.connectTimeout, t.slotConnectTimeout({ slot: at }), true),
						maxRetries: num(d.maxRetries, t.slotRetries({ slot: at }), true),
						reasoningEffort: d.reasoningEffort || undefined,
						structuredOutputs: so
					};
				}),
				params: {
					temperature: num(temperature, t.temperature),
					maxTokens: num(maxTokens, t.maxTokens, true),
					topP: num(topP, t.topP),
					topK: num(topK, t.topK, true),
					frequencyPenalty: num(frequencyPenalty, t.frequencyPenalty),
					presencePenalty: num(presencePenalty, t.presencePenalty),
					stopSequences: stopSequences.split('\n').filter((l) => l !== '')
				},
				timeout: num(timeout, t.timeout, true),
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
		<legend class="caps label">{t.rotation}</legend>
		<p class="muted small">{t.rotationHint}</p>
		<ol class="slots">
			{#each slots as s, i (i)}
				<li class="slot" class:off={!s.enabled}>
					<div class="head">
						<span class="caps muted n">{i + 1}</span>
						<div class="tools">
							<button type="button" class="textbutton small" onclick={() => move(i, -1)} disabled={i === 0}>{m.common.up}</button>
							<button type="button" class="textbutton small" onclick={() => move(i, 1)} disabled={i === slots.length - 1}>{m.common.down}</button>
							<button type="button" class="textbutton small" onclick={() => remove(i)}>{m.common.remove}</button>
						</div>
					</div>
					<label class="field">
						<span class="caps label">{t.model}</span>
						{#if i === 0}
							<input bind:this={first} type="text" bind:value={s.model} list="models-{listOf(s.provider) ? s.provider : 'known'}" placeholder={t.modelPlaceholder} autocomplete="off" spellcheck="false" />
						{:else}
							<input type="text" bind:value={s.model} list="models-{listOf(s.provider) ? s.provider : 'known'}" placeholder={t.modelPlaceholder} autocomplete="off" spellcheck="false" />
						{/if}
						{#if hint(s)}<span class="muted small">{hint(s)}</span>{/if}
					</label>
					<div class="row">
						<label class="field">
							<span class="caps label">{t.provider}</span>
							<select bind:value={s.provider}>
								{#each usableProviders as p (p.id)}<option value={p.id}>{p.hasKey ? p.name : t.noKey({ name: p.name })}</option>{/each}
							</select>
						</label>
						<label class="field">
							<span class="caps label">{t.role}</span>
							<select bind:value={s.role}>
								<option value="">{t.primary}</option>
								<option value="backup">{t.backup}</option>
							</select>
						</label>
						{#if sampling}
							<label class="field short">
								<span class="caps label">{t.temp}</span>
								<input type="text" inputmode="decimal" bind:value={s.temperature} placeholder={temperature || '—'} />
							</label>
						{/if}
					</div>
					<div class="row">
						<label class="check"><input type="checkbox" bind:checked={s.enabled} /> <span>{t.inRotation}</span></label>
						<button type="button" class="textbutton small" onclick={() => (s.more = !s.more)}>{s.more ? m.common.fewerSettings : m.common.moreSettings}</button>
					</div>
					{#if s.more}
						<div class="more">
							<label class="field short">
								<span class="caps label">{t.weight}</span>
								<input type="text" inputmode="decimal" bind:value={s.weight} placeholder="1" />
							</label>
							<label class="field short">
								<span class="caps label">{t.requestTimeout}</span>
								<input type="text" inputmode="numeric" bind:value={s.requestTimeout} placeholder={timeout || '—'} />
							</label>
							<label class="field short">
								<span class="caps label">{t.connectTimeout}</span>
								<input type="text" inputmode="numeric" bind:value={s.connectTimeout} placeholder="—" />
							</label>
							<label class="field short">
								<span class="caps label">{t.retries}</span>
								<input type="text" inputmode="numeric" bind:value={s.maxRetries} placeholder="—" />
							</label>
							<label class="field">
								<span class="caps label">{t.reasoning}</span>
								<select bind:value={s.reasoningEffort}>
									{#each reasoningEfforts as r (r.value)}<option value={r.value}>{r.label}</option>{/each}
								</select>
							</label>
							<label class="field">
								<span class="caps label">{t.structuredOutputs}</span>
								<select bind:value={s.structuredOutputs}>
									<option value="">{t.providerDefault}</option>
									<option value="true">{m.common.isOn}</option>
									<option value="false">{m.common.isOff}</option>
								</select>
							</label>
						</div>
					{/if}
				</li>
			{/each}
		</ol>
		<datalist id="models-known">
			{#each choices as model (model)}<option value={model}></option>{/each}
		</datalist>
		{#each used as p (p.id)}
			{@const list = listOf(p.id)}
			<datalist id="models-{p.id}">
				{#each list?.models ?? [] as info (info.id)}<option value={info.id}>{info.name && info.name !== info.id ? info.name : ''}</option>{/each}
			</datalist>
		{/each}
		<div class="row">
			<button type="button" class="button" onclick={add}>{t.addModel}</button>
		</div>
		<!-- Each provider's list is fetched only when asked: the call goes to the provider, with the key when it has one. -->
		<div class="lists">
			{#each used as p (p.id)}
				{@const list = listOf(p.id)}
				<div class="list">
					<button type="button" class="textbutton" onclick={() => modelLists.fetch(p.id, !!list)} disabled={modelLists.loading[p.id] || p.style !== 'chat' || p.signIn}>
						{modelLists.loading[p.id] ? t.asking({ provider: p.name }) : list ? t.fetchAgain({ provider: p.name }) : t.fetch({ provider: p.name })}
					</button>
					<span class="muted small">
						{#if modelLists.errors[p.id]}
							<span class="error">{modelLists.errors[p.id]}</span>
						{:else if list}
							{t.offered({ count: list.models.length })}
						{:else if p.signIn}
							{t.signIn({ provider: p.name })}
						{:else if p.style !== 'chat'}
							{t.noList({ provider: p.name })}
						{:else}
							{(p.hasKey ? t.asksWithKey : t.asks)({ provider: p.name })}
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</fieldset>

	<fieldset>
		<legend class="caps label">{t.order}</legend>
		<Segmented bind:value={rotationMode} options={orders} label={t.rotationOrder} />
	</fieldset>

	{#if sampling}
		<fieldset class="params">
			<legend class="caps label">{t.forEveryModel}</legend>
			<p class="muted small">{t.forEveryModelHint}</p>
			<div class="grid">
				<label class="field short"><span class="caps label">{t.temperature}</span><input type="text" inputmode="decimal" bind:value={temperature} /></label>
				<label class="field short"><span class="caps label">{t.maxTokens}</span><input type="text" inputmode="numeric" bind:value={maxTokens} /></label>
				<label class="field short"><span class="caps label">{t.timeoutSeconds}</span><input type="text" inputmode="numeric" bind:value={timeout} placeholder={isDialogue ? '' : t.asDialogue} /></label>
				<label class="field short"><span class="caps label">{t.topP}</span><input type="text" inputmode="decimal" bind:value={topP} /></label>
				<label class="field short"><span class="caps label">{t.topK}</span><input type="text" inputmode="numeric" bind:value={topK} /></label>
				<label class="field short"><span class="caps label">{t.frequencyPenalty}</span><input type="text" inputmode="decimal" bind:value={frequencyPenalty} /></label>
				<label class="field short"><span class="caps label">{t.presencePenalty}</span><input type="text" inputmode="decimal" bind:value={presencePenalty} /></label>
			</div>
			<label class="field">
				<span class="caps label">{t.stopSequences}</span>
				<textarea bind:value={stopSequences} rows="2" spellcheck="false"></textarea>
			</label>
		</fieldset>
	{:else}
		<fieldset>
			<legend class="caps label">{t.timeout}</legend>
			<label class="field short"><span class="caps label">{t.seconds}</span><input type="text" inputmode="numeric" bind:value={timeout} placeholder={t.asDialogue} /></label>
		</fieldset>
	{/if}

	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<div class="buttons">
		<button type="submit" class="button primary" disabled={saving}>{m.common.save}</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>{m.common.cancel}</button>
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
