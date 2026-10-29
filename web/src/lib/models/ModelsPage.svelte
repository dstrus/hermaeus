<!-- Models: which model does each of SkyrimNet's tasks. Every task has
     its own rotation; the selected task's is in the pane, and Edit opens it
     as a form. Each save writes OpenRouter.yaml with a backup. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Models, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';
	import { plural } from '#lib/format.ts';
	import { settingsHref } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { modelChoices, modelsDependency, providerName, rotation, saveTask } from './api.ts';
	import TaskForm from './TaskForm.svelte';

	let { save, models }: { save: string; models: Models } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(models.found && !readOnly);
	let choices = $derived(modelChoices(models));
	let dialogue = $derived(models.tasks.find((t) => t.id === 'default'));
	const provider = (id: string) => providerName(models.providers, id);
	// The providers a task's slots use, named once each.
	const providersOf = (t: Task) => [...new Set(t.slots.map((s) => provider(s.provider)))].join(', ');
	const order = (t: Task) => (t.rotationMode || dialogue?.rotationMode || 'sequential') === 'random' ? 'At random' : 'In turn';
	const shortModel = (m: string) => m.replace(/^[^/]+\//, '');

	const selection = new Selection(() => models.tasks, (t) => t.id, { param: 'item' });
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
		const t = current;
		if (!t || saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveTask(t.id, edit);
			notice = `Saved the models for ${t.label}.`;
			await leave('mode'); // settles before the reload, which it would cancel
			await invalidate(modelsDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	// slotNotes are what a slot sets for itself, as a short line.
	function slotNotes(s: Slot, t: Task): string {
		const out: string[] = [];
		if (s.role === 'backup') out.push('backup');
		if (!s.enabled) out.push('out of rotation');
		if (s.temperature !== undefined) out.push(`temperature ${s.temperature}`);
		else if (t.params.temperature !== undefined) out.push(`temperature ${t.params.temperature}`);
		if (s.weight !== undefined) out.push(`weight ${s.weight}`);
		if (s.requestTimeout !== undefined) out.push(`${s.requestTimeout}s`);
		if (s.maxRetries !== undefined) out.push(plural(s.maxRetries, 'retry', 'retries'));
		if (s.reasoningEffort) out.push(`reasoning ${s.reasoningEffort}`);
		if (s.structuredOutputs !== undefined) out.push(s.structuredOutputs ? 'structured outputs' : 'no structured outputs');
		return out.join(' · ');
	}

	useShortcuts(() => {
		const viewing = mode === 'view';
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !viewing, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !viewing, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !viewing, disabled: !can || !current, run: () => setParams({ mode: 'edit' }) },
			{ key: 'Enter', label: 'Save', command: true, hidden: viewing, run: () => document.querySelector<HTMLFormElement>('.pane form')?.requestSubmit() },
			...(viewing ? [] : [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => leave('mode') }])
		];
	});

	let keyed = $derived(models.providers.filter((p) => p.hasKey).map((p) => p.name));
</script>

<div class="bar">
	<h1 class="caps">Models</h1>
	<span class="muted n" use:clipped>
		{#if !models.found}
			Not set up yet
		{:else}
			{plural(models.tasks.length, 'task', 'tasks')}{keyed.length ? ` · keys for ${keyed.join(', ')}` : ' · no provider has a key yet'}
		{/if}
	</span>
</div>
<p class="muted lead">Which model does each of SkyrimNet’s jobs. A task tries its models in rotation, each with its own provider and settings; a task without a model of its own uses Dialogue’s.</p>
{#if !models.found}
	<p class="notice">SkyrimNet hasn’t written <span class="id">{models.path}</span> yet; it will the next time it starts. Until then there’s nothing to change here.</p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<th class="caps task">Task</th>
					<th class="caps">Models</th>
					<th class="caps provider">Provider</th>
					<th class="caps order">Order</th>
				</tr>
			</thead>
			<tbody>
				{#each models.tasks as t (t.id)}
					<tr aria-selected={selection.is(t)} onclick={() => mode === 'view' && selection.select(t)} ondblclick={() => can && mode === 'view' && setParams({ mode: 'edit' })}>
						<td class="mark">{#if selection.is(t)}<Diamond />{/if}</td>
						<td class="task">{t.label}{#if !t.known}<span class="muted small"> · plugin</span>{/if}</td>
						<td class="models">
							<div class="cell">
								<span class="line id">{t.slots.map((s) => shortModel(s.model)).join(', ') || '—'}</span>
								{#if t.slots.some((s) => s.role === 'backup')}<span class="flag">backup</span>{/if}
								{#if t.slots.some((s) => !s.enabled)}<span class="flag">one out</span>{/if}
							</div>
						</td>
						<td class="provider muted">{providersOf(t)}</td>
						<td class="order muted">{t.slots.length > 1 ? order(t) : ''}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">Providers and their keys, and the rest of <span class="id">{models.path}</span>, are under <a href={settingsHref(save, 'Providers')}>Providers</a> and <a href={settingsHref(save, 'OpenRouter')}>OpenRouter</a> in Settings.</p>
	</div>

	<aside class="pane" use:fitToViewport aria-label="Task">
		{#if current}
			{#key `${current.id}:${mode}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{current.label}</h2>
					<p class="muted small">{current.what || `A task added by a plugin, as ${current.id}.`}</p>
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
							<p class="muted">No model of its own: Dialogue’s rotation answers.</p>
						{/if}
						<dl>
							{#if current.slots.length > 1}<dt class="caps">Order</dt><dd>{order(current)}{!current.rotationMode && current.id !== 'default' ? ', as Dialogue' : ''}</dd>{/if}
							{#if current.id !== 'decisions'}
								<dt class="caps">Max tokens</dt><dd>{current.params.maxTokens ?? '—'}</dd>
								{#if current.params.topP !== undefined}<dt class="caps">Top-p</dt><dd>{current.params.topP}</dd>{/if}
								{#if current.params.topK !== undefined}<dt class="caps">Top-k</dt><dd>{current.params.topK}</dd>{/if}
								{#if current.params.frequencyPenalty !== undefined}<dt class="caps">Frequency penalty</dt><dd>{current.params.frequencyPenalty}</dd>{/if}
								{#if current.params.presencePenalty !== undefined}<dt class="caps">Presence penalty</dt><dd>{current.params.presencePenalty}</dd>{/if}
								{#if current.params.stopSequences.length}<dt class="caps">Stops at</dt><dd class="id">{current.params.stopSequences.map((s) => JSON.stringify(s)).join(' ')}</dd>{/if}
							{/if}
							<dt class="caps">Timeout</dt><dd>{current.timeout !== undefined ? `${current.timeout}s` : current.id === 'default' ? '—' : `as Dialogue${dialogue?.timeout !== undefined ? ` (${dialogue.timeout}s)` : ''}`}</dd>
						</dl>
						{#if can}
							<div class="actions"><button type="button" class="button" onclick={() => setParams({ mode: 'edit' })}>Edit</button></div>
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
