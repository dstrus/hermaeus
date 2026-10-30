<!-- A record's form: an existing record, edited as the player's own copy, or
     a new one of a kind with a form. It loads the record as an object, edits
     it in place, and saves it; the server merges it into the file. -->
<script lang="ts">
	import { untrack } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import type { ContentList, RecordForm } from '#lib/api/types/records.ts';
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

	async function submit() {
		if (saving || readOnly || !fk) return;
		error = '';
		if (isNew && !/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/.test(fileName.trim())) {
			error = 'A file name is up to 64 letters, digits, hyphens, and underscores, starting with a letter or digit.';
			return;
		}
		saving = true;
		const had = !!mine;
		try {
			const saved = await saveForm(target, { modified: mine?.modified ?? 0, data: $state.snapshot(draft), new: isNew });
			form = saved;
			onSaved(target, isNew ? `Made ${target}. SkyrimNet reads it from the next time it starts.` : had ? 'Saved your copy. SkyrimNet reads it from the next time it starts.' : 'Saved as your own copy, in place of the one below. SkyrimNet reads it from the next time it starts.');
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	useShortcuts(() => [
		{ key: 'Enter', label: isNew ? 'Make it' : 'Save', command: true, disabled: saving || !usable, run: submit },
		{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: onCancel }
	]);
</script>

<header class="editing">
	<h1 class="caps">{isNew ? `New ${fk?.noun ?? 'record'}` : (path ? stemOf(path) : '')}</h1>
	{#if !isNew}<p class="path">{path}</p>{/if}
	<p class="muted">
		{#if isNew}
			It’s made in your edits, and SkyrimNet reads it in every playthrough.
		{:else if mine}
			Editing <b class="changed">your own copy</b>, which SkyrimNet uses in every playthrough{below ? ` in place of the one from ${layerName(list.layers[below.layer])}` : ''}.
		{:else if below}
			Starting from the copy from {layerName(list.layers[below.layer])}. Saving makes your own copy, which replaces it in every playthrough.
		{/if}
		{#if !isNew}<button type="button" class="textbutton" onclick={onYaml}>Edit as YAML</button>{/if}
	</p>
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	{#if loadError}<p class="notice error">{loadError}</p>{/if}
</header>

{#if isNew}
	<div class="new">
		<label>
			<span class="caps">Kind</span>
			<select value={kind} onchange={(e) => startNew(e.currentTarget.value)}>
				{#each Object.entries(formKinds) as [k, f] (k)}<option value={k}>{f.noun.charAt(0).toUpperCase() + f.noun.slice(1)}</option>{/each}
			</select>
		</label>
		<label>
			<span class="caps">File name</span>
			<input type="text" value={fileName} oninput={(e) => rename(e.currentTarget.value.trim())} placeholder="beast_form" autocomplete="off" spellcheck="false" />
			<span class="muted small">Saved as {kind}/{fileName.trim() || '…'}{suffixOf(kind)}{identity(draft) ? `, and its ${identity(draft)} is the same` : ''}.</span>
		</label>
	</div>
{/if}

{#if form && !usable}
	<p class="notice">{form.error ? `This file can’t be read as a record: ${form.error}.` : 'This one is laid out in a way the form doesn’t cover.'} <button type="button" class="textbutton" onclick={onYaml}>Edit it as YAML</button></p>
{:else if fk && (isNew || form)}
	<form class="record" onsubmit={(e) => (e.preventDefault(), submit())}>
		<FormBlocks blocks={fk.blocks(draft, ctx)} {draft} />
		<div class="buttons" use:aboveHints>
			<button type="submit" class="button primary" disabled={saving || readOnly}>{isNew ? 'Make it' : 'Save'}</button>
			<button type="button" class="button" onclick={onCancel} disabled={saving}>Cancel</button>
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
