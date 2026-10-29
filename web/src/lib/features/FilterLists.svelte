<!-- A filter's lists: what they add up to, and each list, from SkyrimNet,
     a plugin, or you. Editing a built-in or plugin list saves your own copy,
     which replaces it until you remove the copy. Which list is open, and
     whether it's being edited, is in the address. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { tick } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { List, ListsResponse } from '#lib/api/types/filters.ts';
	import { plural } from '#lib/format.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { createList, filterFields, filtersDependency, removeList, saveList, type FilterKind } from './filters.ts';

	let { kind, lists, canEdit }: { kind: FilterKind; lists: ListsResponse; canEdit: boolean } = $props();

	let readOnly = $derived(live.gameRunning || !canEdit);
	let open = $derived(param('list'));
	let mode = $derived(param('mode'));
	let editingId = $derived(mode === 'list' && open && lists.lists.some((l) => l.id === open) ? open : null);
	let adding = $derived(mode === 'list-new');

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	const origin = (l: List) => (l.layerKind === 'overlay' ? (l.replaces ? `Your copy of the one in ${l.replaces}` : 'Yours') : l.layerKind === 'base' ? 'Built-in' : l.source);
	const counts = (l: List) =>
		filterFields
			.filter((f) => l.entries[f.key]?.length)
			.map((f) => `${f.label}: ${l.entries[f.key]!.length}`)
			.join(' · ') || 'Empty';

	// The form, set up when a list opens for editing or a new one starts.
	let name = $state('');
	let description = $state('');
	let fields = $state<Record<string, string>>({});
	let nameField = $state<HTMLInputElement>();
	$effect(() => {
		const l = editingId ? lists.lists.find((x) => x.id === editingId) : undefined;
		if (!l && !adding) return;
		name = l?.name ?? '';
		description = l?.description ?? '';
		fields = Object.fromEntries(filterFields.map((f) => [f.key, (l?.entries[f.key] ?? []).join('\n')]));
		tick().then(() => nameField?.focus());
	});

	async function run(action: () => Promise<{ note: string; opened?: string }>) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const done = await action();
			notice = done.note;
			await (done.opened ? setParams({ list: done.opened, mode: null }, 'replace') : leave('mode'));
			await invalidate(filtersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function edit() {
		return {
			modified: 0,
			name,
			description,
			entries: Object.fromEntries(filterFields.map((f) => [f.key, (fields[f.key] ?? '').split('\n').map((v) => v.trim()).filter(Boolean)]))
		};
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		if (adding) {
			run(async () => {
				const r = await createList(kind, edit());
				return { note: `Added ${r.created.name}.`, opened: r.created.id };
			});
			return;
		}
		const l = lists.lists.find((x) => x.id === editingId)!;
		run(async () => {
			// Your own file carries its time; a first copy of another layer's has none yet.
			await saveList(kind, l.id, { ...edit(), modified: l.layerKind === 'overlay' ? l.modified : 0 });
			return { note: l.layerKind === 'overlay' ? `Saved ${name}.` : `Saved your copy of ${l.name}. It replaces the original until you remove it.` };
		});
	}

	function remove(l: List) {
		run(async () => {
			await removeList(kind, l.id, l.modified);
			return { note: l.replaces ? `Removed your copy of ${l.name}; the one in ${l.replaces} applies again.` : `Removed ${l.name}.` };
		});
	}
</script>

<section class="effect">
	<h2 class="caps">In effect</h2>
	<p class="muted small">What every list below adds up to, with the lists each entry comes from.</p>
	<dl>
		{#each filterFields as f (f.key)}
			{@const entries = lists.inEffect[f.key] ?? []}
			<dt class="caps">{f.label}</dt>
			<dd>
				{#if entries.length}
					<ul>
						{#each entries as e (e.value)}
							<li><span class="id">{e.value}</span><span class="muted small from">· {e.from.map((s) => s.name).join(', ')}</span></li>
						{/each}
					</ul>
				{:else}
					<span class="muted">Nothing</span>
				{/if}
			</dd>
		{/each}
	</dl>
</section>

<section class="lists">
	<div class="head">
		<h2 class="caps">Lists</h2>
		{#if !readOnly && !adding}
			<button type="button" class="button" onclick={() => setParams({ list: null, mode: 'list-new' })}>New list</button>
		{/if}
	</div>
	<p class="muted small">Every list adds its entries. Editing a built-in or plugin list saves your own copy over it; remove the copy to bring the original back. To drop one of them entirely, edit it and empty it.</p>
	{#if notice}<p class="notice" role="status">{notice}</p>{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}

	{#snippet form(title: string)}
		<form class="listform" onsubmit={submit}>
			<h3 class="caps">{title}</h3>
			<label class="field"><span class="caps label">Name</span><input bind:this={nameField} type="text" bind:value={name} autocomplete="off" /></label>
			<label class="field"><span class="caps label">Description</span><input type="text" bind:value={description} autocomplete="off" /></label>
			<div class="grid">
				{#each filterFields as f (f.key)}
					<label class="field">
						<span class="caps label">{f.label}</span>
						<textarea bind:value={fields[f.key]} rows="3" spellcheck="false" placeholder={'placeholder' in f ? f.placeholder : `One ${f.noun} per line, by editor ID`}></textarea>
					</label>
				{/each}
			</div>
			<div class="buttons">
				<button type="submit" class="button primary" disabled={saving}>{adding ? 'Add' : 'Save'}</button>
				<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
			</div>
		</form>
	{/snippet}

	{#if adding}{@render form('New list')}{/if}
	<ul class="rows">
		{#each lists.lists as l (l.id)}
			<li class:open={open === l.id}>
				{#if editingId === l.id}
					{@render form(l.layerKind === 'overlay' ? `Edit ${l.name}` : `Your copy of ${l.name}`)}
				{:else}
					<div class="row">
						<div class="text">
							<span class="name">{l.name}</span>
							<span class="origin small" class:mine={l.layerKind === 'overlay'}>{origin(l)}</span>
							{#if l.description}<p class="muted small">{l.description}</p>{/if}
							<p class="muted small">{counts(l)}</p>
						</div>
						{#if !readOnly}
							<div class="actions">
								<button type="button" class="textbutton small" onclick={() => setParams({ list: l.id, mode: 'list' })}>{l.layerKind === 'overlay' ? 'Edit' : 'Edit a copy'}</button>
								{#if l.layerKind === 'overlay'}
									<button type="button" class="textbutton small" onclick={() => remove(l)} disabled={saving}>{l.replaces ? 'Remove copy' : 'Remove'}</button>
								{/if}
							</div>
						{/if}
					</div>
				{/if}
			</li>
		{:else}
			<li class="muted">No {kind} filter lists in any layer.</li>
		{/each}
	</ul>
	<p class="muted small">{plural(lists.lists.length, 'list', 'lists')}, from SkyrimNet’s defaults, your plugins, and your own.</p>
</section>

<style>
	section {
		margin-top: var(--space-6, 2.5rem);
	}
	h2 {
		font-size: var(--step-0);
		color: var(--accent);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-2);
	}
	h3 {
		font-size: var(--step-0);
		font-weight: 600;
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
	.effect dl {
		display: grid;
		grid-template-columns: 10rem 1fr;
		gap: var(--space-2) var(--space-4);
		margin: var(--space-3) 0 0;
	}
	.effect dt {
		font-size: var(--step--1);
		color: var(--muted);
		padding-top: 0.15em;
	}
	.effect dd {
		margin: 0;
	}
	/* The lists an entry comes from, set a step away from the entry. */
	.from {
		margin-left: var(--space-2);
	}
	.effect ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	.lists .head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-2);
	}
	.lists .head h2 {
		border-bottom: 0;
		margin-bottom: 0;
	}
	.lists .head .button {
		margin-bottom: var(--space-2);
	}
	.notice {
		margin: var(--space-3) 0;
	}
	.rows {
		list-style: none;
		padding: 0;
		margin: var(--space-3) 0;
	}
	.rows > li {
		border-bottom: 1px solid var(--rule);
		padding: var(--space-3) 0;
	}
	.row {
		display: flex;
		gap: var(--space-4);
		align-items: baseline;
		justify-content: space-between;
	}
	.name {
		font-weight: 600;
	}
	/* Where a list comes from, as a word of metadata; yours in copper. */
	.origin {
		margin-left: var(--space-2);
		color: var(--muted);
	}
	.origin::before {
		content: '· ';
	}
	.origin.mine {
		color: var(--changed);
	}
	.actions {
		display: flex;
		gap: var(--space-3);
		flex: none;
	}
	.listform {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		padding: var(--space-3);
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		margin-bottom: var(--space-3);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.field input,
	.field textarea {
		width: 100%;
	}
	textarea {
		resize: vertical;
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}
	.buttons {
		display: flex;
		gap: var(--space-2);
	}
</style>
