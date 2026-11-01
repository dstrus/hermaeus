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
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { pluginHref } from '#lib/plugins/links.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { aboveHints, fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { contentHref } from '#lib/plugins/links.ts';
	import LimitsNotice from './LimitsNotice.svelte';
	import { href } from '#lib/nav.ts';
	import { layerName, recordsDependency, removeRecordCopy, saveVoiceEffect, switchRecord, switched } from './api.ts';
	import Overridden from './Overridden.svelte';

	let { list: loaded }: { list: VoiceEffectList } = $props();
	const msg = m.records.voiceEffects;

	let list = $derived(loaded);
	let save = $derived(page.params.save ?? '');
	let readOnly = $derived(live.gameRunning);
	const kindOf = (t: string): StageKind | undefined => list.kinds.find((k) => k.type === t);
	// A stage's name, what it does, and its settings' names and units, by
	// its type and their names: pitch_shift by stageKinds.pitchShift. A unit
	// with forms is a word that follows the number.
	type Unit = string | ((p: { count: number }) => string);
	type StageCopy = { label: string; what: string; params: { [name: string]: { label: string; unit?: Unit } } };
	const camel = (v: string) => v.replace(/_(.)/g, (_, c: string) => c.toUpperCase());
	const stageCopies: { [type: string]: StageCopy } = msg.stageKinds;
	const stageCopy = (type: string) => (Object.hasOwn(stageCopies, camel(type)) ? stageCopies[camel(type)] : undefined);
	const kindLabel = (type: string) => stageCopy(type)?.label ?? type;
	const paramLabel = (type: string, name: string) => stageCopy(type)?.params[camel(name)]?.label ?? name;
	function unitOf(type: string, name: string, v?: unknown): string {
		const unit = stageCopy(type)?.params[camel(name)]?.unit;
		return typeof unit === 'function' ? unit({ count: typeof v === 'number' ? v : NaN }) : (unit ?? '');
	}
	const layerOf = (e: VoiceEffect) => list.layers[e.providers[0]?.layer ?? -1];
	const mineOf = (e: VoiceEffect) => e.providers.find((p) => list.layers[p.layer]?.kind === 'overlay');

	let query = $state('');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.effects.filter((e) => !q || [e.name, e.id, ...e.stages.map((s) => kindLabel(s.type))].some((v) => v.toLowerCase().includes(q)));
	});
	let off = $derived(list.effects.filter((e) => !e.enabled).length);
	let yours = $derived(list.effects.filter((e) => layerOf(e)?.kind === 'overlay').length);

	const selection = new Selection(() => shown, (e) => e.id, { param: 'item' });
	let current = $derived(selection.current);

	// A number as a recipe's settings read: a true minus sign, and no
	// trailing zeros beyond what's needed.
	const number = (v: unknown) => (typeof v === 'number' ? String(Number(v.toFixed(3))).replace('-', '−') : String(v ?? '—'));
	function paramText(type: string, pm: StageParam, v: unknown): string {
		if (pm.suggest) return String(v);
		const unit = unitOf(type, pm.name, v);
		return unit ? `${number(v)} ${unit}` : number(v);
	}
	// A stage in a line: its settings, named unless there's only one.
	function stageText(s: Stage): string {
		const k = kindOf(s.type);
		if (!k) return Object.entries(s.params).map(([n, v]) => `${n} ${number(v)}`).join(' · ');
		if (k.params.length === 1) return paramText(s.type, k.params[0]!, s.params[k.params[0]!.name]);
		return k.params.map((pm) => `${paramLabel(s.type, pm.name).toLowerCase()} ${paramText(s.type, pm, s.params[pm.name])}`).join(' · ');
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

	function toggle(follow = false) {
		const e = current;
		if (!e) return;
		run(e.id, () => switchRecord(e.path, !e.enabled, follow), switched(!e.enabled, follow));
	}

	function remove() {
		const e = current;
		const mine = e && mineOf(e);
		if (!e || !mine) return;
		const below = e.providers.find((p) => p !== mine);
		const next = below ? msg.nextBelow({ layer: layerName(list.layers[below.layer]) }) : msg.nextNone;
		if (!confirm(m.records.removeCopyConfirm({ name: e.name, next }))) return;
		run(e.id, () => removeRecordCopy(e.path, { modified: mine.modified }), below ? m.records.removedCopyBelow : msg.removed({ name: e.name }));
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
					error = msg.needsNumber({ n: i + 1, stage: kindLabel(d.type).toLowerCase(), param: (pm ? paramLabel(d.type, pm.name) : n).toLowerCase() });
					return;
				}
				params[n] = x;
			}
			stages.push({ type: d.type, params });
		}
		const mine = mode === 'edit' && current ? mineOf(current) : undefined;
		const done = mode === 'new' ? msg.made({ name: draftName.trim() }) : mine ? msg.savedCopy : msg.savedAsCopy;
		if (await run(id, () => saveVoiceEffect(id, { modified: mine?.modified ?? 0, name: draftName, sample: draftSample, stages }), done)) {
			await setParams({ item: id, mode: null });
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		const mine = current && mineOf(current);
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !reading, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !reading, run: () => selection.move(-1) },
			{ key: 'n', shown: 'N', label: msg.newEffect, hidden: !reading, disabled: readOnly, run: startNew },
			{ key: 'e', shown: 'E', label: mine ? m.common.edit : m.records.editCopy, hidden: !reading || !current, disabled: readOnly, run: startEdit },
			{ key: 'o', shown: 'O', label: current?.enabled === false ? m.common.switchOn : m.common.switchOff, hidden: !reading || !current, disabled: readOnly || saving, run: () => toggle() },
			...deleteShortcuts({ label: m.records.removeCopy, hidden: !reading || !mine, disabled: readOnly || saving, run: remove }),
			{ key: 'Enter', label: m.common.save, command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: m.common.cancel, hidden: reading, run: () => leave('mode') }
		];
	});
</script>

{#if mode !== 'read'}
	<header class="editing">
		<LimitsNotice title={m.records.limitsTitle}>{msg.limits}</LimitsNotice>
		<h1 class="caps">{mode === 'new' ? msg.newVoiceEffect : current?.name}</h1>
		{#if mode === 'edit' && current}
			<p class="muted">
				{#if mineOf(current)}<Fill text={msg.editingOwn.text}>{#snippet yourCopy()}<b class="changed">{m.records.yourOwnCopy}</b>{/snippet}</Fill>{:else}{msg.startingFrom({ layer: layerName(layerOf(current)) })}{/if}
			</p>
		{/if}
		{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	</header>
	<form class="form" onsubmit={(e) => (e.preventDefault(), submit())}>
		<div class="fields">
			{#if mode === 'new'}
				<label>
					<span class="caps">{msg.id}</span>
					<input type="text" bind:value={draftId} placeholder="frost_atronach" autocomplete="off" spellcheck="false" />
					<span class="muted small">{msg.idHint}</span>
				</label>
			{/if}
			<label>
				<span class="caps">{msg.name}</span>
				<input type="text" bind:value={draftName} placeholder={msg.namePlaceholder} autocomplete="off" />
			</label>
			<label>
				<span class="caps">{msg.sampleLabel}</span>
				<input type="text" bind:value={draftSample} placeholder={msg.samplePlaceholder} autocomplete="off" spellcheck="false" />
				<span class="muted small">{msg.sampleHint}</span>
			</label>
		</div>

		<h2 class="caps">{msg.stages}</h2>
		<p class="muted small">{msg.order}</p>
		<ol class="stages">
			{#each draftStages as d, i (i)}
				{@const k = kindOf(d.type)}
				<li class="stage">
					<div class="stagehead">
						<span class="n muted">{i + 1}</span>
						<select value={d.type} onchange={(e) => retype(i, e.currentTarget.value)} aria-label={msg.stage({ n: i + 1 })}>
							{#each list.kinds as kk (kk.type)}<option value={kk.type}>{kindLabel(kk.type)}</option>{/each}
							{#if !k}<option value={d.type}>{d.type}</option>{/if}
						</select>
						<span class="muted small what">{k ? (stageCopy(k.type)?.what ?? '') : ''}</span>
						<span class="order">
							<button type="button" class="textbutton" onclick={() => move(i, -1)} disabled={i === 0} aria-label={msg.moveStageUp({ n: i + 1 })}>{m.common.up}</button>
							<button type="button" class="textbutton" onclick={() => move(i, 1)} disabled={i === draftStages.length - 1} aria-label={msg.moveStageDown({ n: i + 1 })}>{m.common.down}</button>
							<button type="button" class="textbutton" onclick={() => (draftStages = draftStages.filter((_, j) => j !== i))} aria-label={msg.removeStage({ n: i + 1 })}>{m.common.remove}</button>
						</span>
					</div>
					<div class="params">
						{#each k?.params ?? [] as pm (pm.name)}
							<label>
								<span class="caps">{paramLabel(d.type, pm.name)}</span>
								{#if pm.suggest}
									<input type="text" list="suggest-{pm.name}" bind:value={d.params[pm.name]} autocomplete="off" spellcheck="false" />
									<datalist id="suggest-{pm.name}">
										{#each pm.suggest as c (c)}<option value={c}></option>{/each}
									</datalist>
								{:else}
									<input type="number" step={pm.whole ? 1 : 'any'} value={d.params[pm.name]} oninput={(e) => (d.params[pm.name] = e.currentTarget.value)} />
									{#if unitOf(d.type, pm.name)}<span class="muted small">{unitOf(d.type, pm.name)}</span>{/if}
								{/if}
							</label>
						{/each}
					</div>
				</li>
			{:else}
				<li class="muted">{msg.noDraftStages}</li>
			{/each}
		</ol>
		<label class="add">
			<span class="sr">{msg.addStage}</span>
			<select bind:value={adding} onchange={() => addStage(adding)}>
				<option value="">{msg.addStageMenu}</option>
				{#each list.kinds as k (k.type)}<option value={k.type}>{kindLabel(k.type)}</option>{/each}
			</select>
		</label>

		<div class="buttons" use:aboveHints>
			<button type="submit" class="button primary" disabled={saving}>{m.common.save}</button>
			<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>{m.common.cancel}</button>
		</div>
	</form>
{:else}
	<div class="bar">
		<h1 class="caps">{msg.title}</h1>
		<span class="muted n" use:clipped>{msg.count({ count: list.effects.length })}{yours ? ` · ${m.records.yoursCount({ count: yours })}` : ''}{off ? ` · ${m.records.offCount({ count: off })}` : ''}</span>
		<button type="button" class="button" onclick={startNew} disabled={readOnly}>{msg.newEffect}</button>
		<input type="search" bind:value={query} placeholder={msg.search} aria-label={msg.search} />
	</div>
	<p class="muted lead">{msg.lead}</p>
	<LimitsNotice title={m.records.limitsTitle}>{msg.limits}</LimitsNotice>
	{#if readOnly}<p class="notice warn">{msg.readOnly}</p>{/if}

	<div class="split">
		<div class="list">
			{#if !list.effects.length}
				<p class="muted">{msg.none}</p>
			{:else if !shown.length}
				<p class="muted">{m.records.noMatch({ query })}</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th class="mark" aria-label={m.common.selected}></th>
							<th class="caps">{msg.effect}</th>
							<th class="caps chain">{msg.stages}</th>
							<th class="caps from">{m.records.from}</th>
						</tr>
					</thead>
					<tbody>
						{#each shown as e (e.id)}
							<tr aria-selected={selection.is(e)} class:off={!e.enabled} onclick={() => selection.select(e)}>
								<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
								<td use:clipped>{e.name}{#if !e.enabled}<span class="flag">{m.common.off}</span>{/if}{#if e.error}<span class="flag bad">{m.records.cantBeRead}</span>{/if}</td>
								<td class="chain muted" use:clipped>{e.stages.map((s) => kindLabel(s.type)).join(' › ') || msg.noChain}</td>
								<td class="from muted" use:clipped>{#if layerOf(e)?.kind === 'overlay'}<span class="changed">{m.records.yourEdits}</span>{:else}{layerName(layerOf(e))}{/if}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>

		<aside class="pane" use:fitToViewport aria-label={msg.voiceEffect}>
			{#if current}
				{@const e = current}
				{@const from = layerOf(e)}
				{#key e.id}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{e.name}</h2>
						<p class="muted small"><span class="id">{e.id}</span> · <a href={contentHref(save, e.path)}>{msg.itsFile}</a></p>
						<p class="small">
							<Fill
								text={(e.providers.length > 1 ? m.records.originInPlaceOf : m.records.origin).text}
								layers={e.providers
									.slice(1)
									.map((p) => layerName(list.layers[p.layer]))
									.join(m.records.thenJoin)}
								>{#snippet source()}{#if from?.kind === 'pack'}<Fill text={m.records.fromPlugin.text}>{#snippet plugin()}<a href={pluginHref(save, from.id)}>{from.source}</a>{/snippet}</Fill>{:else if from?.kind === 'overlay'}<span class="changed">{msg.own}</span>{:else}{m.records.fromDefaults}{/if}{/snippet}</Fill
							>
						</p>
						{#if e.error}<p class="notice error">{m.records.unreadable({ error: e.error })}</p>{/if}
						{#if !e.enabled}<p class="notice">{m.records.isOff}</p>{/if}
						<Overridden override={e.override} disabled={readOnly || saving} onfollow={() => toggle(true)} />
						{#if error}<p class="notice error" role="alert">{error}</p>{/if}
						{#if notice}<p class="notice" role="status">{notice}</p>{/if}
						<div class="buttons">
							<button type="button" class="button" onclick={() => toggle()} disabled={readOnly || saving}>{e.enabled ? m.common.switchOff : m.common.switchOn}</button>
							<button type="button" class="button" onclick={startEdit} disabled={readOnly || !!e.error}>{mineOf(e) ? m.common.edit : m.records.editCopy}</button>
							{#if mineOf(e)}<button type="button" class="button" onclick={remove} disabled={readOnly || saving}>{m.records.removeCopy}</button>{/if}
						</div>
						<h3 class="caps">{msg.stages}</h3>
						{#if e.stages.length}
							<ol class="chainlist">
								{#each e.stages as s, i (i)}
									<li><span class="caps label">{kindLabel(s.type)}</span><span class="small">{stageText(s)}</span></li>
								{/each}
							</ol>
						{:else}
							<p class="muted small">{msg.noStages}</p>
						{/if}
						<h3 class="caps">{msg.sample}</h3>
						<p class="small" class:muted={!e.sample}>{#if e.sample}<span class="id">{e.sample}</span>{:else}{msg.noSample}{/if}</p>
						<h3 class="caps">{msg.usedBy}</h3>
						<ul class="uses small">
							{#if e.id === 'narrator_effect'}<li><Fill text={msg.narrator.text}>{#snippet setting()}<a href="{href(save, 'speech')}?item=tts.enableNarratorAudioEffect">{msg.narratorSetting}</a>{/snippet}</Fill></li>{/if}
							{#if e.id === 'player_thought_effect'}<li><Fill text={msg.thoughts.text}>{#snippet setting()}<a href="{href(save, 'speech')}?item=tts.enablePlayerThoughtAudioEffect">{msg.thoughtsSetting}</a>{/snippet}</Fill></li>{/if}
							{#each e.triggers as t (t.path)}
								<li><Fill text={(t.activates ? msg.switchedOnBy : msg.switchedOffBy).text}>{#snippet trigger()}<a href={href(save, 'triggers', `?item=${encodeURIComponent(t.path)}`)}>{t.name}</a>{/snippet}</Fill><span class="muted"> · {t.source}{t.enabled ? '' : ` · ${m.common.off}`}</span></li>
							{/each}
							{#if !e.triggers.length && e.id !== 'narrator_effect' && e.id !== 'player_thought_effect'}<li class="muted">{msg.noTrigger}</li>{/if}
						</ul>
						<p class="muted small"><Fill text={msg.script.text}>{#snippet id()}<span class="id">{e.id}</span>{/snippet}</Fill></p>
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
	.uses {
		list-style: none;
		padding: 0;
		margin: 0 0 var(--space-2);
	}
	.uses li {
		padding: var(--space-1) 0;
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
