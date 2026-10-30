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
	import type { Snippet } from 'svelte';
	import type { Change, FileValues, Setting } from '#lib/api/types/settings.ts';
	import FilterLists from './FilterLists.svelte';
	import { count } from '#lib/format.ts';
	import { saveSettings, settingsDependency, settingsHref, togglesDependency } from '#lib/settings/api.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { codeOf, keyName, mouseCodeOf, noKey } from '#lib/settings/keys.ts';
	import { blankRow, cellText, draftOf, linesOf, linesValue, pairsOf, pairsValue, parse, pickedOf, rowsOf, rowsValue, setValue, show, valueOf, type Feature, type Field, type Pair, type Row as Entry } from './schema.ts';

	let { save, feature, file, lists, extra, tabs }: { save: string; feature: Feature; file: FileValues; lists?: ListsResponse; extra?: Snippet; tabs?: Snippet } = $props();

	let readOnly = $derived(live.gameRunning);
	let can = $derived(file.file.exists && !readOnly);
	let settings = $derived(new Map(file.values.map((s) => [s.key.toLowerCase(), s])));
	const settingOf = (f: Field): Setting | undefined => settings.get(f.key.toLowerCase());

	// The page's rows: every field, in its section. A map field stands for
	// the settings under its key, each its own setting in the file.
	type Row = { field: Field; section: string; setting?: Setting; children?: Setting[] };
	const childrenOf = (f: Field): Setting[] => {
		const prefix = f.key.toLowerCase() + '.';
		const leaf = f.leaf ? '.' + f.leaf.toLowerCase() : '';
		return file.values.filter((s) => {
			const k = s.key.toLowerCase();
			return k.startsWith(prefix) && (leaf ? k.endsWith(leaf) && !k.slice(prefix.length, -leaf.length).includes('.') : true);
		});
	};
	// nameOf is a map entry's name, from its setting's key.
	const nameOf = (f: Field, key: string) => key.slice(f.key.length + 1, f.leaf ? -(f.leaf.length + 1) : undefined);
	let rows = $derived<Row[]>(
		feature.sections.flatMap((s) => s.fields.map((field) => (field.kind === 'map' ? { field, section: s.title, children: childrenOf(field) } : { field, section: s.title, setting: settingOf(field) })))
	);
	const isChanged = (r: Row) => (r.children ? r.children.some((c) => c.changed) : !!r.setting?.changed);
	let changed = $derived(rows.filter(isChanged).length);
	// valueOfRow is a row's value; a map's is its settings by name.
	const valueOfRow = (r: Row): unknown => (r.children ? Object.fromEntries(r.children.map((c) => [nameOf(r.field, c.key), valueOf(c)])) : valueOf(r.setting));

	const selection = new Selection(() => rows, (r) => r.field.key, { param: 'item' });
	let current = $derived(selection.current);
	let value = $derived(current ? valueOfRow(current) : undefined);
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
	let pairs = $state<Pair[]>([]); // a pairs field's rows, or a map's names and values
	let entries = $state<Entry[]>([]); // a rows field's entries
	let areaField = $state<HTMLTextAreaElement>();
	let field = $state<HTMLInputElement | HTMLSelectElement | HTMLButtonElement>();
	$effect(() => {
		if (!editing) return;
		const f = current!.field;
		draft = f.kind === 'lines' ? linesOf(value) : draftOf(f, value);
		picked = pickedOf(f, value);
		const rows = f.kind === 'pairs' ? pairsOf(f, value) : f.kind === 'map' ? Object.entries((value as Record<string, unknown>) ?? {}).map(([a, b]) => ({ a, b: b == null ? '' : String(b) })) : [];
		pairs = f.kind === 'pairs' && !rows.length ? [{ a: '', b: '' }] : rows;
		entries = f.kind === 'rows' ? rowsOf(f, value) : [];
		tick().then(() => (f.kind === 'lines' ? areaField : field)?.focus());
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

	async function apply(change: Change | Change[], done: string) {
		if (saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await saveSettings(feature.file, { modified: file.file.modified, changes: Array.isArray(change) ? change : [change] });
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
		const f = current.field;
		// A map saves each name whose value changed, as its own setting.
		if (f.kind === 'map') {
			const was = (value as Record<string, unknown>) ?? {};
			const edited = pairs.filter((p) => String(was[p.a] ?? '') !== p.b);
			if (!edited.length) {
				leave('mode');
				return;
			}
			const changes = edited.map((p) => ({ key: `${f.key}.${p.a}${f.leaf ? `.${f.leaf}` : ''}`, value: p.b }));
			apply(changes, `${f.label}: ${edited.length === 1 ? `${edited[0]!.a} changed` : `${edited.length} changed`}.`);
			return;
		}
		let v: unknown;
		try {
			v = f.kind === 'set' ? setValue(f, value, picked) : f.kind === 'lines' ? linesValue(draft) : f.kind === 'pairs' ? pairsValue(f, pairs) : f.kind === 'rows' ? rowsValue(f, entries) : parse(f, draft);
			// A plugin's hotkey goes back as the file keeps it: as text when it's
			// text, unset as empty text, the way SkyrimNet writes it; otherwise as
			// the key code, -1 when unset, as SkyrimNet's dashboard writes it.
			if (current.field.pluginHotkey && typeof v === 'number' && typeof value === 'string') v = v === noKey ? '' : String(v);
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
	{#if extra}<span class="extra">{@render extra()}</span>{/if}
</div>
<p class="muted lead">{feature.what}{#if feature.see}{' '}{feature.see.before}<a href="/{encodeURIComponent(save)}/{feature.see.path}">{feature.see.label}</a>{feature.see.after}{/if}</p>
{#if tabs}<div class="tabs">{@render tabs()}</div>{/if}
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
						<td class="value" class:changed={isChanged(r)}>{show(r.field, valueOfRow(r))}</td>
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
							{#each f.choices as c (c.value)}<li>{#if c.label !== c.value}{c.label}<span class="id muted what-choice">{c.value}</span>{:else}<span class="id">{c.value}</span>{/if}{#if c.what}<span class="muted what-choice">· {c.what}</span>{/if}</li>{/each}
						</ul>
					{/if}
					<p class="muted small"><span class="id">{f.key}</span> in {file.file.path}</p>
					{#if editing}
						<form onsubmit={submit}>
							<label class="field">
								<span class="caps label" class:sr={f.kind === 'pairs' || f.kind === 'map' || f.kind === 'rows'}>{f.kind === 'percent' ? 'Percent' : f.kind === 'minutes' ? 'Minutes' : f.unit ? f.unit.charAt(0).toUpperCase() + f.unit.slice(1) : f.kind === 'set' ? 'Chosen' : f.kind === 'hotkey' ? 'Key' : 'Value'}</span>
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
								{:else if f.kind === 'secret'}
									<input bind:this={field} type="password" bind:value={draft} placeholder={show(f, value) === 'Set' ? 'Type a new one to replace it' : ''} autocomplete="off" spellcheck="false" />
									<span class="muted small">Saved into {file.file.path}. Hermaeus never shows it again; save it empty to clear it.</span>
								{:else if f.kind === 'lines'}
									<textarea bind:this={areaField} bind:value={draft} rows="8" spellcheck="false" placeholder="One per line"></textarea>
									<span class="muted small">One per line. Blank lines are left out.</span>
								{:else if f.kind === 'rows'}
									<!-- One card per entry: its name and switch, then its other fields. -->
									<ol class="cards">
										{#each entries as r, i (i)}
											<li>
												<div class="cardhead">
													{#each (f.cols ?? []).filter((c) => c.kind === 'text').slice(0, 1) as c (c.key)}
														<input type="text" bind:value={r.cells[c.key]} placeholder={c.placeholder} aria-label="{c.label} {i + 1}" autocomplete="off" spellcheck="false" />
													{/each}
													{#each (f.cols ?? []).filter((c) => c.kind === 'switch') as c (c.key)}
														<label class="check"><input type="checkbox" bind:checked={r.cells[c.key] as boolean} /><span>{c.label}</span></label>
													{/each}
													<button type="button" class="textbutton" onclick={() => (entries = entries.filter((_, j) => j !== i))} aria-label="Remove entry {i + 1}">Remove</button>
												</div>
												<div class="cardcells">
													{#each (f.cols ?? []).filter((c, k) => c.kind !== 'switch' && !(c.kind === 'text' && k === (f.cols ?? []).findIndex((x) => x.kind === 'text'))) as c (c.key)}
														<label>
															<span class="caps">{c.label}{c.kind === 'percent' ? ' %' : c.unit ? ` (${c.unit})` : ''}</span>
															{#if c.kind === 'choice'}
																<select bind:value={r.cells[c.key]}>
																	{#each c.choices ?? [] as ch (ch.value)}<option value={ch.value}>{ch.label}</option>{/each}
																</select>
															{:else}
																<input type="text" inputmode={c.kind === 'text' ? undefined : 'decimal'} bind:value={r.cells[c.key]} autocomplete="off" />
															{/if}
														</label>
													{/each}
												</div>
											</li>
										{/each}
									</ol>
									<span class="muted small"><button type="button" class="textbutton" onclick={() => (entries = [...entries, blankRow(f)])}>Add {f.noun?.[0] ?? 'an entry'}</button></span>
								{:else if f.kind === 'map' && f.leaf}
									<ul class="entries">
										{#each pairs as p (p.a)}
											<li>
												<label>
													<span class="id">{p.a}</span>
													<input type="text" bind:value={p.b} autocomplete="off" spellcheck="false" />
												</label>
											</li>
										{/each}
									</ul>
								{:else if f.kind === 'pairs' || f.kind === 'map'}
									<table class="rows">
										<thead><tr><th class="caps">{f.columns?.[0] ?? 'Name'}</th><th class="caps">{f.columns?.[1] ?? 'Value'}</th>{#if f.kind === 'pairs'}<th><span class="sr">Remove</span></th>{/if}</tr></thead>
										<tbody>
											{#each pairs as p, i (i)}
												<tr>
													{#if f.kind === 'map'}
														<td class="id">{p.a}</td>
														<td><input type="text" bind:value={p.b} aria-label="{f.columns?.[1] ?? 'Value'} for {p.a}" autocomplete="off" spellcheck="false" /></td>
													{:else}
														<td><input type="text" bind:value={p.a} aria-label="{f.columns?.[0] ?? 'Name'} {i + 1}" autocomplete="off" spellcheck="false" /></td>
														<td><input type="text" bind:value={p.b} aria-label="{f.columns?.[1] ?? 'Value'} {i + 1}" autocomplete="off" spellcheck="false" /></td>
														<td><button type="button" class="textbutton" onclick={() => (pairs = pairs.filter((_, j) => j !== i))} aria-label="Remove {p.a || `entry ${i + 1}`}">Remove</button></td>
													{/if}
												</tr>
											{/each}
										</tbody>
									</table>
									{#if f.kind === 'pairs'}
										<span class="muted small"><button type="button" class="textbutton" onclick={() => (pairs = [...pairs, { a: '', b: '' }])}>Add {f.noun?.[0] ?? 'an entry'}</button> · Spaces count, so a phrase like “ - ” matches only with its spaces.</span>
									{/if}
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
									<input bind:this={field} type="text" bind:value={draft} placeholder={f.placeholder} autocomplete="off" />
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
								{#if f.kind === 'rows' && Array.isArray(value) && value.length}
									{@const cols = f.cols ?? []}
									<ul class="entries small">
										{#each value as e, i (i)}
											{@const row = (e ?? {}) as Record<string, unknown>}
											<li class:off={cols.some((c) => c.kind === 'switch' && row[c.key] === false)}>
												<span class="id">{cellText(cols[0]!, row[cols[0]!.key])}</span>
												<span class="muted">{cols
														.slice(1)
														.map((c) => (c.kind === 'switch' ? cellText(c, row[c.key]) : `${c.label.toLowerCase()} ${cellText(c, row[c.key])}`))
														.join(' · ')}</span>
											</li>
										{/each}
									</ul>
								{:else if f.kind === 'lines' && Array.isArray(value) && value.length}
									<ul class="setlist small">
										{#each value as l, i (i)}<li>{String(l)}</li>{/each}
									</ul>
								{:else if (f.kind === 'pairs' || f.kind === 'map') && value}
									{@const list = f.kind === 'pairs' ? pairsOf(f, value) : Object.entries(value as Record<string, unknown>).map(([a, b]) => ({ a, b: b == null ? '' : String(b) }))}
									{#if list.length && f.leaf}
										<ul class="entries small">
											{#each list as p, i (i)}<li><span class="id">{p.a}</span><span class="muted">{p.b}</span></li>{/each}
										</ul>
									{:else if list.length}
										<table class="rows small">
											<thead><tr><th class="caps">{f.columns?.[0] ?? 'Name'}</th><th class="caps">{f.columns?.[1] ?? 'Value'}</th></tr></thead>
											<tbody>
												{#each list as p, i (i)}<tr><td class:id={f.kind === 'map'}>{p.a}</td><td>{p.b}</td></tr>{/each}
											</tbody>
										</table>
									{/if}
								{/if}
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
	.what-choice {
		margin-left: var(--space-2);
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
	.tabs {
		margin-bottom: var(--space-4);
	}
	.extra {
		margin-left: auto;
	}
	.field textarea {
		resize: vertical;
		line-height: 1.5;
	}
	/* A pairs or map field's entries: two columns, read or edited in place. */
	.rows {
		width: 100%;
		margin-top: var(--space-2);
		border-collapse: collapse;
		table-layout: auto;
	}
	.rows th {
		padding: var(--space-1) var(--space-2) var(--space-1) 0;
		border-bottom: 1px solid var(--rule);
	}
	.rows td {
		padding: var(--space-1) var(--space-2) var(--space-1) 0;
		border-bottom: 0;
		white-space: normal;
		overflow-wrap: anywhere;
	}
	.rows td:last-child:has(.textbutton) {
		width: 1%;
		white-space: nowrap;
	}
	.rows input {
		width: 100%;
		min-width: 0;
		padding: var(--space-1) var(--space-2);
	}
	/* A map whose values are long, such as file paths: each entry's name,
	   then its value beneath. */
	.entries {
		list-style: none;
		padding: 0;
		margin: var(--space-2) 0 0;
	}
	.entries li {
		padding: var(--space-1) 0;
		border-bottom: 1px solid var(--rule);
	}
	.entries li > span,
	.entries label > span {
		display: block;
		overflow-wrap: anywhere;
	}
	.entries label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.entries input {
		width: 100%;
		padding: var(--space-1) var(--space-2);
		font-size: var(--step--1);
	}
	/* A rows field's entries, edited one card each. */
	.cards {
		list-style: none;
		padding: 0;
		margin: var(--space-2) 0;
	}
	.cards li {
		padding: var(--space-2) 0 var(--space-3);
		border-bottom: 1px solid var(--rule);
	}
	.cardhead {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}
	.cardhead input[type='text'] {
		flex: 1;
		min-width: 0;
		padding: var(--space-1) var(--space-2);
	}
	.cardhead .check {
		display: flex;
		gap: var(--space-1);
		align-items: center;
		font-size: var(--step--1);
	}
	.cardcells {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: var(--space-2);
		margin-top: var(--space-2);
	}
	.cardcells label {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.cardcells .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.cardcells input,
	.cardcells select {
		width: 100%;
		min-width: 0;
		padding: var(--space-1) var(--space-2);
	}
	.entries li.off .id {
		text-decoration: line-through;
		text-decoration-color: var(--muted);
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
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
