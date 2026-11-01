<!-- Models: which model does each of SkyrimNet's tasks. Every task has
     its own rotation; the selected task's is in the pane, and Edit opens it
     as a form. Each save writes OpenRouter.yaml with a backup. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Models, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { providersHref } from '#lib/providers/api.ts';
	import { settingsHref } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { modelChoices, modelsDependency, providerName, rotation, saveTask, taskLabel, taskWhat } from './api.ts';
	import { modelLists } from './lists.svelte.ts';
	import { page } from '$app/state';
	import type { PluginSchema } from '#lib/api/types/settings.ts';
	import { pluginHref } from '#lib/plugins/links.ts';
	import TaskForm from './TaskForm.svelte';
	import ModelTest from './ModelTest.svelte';

	const t = m.models;

	let { save, models }: { save: string; models: Models } = $props();

	let readOnly = $derived(live.gameRunning);
	// The plugin that added a task, when its settings schema names it.
	const addedBy = (task: string) => (page.data.pluginSettings as PluginSchema[] | undefined)?.find((s) => s.variants.includes(task) && s.plugin);
	let can = $derived(models.found && !readOnly);
	let choices = $derived(modelChoices(models));
	let dialogue = $derived(models.tasks.find((task) => task.id === 'default'));
	const provider = (id: string) => providerName(models.providers, id);
	// The providers a task's slots use, named once each.
	const providersOf = (task: Task) => [...new Set(task.slots.map((s) => provider(s.provider)))].join(', ');
	const order = (task: Task) => ((task.rotationMode || dialogue?.rotationMode || 'sequential') === 'random' ? t.atRandom : t.inTurn);
	const shortModel = (model: string) => model.replace(/^[^/]+\//, '');

	const selection = new Selection(() => models.tasks, (task) => task.id, { param: 'item' });
	let current = $derived(selection.current);

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	// Editing is a state of the address, so it has its own URL and Back cancels it.
	let mode = $derived<'view' | 'edit'>(param('mode') === 'edit' && current && models.found ? 'edit' : 'view');
	$effect(() => {
		if (mode === 'view' && param('mode')) setParams({ mode: null }, 'replace');
	});

	async function submit(edit: TaskEdit) {
		const task = current;
		if (!task || saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveTask(task.id, edit);
			notice = t.saved({ task: taskLabel(task.id) });
			await leave('mode'); // settles before the reload, which it would cancel
			await invalidate(modelsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	// slotNotes are what a slot sets for itself, as a short line.
	function slotNotes(s: Slot, task: Task): string {
		const n = t.slot;
		const out: string[] = [];
		if (s.role === 'backup') out.push(n.backup);
		if (!s.enabled) out.push(n.outOfRotation);
		if (s.temperature !== undefined) out.push(n.temperature({ value: s.temperature }));
		else if (task.params.temperature !== undefined) out.push(n.temperature({ value: task.params.temperature }));
		if (s.weight !== undefined) out.push(n.weight({ value: s.weight }));
		if (s.requestTimeout !== undefined) out.push(t.seconds({ seconds: s.requestTimeout }));
		if (s.maxRetries !== undefined) out.push(n.retries({ count: s.maxRetries }));
		if (s.reasoningEffort) out.push(n.reasoning({ effort: s.reasoningEffort }));
		if (s.structuredOutputs !== undefined) out.push(s.structuredOutputs ? n.structured : n.unstructured);
		return out.join(' · ');
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: m.common.edit, hidden: !viewing, disabled: !can || !current, run: () => setParams({ mode: 'edit' }) },
			{ key: 'Enter', label: m.common.save, command: true, hidden: viewing, run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit() },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => leave('mode') }])
		];
	});

	// The row above the tasks shows the providers a task can use: those with
	// a key. The rest wait on the Providers page.
	// Fateless needs no key here: SkyrimNet reaches it through its own sign-in.
	let ready = $derived(models.providers.filter((p) => p.hasKey && p.style === 'chat' && !p.signIn));
	let waiting = $derived(models.providers.filter((p) => !p.hasKey && p.style === 'chat' && !p.signIn));
	// A provider's model list, as a few words.
	function listState(id: string): string {
		if (modelLists.loading[id]) return t.lists.asking;
		const list = modelLists.lists[id];
		return list ? t.lists.fetched({ count: list.models.length }) : t.lists.notFetched;
	}
</script>

<div class="bar">
	<h1 class="caps">{t.title}</h1>
	<span class="muted n" use:clipped>
		{#if !models.found}
			{m.common.notSetUp}
		{:else}
			{t.tasksCount({ count: models.tasks.length })} · {t.providersCount({ count: models.providers.length })}
		{/if}
	</span>
</div>
<p class="muted lead">{t.lead}</p>
{#if models.providers.length}
	<!-- The providers with a key, each with its model list to fetch here,
	     before any form needs it; a fetched list is offered as you type. -->
	<div class="providers" role="group" aria-label={t.providers}>
		<span class="caps muted group" aria-hidden="true">{t.providers}</span>
		{#each ready as p (p.id)}
			<span class="item">
				<span class="name">{p.name}</span>
				<span class="muted small">{listState(p.id)}</span>
				<button type="button" class="textbutton small" onclick={() => modelLists.fetch(p.id, !!modelLists.lists[p.id])} disabled={modelLists.loading[p.id]}>
					{modelLists.lists[p.id] ? t.lists.fetchAgain : t.lists.fetch}
				</button>
			</span>
		{/each}
		<a class="small" href={providersHref(save)}>
			{#if !ready.length}
				{t.noKeyYet}
			{:else if waiting.length}
				{t.moreWithoutKey({ count: waiting.length })}
			{:else}
				{t.allProviders}
			{/if}
		</a>
		{#each ready.filter((p) => modelLists.errors[p.id]) as p (p.id)}
			<span class="notice error small" role="alert">{p.name}: {modelLists.errors[p.id]}</span>
		{/each}
	</div>
{/if}
{#if !models.found}
	<p class="notice"><Fill text={m.common.notWrittenYet.text}>{#snippet path()}<span class="id">{models.path}</span>{/snippet}</Fill></p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label={m.common.selected}></th>
					<th class="caps task">{t.task}</th>
					<th class="caps">{t.title}</th>
					<th class="caps provider">{t.provider}</th>
					<th class="caps order">{t.order}</th>
				</tr>
			</thead>
			<tbody>
				{#each models.tasks as task (task.id)}
					<tr aria-selected={selection.is(task)} onclick={() => mode === 'view' && selection.select(task)} ondblclick={() => can && mode === 'view' && setParams({ mode: 'edit' })}>
						<td class="mark">{#if selection.is(task)}<Diamond />{/if}</td>
						<td class="task">{taskLabel(task.id)}{#if !task.known}<span class="muted small"> · {t.plugin}</span>{/if}</td>
						<td class="models">
							<div class="cell">
								<span class="line id">{task.slots.map((s) => shortModel(s.model)).join(', ') || '—'}</span>
								{#if task.slots.some((s) => s.role === 'backup')}<span class="flag">{t.backup}</span>{/if}
								{#if task.slots.some((s) => !s.enabled)}<span class="flag">{t.oneOut}</span>{/if}
							</div>
						</td>
						<td class="provider muted">{providersOf(task)}</td>
						<td class="order muted">{task.slots.length > 1 ? order(task) : ''}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">
			<Fill text={t.note.text}>
				{#snippet providers()}<a href={providersHref(save)}>{t.providers}</a>{/snippet}
				{#snippet path()}<span class="id">{models.path}</span>{/snippet}
				{#snippet openRouter()}<a href={settingsHref(save, 'OpenRouter')}>OpenRouter</a>{/snippet}
			</Fill>
		</p>
	</div>

	<aside class="pane" use:fitToViewport aria-label={t.task}>
		{#if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{taskLabel(current.id)}</h2>
					{#if taskWhat(current.id)}
						<p class="muted small">{taskWhat(current.id)}</p>
					{:else}
						{@const by = addedBy(current.id)}
						<p class="muted small">
							<Fill text={t.addedBy.text}>
								{#snippet plugin()}{#if by}<a href={pluginHref(save, by.plugin ?? '')}>{by.title}</a>{:else}{t.aPlugin}{/if}{/snippet}
								{#snippet id()}<span class="id">{current.id}</span>{/snippet}
							</Fill>
						</p>
					{/if}
					{#if mode === 'edit'}
						<TaskForm task={current} modified={models.modified} providers={models.providers} {choices} {saving} onsave={submit} oncancel={() => leave('mode')} />
					{:else}
						{#if current.slots.length}
							<ol class="slots">
								{#each current.slots as s, i (i)}
									<li class:off={!s.enabled}>
										<span class="id model">{s.model}</span>
										<span class="muted small">{provider(s.provider)}{slotNotes(s, current) ? ` · ${slotNotes(s, current)}` : ''}</span>
									</li>
								{/each}
							</ol>
						{:else}
							<p class="muted">{t.noOwnModel}</p>
						{/if}
						<dl>
							{#if current.slots.length > 1}<dt class="caps">{t.order}</dt><dd>{!current.rotationMode && current.id !== 'default' ? t.orderAsDialogue({ order: order(current) }) : order(current)}</dd>{/if}
							{#if current.id !== 'decisions'}
								<dt class="caps">{t.maxTokens}</dt><dd>{current.params.maxTokens ?? '—'}</dd>
								{#if current.params.topP !== undefined}<dt class="caps">{t.topP}</dt><dd>{current.params.topP}</dd>{/if}
								{#if current.params.topK !== undefined}<dt class="caps">{t.topK}</dt><dd>{current.params.topK}</dd>{/if}
								{#if current.params.frequencyPenalty !== undefined}<dt class="caps">{t.frequencyPenalty}</dt><dd>{current.params.frequencyPenalty}</dd>{/if}
								{#if current.params.presencePenalty !== undefined}<dt class="caps">{t.presencePenalty}</dt><dd>{current.params.presencePenalty}</dd>{/if}
								{#if current.params.stopSequences.length}<dt class="caps">{t.stopsAt}</dt><dd class="id">{current.params.stopSequences.map((s) => JSON.stringify(s)).join(' ')}</dd>{/if}
							{/if}
							<dt class="caps">{t.timeout}</dt><dd>{current.timeout !== undefined ? t.seconds({ seconds: current.timeout }) : current.id === 'default' ? '—' : dialogue?.timeout !== undefined ? t.asDialogueSeconds({ seconds: dialogue.timeout }) : t.asDialogue}</dd>
						</dl>
						{#if can}
							<div class="actions"><button type="button" class="button" onclick={() => setParams({ mode: 'edit' })}>{m.common.edit}</button></div>
						{/if}
						{#if current.id !== 'decisions'}<ModelTest task={current} providers={models.providers} />{/if}
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
		margin-right: calc(clamp(22rem, 34vw, 32rem) + var(--space-5));
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
	/* The providers run in one row, as the quick toggles do, each with
	   its model list's state and the one action on it. */
	.providers {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-1) var(--space-5);
		margin-bottom: var(--space-4);
	}
	.providers .group {
		font-size: var(--step--1);
	}
	.providers .item {
		display: inline-flex;
		align-items: baseline;
		gap: var(--space-2);
		white-space: nowrap;
	}
	.providers .notice {
		flex-basis: 100%;
		margin: 0;
	}
	/* The pane is wider here: a rotation is a form of several fields per model. */
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(22rem, 34vw, 32rem);
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
	.task {
		width: 11rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.provider {
		width: 8rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.order {
		width: 6.5rem;
		white-space: nowrap;
	}
	/* The models take what's left, clipped, and a flag follows right after. */
	.models {
		overflow: hidden;
	}
	.models .cell {
		display: flex;
		align-items: baseline;
	}
	.models .line {
		flex: 0 1 auto;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	/* A flag is a word of metadata, as in the other tables: never a button. */
	.flag {
		flex: none;
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.flag::before {
		content: '· ';
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
	.pane > div > .small {
		margin-bottom: var(--space-4);
	}
	.slots {
		list-style: none;
		padding: 0;
		margin: 0 0 var(--space-4);
		counter-reset: slot;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.slots li {
		display: grid;
		grid-template-columns: 1.5rem 1fr;
		gap: 0 var(--space-2);
	}
	.slots li::before {
		counter-increment: slot;
		content: counter(slot);
		color: var(--muted);
		font-size: var(--step--1);
		font-variant-numeric: tabular-nums;
		grid-row: span 2;
		padding-top: 0.15em;
	}
	.slots li.off .model {
		color: var(--muted);
		text-decoration: line-through;
	}
	.model {
		overflow-wrap: anywhere;
	}
	.pane dl {
		display: grid;
		grid-template-columns: 9rem 1fr;
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
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.actions {
		display: flex;
		gap: var(--space-2);
	}
	/* Narrower than 90rem the models need the room; the provider and the
	   order are in the pane anyway. */
	@media (max-width: 90rem) {
		.provider,
		.order {
			display: none;
		}
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
