<!-- Universal Translator: the switch, the rules in priority order, the
     pattern for everyone else, and the plugin filter. Every change saves at
     once, with a backup; which rule is open, and what the pane is doing, is
     in the address. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Catalog } from '#lib/api/types/bios.ts';
import type { Character } from '#lib/api/types/characters.ts';
	import type { Rule, Translator, TranslatorEdit } from '#lib/api/types/settings.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { saveSettings, settingsDependency, settingsHref, togglesDependency } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { blankRule, editOf, entityLabel, importRules, modelHref, npcChoices, npcOf, pluginChoices, saveTranslator, translatorDependency } from './api.ts';
	import { templateName } from '#lib/bios/api.ts';
	import RuleForm from './RuleForm.svelte';

	const t = m.translator;

	let { save, translator, characters, catalog }: { save: string; translator: Translator; characters: Character[]; catalog: Catalog } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(translator.found && !readOnly);
	let npcs = $derived(npcChoices(characters, catalog));
	let plugins = $derived(pluginChoices(characters));
	// An NPC's rule is named for the character: as SkyrimNet recorded them
	// in this playthrough, or as their bio template spells it. The template
	// shows beside the name, since the rule is keyed by it.
	let ruleTemplate = (r: Rule) => r.name || r.entityId;
	let npcName = (r: Rule) => npcOf(r, npcs)?.name ?? templateName(ruleTemplate(r));
	let here = (r: Rule) => npcOf(r, npcs)?.here ?? false;
	let label = (r: Rule) => (r.entityType === 'npc' ? npcName(r) : r.name || r.entityId);

	// The global pattern is listed last as a rule of its own, under its key.
	const GLOBAL = 'global';
	type Row = { key: string; rule?: Rule; at: number };
	let rows = $derived<Row[]>([...translator.rules.map((rule, at) => ({ key: rule.id, rule, at })), { key: GLOBAL, at: translator.rules.length }]);
	const selection = new Selection(() => rows, (r) => r.key, { param: 'item' });
	let current = $derived(selection.current);
	let isGlobal = $derived(current?.key === GLOBAL);

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	// What the pane is doing, kept in the address so each state has its own
	// URL and Back closes a form.
	type Mode = 'view' | 'edit' | 'new' | 'delete' | 'filter' | 'import';
	let mode = $derived.by((): Mode => {
		const asked = param('mode');
		if (!translator.found) return 'view';
		if ((asked === 'edit' || asked === 'delete') && current && !(asked === 'delete' && isGlobal)) return asked;
		if (asked === 'new' || asked === 'filter' || asked === 'import') return asked;
		return 'view';
	});
	const open = (to: Mode) => setParams({ mode: to });
	const close = (opened?: string) => (opened ? setParams({ item: opened, mode: null }, 'replace') : leave('mode'));
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	// commit saves the configuration with one change made to it.
	async function commit(change: (e: TranslatorEdit) => string | { note: string; opened: string }, reload = true) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const e = editOf(translator);
			const done = change(e);
			await saveTranslator(e);
			notice = typeof done === 'string' ? done : done.note;
			await close(typeof done === 'string' ? undefined : done.opened);
			if (reload) await Promise.all([invalidate(translatorDependency), invalidate(togglesDependency)]);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function saveRule(rule: Rule) {
		commit((e) => {
			if (isGlobal && mode === 'edit') {
				e.global = rule.pattern;
				return t.savedGlobal;
			}
			if (mode === 'edit') {
				e.rules[current!.at] = rule;
				return t.savedRule({ name: label(rule) });
			}
			// A new rule goes first: it applies before the rest, as the dashboard adds them.
			const id = `rule_${Date.now()}`;
			e.rules.unshift({ ...rule, id });
			return { note: t.added({ name: label(rule) }), opened: id };
		});
	}

	function move(by: number) {
		const r = current;
		if (!r?.rule || !can) return;
		const to = r.at + by;
		if (to < 0 || to >= translator.rules.length) return;
		commit((e) => {
			const [rule] = e.rules.splice(r.at, 1);
			e.rules.splice(to, 0, rule!);
			return by < 0 ? t.movedUp : t.movedDown;
		});
	}

	function toggleRule() {
		const r = current?.rule;
		if (!r || !can) return;
		commit((e) => {
			e.rules[current!.at]!.enabled = !r.enabled;
			return r.enabled ? t.switchedOff : t.switchedOn;
		});
	}

	function remove() {
		const r = current?.rule;
		if (!r) return;
		commit((e) => {
			e.rules.splice(current!.at, 1);
			return t.deleted({ name: label(r) });
		});
	}

	// The plugin filter, edited as a mode and a list of plugin names.
	let filterMode = $state('blacklist');
	let filterText = $state('');
	$effect(() => {
		if (mode !== 'filter') return;
		filterMode = translator.filterMode;
		filterText = translator.filterPlugins.join('\n');
	});
	let filterList = $derived([...new Set(filterText.split('\n').map((l) => l.trim()).filter(Boolean))]);
	function addPlugin(p: string) {
		if (!filterList.includes(p)) filterText = [...filterList, p].join('\n');
	}
	function saveFilter(e: SubmitEvent) {
		e.preventDefault();
		commit((edit) => {
			edit.filterMode = filterMode;
			edit.filterPlugins = filterList;
			return filterList.length ? t.filterSaved({ plugins: t.pluginCount({ count: filterList.length }) }) : t.filterCleared;
		});
	}

	// Switching the translator, and the SkyrimNet setting it depends on.
	function switchTranslator() {
		commit((e) => {
			e.enabled = !translator.enabled;
			return e.enabled ? t.turnedOn : t.turnedOff;
		});
	}
	async function switchVoices() {
		const dep = translator.voiceSilentNPCs;
		if (saving || !dep.available || readOnly) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveSettings('game', { modified: dep.modified, changes: [{ key: dep.key, value: !dep.on }] });
			notice = dep.on ? t.voicesOff : t.voicesOn;
			await Promise.all([invalidate(translatorDependency), invalidate(settingsDependency)]);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	// Importing rules from a YAML file.
	let file = $state<HTMLInputElement>();
	async function runImport(e: SubmitEvent) {
		e.preventDefault();
		const f = file?.files?.[0];
		if (!f || saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const res = await importRules({ modified: translator.modified, yaml: await f.text() });
			const parts = [];
			if (res.added) parts.push(t.rulesAdded({ count: res.added }));
			if (res.updated) parts.push(t.rulesUpdated({ count: res.updated }));
			notice = t.imported({ file: f.name, parts: parts.join(', ') || t.nothingNew });
			await close();
			await invalidate(translatorDependency);
		} catch (err) {
			error = err instanceof ApiError ? err.message : String(err);
		} finally {
			saving = false;
		}
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const r = current;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: m.common.edit, hidden: !viewing, disabled: !can || !r, run: () => open('edit') },
			{ key: 'n', shown: 'N', label: t.newRule, hidden: !viewing, disabled: !can, run: () => open('new') },
			{ key: 'u', shown: 'U', label: m.common.moveUp, hidden: !viewing, disabled: !can || !r?.rule || r.at === 0, run: () => move(-1) },
			{ key: 'd', shown: 'D', label: m.common.moveDown, hidden: !viewing, disabled: !can || !r?.rule || r.at >= translator.rules.length - 1, run: () => move(1) },
			{ key: 'o', shown: 'O', label: r?.rule?.enabled === false ? m.common.switchOn : m.common.switchOff, hidden: !viewing, disabled: !can || !r?.rule, run: toggleRule },
			{ key: 'f', shown: 'F', label: t.filter, hidden: !viewing, disabled: !can, run: () => open('filter') },
			{ key: 'i', shown: 'I', label: t.import, hidden: !viewing, disabled: !can, run: () => open('import') },
			...deleteShortcuts({ label: m.common.delete, hidden: !viewing, disabled: !can || !r?.rule, run: () => open('delete') }),
			{ key: 'Enter', label: m.common.save, command: true, hidden: !(mode === 'edit' || mode === 'new' || mode === 'filter'), run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit() },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => close() }])
		];
	});

	let filterSummary = $derived(
		translator.filterPlugins.length === 0
			? t.noFilter
			: (translator.filterMode === 'whitelist' ? t.filterOnly : t.filterAllBut)({ plugins: t.pluginCount({ count: translator.filterPlugins.length }), list: translator.filterPlugins.join(', ') })
	);
</script>

<div class="bar">
	<h1 class="caps">{t.title}</h1>
	<span class="muted n" use:clipped>
		{#if !translator.found}
			{m.common.notSetUp}
		{:else if translator.enabled}
			{m.common.isOn} · {t.rules({ count: translator.rules.length })}{translator.global ? ` · ${t.globalPattern}` : ''}
		{:else}
			{m.common.isOff}
		{/if}
	</span>
</div>
<p class="muted lead">{t.lead}</p>

<div class="switches" role="group" aria-label={t.switches}>
	<button type="button" role="switch" aria-checked={translator.enabled} disabled={!can || saving} onclick={switchTranslator}>
		<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
		<span class="caps">{t.title}</span>
	</button>
	<button type="button" role="switch" aria-checked={translator.voiceSilentNPCs.on} disabled={!translator.voiceSilentNPCs.available || readOnly || saving} onclick={switchVoices} title={t.voiceSilentTitle({ key: translator.voiceSilentNPCs.key })}>
		<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
		<span class="caps">{t.voiceSilent}</span>
	</button>
	<span class="muted small">{t.speaksThrough}</span>
</div>
{#if translator.found && translator.enabled && !translator.voiceSilentNPCs.on}
	<p class="notice warn">{t.unheard}</p>
{/if}
{#if !translator.found}
	<p class="notice"><Fill text={m.common.notWrittenYet.text}>{#snippet path()}<span class="id">{translator.path}</span>{/snippet}</Fill></p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label={m.common.selected}></th>
					<th class="caps at">{t.order}</th>
					<th class="caps">{t.for}</th>
					<th class="caps text">{t.pattern}</th>
				</tr>
			</thead>
			<tbody>
				{#each rows as row (row.key)}
					{@const r = row.rule}
					<tr aria-selected={selection.is(row)} class:off={r && !r.enabled} class:global={!r} onclick={() => mode === 'view' && selection.select(row)} ondblclick={() => can && mode === 'view' && open('edit')}>
						<td class="mark">{#if selection.is(row)}<Diamond />{/if}</td>
						<td class="at n">{r ? row.at + 1 : '∞'}</td>
						<td class="who">
							{#if r}
								<span class="kind caps">{entityLabel(r.entityType)}</span>
								<span class="name">{label(r)}</span>
								{#if r.entityType === 'npc'}<span class="muted id"> · {ruleTemplate(r)}</span>{/if}
							{:else}
								<span class="kind caps">{t.everyoneElse}</span>
							{/if}
						</td>
						<td class="text">
							<span class="line">{(r ? r.pattern : translator.global) || ''}</span>
							{#if r && !r.pattern}<span class="flag">{t.leftAsWritten}</span>{/if}
							{#if !r && !translator.global}<span class="flag">{t.asWritten}</span>{/if}
							{#if r && !r.enabled}<span class="flag">{m.common.off}</span>{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">{filterSummary}</p>
		<p class="muted small"><Fill text={t.modelNote.text}>{#snippet models()}<a href={modelHref(save)}>{m.models.title}</a>{/snippet}</Fill></p>
	</div>

	<aside class="pane" use:fitToViewport aria-label={t.rule}>
		{#if mode === 'new'}
			<h2 class="caps">{t.newRule}</h2>
			<p class="muted small">{t.newRuleHint}</p>
			<RuleForm rule={blankRule()} {npcs} {saving} onsave={saveRule} oncancel={() => close()} />
		{:else if mode === 'filter'}
			<h2 class="caps">{t.filter}</h2>
			<p class="muted small">{t.filterHint}</p>
			<form class="filter" onsubmit={saveFilter}>
				<Segmented bind:value={filterMode} options={[{ value: 'blacklist', label: t.allButThese }, { value: 'whitelist', label: t.onlyThese }]} label={t.howApplies} />
				<label class="field">
					<span class="caps label">{t.pluginsPerLine}</span>
					<textarea bind:value={filterText} rows="6" placeholder="Skyrim.esm" spellcheck="false"></textarea>
				</label>
				{#if plugins.length}
					<div class="choices">
						<span class="caps label">{t.inPlaythrough}</span>
						{#each plugins as p (p)}
							<button type="button" class="textbutton" disabled={filterList.includes(p)} onclick={() => addPlugin(p)}>{p}</button>
						{/each}
					</div>
				{/if}
				<div class="buttons">
					<button type="submit" class="button primary" disabled={saving}>{m.common.save}</button>
					<button type="button" class="button" onclick={() => close()} disabled={saving}>{m.common.cancel}</button>
				</div>
			</form>
		{:else if mode === 'import'}
			<h2 class="caps">{t.importTitle}</h2>
			<p class="muted small">{t.importHint}</p>
			<form class="import" onsubmit={runImport}>
				<input bind:this={file} type="file" accept=".yaml,.yml,text/yaml" aria-label={t.yamlFile} />
				<div class="buttons">
					<button type="submit" class="button primary" disabled={saving}>{t.import}</button>
					<button type="button" class="button" onclick={() => close()} disabled={saving}>{m.common.cancel}</button>
				</div>
			</form>
		{:else if current}
			{#key `${current.key}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						{#if isGlobal}
							<h2 class="caps">{t.everyoneElse}</h2>
							<RuleForm rule={{ ...blankRule(), entityType: 'global', entityId: 'global', pattern: translator.global }} global {saving} onsave={saveRule} oncancel={() => close()} />
						{:else}
							<h2 class="caps">{t.editRule}</h2>
							<RuleForm rule={current.rule!} {npcs} {saving} onsave={saveRule} oncancel={() => close()} />
						{/if}
					{:else if isGlobal}
						<h2 class="caps">{t.everyoneElse}</h2>
						<p class="muted small">{t.unnamed}</p>
						{#if translator.global}
							<p class="pattern">{translator.global}</p>
						{:else}
							<p class="muted">{t.noGlobal}</p>
						{/if}
						{#if can}
							<div class="actions"><button type="button" class="button" onclick={() => open('edit')}>{m.common.edit}</button></div>
						{/if}
					{:else}
						{@const r = current.rule!}
						<h2 class="caps">{label(r)}</h2>
						<p class="muted small">
							{#if r.entityType === 'npc'}
								<Fill text={(here(r) ? t.npcHere : t.npc).text}>{#snippet template()}<span class="id">{ruleTemplate(r)}</span>{/snippet}</Fill>
							{:else}
								{entityLabel(r.entityType)} · <span class="id">{r.entityId}</span>
							{/if}
						</p>
						{#if r.pattern}
							<p class="pattern">{r.pattern}</p>
						{:else}
							<p class="muted">{t.noPattern}</p>
						{/if}
						{#if !r.enabled}<p class="notice">{t.isOff}</p>{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>{t.confirmDelete({ name: label(r) })}</p>
								<button type="button" class="button primary" onclick={remove} disabled={saving}>{m.common.delete}</button>
								<button type="button" class="button" onclick={() => close()}>{m.common.keep}</button>
							</div>
						{:else if can}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('edit')}>{m.common.edit}</button>
								<button type="button" class="button" onclick={toggleRule}>{r.enabled ? m.common.switchOff : m.common.switchOn}</button>
								<button type="button" class="button" onclick={() => move(-1)} disabled={current.at === 0}>{m.common.moveUp}</button>
								<button type="button" class="button" onclick={() => move(1)} disabled={current.at >= translator.rules.length - 1}>{m.common.moveDown}</button>
								<button type="button" class="button" onclick={() => open('delete')}>{t.deleteMore}</button>
							</div>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">{m.common.gameRunning}</p>{/if}
	</aside>
</div>

<style>
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
		margin-right: calc(clamp(18rem, 26vw, 24rem) + var(--space-5));
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
	}
	/* The switches, drawn as the quick toggles are: check boxes from the
	   game's own option menus. */
	.switches {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1) var(--space-4);
		margin-bottom: var(--space-4);
	}
	.switches button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		background: none;
		border: 0;
		padding: var(--space-1) 0;
		color: var(--muted);
		font-size: var(--step--1);
		cursor: pointer;
	}
	.switches button[aria-checked='true'],
	.switches button:hover:not(:disabled) {
		color: var(--ink);
	}
	.switches button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.switches .mark {
		display: inline-grid;
		place-items: center;
		width: 0.9rem;
		height: 0.9rem;
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		color: var(--on-accent);
	}
	.switches .mark svg {
		opacity: 0;
	}
	.switches button:hover:not(:disabled) .mark {
		border-color: var(--ink);
	}
	.switches button[aria-checked='true'] .mark {
		background: var(--accent);
		border-color: var(--accent);
	}
	.switches button[aria-checked='true'] .mark svg {
		opacity: 1;
	}
	.switches .small {
		flex-basis: 100%;
	}
	.notice {
		margin-bottom: var(--space-4);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	.list > .small {
		margin-top: var(--space-4);
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
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		vertical-align: baseline;
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
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	th.mark {
		border-bottom: 1px solid var(--rule-strong);
	}
	.at {
		width: 4.5rem;
	}
	td.at {
		text-align: right;
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}
	.who {
		width: 21rem; /* a name and a bio template, most of the time */
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.kind {
		font-size: var(--step--1);
		color: var(--muted);
		margin-right: var(--space-2);
	}
	.text {
		overflow: hidden;
	}
	.text .line {
		display: inline-block;
		max-width: 100%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		vertical-align: bottom;
	}
	/* A flag is a word of metadata, as in the other tables: never a button. */
	.flag {
		display: inline-block;
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.text .line:not(:empty) + .flag::before {
		content: '· ';
	}
	tr.off td {
		color: var(--muted);
	}
	/* The global rule stands a little apart from those above it. */
	tr.global td {
		border-top: 1px solid var(--rule-strong);
	}
	/* The pane is fitted to the window and scrolls on its own, so a form's
	   Save stays at its bottom edge, above the hint bar. */
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
	}
	.pane > div > .small,
	.pane > .small {
		margin-bottom: var(--space-4);
	}
	.pattern {
		line-height: 1.55;
		margin: var(--space-3) 0 var(--space-4);
		white-space: pre-wrap;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.actions,
	.confirm {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}
	.confirm p {
		flex-basis: 100%;
	}
	form.filter,
	form.import {
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
	textarea {
		width: 100%;
		resize: vertical;
		font-family: var(--font-mono);
		font-size: var(--step--1);
		line-height: 1.5;
	}
	.choices {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1) var(--space-3);
	}
	.choices .label {
		flex-basis: 100%;
	}
	.choices button:disabled {
		color: var(--muted);
		text-decoration: none;
		cursor: default;
	}
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0;
		background: var(--panel);
	}
	@media (max-width: 60rem) {
		.bar {
			margin-right: 0;
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
