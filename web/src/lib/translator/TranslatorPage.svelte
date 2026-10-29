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
	import { plural } from '#lib/format.ts';
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

	let { save, translator, characters, catalog }: { save: string; translator: Translator; characters: Character[]; catalog: Catalog } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(translator.found && !readOnly);
	let npcs = $derived(npcChoices(characters, catalog));
	let plugins = $derived(pluginChoices(characters));
	// An NPC's rule is named for the character: as SkyrimNet recorded them
	// in this playthrough, or as their bio template spells it. The template
	// shows beside the name, since the rule is keyed by it.
	let template = (r: Rule) => r.name || r.entityId;
	let npcName = (r: Rule) => npcOf(r, npcs)?.name ?? templateName(template(r));
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
		const m = param('mode');
		if (!translator.found) return 'view';
		if ((m === 'edit' || m === 'delete') && current && !(m === 'delete' && isGlobal)) return m;
		if (m === 'new' || m === 'filter' || m === 'import') return m;
		return 'view';
	});
	const open = (m: Mode) => setParams({ mode: m });
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
				return 'Saved the pattern for everyone else.';
			}
			if (mode === 'edit') {
				e.rules[current!.at] = rule;
				return `Saved the rule for ${label(rule)}.`;
			}
			// A new rule goes first: it applies before the rest, as the dashboard adds them.
			const id = `rule_${Date.now()}`;
			e.rules.unshift({ ...rule, id });
			return { note: `Added a rule for ${label(rule)}.`, opened: id };
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
			return by < 0 ? 'Moved up: it applies before the one below.' : 'Moved down: the one above applies first.';
		});
	}

	function toggleRule() {
		const r = current?.rule;
		if (!r || !can) return;
		commit((e) => {
			e.rules[current!.at]!.enabled = !r.enabled;
			return r.enabled ? 'Switched the rule off; it stays here for later.' : 'Switched the rule on.';
		});
	}

	function remove() {
		const r = current?.rule;
		if (!r) return;
		commit((e) => {
			e.rules.splice(current!.at, 1);
			return `Deleted the rule for ${label(r)}.`;
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
			return filterList.length ? `Saved the plugin filter: ${plural(filterList.length, 'plugin', 'plugins')}.` : 'Cleared the plugin filter.';
		});
	}

	// Switching the translator, and the SkyrimNet setting it depends on.
	function switchTranslator() {
		commit((e) => {
			e.enabled = !translator.enabled;
			return e.enabled ? 'Universal Translator is on.' : 'Universal Translator is off.';
		});
	}
	async function switchVoices() {
		const dep = translator.voiceSilentNPCs;
		if (saving || !dep.available || readOnly) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveSettings('game', { modified: dep.modified, changes: [{ key: 'dbvo.voiceSilentNPCs', value: !dep.on }] });
			notice = dep.on ? 'SkyrimNet no longer voices silent NPCs.' : 'SkyrimNet voices silent NPCs again.';
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
			if (res.added) parts.push(`${plural(res.added, 'rule', 'rules')} added`);
			if (res.updated) parts.push(`${plural(res.updated, 'rule', 'rules')} updated`);
			notice = `Imported ${f.name}: ${parts.join(', ') || 'nothing new'}.`;
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
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !can || !r, run: () => open('edit') },
			{ key: 'n', shown: 'N', label: 'New rule', hidden: !viewing, disabled: !can, run: () => open('new') },
			{ key: 'u', shown: 'U', label: 'Move up', hidden: !viewing, disabled: !can || !r?.rule || r.at === 0, run: () => move(-1) },
			{ key: 'd', shown: 'D', label: 'Move down', hidden: !viewing, disabled: !can || !r?.rule || r.at >= translator.rules.length - 1, run: () => move(1) },
			{ key: 'o', shown: 'O', label: r?.rule?.enabled === false ? 'Switch on' : 'Switch off', hidden: !viewing, disabled: !can || !r?.rule, run: toggleRule },
			{ key: 'f', shown: 'F', label: 'Plugin filter', hidden: !viewing, disabled: !can, run: () => open('filter') },
			{ key: 'i', shown: 'I', label: 'Import', hidden: !viewing, disabled: !can, run: () => open('import') },
			...deleteShortcuts({ label: 'Delete', hidden: !viewing, disabled: !can || !r?.rule, run: () => open('delete') }),
			{ key: 'Enter', label: 'Save', command: true, hidden: !(mode === 'edit' || mode === 'new' || mode === 'filter'), run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit() },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => close() }])
		];
	});

	let filterSummary = $derived(
		translator.filterPlugins.length === 0
			? 'No plugin filter: every voice line counts.'
			: `${translator.filterMode === 'whitelist' ? 'Only' : 'All but'} ${plural(translator.filterPlugins.length, 'plugin', 'plugins')}: ${translator.filterPlugins.join(', ')}.`
	);
</script>

<div class="bar">
	<h1 class="caps">Universal Translator</h1>
	<span class="muted n" use:clipped>
		{#if !translator.found}
			Not set up yet
		{:else if translator.enabled}
			On · {plural(translator.rules.length, 'rule', 'rules')}{translator.global ? ' · a pattern for everyone else' : ''}
		{:else}
			Off
		{/if}
	</span>
</div>
<p class="muted lead">Rewrites what NPCs say, line by line, in a speech pattern: the first rule that names them applies, and the rest speak as everyone else does.</p>

<div class="switches" role="group" aria-label="Switches">
	<button type="button" role="switch" aria-checked={translator.enabled} disabled={!can || saving} onclick={switchTranslator}>
		<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
		<span class="caps">Universal Translator</span>
	</button>
	<button type="button" role="switch" aria-checked={translator.voiceSilentNPCs.on} disabled={!translator.voiceSilentNPCs.available || readOnly || saving} onclick={switchVoices} title="dbvo.voiceSilentNPCs in SkyrimNet.yaml">
		<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
		<span class="caps">Voice silent NPCs</span>
	</button>
	<span class="muted small">The translator speaks through SkyrimNet’s voicing of silent NPCs, which needs DBVO installed with DBVO.esp disabled.</span>
</div>
{#if translator.found && translator.enabled && !translator.voiceSilentNPCs.on}
	<p class="notice warn">Universal Translator is on, but SkyrimNet isn’t voicing silent NPCs, so nothing it rewrites can be heard.</p>
{/if}
{#if !translator.found}
	<p class="notice">SkyrimNet hasn’t written <span class="id">{translator.path}</span> yet; it will the next time it starts. Until then there’s nothing to change here.</p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<th class="caps at">Order</th>
					<th class="caps">For</th>
					<th class="caps text">Speech pattern</th>
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
								{#if r.entityType === 'npc'}<span class="muted id"> · {template(r)}</span>{/if}
							{:else}
								<span class="kind caps">Everyone else</span>
							{/if}
						</td>
						<td class="text">
							<span class="line">{(r ? r.pattern : translator.global) || ''}</span>
							{#if r && !r.pattern}<span class="flag">left as written</span>{/if}
							{#if !r && !translator.global}<span class="flag">as written</span>{/if}
							{#if r && !r.enabled}<span class="flag">off</span>{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">{filterSummary}</p>
		<p class="muted small">The lines are rewritten by the model chosen for the translator on the <a href={modelHref(save)}>Models</a> page.</p>
	</div>

	<aside class="pane" use:fitToViewport aria-label="Rule">
		{#if mode === 'new'}
			<h2 class="caps">New rule</h2>
			<p class="muted small">It goes first, so it applies before every rule below it.</p>
			<RuleForm rule={blankRule()} {npcs} {saving} onsave={saveRule} oncancel={() => close()} />
		{:else if mode === 'filter'}
			<h2 class="caps">Plugin filter</h2>
			<p class="muted small">Which plugins’ voice lines the translator touches, by the plugin the line’s voice file comes from.</p>
			<form class="filter" onsubmit={saveFilter}>
				<Segmented bind:value={filterMode} options={[{ value: 'blacklist', label: 'All but these' }, { value: 'whitelist', label: 'Only these' }]} label="How the list applies" />
				<label class="field">
					<span class="caps label">Plugins, one per line</span>
					<textarea bind:value={filterText} rows="6" placeholder="Skyrim.esm" spellcheck="false"></textarea>
				</label>
				{#if plugins.length}
					<div class="choices">
						<span class="caps label">In this playthrough</span>
						{#each plugins as p (p)}
							<button type="button" class="textbutton" disabled={filterList.includes(p)} onclick={() => addPlugin(p)}>{p}</button>
						{/each}
					</div>
				{/if}
				<div class="buttons">
					<button type="submit" class="button primary" disabled={saving}>Save</button>
					<button type="button" class="button" onclick={() => close()} disabled={saving}>Cancel</button>
				</div>
			</form>
		{:else if mode === 'import'}
			<h2 class="caps">Import rules</h2>
			<p class="muted small">From a YAML file: one SkyrimNet’s dashboard exported, or UniversalTranslator.yaml from another install. A rule for an entity already listed replaces its pattern; the others go last, before everyone else.</p>
			<form class="import" onsubmit={runImport}>
				<input bind:this={file} type="file" accept=".yaml,.yml,text/yaml" aria-label="YAML file" />
				<div class="buttons">
					<button type="submit" class="button primary" disabled={saving}>Import</button>
					<button type="button" class="button" onclick={() => close()} disabled={saving}>Cancel</button>
				</div>
			</form>
		{:else if current}
			{#key `${current.key}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if mode === 'edit'}
						{#if isGlobal}
							<h2 class="caps">Everyone else</h2>
							<RuleForm rule={{ ...blankRule(), entityType: 'global', entityId: 'global', pattern: translator.global }} global {saving} onsave={saveRule} oncancel={() => close()} />
						{:else}
							<h2 class="caps">Edit rule</h2>
							<RuleForm rule={current.rule!} {npcs} {saving} onsave={saveRule} oncancel={() => close()} />
						{/if}
					{:else if isGlobal}
						<h2 class="caps">Everyone else</h2>
						<p class="muted small">NPCs no rule names, when the translator is on.</p>
						{#if translator.global}
							<p class="pattern">{translator.global}</p>
						{:else}
							<p class="muted">No pattern: they speak as written. Only the NPCs the rules name are translated.</p>
						{/if}
						{#if can}
							<div class="actions"><button type="button" class="button" onclick={() => open('edit')}>Edit</button></div>
						{/if}
					{:else}
						{@const r = current.rule!}
						<h2 class="caps">{label(r)}</h2>
						<p class="muted small">
							{#if r.entityType === 'npc'}
								NPC{here(r) ? ' in this playthrough' : ''}, by bio template <span class="id">{template(r)}</span>
							{:else}
								{entityLabel(r.entityType)} · <span class="id">{r.entityId}</span>
							{/if}
						</p>
						{#if r.pattern}
							<p class="pattern">{r.pattern}</p>
						{:else}
							<p class="muted">No pattern: they speak as written, whatever everyone else does.</p>
						{/if}
						{#if !r.enabled}<p class="notice">Switched off: skipped, so the rules below it can apply.</p>{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>Delete the rule for {label(r)}? A backup of the file is kept.</p>
								<button type="button" class="button primary" onclick={remove} disabled={saving}>Delete</button>
								<button type="button" class="button" onclick={() => close()}>Keep</button>
							</div>
						{:else if can}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('edit')}>Edit</button>
								<button type="button" class="button" onclick={toggleRule}>{r.enabled ? 'Switch off' : 'Switch on'}</button>
								<button type="button" class="button" onclick={() => move(-1)} disabled={current.at === 0}>Move up</button>
								<button type="button" class="button" onclick={() => move(1)} disabled={current.at >= translator.rules.length - 1}>Move down</button>
								<button type="button" class="button" onclick={() => open('delete')}>Delete…</button>
							</div>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
		{#if readOnly}<p class="notice warn">Skyrim is running, so nothing can be changed until you close it.</p>{/if}
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
