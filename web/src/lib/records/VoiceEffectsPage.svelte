<!-- Voice effects: recipes that change how a voice sounds, each a chain of
     stages applied in order. A recipe from SkyrimNet or a plugin is edited
     as the player's own copy; a new one is the player's from the start. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Stage, StageKind, StageParam, VoiceEffect, VoiceEffectList } from '#lib/api/types/records.ts';
	import { plural } from '#lib/format.ts';
	import { pluginHref } from '#lib/plugins/links.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { aboveHints, fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { contentHref } from '#lib/plugins/links.ts';
	import LimitsNotice from './LimitsNotice.svelte';
	import { layerName, recordsDependency, removeRecordCopy, saveVoiceEffect, switchRecord } from './api.ts';

	let { list: loaded }: { list: VoiceEffectList } = $props();

	// SkyrimNet limits each stage's settings, and what a sample can be, in
	// ways that aren't published.
	const limits = 'SkyrimNet limits what each stage’s settings and a voice sample can be, in ways that aren’t published, so Hermaeus can’t show those limits or check them. An effect outside them may not work in the game. SkyrimNet’s own dashboard shows the limits as you edit.';

	let list = $derived(loaded);
	let save = $derived(page.params.save ?? '');
	let readOnly = $derived(live.gameRunning);
	const kindOf = (t: string): StageKind | undefined => list.kinds.find((k) => k.type === t);
	const layerOf = (e: VoiceEffect) => list.layers[e.providers[0]?.layer ?? -1];
	const mineOf = (e: VoiceEffect) => e.providers.find((p) => list.layers[p.layer]?.kind === 'overlay');

	let query = $state('');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.effects.filter((e) => !q || [e.name, e.id, ...e.stages.map((s) => kindOf(s.type)?.label ?? s.type)].some((v) => v.toLowerCase().includes(q)));
	});
	let off = $derived(list.effects.filter((e) => !e.enabled).length);
	let yours = $derived(list.effects.filter((e) => layerOf(e)?.kind === 'overlay').length);

	const selection = new Selection(() => shown, (e) => e.id, { param: 'item' });
	let current = $derived(selection.current);

	// A number as a recipe's settings read: a true minus sign, and no
	// trailing zeros beyond what's needed.
	const number = (v: unknown) => (typeof v === 'number' ? String(Number(v.toFixed(3))).replace('-', '−') : String(v ?? '—'));
	function paramText(pm: StageParam, v: unknown): string {
		if (pm.suggest) return String(v);
		const unit = pm.unit ? ` ${v === 1 && pm.unit.endsWith('s') ? pm.unit.slice(0, -1) : pm.unit}` : '';
		return `${number(v)}${unit}`;
	}
	// A stage in a line: its settings, named unless there's only one.
	function stageText(s: Stage): string {
		const k = kindOf(s.type);
		if (!k) return Object.entries(s.params).map(([n, v]) => `${n} ${number(v)}`).join(' · ');
		if (k.params.length === 1) return paramText(k.params[0]!, s.params[k.params[0]!.name]);
		return k.params.map((pm) => `${pm.label.toLowerCase()} ${paramText(pm, s.params[pm.name])}`).join(' · ');
	}

	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let noticeFor = $state('');
	$effect(() => {
		if ((current?.id ?? '') !== noticeFor) {
			error = '';
			notice = '';
		}
	});

	async function run(id: string, change: () => Promise<unknown>, done: string): Promise<boolean> {
		if (saving || readOnly) return false;
		saving = true;
		error = '';
		notice = '';
		noticeFor = id;
		try {
			await change();
			notice = done;
			await invalidate(recordsDependency);
			return true;
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
			return false;
		} finally {
			saving = false;
		}
	}

	function toggle() {
		const e = current;
		if (!e) return;
		const made = mineOf(e) ? '' : ' Hermaeus made your own copy of it, with only its switch changed.';
		run(e.id, () => switchRecord(e.path, !e.enabled), e.enabled ? `Switched off.${made} SkyrimNet stops using it from the next time it starts.` : `Switched on.${made}`);
	}

	function remove() {
		const e = current;
		const mine = e && mineOf(e);
		if (!e || !mine) return;
		const below = e.providers.find((p) => p !== mine);
		const next = below ? `the one from ${layerName(list.layers[below.layer])} applies again` : 'the effect is gone, since only you had it';
		if (!confirm(`Remove your copy of ${e.name}? A backup is kept, and ${next}.`)) return;
		run(e.id, () => removeRecordCopy(e.path, { modified: mine.modified }), below ? 'Removed your copy. The one below applies again.' : `Removed ${e.name}.`);
	}

	// Editing (?mode=edit) and making a new effect (?mode=new) are states of
	// the address, so Back cancels them.
	let mode = $derived(param('mode') === 'edit' && current ? 'edit' : param('mode') === 'new' ? 'new' : 'read');
	type DraftStage = { type: string; params: Record<string, string> };
	let draftId = $state('');
	let draftName = $state('');
	let draftSample = $state('');
	let draftStages = $state<DraftStage[]>([]);
	let adding = $state('');

	const draftOf = (s: Stage): DraftStage => ({ type: s.type, params: Object.fromEntries(Object.entries(s.params).map(([n, v]) => [n, String(v ?? '')])) });
	function fresh(type: string): DraftStage {
		const k = kindOf(type);
		return { type, params: Object.fromEntries((k?.params ?? []).map((pm) => [pm.name, String(pm.default ?? '')])) };
	}
	$effect(() => {
		if (mode === 'read') return;
		untrack(() => {
			const e = mode === 'edit' ? current : undefined;
			draftId = e?.id ?? '';
			draftName = e?.name ?? '';
			draftSample = e?.sample ?? '';
			draftStages = e ? e.stages.map(draftOf) : [];
			adding = '';
			error = '';
			notice = '';
		});
	});

	function startEdit() {
		if (current && !readOnly) setParams({ item: current.id, mode: 'edit' });
	}
	function startNew() {
		if (!readOnly) setParams({ mode: 'new' });
	}
	function move(i: number, by: number) {
		const j = i + by;
		if (j < 0 || j >= draftStages.length) return;
		const next = [...draftStages];
		[next[i], next[j]] = [next[j]!, next[i]!];
		draftStages = next;
	}
	function addStage(type: string) {
		if (!type) return;
		draftStages = [...draftStages, fresh(type)];
		adding = '';
	}
	// Changing a stage's kind starts it from that kind's defaults.
	function retype(i: number, type: string) {
		draftStages[i] = fresh(type);
	}

	async function submit() {
		const id = mode === 'new' ? draftId.trim() : (current?.id ?? '');
		const stages: Stage[] = [];
		for (const [i, d] of draftStages.entries()) {
			const k = kindOf(d.type);
			const params: Record<string, unknown> = {};
			for (const [n, v] of Object.entries(d.params)) {
				const pm = k?.params.find((p) => p.name === n);
				// Text settings, and an unknown stage's settings that aren't numbers, stay text.
				if (pm?.suggest || (!pm && v.trim() !== '' && Number.isNaN(Number(v)))) {
					params[n] = v;
					continue;
				}
				const x = Number(v);
				if (v.trim() === '' || Number.isNaN(x)) {
					error = `Stage ${i + 1} (${(k?.label ?? d.type).toLowerCase()}): ${(pm?.label ?? n).toLowerCase()} needs a number.`;
					return;
				}
				params[n] = x;
			}
			stages.push({ type: d.type, params });
		}
		const mine = mode === 'edit' && current ? mineOf(current) : undefined;
		const done = mode === 'new' ? `Made ${draftName.trim()}.` : mine ? 'Saved your copy.' : 'Saved as your own copy, in place of the one below.';
		if (await run(id, () => saveVoiceEffect(id, { modified: mine?.modified ?? 0, name: draftName, sample: draftSample, stages }), `${done} SkyrimNet uses it from the next time it starts.`)) {
			await setParams({ item: id, mode: null });
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		const mine = current && mineOf(current);
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !reading, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !reading, run: () => selection.move(-1) },
			{ key: 'n', shown: 'N', label: 'New effect', hidden: !reading, disabled: readOnly, run: startNew },
			{ key: 'e', shown: 'E', label: mine ? 'Edit' : 'Edit as your copy', hidden: !reading || !current, disabled: readOnly, run: startEdit },
			{ key: 'o', shown: 'O', label: current?.enabled === false ? 'Switch on' : 'Switch off', hidden: !reading || !current, disabled: readOnly || saving, run: toggle },
			...deleteShortcuts({ label: 'Remove your copy', hidden: !reading || !mine, disabled: readOnly || saving, run: remove }),
			{ key: 'Enter', label: 'Save', command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: 'Cancel', hidden: reading, run: () => leave('mode') }
		];
	});
</script>

{#if mode !== 'read'}
	<header class="editing">
		<LimitsNotice title="Limits Hermaeus can’t check">{limits}</LimitsNotice>
		<h1 class="caps">{mode === 'new' ? 'New voice effect' : current?.name}</h1>
		{#if mode === 'edit' && current}
			<p class="muted">
				{#if mineOf(current)}Editing <b class="changed">your own copy</b>, which SkyrimNet uses in place of any other.{:else}Starting from the one from {layerName(layerOf(current))}. Saving makes your own copy, which replaces it.{/if}
			</p>
		{/if}
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	</header>
	<form class="form" onsubmit={(e) => (e.preventDefault(), submit())}>
		<div class="fields">
			{#if mode === 'new'}
				<label>
					<span class="caps">Id</span>
					<input type="text" bind:value={draftId} placeholder="frost_atronach" autocomplete="off" spellcheck="false" />
					<span class="muted small">Its file’s name, and how other records refer to it. Lowercase letters, digits, and underscores; it can’t change later.</span>
				</label>
			{/if}
			<label>
				<span class="caps">Name</span>
				<input type="text" bind:value={draftName} placeholder="Frost atronach" autocomplete="off" />
			</label>
			<label>
				<span class="caps">Sample</span>
				<input type="text" bind:value={draftSample} placeholder="None" autocomplete="off" spellcheck="false" />
				<span class="muted small">A recording that replaces the speaker’s own voice sample for cloning while the effect is on. Leave it empty to keep theirs.</span>
			</label>
		</div>

		<h2 class="caps">Stages</h2>
		<p class="muted small">Applied to the voice in order, top to bottom.</p>
		<ol class="stages">
			{#each draftStages as d, i (i)}
				{@const k = kindOf(d.type)}
				<li class="stage">
					<div class="stagehead">
						<span class="n muted">{i + 1}</span>
						<select value={d.type} onchange={(e) => retype(i, e.currentTarget.value)} aria-label="Stage {i + 1}">
							{#each list.kinds as kk (kk.type)}<option value={kk.type}>{kk.label}</option>{/each}
							{#if !k}<option value={d.type}>{d.type}</option>{/if}
						</select>
						<span class="muted small what">{k?.what ?? ''}</span>
						<span class="order">
							<button type="button" class="textbutton" onclick={() => move(i, -1)} disabled={i === 0} aria-label="Move stage {i + 1} up">Up</button>
							<button type="button" class="textbutton" onclick={() => move(i, 1)} disabled={i === draftStages.length - 1} aria-label="Move stage {i + 1} down">Down</button>
							<button type="button" class="textbutton" onclick={() => (draftStages = draftStages.filter((_, j) => j !== i))} aria-label="Remove stage {i + 1}">Remove</button>
						</span>
					</div>
					<div class="params">
						{#each k?.params ?? [] as pm (pm.name)}
							<label>
								<span class="caps">{pm.label}</span>
								{#if pm.suggest}
									<input type="text" list="suggest-{pm.name}" bind:value={d.params[pm.name]} autocomplete="off" spellcheck="false" />
									<datalist id="suggest-{pm.name}">
										{#each pm.suggest as c (c)}<option value={c}></option>{/each}
									</datalist>
								{:else}
									<input type="number" step={pm.whole ? 1 : 'any'} value={d.params[pm.name]} oninput={(e) => (d.params[pm.name] = e.currentTarget.value)} />
									{#if pm.unit}<span class="muted small">{pm.unit}</span>{/if}
								{/if}
							</label>
						{/each}
					</div>
				</li>
			{:else}
				<li class="muted">No stages: the voice is left as it is.</li>
			{/each}
		</ol>
		<label class="add">
			<span class="sr">Add a stage</span>
			<select bind:value={adding} onchange={() => addStage(adding)}>
				<option value="">Add a stage…</option>
				{#each list.kinds as k (k.type)}<option value={k.type}>{k.label}</option>{/each}
			</select>
		</label>

		<div class="buttons" use:aboveHints>
			<button type="submit" class="button primary" disabled={saving}>Save</button>
			<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
		</div>
	</form>
{:else}
	<div class="bar">
		<h1 class="caps">Voice effects</h1>
		<span class="muted n" use:clipped>{plural(list.effects.length, 'effect', 'effects')}{yours ? ` · ${yours} yours` : ''}{off ? ` · ${off} off` : ''}</span>
		<button type="button" class="button" onclick={startNew} disabled={readOnly}>New effect</button>
		<input type="search" bind:value={query} placeholder="Search effects" aria-label="Search effects" />
	</div>
	<p class="muted lead">Recipes that change how a voice sounds, such as a draugr’s or a ghost’s, applied to speech one stage at a time. Each applies wherever SkyrimNet assigns it, such as to your thoughts or the narrator.</p>
	<LimitsNotice title="Limits Hermaeus can’t check">{limits}</LimitsNotice>
	{#if readOnly}<p class="notice warn">Skyrim is running, so effects can only be read.</p>{/if}

	<div class="split">
		<div class="list">
			{#if !list.effects.length}
				<p class="muted">No voice effects in SkyrimNet’s defaults, your plugins, or your own edits.</p>
			{:else if !shown.length}
				<p class="muted">Nothing matches “{query}”.</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th class="mark" aria-label="Selected"></th>
							<th class="caps">Effect</th>
							<th class="caps chain">Stages</th>
							<th class="caps from">From</th>
						</tr>
					</thead>
					<tbody>
						{#each shown as e (e.id)}
							<tr aria-selected={selection.is(e)} class:off={!e.enabled} onclick={() => selection.select(e)}>
								<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
								<td use:clipped>{e.name}{#if !e.enabled}<span class="flag">off</span>{/if}{#if e.error}<span class="flag bad">can’t be read</span>{/if}</td>
								<td class="chain muted" use:clipped>{e.stages.map((s) => kindOf(s.type)?.label ?? s.type).join(' › ') || 'None'}</td>
								<td class="from muted" use:clipped>{#if layerOf(e)?.kind === 'overlay'}<span class="changed">Your edits</span>{:else}{layerName(layerOf(e))}{/if}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>

		<aside class="pane" use:fitToViewport aria-label="Voice effect">
			{#if current}
				{@const e = current}
				{@const from = layerOf(e)}
				{#key e.id}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{e.name}</h2>
						<p class="muted small"><span class="id">{e.id}</span> · <a href={contentHref(save, e.path)}>its file</a></p>
						<p class="small">
							{#if from?.kind === 'pack'}From <a href={pluginHref(save, from.id)}>{from.source}</a>{:else if from?.kind === 'overlay'}<span class="changed">Your own</span>{:else}From SkyrimNet’s defaults{/if}{#if e.providers.length > 1}, in place of {e.providers
									.slice(1)
									.map((p) => layerName(list.layers[p.layer]))
									.join(', then ')}{/if}.
						</p>
						{#if e.error}<p class="notice error">SkyrimNet can’t read this file: {e.error}</p>{/if}
						{#if !e.enabled}<p class="notice">Off: SkyrimNet doesn’t use it.</p>{/if}
						{#if error}<p class="notice error" role="alert">{error}</p>{/if}
						{#if notice}<p class="notice" role="status">{notice}</p>{/if}
						<div class="buttons">
							<button type="button" class="button" onclick={toggle} disabled={readOnly || saving}>{e.enabled ? 'Switch off' : 'Switch on'}</button>
							<button type="button" class="button" onclick={startEdit} disabled={readOnly || !!e.error}>{mineOf(e) ? 'Edit' : 'Edit as your copy'}</button>
							{#if mineOf(e)}<button type="button" class="button" onclick={remove} disabled={readOnly || saving}>Remove your copy</button>{/if}
						</div>
						<h3 class="caps">Stages</h3>
						{#if e.stages.length}
							<ol class="chainlist">
								{#each e.stages as s, i (i)}
									<li><span class="caps label">{kindOf(s.type)?.label ?? s.type}</span><span class="small">{stageText(s)}</span></li>
								{/each}
							</ol>
						{:else}
							<p class="muted small">None: the voice is left as it is.</p>
						{/if}
						<h3 class="caps">Sample</h3>
						<p class="small" class:muted={!e.sample}>{#if e.sample}<span class="id">{e.sample}</span>{:else}None: the speaker’s own sample is used.{/if}</p>
					</div>
				{/key}
			{/if}
		</aside>
	</div>
{/if}

<style>
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
		margin-right: calc(var(--pane) + var(--space-5));
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		position: relative;
		top: 0.21em;
	}
	.bar .n {
		white-space: nowrap;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		margin-right: auto;
	}
	.bar h1,
	.bar .button {
		flex: none;
		white-space: nowrap;
	}
	.bar input {
		flex: 0 1 14rem;
		min-width: 8rem;
	}
	.lead {
		max-width: 65ch;
		margin-bottom: var(--space-4);
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
		overflow-wrap: anywhere;
	}
	.changed {
		font-weight: 400;
		color: var(--changed);
	}
	.bar,
	.split {
		--pane: clamp(20rem, 30vw, 28rem);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) var(--pane);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
	}
	th {
		padding: var(--space-2) var(--space-3);
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
	}
	th.chain {
		width: 45%;
	}
	th.from {
		width: 22%;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	tbody tr {
		cursor: default;
	}
	tbody tr[aria-selected]:not([aria-selected='true']):hover td {
		background: var(--glass);
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	tr.off td:nth-child(2) {
		color: var(--muted);
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	/* A flag is a word of metadata, never a button. */
	.flag {
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.flag::before {
		content: '· ';
	}
	.flag.bad {
		color: var(--danger);
	}
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
		overflow-wrap: anywhere;
	}
	.pane h3 {
		font-size: var(--step--1);
		color: var(--muted);
		margin: var(--space-4) 0 var(--space-2);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.pane .small {
		margin-bottom: var(--space-2);
		overflow-wrap: anywhere;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin: var(--space-3) 0;
	}
	/* The chain, read top to bottom: each stage's kind, then its settings. */
	.chainlist {
		list-style: none;
		padding: 0;
		margin: 0;
		counter-reset: stage;
	}
	.chainlist li {
		display: grid;
		grid-template-columns: 7.5rem minmax(0, 1fr);
		gap: var(--space-3);
		padding: var(--space-1) 0;
		align-items: baseline;
	}
	.chainlist .label {
		font-size: var(--step--1);
	}
	.pane > div > p.small:last-child {
		padding-bottom: var(--space-5);
	}
	/* The editor: the recipe's fields, then its stages as numbered cards. */
	.editing {
		margin-bottom: var(--space-4);
	}
	.editing h1 {
		top: 0;
		margin-bottom: var(--space-2);
	}
	.form {
		max-width: 56rem;
	}
	.fields {
		display: grid;
		gap: var(--space-3);
		max-width: 36rem;
		margin-bottom: var(--space-5);
	}
	.fields label,
	.params label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.fields .caps,
	.params .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.form h2 {
		font-size: var(--step-1);
		font-weight: 600;
	}
	.stages {
		list-style: none;
		padding: 0;
		margin: var(--space-3) 0;
	}
	.stage {
		padding: var(--space-3) 0 var(--space-4);
		border-top: 1px solid var(--rule);
	}
	.stagehead {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-2);
	}
	.stagehead .n {
		width: 1.25rem;
		font-variant-numeric: tabular-nums;
	}
	.stagehead .what {
		flex: 1;
		min-width: 0;
	}
	.order {
		display: flex;
		gap: var(--space-3);
		font-size: var(--step--1);
	}
	.order .textbutton:disabled {
		color: var(--muted);
		text-decoration: none;
		cursor: default;
	}
	.params {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
		gap: var(--space-3);
		padding-left: calc(1.25rem + var(--space-3));
	}
	.params input {
		width: 100%;
	}
	.add select {
		min-width: 14rem;
	}
	.form > .buttons {
		position: sticky;
		bottom: 0;
		margin: var(--space-4) 0 0;
		padding: var(--space-3) 0 var(--space-5);
		background: var(--ground);
		border-top: 1px solid var(--rule);
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
	}
	@media (max-width: 60rem) {
		.bar {
			margin-right: 0;
			flex-wrap: wrap;
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
