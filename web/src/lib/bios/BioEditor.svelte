<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Earlier, Version } from '#lib/api/types/bios.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { realTime } from '#lib/format.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { bioDependency, createBio, getHistory, kinds, restoreBio, saveBio, sectionLabel } from './api.ts';
	import BioDiff from './BioDiff.svelte';
	import CodeArea from '#lib/ui/CodeArea.svelte';
	import Inja from '#lib/ui/Inja.svelte';
	import { parseSections } from './diff.ts';

	let {
		save,
		template,
		name,
		versions
	}: {
		save: string;
		template: string;
		name: string; // the character's, as shown
		versions: Version[]; // the dynamic bio, the baseline, the default: whichever exist
	} = $props();

	let readOnly = $derived(live.gameRunning);
	// Which bio is shown is in the address (?bio=baseline); it starts on the
	// one NPCs use, and falls back to it when the one named isn't there.
	let kind = $derived.by(() => {
		const k = param('bio');
		return k && versions.some((v) => v.kind === k) ? k : (versions[0]?.kind ?? 'dynamic');
	});
	let version = $derived(versions.find((v) => v.kind === kind));
	let editable = $derived(!!version && version.kind !== 'default' && !readOnly);
	let hasBaseline = $derived(versions.some((v) => v.kind === 'baseline'));

	// The bio reads as one document. One section is selected (for the keys);
	// editing turns that section into a field where it is. Raw, history, and
	// diff replace the document. The mode is in the address (?mode=raw), so
	// each has its own URL and Back leaves it; one the bio can't be in is
	// dropped again.
	type Mode = 'read' | 'edit' | 'raw' | 'history' | 'diff';
	let mode = $derived.by((): Mode => {
		const p = param('mode');
		if (p === 'edit' && editable && section) return 'edit';
		if (p === 'raw' && editable) return 'raw';
		if (p === 'history' && version && version.kind !== 'default') return 'history';
		if (p === 'diff' && pairs.length) return 'diff';
		return 'read';
	});
	$effect(() => {
		if (mode === 'read' && (param('mode') || param('version'))) setParams({ mode: null, version: null }, 'replace');
	});
	// back returns to reading, and to the entry the mode was opened from.
	const back = () => leave('mode', 'version');

	// The diff of two of the bios: the later one against the one beneath it,
	// so what SkyrimNet's rewrite did (dynamic against baseline) reads first.
	const order = ['dynamic', 'baseline', 'default'];
	let pairs = $derived.by(() => {
		const have = order.filter((k) => versions.some((v) => v.kind === k));
		const out: { value: string; label: string }[] = [];
		for (let i = 0; i < have.length; i++) for (let j = i + 1; j < have.length; j++) out.push({ value: `${have[i]}|${have[j]}`, label: m.bios.editor.pair({ to: `${kinds[have[i]!]?.label}`, from: `${kinds[have[j]!]?.label.toLowerCase()}` }) });
		return out;
	});
	let pair = $state('');
	let compared = $derived.by(() => {
		const chosenPair = pairs.some((p) => p.value === pair) ? pair : (pairs[0]?.value ?? '');
		const [toKind, fromKind] = chosenPair.split('|');
		const to = versions.find((v) => v.kind === toKind);
		const from = versions.find((v) => v.kind === fromKind);
		return to && from ? { to, from } : undefined;
	});
	// In history, the earlier version reads as text or as the changes since.
	let historyView = $state('text');
	let draft = $state('');
	let saving = $state(false);
	let notice = $state('');
	let error = $state('');

	const selection = new Selection(() => version?.sections ?? [], (s) => s.name, { marker: '[data-selected="true"]', param: 'section' });
	let section = $derived(selection.current);

	// Which section is in view, for the index, from where the page has scrolled to.
	let inView = $state<string | null>(null);
	let doc = $state<HTMLElement>();
	$effect(() => {
		if (!doc || mode === 'raw' || mode === 'history' || mode === 'diff') return;
		const seen = new Map<string, number>();
		const watch = new IntersectionObserver(
			(entries) => {
				for (const e of entries) seen.set((e.target as HTMLElement).dataset.section!, e.intersectionRatio);
				let best: string | null = null;
				let most = 0;
				for (const [name, ratio] of seen) if (ratio > most) ((most = ratio), (best = name));
				inView = best;
			},
			{ rootMargin: '-10% 0px -60% 0px', threshold: [0, 0.25, 0.5, 1] }
		);
		for (const el of doc.querySelectorAll<HTMLElement>('[data-section]')) watch.observe(el);
		return () => watch.disconnect();
	});

	let history = $state<Earlier[]>([]);
	const chosen = new Selection(() => history, (e) => e.id, { param: 'version' });
	let earlier = $derived(chosen.current);

	// The draft, or the history, is set up when the mode starts, however it
	// starts; it follows the bio shown, should that change underneath.
	$effect(() => {
		if (mode === 'edit') draft = section?.text ?? '';
		else if (mode === 'raw') draft = version?.raw ?? '';
		else if (mode === 'history') untrack(loadHistory);
	});

	async function run(action: () => Promise<string>) {
		error = '';
		notice = '';
		saving = true;
		try {
			notice = await action();
			await back(); // settles before the reload, which it would cancel
			await invalidate(bioDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	function startEdit(name?: string) {
		if (!editable) return;
		if (name) selection.key = name;
		if (selection.current) setParams({ mode: 'edit' });
	}

	function startRaw() {
		if (editable && version) setParams({ mode: 'raw' });
	}

	function startHistory() {
		if (version && version.kind !== 'default') setParams({ mode: 'history' });
	}

	async function loadHistory() {
		if (!version) return;
		error = '';
		try {
			history = (await getHistory(save, template, version.kind)).earlier;
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	function submit() {
		if (!version) return;
		const v = version;
		const edit = mode === 'raw' ? { modified: v.modified, raw: draft } : { modified: v.modified, sections: { [section!.name]: draft } };
		run(async () => {
			const r = await saveBio(save, template, v.kind, edit);
			return [m.bios.editor.saved({ kind: kinds[v.kind]?.label ?? v.kind }), r.warning].filter(Boolean).join(' ');
		});
	}

	function create() {
		run(async () => {
			await createBio(save, template);
			setParams({ bio: 'baseline' }, 'replace');
			return m.bios.editor.created({ name });
		});
	}

	function restore() {
		if (!version || !earlier) return;
		const v = version;
		const e = earlier;
		run(async () => {
			await restoreBio(save, template, v.kind, { id: e.id, modified: v.modified });
			return m.bios.editor.restored({ when: realTime(e.replaced) });
		});
	}

	function jump(name: string) {
		selection.key = name;
		doc?.querySelector(`[data-section="${name}"]`)?.scrollIntoView({ block: 'start', behavior: 'smooth' });
	}

	let kindOptions = $derived(versions.map((v) => ({ value: v.kind, label: kinds[v.kind]?.label ?? v.kind, title: kinds[v.kind]?.what })));

	useShortcuts(() => {
		const reading = mode === 'read';
		const editing = mode === 'edit' || mode === 'raw';
		const listing = reading || mode === 'history';
		return [
			{ key: 'ArrowDown', label: m.common.next, hidden: true, disabled: !listing, run: () => (mode === 'history' ? chosen.move(1) : selection.move(1)) },
			{ key: 'ArrowUp', label: m.common.previous, hidden: true, disabled: !listing, run: () => (mode === 'history' ? chosen.move(-1) : selection.move(-1)) },
			{ key: 'e', shown: 'E', label: m.bios.editor.editSection, hidden: !reading, disabled: !editable || !section, run: () => startEdit() },
			{ key: 'r', shown: 'R', label: m.bios.editor.rawFile, hidden: !reading, disabled: !editable, run: startRaw },
			{ key: 'h', shown: 'H', label: m.bios.editor.history, hidden: !reading, disabled: !version || version.kind === 'default', run: startHistory },
			{ key: 'd', shown: 'D', label: m.bios.editor.diff, hidden: !reading, disabled: pairs.length === 0, run: () => setParams({ mode: 'diff' }) },
			{ key: 'c', shown: 'C', label: m.bios.editor.createShortcut, hidden: !reading || version?.kind !== 'default' || hasBaseline, disabled: readOnly, run: create },
			{ key: 'Enter', label: m.common.save, command: true, hidden: !editing, disabled: saving, run: submit },
			{ key: 'Enter', label: m.bios.editor.restoreShortcut, hidden: mode !== 'history', disabled: readOnly || !earlier || earlier.current, run: restore },
			...(reading ? [] : [{ key: 'Escape', shown: 'Esc', label: mode === 'history' || mode === 'diff' ? m.bios.editor.back : m.common.cancel, run: back }])
		];
	});
</script>

<div class="tools">
	{#if versions.length > 1}
		<Segmented value={kind} onchange={(k) => setParams({ bio: k }, 'replace')} options={kindOptions} label={m.bios.editor.which} />
	{:else if version}
		<span class="caps muted">{m.bios.kindBio({ kind: kinds[version.kind]?.label ?? version.kind })}</span>
	{/if}
	{#if version}
		<span class="muted">{kinds[version.kind]?.what}{#if version.source}{' '}{version.shared ? m.bios.editor.fromShared({ source: version.source, shared: version.shared }) : m.bios.editor.from({ source: version.source })}{/if}</span>
	{/if}
	{#if version?.kind === 'default' && !hasBaseline && !readOnly && mode === 'read'}
		<button type="button" class="button primary" onclick={create} disabled={saving}>{m.bios.editor.createFromThis}</button>
	{/if}
</div>
{#if notice}<p class="notice" role="status">{notice}</p>{/if}
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !version}
	<p class="muted"><Fill text={m.bios.editor.noBio.text} {name}>{#snippet bioTemplate()}<span class="id">{template}</span>{/snippet}</Fill></p>
{:else if mode === 'raw'}
	<div class="raw" use:fitToViewport>
		<p class="muted small">{m.bios.editor.rawHint}</p>
		<CodeArea bind:value={draft} label={m.bios.editor.rawLabel} />
		<div class="buttons">
			<button type="button" class="button primary" onclick={submit} disabled={saving}>{m.common.save}</button>
			<button type="button" class="button" onclick={back} disabled={saving}>{m.common.cancel}</button>
		</div>
	</div>
{:else if mode === 'diff' && compared}
	<div class="diff">
		<div class="which">
			{#if pairs.length > 1}
				<Segmented bind:value={pair} options={pairs} label={m.bios.editor.whichTwo} />
			{:else}
				<span class="caps muted">{pairs[0]?.label}</span>
			{/if}
			<span class="muted small">{m.bios.editor.diffLead({ to: kinds[compared.to.kind]?.label.toLowerCase() ?? '', from: kinds[compared.from.kind]?.label.toLowerCase() ?? '' })}</span>
		</div>
		<BioDiff from={compared.from.sections} to={compared.to.sections} fromLabel={kinds[compared.from.kind]?.label ?? compared.from.kind} toLabel={kinds[compared.to.kind]?.label ?? compared.to.kind} />
	</div>
{:else if mode === 'history'}
	<div class="split">
		<table>
			<thead>
				<tr><th class="mark" aria-label={m.common.selected}></th><th class="caps">{m.bios.editor.replaced}</th><th class="caps">{m.bios.editor.by}</th><th class="caps">{m.bios.editor.playthrough}</th></tr>
			</thead>
			<tbody>
				{#if !history.length}
					<tr><td colspan="4" class="muted">{m.bios.editor.noHistory}</td></tr>
				{/if}
				{#each history as e (e.id)}
					<tr aria-selected={chosen.is(e)} onclick={() => chosen.select(e)}>
						<td class="mark">{#if chosen.is(e)}<Diamond />{/if}</td>
						<td>{realTime(e.replaced)}{#if e.current}<span class="muted"> · {m.bios.editor.sameAsNow}</span>{/if}</td>
						<td>{e.source === 'skyrimnet' ? m.bios.editor.bySkyrimNet : m.bios.editor.byHermaeus}</td>
						<td class="muted">{e.playthrough === 'this' ? m.bios.editor.thisOne : e.playthrough === 'other' ? e.save : m.bios.editor.unclear}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<aside class="pane" use:fitToViewport aria-label={m.bios.editor.earlierVersion}>
			{#if earlier}
				{#key earlier.id}
					<div in:fly={{ y: 6, duration: 180 }}>
						<h2 class="caps">{realTime(earlier.replaced)}</h2>
						<p class="muted small">{earlier.source === 'skyrimnet' ? m.bios.editor.keptBySkyrimNet : m.bios.editor.keptByHermaeus}</p>
						{#if !earlier.current}
							<div class="which">
								<Segmented bind:value={historyView} options={[{ value: 'text', label: m.bios.editor.text }, { value: 'changes', label: m.bios.editor.changesSince }]} label={m.bios.editor.howToRead} />
							</div>
						{/if}
						{#if historyView === 'changes' && !earlier.current}
							<BioDiff from={parseSections(earlier.raw)} to={version.sections} fromLabel={m.bios.editor.thisVersion} toLabel={m.bios.editor.bioNow} />
						{:else}
							<pre class="file"><Inja text={earlier.raw} /></pre>
						{/if}
						<div class="buttons">
							{#if !earlier.current && !readOnly}
								<button type="button" class="button primary" onclick={restore} disabled={saving}>{m.bios.editor.restore}</button>
							{/if}
							<button type="button" class="button" onclick={back}>{m.bios.editor.back}</button>
						</div>
					</div>
				{/key}
			{:else}
				<p class="muted">{m.bios.editor.chooseEarlier}</p>
			{/if}
		</aside>
	</div>
{:else}
	<div class="reading">
		<article class="document" bind:this={doc}>
			{#each version.sections as s (s.name)}
				{@const selected = selection.is(s)}
				{@const editing = mode === 'edit' && selected}
				<!-- The label's button is the accessible control; clicking anywhere in
				     the section is a larger target for the same thing, and a double
				     click opens the editor. -->
				<!-- svelte-ignore a11y_no_static_element_interactions, a11y_click_events_have_key_events -->
				<section
					data-section={s.name}
					data-selected={selected}
					class:editing
					onclick={() => mode === 'read' && (selection.key = s.name)}
					ondblclick={() => mode === 'read' && startEdit(s.name)}
				>
					<div class="label">
						<h2 class="caps">
							<button type="button" onclick={() => mode === 'read' && (selection.key = s.name)} disabled={mode !== 'read'}>
								<span class="mark">{#if selected}<Diamond />{/if}</span>{sectionLabel(s.name)}
							</button>
						</h2>
						{#if selected && editable && mode === 'read'}
							<button type="button" class="textbutton small" onclick={() => startEdit(s.name)}>{m.common.edit}</button>
						{/if}
					</div>
					{#if editing}
						<div class="field">
							<textarea bind:value={draft} rows={Math.max(6, Math.min(28, Math.ceil(draft.length / 80) + draft.split('\n').length + 1))}></textarea>
							<div class="buttons">
								<button type="button" class="button primary" onclick={submit} disabled={saving}>{m.common.save}</button>
								<button type="button" class="button" onclick={back} disabled={saving}>{m.common.cancel}</button>
							</div>
						</div>
					{:else}
						<p class="text"><Inja text={s.text} /></p>
					{/if}
				</section>
			{:else}
				<p class="muted">{m.bios.editor.noSections}</p>
			{/each}
			{#if readOnly}<p class="notice warn">{m.bios.editor.gameRunning}</p>{/if}
		</article>
		<nav class="index" aria-label={m.bios.editor.sections}>
			<ul>
				{#each version.sections as s (s.name)}
					<li>
						<button type="button" class="caps" class:here={inView === s.name} aria-current={selection.is(s) ? 'true' : undefined} onclick={() => jump(s.name)}>
							{sectionLabel(s.name)}
						</button>
					</li>
				{/each}
			</ul>
		</nav>
	</div>
{/if}

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.id {
		font-variant-numeric: tabular-nums;
	}
	.small {
		font-size: var(--step--1);
	}

	/* Reading: the document, with the index alongside. */
	.reading {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 11rem;
		gap: var(--space-7);
		align-items: start;
	}
	.document {
		max-width: 52rem;
	}
	section {
		display: grid;
		grid-template-columns: 10rem minmax(0, 1fr);
		gap: var(--space-2) var(--space-4);
		padding: var(--space-3) 0 var(--space-3) var(--space-2);
		border-top: 1px solid var(--rule);
		scroll-margin-top: var(--space-4);
	}
	section:first-child {
		border-top: 0;
	}
	section[data-selected='true'] {
		background: var(--select);
	}
	section:not(.editing):hover {
		background: var(--glass);
	}
	section[data-selected='true']:not(.editing):hover {
		background: var(--select);
	}
	.label {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-1);
	}
	section h2 {
		font-size: var(--step--1);
		padding-top: 0.35em;
	}
	section h2 button {
		display: flex;
		gap: var(--space-2);
		align-items: baseline;
		background: none;
		border: 0;
		padding: 0;
		color: var(--muted);
		font: inherit;
		letter-spacing: inherit;
		text-transform: inherit;
		text-align: left;
		cursor: pointer;
	}
	section h2 button:hover:not(:disabled) {
		color: var(--ink);
	}
	section h2 .mark {
		width: 0.75rem;
		flex: none;
	}
	.label .textbutton {
		margin-left: calc(0.75rem + var(--space-2));
	}
	.text {
		white-space: pre-line;
		line-height: 1.6;
		max-width: 65ch;
		margin-right: var(--space-4);
	}
	.field {
		margin-right: var(--space-4);
	}
	.field textarea {
		width: 100%;
		resize: vertical;
		line-height: 1.55;
	}
	.index {
		position: sticky;
		top: var(--space-4);
	}
	.index ul {
		list-style: none;
		margin: 0;
		padding: 0;
		border-left: 1px solid var(--rule-strong);
	}
	.index button {
		display: block;
		width: 100%;
		background: none;
		border: 0;
		padding: var(--space-1) var(--space-3);
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		cursor: pointer;
		margin-left: -1px;
		border-left: 1px solid transparent;
	}
	.index button:hover {
		color: var(--ink);
	}
	.index button.here {
		color: var(--ink);
		border-left-color: var(--accent);
	}
	.index button[aria-current='true'] {
		color: var(--accent);
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}

	/* The whole file fills the window down to the hint strip, with Save in
	   view beneath it; the text scrolls inside, not the page. */
	.raw {
		position: sticky;
		top: 0;
		display: flex;
		flex-direction: column;
		max-width: 64rem; /* fitToViewport sets its height */
	}
	.raw :global(.code) {
		flex: 1;
		margin-top: var(--space-3);
	}
	.raw .buttons {
		padding-bottom: var(--space-4);
	}

	/* The diff of two bios. */
	.diff .which,
	.pane .which {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-3) var(--space-4);
		margin-bottom: var(--space-4);
	}
	.pane .which {
		margin: var(--space-3) 0;
	}

	/* Earlier versions: the item table and the pane. */
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(22rem, 34vw, 36rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule-strong);
		font-weight: inherit;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
	}
	tbody tr {
		cursor: default;
	}
	/* A row you can select tints under the pointer, as a bio's sections do. */
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
	.pane {
		position: sticky;
		top: 0;
		overflow: auto; /* fitToViewport sets its height */
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.file {
		margin: var(--space-3) 0;
		font: 400 var(--step--1) / 1.55 var(--font-mono);
		white-space: pre-wrap;
	}
	.pane .buttons {
		position: sticky;
		bottom: 0;
		padding: var(--space-3) 0 var(--space-5);
		margin-top: var(--space-3);
		background: var(--panel);
		border-top: 1px solid var(--rule);
	}

	@media (max-width: 70rem) {
		.reading {
			grid-template-columns: 1fr;
		}
		.index {
			display: none;
		}
		.split {
			grid-template-columns: minmax(0, 1fr) minmax(18rem, 24rem);
		}
	}
	@media (max-width: 60rem) {
		section {
			grid-template-columns: 1fr;
			gap: var(--space-1);
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
