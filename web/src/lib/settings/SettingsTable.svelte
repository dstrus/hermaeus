<!-- The settings of one file (or everything the player changed, across
     files) as a table, with the selected one edited in the pane. -->
<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { tick } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { File } from '#lib/api/types/settings.ts';
	import { count } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { editable, fileLabel, kindOf, saveSettings, settingsDependency, settingsHref, show, togglesDependency, words, type Kind, type Row } from './api.ts';
	import { codeOf, keyName, mouseCodeOf, noKey } from './keys.ts';

	let {
		save,
		title,
		file,
		rows,
		byFile
	}: {
		save: string;
		title: string;
		file: File | null; // the one file shown, or null for the changed page
		rows: Row[];
		byFile: boolean; // group rows under their file rather than their key's first part
	} = $props();

	let query = $state('');
	let error = $state('');
	let notice = $state('');
	let readOnly = $derived(live.gameRunning);

	// The group a row sits under: its file, or the part of its key before the first dot.
	const groupOf = (r: Row) => (byFile ? fileLabel(r.file) : r.setting.key.includes('.') ? r.setting.key.slice(0, r.setting.key.indexOf('.')) : '');
	const rowKey = (r: Row) => `${r.file.name}\u0000${r.setting.key}`;
	// Hotkey.yaml holds key codes; they're shown and edited as keys.
	const valueOf = (r: Row) => (r.setting.set ? r.setting.value : r.setting.default);
	const hotkey = (r: Row) => r.file.name.toLowerCase() === 'hotkey' && typeof valueOf(r) === 'number';
	const showValue = (r: Row, v: unknown) => (hotkey(r) && typeof v === 'number' ? keyName(v) : show(v));

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return rows.filter((r) => !q || r.setting.key.toLowerCase().includes(q) || showValue(r, valueOf(r)).toLowerCase().includes(q));
	});
	const selection = new Selection(() => shown, rowKey);
	let current = $derived(selection.current);
	let kind = $derived<Kind>(!current ? 'none' : hotkey(current) ? 'hotkey' : kindOf(current.setting));
	// What's holding a secret can't be changed here; a file SkyrimNet hasn't
	// written yet can't be either.
	let locked = $derived(!current || kind === 'secret' || current.setting.hasSecrets || !current.file.exists);
	let editableNow = $derived(!!current && !locked && !readOnly && kind !== 'none');
	let resettable = $derived(!!current && !locked && !readOnly && current.setting.hasDefault && (current.setting.changed || !current.setting.set));

	let editing = $state(false);
	let draft = $state('');
	let saving = $state(false);
	let field = $state<HTMLTextAreaElement | HTMLInputElement | HTMLButtonElement>();

	let list = $derived(settingsHref(save));

	async function startEdit() {
		if (!editableNow || !current) return;
		const s = current.setting;
		draft = editable(s.set ? s.value : s.default);
		error = '';
		notice = '';
		editing = true;
		await tick();
		field?.focus();
	}

	// parse turns the form's text into the value the setting takes.
	function parse(): unknown {
		switch (kind) {
			case 'switch':
				return draft === 'true';
			case 'number':
			case 'hotkey': {
				const n = Number(draft.trim());
				if (draft.trim() === '' || Number.isNaN(n)) throw new Error('Enter a number.');
				return n;
			}
			case 'text':
				return draft;
			case 'list':
			case 'group': {
				let v: unknown;
				try {
					v = JSON.parse(draft);
				} catch {
					throw new Error(`Enter the ${kind === 'list' ? 'list' : 'group'} as JSON.`);
				}
				if (kind === 'list' ? !Array.isArray(v) : typeof v !== 'object' || v === null || Array.isArray(v)) throw new Error(`This setting is a ${kind}.`);
				return v;
			}
		}
		throw new Error('This setting can’t be edited here.');
	}

	async function apply(change: { key: string; value?: unknown; reset?: boolean }, done: string) {
		if (!current || saving) return;
		saving = true;
		error = '';
		try {
			await saveSettings(current.file.name, { modified: current.file.modified, changes: [change] });
			notice = done;
			editing = false;
			await invalidate(settingsDependency);
			await invalidate(togglesDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function submit() {
		if (!current) return;
		let value: unknown;
		try {
			value = parse();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
			return;
		}
		apply({ key: current.setting.key, value }, `Saved ${current.setting.key}.`);
	}

	// A hotkey is set by pressing it in the capture box. Tab, Esc, and the
	// save shortcut keep their jobs; a mouse button other than the left one
	// (which focuses the box) counts too.
	function captureKey(e: KeyboardEvent) {
		if (e.key === 'Tab' || e.key === 'Escape' || (e.key === 'Enter' && (e.metaKey || e.ctrlKey))) return;
		const code = codeOf(e);
		e.preventDefault();
		e.stopPropagation();
		if (code !== undefined) draft = String(code);
	}
	function captureMouse(e: MouseEvent) {
		const code = mouseCodeOf(e);
		if (e.button === 0 || code === undefined) return;
		e.preventDefault();
		draft = String(code);
	}

	function reset() {
		if (!current || !resettable) return;
		apply({ key: current.setting.key, reset: true }, `${current.setting.key} is back to its default.`);
	}

	useShortcuts(() => [
		{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: editing, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: editing, run: () => selection.move(-1) },
		{ key: 'e', shown: 'E', label: 'Edit', hidden: editing, disabled: !editableNow, run: startEdit },
		{ key: 'r', shown: 'R', label: 'Reset to default', hidden: editing, disabled: !resettable, run: reset },
		{ key: 'Enter', label: 'Save', command: true, hidden: !editing, disabled: saving, run: submit },
		{ key: 'Escape', shown: 'Esc', label: editing ? 'Cancel' : 'Settings', run: () => (editing ? (editing = false) : goto(list)) }
	]);
</script>

<div class="bar">
	<h1 class="caps">{title}</h1>
	{#if file}
		<span class="path n">{file.path}</span>
	{:else}
		<span class="muted n">{count(rows.length)} across {new Set(rows.map((r) => r.file.name)).size} files</span>
	{/if}
	<input type="search" bind:value={query} placeholder="Filter by key or value" aria-label="Filter settings" />
</div>
{#if file && !file.exists}
	<p class="notice warn">SkyrimNet hasn’t created this file yet, so these are its defaults; they can be changed once it has.</p>
{/if}

{#if !rows.length}
	<p class="muted">{file ? 'This file has no settings.' : 'Nothing differs from the defaults.'}</p>
{:else if !shown.length}
	<p class="muted">Nothing matches “{query}”.</p>
{:else}
	<div class="split">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<th class="caps">Setting</th>
					<th class="caps">Value</th>
					<th class="caps default">Default</th>
				</tr>
			</thead>
			<tbody>
				{#each shown as r, i (rowKey(r))}
					{#if groupOf(r) && (i === 0 || groupOf(r) !== groupOf(shown[i - 1]!))}
						<tr class="group"><th colspan="4" class="caps" scope="rowgroup">{byFile ? groupOf(r) : words(groupOf(r))}</th></tr>
					{/if}
					<tr aria-selected={selection.is(r)} onclick={() => selection.select(r)} ondblclick={startEdit}>
						<td class="mark">{#if selection.is(r)}<Diamond />{/if}</td>
						<td class="key">{byFile ? r.setting.key : r.setting.key.slice(groupOf(r) ? groupOf(r).length + 1 : 0)}</td>
						<td class="value" class:changed={r.setting.changed} class:muted={!r.setting.set}>{showValue(r, valueOf(r))}</td>
						<td class="default muted">{r.setting.hasDefault ? showValue(r, r.setting.default) : ''}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" aria-label="Setting" use:fitToViewport>
			{#if current}
				<h2 class="key">{current.setting.key}</h2>
				<p class="muted small">{byFile ? fileLabel(current.file) : current.file.path}</p>
				{#if error}<p class="notice error" role="alert">{error}</p>{/if}
				{#if notice}<p class="notice" role="status">{notice}</p>{/if}
				{#if editing}
					<form class="edit" onsubmit={(e) => (e.preventDefault(), submit())}>
						{#if kind === 'switch'}
							<Segmented bind:value={draft} options={[{ value: 'true', label: 'On' }, { value: 'false', label: 'Off' }]} label="Value" />
						{:else if kind === 'number'}
							<input bind:this={field} bind:value={draft} type="text" inputmode="decimal" aria-label="Value" autocomplete="off" />
						{:else if kind === 'hotkey'}
							<button
								type="button"
								class="capture"
								bind:this={field}
								aria-label="Press the key to use"
								onkeydown={captureKey}
								onmousedown={captureMouse}
								oncontextmenu={(e) => e.preventDefault()}>{keyName(Number(draft))}</button
							>
							<p class="muted small">Press a key, or click the box with a mouse button.</p>
							<button type="button" class="textbutton" onclick={() => (draft = String(noKey))}>No key</button>
						{:else}
							<textarea bind:this={field} bind:value={draft} rows={kind === 'text' ? 4 : 12} spellcheck="false" aria-label="Value"></textarea>
						{/if}
						{#if kind === 'list' || kind === 'group'}<p class="muted small">As JSON.</p>{/if}
						<div class="buttons">
							<button type="submit" class="button primary" disabled={saving}>Save</button>
							<button type="button" class="button" onclick={() => (editing = false)} disabled={saving}>Cancel</button>
						</div>
					</form>
				{:else}
					<dl>
						<dt class="caps">Value</dt>
						<dd class:changed={current.setting.changed}><pre>{showValue(current, valueOf(current))}</pre>{#if !current.setting.set}<span class="muted"> (the default; the file doesn’t set it)</span>{/if}</dd>
						{#if current.setting.hasDefault}
							<dt class="caps">Default</dt><dd><pre>{showValue(current, current.setting.default)}</pre></dd>
						{/if}
					</dl>
					{#if kind === 'secret'}
						<p class="notice">API keys and other secrets are never shown or changed here. Use SkyrimNet’s dashboard.</p>
					{:else if current.setting.hasSecrets}
						<p class="notice">This holds API keys inside it, so it can’t be changed here. Use SkyrimNet’s dashboard.</p>
					{:else if readOnly}
						<p class="notice warn">Close Skyrim to change settings.</p>
					{/if}
					<div class="buttons">
						<button type="button" class="button primary" disabled={!editableNow} onclick={startEdit}>Edit</button>
						{#if current.setting.hasDefault}
							<button type="button" class="button" disabled={!resettable} onclick={reset}>Reset to default</button>
						{/if}
					</div>
				{/if}
			{/if}
		</aside>
	</div>
{/if}

<style>
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
		margin-right: calc(clamp(20rem, 28vw, 26rem) + var(--space-5));
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	.bar .n {
		white-space: nowrap;
	}
	.path {
		font: 400 var(--step--1) / 1.5 var(--font-mono);
		color: var(--muted);
	}
	.bar input {
		margin-left: auto;
		width: 18rem;
	}
	.small {
		font-size: var(--step--1);
	}
	.changed {
		color: var(--changed);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(20rem, 28vw, 26rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
	}
	th {
		text-align: left;
		padding: var(--space-2) var(--space-3);
		font-size: var(--step--1);
		color: var(--muted);
		font-weight: inherit;
		border-bottom: 1px solid var(--rule-strong);
	}
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	tbody tr[aria-selected]:not([aria-selected='true']):hover td {
		background: var(--glass);
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	tbody tr {
		cursor: default;
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
		border-bottom: 1px solid var(--rule-strong);
	}
	tbody .mark {
		border-bottom-color: var(--rule);
	}
	.key {
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	th.default,
	td.default {
		width: 9rem;
	}
	/* The pane: the setting itself, read or edited; sized to end at the hint strip. */
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	.pane h2 {
		font-size: var(--step-0);
		font-weight: 500;
		margin-bottom: var(--space-1);
		overflow-wrap: anywhere;
	}
	.pane dl {
		margin: var(--space-4) 0;
	}
	.pane dt {
		color: var(--muted);
		font-size: var(--step--1);
	}
	.pane dd {
		margin: 0 0 var(--space-3);
	}
	.pane pre {
		display: inline;
		margin: 0;
		font: inherit;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.edit {
		margin-top: var(--space-4);
	}
	.edit input,
	.edit textarea {
		display: block;
		width: 100%;
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	/* The box a hotkey is pressed into: a field that listens rather than types. */
	.capture {
		display: block;
		width: 100%;
		padding: var(--space-2) var(--space-3);
		background: var(--glass-strong);
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		text-align: left;
		cursor: pointer;
	}
	.capture:focus-visible {
		border-color: var(--accent);
		outline: none;
	}
	.edit .small {
		margin-top: var(--space-2);
	}
	.edit .textbutton {
		margin-top: var(--space-1);
		font-size: var(--step--1);
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-3);
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
		th.default,
		td.default {
			display: none;
		}
	}
</style>
