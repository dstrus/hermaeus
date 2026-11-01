<!-- A record's form: an existing record, edited as the player's own copy, or
     a new one of a kind with a form. It loads the record as an object, edits
     it in place, and saves it; the server merges it into the file. -->
<script lang="ts">
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import type { ContentList, RecordForm } from '#lib/api/types/records.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { aboveHints } from '#lib/ui/fit.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { getActions, getForm, layerName, saveForm } from './api.ts';
	import FormBlocks from './FormBlocks.svelte';
	import { fits, formKinds, setAt, type FormContext, type Obj } from './forms.ts';

	let {
		list,
		path,
		readOnly,
		onSaved,
		onCancel,
		onYaml
	}: {
		list: ContentList;
		path?: string; // the record to edit; none makes a new one
		readOnly: boolean;
		onSaved: (path: string, done: string) => void;
		onCancel: () => void;
		onYaml: () => void; // edit the record as YAML instead
	} = $props();

	// The editor is made for one record (or one new one), so these are read once.
	const isNew = untrack(() => !path);
	const stemOf = (p: string) => p.slice(p.lastIndexOf('/') + 1).replace(/(\.entity)?\.yaml$/, '');
	// SkyrimNet reads an entity only from a .entity.yaml file.
	const suffixOf = (k: string) => (k === 'entities' ? '.entity.yaml' : '.yaml');

	let form = $state<RecordForm | null>(null);
	let loadError = $state('');
	let error = $state('');
	let saving = $state(false);
	let draft = $state<Obj>({});
	let kind = $state(untrack(() => (path ? path.slice(0, path.indexOf('/')) : 'triggers')));
	let fileName = $state('');

	// Names the form suggests: voice effects and action categories.
	let ctx = $state<FormContext>({ effects: [], categories: [], decorators: [] });
	$effect(() => {
		ctx.effects = list.records.filter((r) => r.kind === 'voice_effects').map((r) => stemOf(r.path));
		if (kind === 'actions')
			getActions().then(
				(a) => (ctx.categories = [...new Set(a.actions.map((x) => x.category ?? '').filter(Boolean))].sort()),
				() => {}
			);
	});

	// An existing record loads as its object; a new one starts blank.
	$effect(() => {
		if (!path) return;
		getForm(path).then(
			(f) => {
				form = f;
				if (f.data) draft = JSON.parse(JSON.stringify(f.data));
			},
			(e) => (loadError = e instanceof ApiError ? e.message : String(e))
		);
	});
	function startNew(k: string) {
		kind = k;
		draft = formKinds[k]!.blank(fileName);
	}
	if (isNew) untrack(() => startNew(kind));

	// The field that names the file follows the file's name.
	function identity(d: Obj): string | undefined {
		const fk = formKinds[kind];
		if (fk?.nameKey) return fk.nameKey;
		if (kind === 'filters' && (d.kind === 'dialogue_rule' || d.kind === 'tts_rule')) return 'id';
		if (kind === 'dialogue_actions' && d.kind === 'instruction') return 'key';
		return undefined;
	}
	function rename(name: string) {
		fileName = name;
		const key = identity(draft);
		if (key) setAt(draft, key, name);
	}
	$effect(() => {
		// A filter or dialogue action whose kind changes takes its new
		// identity field from the file's name.
		const key = identity(draft);
		if (isNew && key && draft[key] !== fileName) setAt(draft, key, fileName);
	});

	let fk = $derived(formKinds[kind]);
	let usable = $derived(isNew || (form !== null && fits(kind, form.data ?? null)));
	let mine = $derived(form?.versions.find((v) => v.kind === 'overlay'));
	let below = $derived(form?.versions.find((v) => v.kind !== 'overlay'));
	let target = $derived(isNew ? `${kind}/${fileName.trim()}${suffixOf(kind)}` : (path ?? ''));
	// Where a new record is saved, and the field that names it.
	let savedAs = $derived.by(() => {
		const at = `${kind}/${fileName.trim() || '…'}${suffixOf(kind)}`;
		const key = identity(draft);
		return key ? m.records.editor.savedAsSame({ path: at, key }) : m.records.editor.savedAs({ path: at });
	});

	async function submit() {
		if (saving || readOnly || !fk) return;
		error = '';
		if (isNew && !/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/.test(fileName.trim())) {
			error = m.records.editor.badFileName;
			return;
		}
		saving = true;
		const had = !!mine;
		try {
			const saved = await saveForm(target, { modified: mine?.modified ?? 0, data: $state.snapshot(draft), new: isNew });
			form = saved;
			onSaved(target, isNew ? m.records.editor.made({ path: target }) : had ? m.records.savedCopy : m.records.editor.savedAsCopy);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	useShortcuts(() => [
		{ key: 'Enter', label: isNew ? m.records.editor.makeIt : m.common.save, command: true, disabled: saving || !usable, run: submit },
		{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: onCancel }
	]);
</script>

<header class="editing">
	<h1 class="caps">{isNew ? m.records.editor.newKind({ noun: fk?.noun ?? m.records.editor.record }) : path ? stemOf(path) : ''}</h1>
	{#if !isNew}<p class="path">{path}</p>{/if}
	<p class="muted">
		{#if isNew}
			{m.records.editor.madeInEdits}
		{:else if mine}
			<Fill text={(below ? m.records.editingCopyInPlaceOf : m.records.editingCopy).text} layer={below ? layerName(list.layers[below.layer]) : ''}>{#snippet yourCopy()}<b class="changed">{m.records.yourOwnCopy}</b>{/snippet}</Fill>
		{:else if below}
			{m.records.startingFrom({ layer: layerName(list.layers[below.layer]) })}
		{/if}
		{#if !isNew}<button type="button" class="textbutton" onclick={onYaml}>{m.records.editor.editYaml}</button>{/if}
	</p>
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	{#if loadError}<p class="notice error">{loadError}</p>{/if}
</header>

{#if isNew}
	<div class="new">
		<label>
			<span class="caps">{m.records.kind}</span>
			<select value={kind} onchange={(e) => startNew(e.currentTarget.value)}>
				{#each Object.entries(formKinds) as [k, f] (k)}<option value={k}>{f.noun.charAt(0).toUpperCase() + f.noun.slice(1)}</option>{/each}
			</select>
		</label>
		<label>
			<span class="caps">{m.records.editor.fileName}</span>
			<input type="text" value={fileName} oninput={(e) => rename(e.currentTarget.value.trim())} placeholder="beast_form" autocomplete="off" spellcheck="false" />
			<span class="muted small">{savedAs}</span>
		</label>
	</div>
{/if}

{#if form && !usable}
	<p class="notice">{form.error ? m.records.editor.unreadable({ error: form.error }) : m.records.editor.notCovered} <button type="button" class="textbutton" onclick={onYaml}>{m.records.editor.editItYaml}</button></p>
{:else if fk && (isNew || form)}
	<form class="record" onsubmit={(e) => (e.preventDefault(), submit())}>
		<FormBlocks blocks={fk.blocks(draft, ctx)} {draft} />
		<div class="buttons" use:aboveHints>
			<button type="submit" class="button primary" disabled={saving || readOnly}>{isNew ? m.records.editor.makeIt : m.common.save}</button>
			<button type="button" class="button" onclick={onCancel} disabled={saving}>{m.common.cancel}</button>
		</div>
	</form>
{/if}

<style>
	.editing {
		margin-bottom: var(--space-4);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		margin-bottom: var(--space-2);
	}
	.path {
		font: 400 var(--step--1) / 1.5 var(--font-mono);
		color: var(--muted);
		margin-bottom: var(--space-2);
	}
	.editing .muted {
		max-width: 64rem;
	}
	.editing .textbutton {
		margin-left: var(--space-2);
	}
	.changed {
		font-weight: 400;
		color: var(--changed);
	}
	.new {
		display: grid;
		grid-template-columns: 14rem minmax(0, 24rem);
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.new label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.new .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.small {
		font-size: var(--step--1);
	}
	.record {
		max-width: 64rem;
	}
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0 var(--space-5);
		background: var(--ground);
		border-top: 1px solid var(--rule);
	}
</style>
