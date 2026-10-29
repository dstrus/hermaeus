<!-- A feature's settings, in sections: the selected one in the pane with
     what it does, and an editor on request. Each change saves at once,
     with a backup; the selected setting and the editor are in the address. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { tick } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { ListsResponse } from '#lib/api/types/filters.ts';
	import type { FileValues, Setting } from '#lib/api/types/settings.ts';
	import FilterLists from './FilterLists.svelte';
	import { count } from '#lib/format.ts';
	import { saveSettings, settingsDependency, settingsHref, togglesDependency } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { codeOf, keyName, mouseCodeOf, noKey } from '#lib/settings/keys.ts';
	import { draftOf, parse, pickedOf, setValue, show, valueOf, type Feature, type Field } from './schema.ts';

	let { save, feature, file, lists }: { save: string; feature: Feature; file: FileValues; lists?: ListsResponse } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(file.file.exists && !readOnly);
	let settings = $derived(new Map(file.values.map((s) => [s.key.toLowerCase(), s])));
	const settingOf = (f: Field): Setting | undefined => settings.get(f.key.toLowerCase());

	// The page's rows: every field, in its section.
	type Row = { field: Field; section: string; setting?: Setting };
	let rows = $derived<Row[]>(feature.sections.flatMap((s) => s.fields.map((field) => ({ field, section: s.title, setting: settingOf(field) }))));
	let changed = $derived(rows.filter((r) => r.setting?.changed).length);

	const selection = new Selection(() => rows, (r) => r.field.key, { param: 'item' });
	let current = $derived(selection.current);
	let value = $derived(current ? valueOf(current.setting) : undefined);
	let resettable = $derived(!!current?.setting?.hasDefault && (current.setting.changed || !current.setting.set));

	let notice = $state('');
	let error = $state('');
	let saving = $state(false);

	// Editing is a state of the address (?mode=edit); a switch has no form,
	// it flips on E.
	let editing = $derived(param('mode') === 'edit' && !!current && can && current.field.kind !== 'switch');
	$effect(() => {
		if (!editing && param('mode') === 'edit') setParams({ mode: null }, 'replace');
	});
	// A filter list's form is open (see FilterLists), so the keys are its.
	let listForm = $derived(param('mode') === 'list' || param('mode') === 'list-new');
	let draft = $state('');
	let picked = $state<string[]>([]); // a set's chosen values
	let field = $state<HTMLInputElement | HTMLSelectElement | HTMLButtonElement>();
	$effect(() => {
		if (!editing) return;
		draft = draftOf(current!.field, value);
		picked = pickedOf(current!.field, value);
		tick().then(() => field?.focus());
	});
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
	function pick(v: string, on: boolean) {
		picked = on ? [...picked, v] : picked.filter((p) => p !== v);
	}

	async function apply(change: { key: string; value?: unknown; reset?: boolean }, done: string) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveSettings(feature.file, { modified: file.file.modified, changes: [change] });
			notice = done;
			await leave('mode'); // settles before the reload, which it would cancel
			await Promise.all([invalidate(settingsDependency), invalidate(togglesDependency)]);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function startEdit() {
		if (!can || !current) return;
		if (current.field.kind === 'switch') {
			const on = value !== true;
			apply({ key: current.field.key, value: on }, `${current.field.label}: ${on ? 'on' : 'off'}.`);
			return;
		}
		setParams({ mode: 'edit' });
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!current) return;
		let v: unknown;
		try {
			v = current.field.kind === 'set' ? setValue(current.field, value, picked) : parse(current.field, draft);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			return;
		}
		apply({ key: current.field.key, value: v }, `${current.field.label}: ${show(current.field, v)}.`);
	}

	function reset() {
		if (!current || !resettable) return;
		apply({ key: current.field.key, reset: true }, `${current.field.label} is back to its default.`);
	}


	useShortcuts(() => {
		const busy = editing || listForm;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: busy, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: busy, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: current?.field.kind === 'switch' ? (value === true ? 'Switch off' : 'Switch on') : 'Edit', hidden: busy, disabled: !can || !current, run: startEdit },
			{ key: 'r', shown: 'R', label: 'Reset to default', hidden: busy, disabled: !can || !resettable, run: reset },
			{ key: 'Enter', label: 'Save', command: true, hidden: !busy, disabled: saving, run: () => document.querySelector<HTMLFormElement>(listForm ? '.listform' : '.pane form')?.requestSubmit() },
			...(busy ? [{ key: 'Escape', shown: 'Esc', label: 'Cancel', run: () => leave('mode') }] : [])
		];
	});
</script>

<div class="bar">
	<h1 class="caps">{feature.title}</h1>
	<span class="muted n">
		{#if !file.file.exists}
			Not written yet
		{:else if changed}
			<span class="changed">{count(changed)} changed from the defaults</span>
		{:else}
			Nothing differs from the defaults
		{/if}
	</span>
</div>
<p class="muted lead">{feature.what}</p>
{#if !file.file.exists}
	<p class="notice">SkyrimNet hasn’t written <span class="id">{file.file.path}</span> yet; it will the next time it starts. Until then the defaults can be read but not changed.</p>
{/if}
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

<div class="split">
	<div class="list">
		<table>
			<thead>
				<tr>
					<th class="mark" aria-label="Selected"></th>
					<th class="caps">Setting</th>
					<th class="caps value">Value</th>
					<th class="caps value">Default</th>
				</tr>
			</thead>
			<tbody>
				{#each rows as r, i (r.field.key)}
					{#if i === 0 || rows[i - 1]!.section !== r.section}
						<tr class="group"><th colspan="4" class="caps" scope="rowgroup">{r.section}</th></tr>
					{/if}
					<tr aria-selected={selection.is(r)} onclick={() => !editing && selection.select(r)} ondblclick={startEdit}>
						<td class="mark">{#if selection.is(r)}<Diamond />{/if}</td>
						<td class="name">{r.field.label}</td>
						<td class="value" class:changed={r.setting?.changed}>{show(r.field, valueOf(r.setting))}</td>
						<td class="value muted">{r.setting?.hasDefault ? show(r.field, r.setting.default) : ''}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="muted small">Everything else in <span class="id">{file.file.path}</span> is on its <a href={settingsHref(save, feature.file)}>full settings page</a>.</p>
		{#if feature.lists && lists}
			<FilterLists kind={feature.lists} {lists} canEdit={file.file.exists} />
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Setting">
		{#if current}
			{@const f = current.field}
			{@const s = current.setting}
			{#key `${f.key}:${editing}`}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{f.label}</h2>
					<p class="small what">{f.what}</p>
					{#if f.kind === 'choice' && f.choices?.length}
						<ul class="choices small">
							{#each f.choices as c (c.value)}<li><span class="id">{c.value}</span>{#if c.what}<span class="muted"> · {c.what}</span>{/if}</li>{/each}
						</ul>
					{/if}
					<p class="muted small"><span class="id">{f.key}</span> in {file.file.path}</p>
					{#if editing}
						<form onsubmit={submit}>
							<label class="field">
								<span class="caps label">{f.kind === 'percent' ? 'Percent' : f.kind === 'minutes' ? 'Minutes' : f.unit ? f.unit.charAt(0).toUpperCase() + f.unit.slice(1) : f.kind === 'set' ? 'Chosen' : f.kind === 'hotkey' ? 'Key' : 'Value'}</span>
								{#if f.kind === 'hotkey'}
									<button
										type="button"
										class="capture"
										bind:this={field}
										aria-label="Press the key to use"
										onkeydown={captureKey}
										onmousedown={captureMouse}
										oncontextmenu={(e) => e.preventDefault()}>{keyName(Number(draft))}</button
									>
									<span class="muted small">Press a key, or click the box with a mouse button. <button type="button" class="textbutton" onclick={() => (draft = String(noKey))}>No key</button></span>
								{:else if f.kind === 'set'}
									<ul class="picks">
										{#each f.choices ?? [] as c, i (c.value)}
											<li>
												<label class="pick">
													{#if i === 0}
														<input bind:this={field} type="checkbox" checked={picked.includes(c.value)} onchange={(e) => pick(c.value, e.currentTarget.checked)} />
													{:else}
														<input type="checkbox" checked={picked.includes(c.value)} onchange={(e) => pick(c.value, e.currentTarget.checked)} />
													{/if}
													<span>{c.label}</span>
												</label>
											</li>
										{/each}
									</ul>
								{:else if f.kind === 'choice' && !f.custom}
									<select bind:this={field} bind:value={draft}>
										{#each f.choices ?? [] as c (c.value)}<option value={c.value}>{c.label}</option>{/each}
									</select>
								{:else if f.kind === 'choice'}
									<input bind:this={field} type="text" bind:value={draft} list="feature-choices" autocomplete="off" spellcheck="false" />
									<datalist id="feature-choices">
										{#each f.choices ?? [] as c (c.value)}<option value={c.value}>{c.label}</option>{/each}
									</datalist>
								{:else if f.kind === 'text'}
									<input bind:this={field} type="text" bind:value={draft} autocomplete="off" />
								{:else}
									<input bind:this={field} type="text" inputmode="decimal" bind:value={draft} autocomplete="off" />
									<span class="muted small">
										{#if f.kind === 'percent'}0 to 100{:else if f.min !== undefined && f.max !== undefined}{f.min} to {f.max}{/if}{#if s?.hasDefault} · default {show(f, s.default)}{/if}
									</span>
								{/if}
							</label>
							<div class="buttons">
								<button type="submit" class="button primary" disabled={saving}>Save</button>
								<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
							</div>
						</form>
					{:else}
						<dl>
							<dt class="caps">Now</dt>
							<dd class:changed={s?.changed}>
								{show(f, value)}{#if s && !s.set}<span class="muted"> · not in the file, so the default</span>{/if}
								{#if f.kind === 'set'}
									{@const on = pickedOf(f, value)}
									<!-- Every choice, with the ones left out struck through, so what isn't chosen shows at a glance. -->
									<ul class="setlist small">
										{#each f.choices ?? [] as c (c.value)}
											<li class:off={!on.includes(c.value)}>{c.label}</li>
										{/each}
									</ul>
								{/if}
							</dd>
							{#if s?.hasDefault}<dt class="caps">Default</dt><dd>{show(f, s.default)}</dd>{/if}
						</dl>
						{#if can}
							<div class="actions">
								<button type="button" class="button" onclick={startEdit}>{f.kind === 'switch' ? (value === true ? 'Switch off' : 'Switch on') : 'Edit'}</button>
								{#if resettable}<button type="button" class="button" onclick={reset}>Reset to default</button>{/if}
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
	.changed {
		color: var(--changed);
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
	.group th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
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
	.value {
		width: 10.5rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-variant-numeric: tabular-nums;
	}
	.name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
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
		margin-bottom: var(--space-2);
	}
	.what {
		margin-bottom: var(--space-3);
		line-height: 1.5;
	}
	.choices {
		list-style: none;
		padding: 0;
		margin: 0 0 var(--space-3);
	}
	.pane > div > p.muted {
		margin-bottom: var(--space-4);
		overflow-wrap: anywhere;
	}
	.pane dl {
		display: grid;
		grid-template-columns: 5rem 1fr;
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
		flex-wrap: wrap;
		gap: var(--space-2);
	}
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.field input:not([type='checkbox']),
	.field select {
		width: 100%;
	}
	.pick span {
		white-space: nowrap;
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
	}
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
	.field .small .textbutton {
		font-size: inherit;
	}
	.picks {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.pick {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}
	.setlist {
		list-style: none;
		padding: 0;
		margin: var(--space-1) 0 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0 var(--space-3);
	}
	.setlist .off {
		color: var(--muted);
		text-decoration: line-through;
		text-decoration-color: color-mix(in srgb, var(--muted) 60%, transparent);
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
