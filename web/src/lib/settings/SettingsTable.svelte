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
	import ListEditor from './ListEditor.svelte';
	import MapEditor from './MapEditor.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { editable, fileLabel, fromEntries, kindOf, parseJSONText, saveSettings, settingsDependency, settingsHref, show, toEntries, togglesDependency, words, type Entry, type Kind, type Row } from './api.ts';
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
	// A file's settings sit under a heading per group: everything before the
	// key's last part, so openrouter.variants.vision.timeout is the field
	// "timeout" under "Openrouter › Variants › Vision". A root every key
	// shares is left out of the headings. The changed page groups by file.
	const parentOf = (key: string) => (key.includes('.') ? key.slice(0, key.lastIndexOf('.')) : '');
	const lastOf = (key: string) => key.slice(key.lastIndexOf('.') + 1);
	const groupOf = (r: Row) => (byFile ? fileLabel(r.file) : parentOf(r.setting.key));
	let root = $derived.by(() => {
		const first = rows[0]?.setting.key.split('.')[0] ?? '';
		return !byFile && first && rows.every((r) => r.setting.key.startsWith(first + '.')) ? first : '';
	});
	function heading(group: string): string {
		if (byFile) return group;
		let parts = group.split('.').filter(Boolean);
		if (root && parts[0] === root) parts = parts.slice(1);
		return parts.length ? parts.map(words).join(' › ') : words(root);
	}
	const rowKey = (r: Row) => `${r.file.name}\u0000${r.setting.key}`;
	// Hotkey.yaml holds key codes; they're shown and edited as keys.
	const valueOf = (r: Row) => (r.setting.set ? r.setting.value : r.setting.default);
	const hotkey = (r: Row) => r.file.name.toLowerCase() === 'hotkey' && typeof valueOf(r) === 'number';
	const showValue = (r: Row, v: unknown) => (hotkey(r) && typeof v === 'number' ? keyName(v) : show(v));

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const found = rows.filter((r) => !q || r.setting.key.toLowerCase().includes(q) || showValue(r, valueOf(r)).toLowerCase().includes(q));
		// Settings the file doesn't set come last from the server; each
		// goes under its group's heading instead, where the group first appears.
		const order = new Map<string, number>();
		for (const r of found) if (!order.has(groupOf(r))) order.set(groupOf(r), order.size);
		return found.toSorted((a, b) => order.get(groupOf(a))! - order.get(groupOf(b))!);
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
	// A list or map is edited by its shape: one line per item for a list of
	// words or numbers, a form per item for a list of records, key and value
	// rows for a map, and JSON for the rest or on request. A text setting
	// that holds JSON is edited the same way and saved back as JSON text.
	type Scalar = string | number | boolean | null;
	type Shape = 'plain' | 'lines' | 'records' | 'map' | 'json';
	let shape = $state<Shape>('plain');
	let asText = $state(false); // the setting is text holding JSON
	let lines = $state('');
	let numbers = $state(false); // the list held numbers, so lines are read as numbers
	let records = $state<Record<string, Scalar>[]>([]);
	let entries = $state<Entry[]>([]);
	const isScalar = (v: unknown): v is Scalar => v === null || ['string', 'number', 'boolean'].includes(typeof v);
	const isMap = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null && !Array.isArray(v);
	function shapeOf(v: unknown[]): Shape {
		if (v.every(isScalar)) return 'lines'; // an empty list too: one per line is the likeliest
		if (v.every((x) => isMap(x) && Object.values(x).every(isScalar))) return 'records';
		return 'json';
	}
	// toFields sets the editors up for a list or map value.
	function toFields(v: unknown) {
		shape = 'json';
		if (Array.isArray(v)) {
			shape = shapeOf(v);
			numbers = v.length > 0 && v.every((x) => typeof x === 'number');
			lines = shape === 'lines' ? v.map((x) => (x === null ? '' : String(x))).join('\n') : '';
			records = shape === 'records' ? (v as Record<string, Scalar>[]).map((x) => ({ ...x })) : [];
		} else if (isMap(v)) {
			shape = 'map';
			entries = toEntries(v);
		}
	}
	// fromFields is the value the editors hold.
	function fromFields(): unknown {
		switch (shape) {
			case 'lines': {
				const items = lines.split('\n').map((l) => l.trim()).filter(Boolean);
				if (!numbers) return items;
				return items.map((l) => {
					const n = Number(l);
					if (Number.isNaN(n)) throw new Error(`“${l}” isn’t a number; this list holds numbers.`);
					return n;
				});
			}
			case 'records': {
				// A number field left as text didn't parse (the editor keeps
				// numbers as numbers); an empty one means no value.
				const numeric = new Set(records.flatMap((r) => Object.keys(r).filter((k) => typeof r[k] === 'number')));
				return records.map((r) => {
					const out: Record<string, Scalar> = {};
					for (const [key, v] of Object.entries(r)) {
						if (numeric.has(key) && typeof v === 'string') {
							if (v.trim() !== '') throw new Error(`${key} holds numbers; “${v}” isn’t one.`);
							out[key] = null;
						} else out[key] = v;
					}
					return out;
				});
			}
			case 'map':
				return fromEntries(entries);
		}
		let v: unknown;
		try {
			v = JSON.parse(draft);
		} catch {
			throw new Error('Enter valid JSON.');
		}
		if (kind === 'list' && !Array.isArray(v)) throw new Error('This setting is a list.');
		if (kind === 'group' && !isMap(v)) throw new Error('This setting is a group of settings.');
		return v;
	}
	// Switching between the fields and the JSON text keeps what's been typed.
	function showAsText() {
		try {
			draft = asText ? JSON.stringify(fromFields()) : JSON.stringify(fromFields(), null, 2);
		} catch {
			// keep the draft as it was
		}
		shape = asText ? 'plain' : 'json';
	}
	function showAsFields() {
		try {
			const v = asText ? parseJSONText(draft) : fromFields();
			if (v === undefined) throw new Error('The text isn’t a JSON list or map.');
			toFields(v);
			error = '';
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}
	let field = $state<HTMLTextAreaElement | HTMLInputElement | HTMLButtonElement>();

	let list = $derived(settingsHref(save));

	async function startEdit() {
		if (!editableNow || !current) return;
		const s = current.setting;
		const v = s.set ? s.value : s.default;
		draft = editable(v);
		shape = 'plain';
		asText = false;
		if (kind === 'list' || kind === 'group') toFields(v);
		else if (kind === 'text') {
			const held = parseJSONText(v);
			if (held !== undefined) {
				asText = true;
				toFields(held);
			}
		}
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
				return asText && shape !== 'plain' ? JSON.stringify(fromFields()) : draft;
			case 'list':
			case 'group':
				return fromFields();
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
					{#if (groupOf(r) || root) && (i === 0 || groupOf(r) !== groupOf(shown[i - 1]!))}
						<tr class="group"><th colspan="4" class="caps" scope="rowgroup">{heading(groupOf(r))}</th></tr>
					{/if}
					<tr aria-selected={selection.is(r)} onclick={() => selection.select(r)} ondblclick={startEdit}>
						<td class="mark">{#if selection.is(r)}<Diamond />{/if}</td>
						<td class="key">{byFile ? r.setting.key : lastOf(r.setting.key)}</td>
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
						{:else if shape === 'lines'}
							<textarea bind:this={field} bind:value={lines} rows={Math.min(16, Math.max(4, lines.split('\n').length + 1))} spellcheck="false" aria-label="Items, one per line"></textarea>
							<p class="muted small">One {numbers ? 'number' : 'item'} per line. <button type="button" class="textbutton" onclick={showAsText}>Edit as {asText ? 'text' : 'JSON'}</button></p>
						{:else if shape === 'records'}
							<ListEditor bind:items={records} />
							<p class="muted small"><button type="button" class="textbutton" onclick={showAsText}>Edit as {asText ? 'text' : 'JSON'}</button></p>
						{:else if shape === 'map'}
							<MapEditor bind:entries />
							<p class="muted small"><button type="button" class="textbutton" onclick={showAsText}>Edit as {asText ? 'text' : 'JSON'}</button></p>
						{:else}
							<textarea bind:this={field} bind:value={draft} rows={kind === 'text' && !asText ? 4 : 12} spellcheck="false" aria-label="Value"></textarea>
							{#if shape === 'json' || asText}
								<p class="muted small">As JSON. <button type="button" class="textbutton" onclick={showAsFields}>Edit as fields</button></p>
							{/if}
						{/if}
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
	.edit .small .textbutton {
		font-size: inherit;
		margin: 0;
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
