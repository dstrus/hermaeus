<!-- Providers: the services SkyrimNet sends its tasks to. Each has a key,
     set here and never shown again, and a model list to fetch on request.
     Every change writes Providers.yaml with a backup. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { tick } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { ProviderDetails, ProviderEdit, ProvidersResponse } from '#lib/api/types/settings.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { modelsDependency, modelsHref, taskLabel } from '#lib/models/api.ts';
	import { modelLists } from '#lib/models/lists.svelte.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { addProvider, blankProvider, editOf, keyState, providersDependency, removeProvider, saveProvider, setKey, styleLabel } from './api.ts';
	import ProviderForm from './ProviderForm.svelte';

	const t = m.providers;

	let { save, providers: loaded }: { save: string; providers: ProvidersResponse } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(loaded.found && !readOnly);
	let all = $derived(loaded.providers);
	const selection = new Selection(() => all, (p) => p.id, { param: 'item' });
	let current = $derived(selection.current);
	let keyed = $derived(all.filter((p) => keyState(p, all) !== 'none').length);

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	// What the pane is doing, in the address: each state has its own URL,
	// and Back closes a form.
	type Mode = 'view' | 'edit' | 'new' | 'key' | 'delete';
	let mode = $derived.by((): Mode => {
		const asked = param('mode');
		if (!loaded.found) return 'view';
		if (asked === 'new') return 'new';
		if ((asked === 'edit' || asked === 'key') && current) return asked;
		if (asked === 'delete' && current && !current.builtin) return 'delete';
		return 'view';
	});
	const open = (to: Mode) => setParams({ mode: to });
	const close = (opened?: string) => (opened ? setParams({ item: opened, mode: null }, 'replace') : leave('mode'));
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	type Done = string | { note: string; opened: string };
	async function run(action: () => Promise<Done>) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const done = await action();
			notice = typeof done === 'string' ? done : done.note;
			await close(typeof done === 'string' ? undefined : done.opened); // settles before the reload, which it would cancel
			await Promise.all([invalidate(providersDependency), invalidate(modelsDependency)]);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function saveForm(edit: ProviderEdit) {
		const p = current;
		run(async () => {
			if (mode === 'new') {
				const r = await addProvider(edit);
				const added = r.providers.at(-1);
				return { note: t.added({ name: edit.name }), opened: added?.id ?? '' };
			}
			await saveProvider(p!.id, edit);
			return t.saved({ name: edit.name });
		});
	}

	// The key: typed into a field that's cleared as soon as it's sent.
	let keyDraft = $state('');
	let keyField = $state<HTMLInputElement>();
	$effect(() => {
		if (mode === 'key') tick().then(() => keyField?.focus());
	});
	function saveKey(e: SubmitEvent) {
		e.preventDefault();
		const p = current;
		const key = keyDraft.trim();
		keyDraft = '';
		run(async () => {
			await setKey(p!.id, { modified: loaded.modified, key });
			return key ? t.keyWasSet({ name: p!.name }) : t.keyCleared({ name: p!.name });
		});
	}
	function clearKey() {
		const p = current;
		run(async () => {
			await setKey(p!.id, { modified: loaded.modified, key: '' });
			return t.keyCleared({ name: p!.name });
		});
	}

	function remove() {
		const p = current;
		run(async () => {
			await removeProvider(p!.id, loaded.modified);
			return t.removed({ name: p!.name });
		});
	}

	function listState(p: ProviderDetails): string {
		if (p.style !== 'chat') return '';
		if (modelLists.loading[p.id]) return m.models.lists.asking;
		const list = modelLists.lists[p.id];
		return list ? m.models.lists.fetched({ count: list.models.length }) : '';
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		const p = current;
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: m.common.edit, hidden: !viewing, disabled: !can || !p, run: () => open('edit') },
			{ key: 'k', shown: 'K', label: t.setKey, hidden: !viewing, disabled: !can || !p, run: () => open('key') },
			{ key: 'n', shown: 'N', label: t.newProvider, hidden: !viewing, disabled: !can, run: () => open('new') },
			{ key: 'f', shown: 'F', label: m.models.lists.fetch, hidden: !viewing, disabled: !p || p.style !== 'chat' || modelLists.loading[p.id], run: () => p && modelLists.fetch(p.id, !!modelLists.lists[p.id]) },
			...deleteShortcuts({ label: m.common.remove, hidden: !viewing || !p || p.builtin, disabled: !can || !p || p.builtin, run: () => open('delete') }),
			{ key: 'Enter', label: m.common.save, command: true, hidden: viewing || mode === 'delete', run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit() },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => close() }])
		];
	});
</script>

<div class="bar">
	<h1 class="caps">{t.title}</h1>
	<span class="muted n" use:clipped>
		{#if !loaded.found}
			{m.common.notSetUp}
		{:else}
			{t.count({ count: all.length })} · {keyed === all.length ? t.allKeyed : keyed ? t.keyed({ count: keyed }) : t.noneKeyed}
		{/if}
	</span>
</div>
<p class="muted lead">{t.lead}</p>
{#if !loaded.found}
	<p class="notice"><Fill text={m.common.notWrittenYet.text}>{#snippet path()}<span class="id">{loaded.path}</span>{/snippet}</Fill></p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label={m.common.selected}></th>
					<th class="caps name">{t.provider}</th>
					<th class="caps key">{t.key}</th>
					<th class="caps style">{t.api}</th>
					<th class="caps">{t.usedBy}</th>
				</tr>
			</thead>
			<tbody>
				{#each all as p (p.id)}
					{@const key = keyState(p, all)}
					<tr aria-selected={selection.is(p)} class:unkeyed={key === 'none'} onclick={() => mode === 'view' && selection.select(p)} ondblclick={() => can && mode === 'view' && open('edit')}>
						<td class="mark">{#if selection.is(p)}<Diamond />{/if}</td>
						<td class="name">{p.name}{#if !p.builtin}<span class="muted small"> · {t.yours}</span>{/if}</td>
						<td class="key">{key === 'own' ? t.keySet : key === 'borrowed' ? t.borrowed({ provider: all.find((o) => o.id === p.keyFrom)?.name ?? p.keyFrom ?? '' }) : t.none}</td>
						<td class="style">{styleLabel(p.style)}</td>
						<td class="tasks" use:clipped>{p.tasks.map(taskLabel).join(', ')}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small"><Fill text={t.modelsNote.text}>{#snippet models()}<a href={modelsHref(save)}>{m.models.title}</a>{/snippet}</Fill></p>
	</div>

	<aside class="pane" use:fitToViewport aria-label={t.provider}>
		{#if mode === 'new'}
			<h2 class="caps">{t.newProvider}</h2>
			<p class="muted small">{t.newHint}</p>
			<ProviderForm edit={blankProvider(loaded.modified)} isNew {saving} onsave={saveForm} oncancel={() => close()} />
		{:else if current}
			{@const p = current}
			{@const key = keyState(p, all)}
			{#key `${p.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{p.name}</h2>
					<p class="muted small">{t.styleApi({ style: styleLabel(p.style) })} · <span class="id">{p.id}</span></p>
					{#if mode === 'edit'}
						<ProviderForm edit={editOf(p, loaded.modified)} isOpenRouter={p.id === 'openrouter'} {saving} onsave={saveForm} oncancel={() => close()} />
					{:else if mode === 'key'}
						<form class="keyform" onsubmit={saveKey}>
							<label class="field">
								<span class="caps label">{p.hasKey ? t.newKey : t.key}</span>
								<input bind:this={keyField} type="password" bind:value={keyDraft} autocomplete="off" spellcheck="false" placeholder={t.pasteKey} />
								<span class="muted small"><Fill text={t.keySaved.text}>{#snippet path()}<span class="id">{loaded.path}</span>{/snippet}</Fill></span>
							</label>
							<div class="buttons">
								<button type="submit" class="button primary" disabled={saving || !keyDraft.trim()}>{t.saveKey}</button>
								<button type="button" class="button" onclick={() => close()} disabled={saving}>{m.common.cancel}</button>
							</div>
						</form>
					{:else}
						{@const lender = all.find((o) => o.id === p.keyFrom)?.name ?? p.keyFrom ?? ''}
						<dl>
							<dt class="caps">{t.endpoint}</dt><dd class="id endpoint">{p.endpoint}</dd>
							<dt class="caps">{t.key}</dt>
							<dd>
								{#if key === 'own'}{t.keySet}{:else if key === 'borrowed'}{t.borrows({ provider: lender })}{:else if p.keyFrom}{t.noneYetBorrow({ provider: lender })}{:else}{t.noneYet}{/if}
							</dd>
							<dt class="caps">{t.streaming}</dt><dd>{p.streaming ? m.common.isOn : m.common.isOff}</dd>
							<dt class="caps">{t.structuredOutputs}</dt><dd>{p.structuredOutputs ? m.common.isOn : m.common.isOff}</dd>
							<dt class="caps">{t.reasoning}</dt><dd>{p.reasoning ? m.common.isOn : m.common.isOff}</dd>
							{#if p.id === 'openrouter' && p.sorting && p.sorting !== 'default'}<dt class="caps">{t.upstreams}</dt><dd>{t.bySorting({ sorting: p.sorting })}</dd>{/if}
							{#if p.cacheTTL === '1h'}<dt class="caps">{t.cache}</dt><dd>{t.anHour}</dd>{/if}
							{#if p.settings}<dt class="caps">{t.settings}</dt><dd class="id settings">{p.settings}</dd>{/if}
							<dt class="caps">{t.usedBy}</dt><dd>{p.tasks.length ? p.tasks.map(taskLabel).join(', ') : t.noTask}</dd>
							{#if p.style === 'chat'}
								<dt class="caps">{t.models}</dt>
								<dd>
									{#if modelLists.errors[p.id]}
										<span class="error">{modelLists.errors[p.id]}</span>
									{:else if listState(p)}
										{t.listOffered({ list: listState(p) })}
									{:else}
										{(key === 'none' ? t.notFetched : t.notFetchedKey)({ provider: p.name })}
									{/if}
								</dd>
							{/if}
						</dl>
						{#if key === 'none' && p.style === 'chat'}
							<p class="notice warn">{t.noKeyWarn({ provider: p.name })}</p>
						{/if}
						{#if mode === 'delete'}
							<div class="confirm" role="alert">
								<p>{t.confirmRemove({ provider: p.name })}</p>
								<button type="button" class="button primary" onclick={remove} disabled={saving}>{m.common.remove}</button>
								<button type="button" class="button" onclick={() => close()}>{m.common.keep}</button>
							</div>
						{:else if can}
							<div class="actions">
								<button type="button" class="button" onclick={() => open('key')}>{p.hasKey ? t.changeKey : t.setKey}</button>
								{#if p.hasKey}<button type="button" class="button" onclick={clearKey}>{t.clearKey}</button>{/if}
								<button type="button" class="button" onclick={() => open('edit')}>{m.common.edit}</button>
								{#if p.style === 'chat'}
									<button type="button" class="button" onclick={() => modelLists.fetch(p.id, !!modelLists.lists[p.id])} disabled={modelLists.loading[p.id]}>{modelLists.lists[p.id] ? t.fetchModelsAgain : m.models.lists.fetch}</button>
								{/if}
								{#if !p.builtin}<button type="button" class="button" onclick={() => open('delete')}>{t.removeMore}</button>{/if}
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
	.name {
		width: 13rem;
	}
	.key {
		width: 7rem;
	}
	.style {
		width: 7rem;
	}
	.tasks {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--muted);
	}
	tr.unkeyed .name,
	tr.unkeyed .key {
		color: var(--muted);
	}
	/* The pane is fitted to the window and scrolls on its own. */
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
	.pane dl {
		display: grid;
		grid-template-columns: 7.5rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: 0 0 var(--space-4);
	}
	.pane dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	.pane dd {
		margin: 0;
		min-width: 0;
	}
	.endpoint,
	.settings {
		overflow-wrap: anywhere;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.error {
		color: var(--warn);
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
	.keyform {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.field input {
		width: 100%;
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
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
