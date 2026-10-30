<!-- OmniSight's screenshots: a picture of each place SkyrimNet has seen,
     with the description a vision model wrote of it, as a grid with the
     selected one in the pane. A description can be rewritten, and kept from
     being rewritten by OmniSight. -->
<script lang="ts">
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError, mediaURL } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Changes, List, Screenshot } from '#lib/api/types/screenshots.ts';
	import { plural, realTime } from '#lib/format.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import MediaImage from '#lib/ui/MediaImage.svelte';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { imagePath, saveScreenshot } from './api.ts';

	let { list: loaded }: { list: List } = $props();

	// The list as last loaded or saved.
	let list = $derived(loaded);
	let save = $derived(page.params.save ?? '');
	let readOnly = $derived(live.gameRunning);

	const typeLabel = (t: string) => ({ location: 'Places', actor: 'People', furniture: 'Furniture' })[t] ?? t.charAt(0).toUpperCase() + t.slice(1);
	let types = $derived([...new Set(list.screenshots.map((s) => s.subjectType))]);
	let type = $derived(param('type') ?? '');
	let query = $state('');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.screenshots.filter((s) => (!type || s.subjectType === type) && (!q || [s.name, s.location, s.cell, s.description].some((v) => v.toLowerCase().includes(q))));
	});
	let locked = $derived(list.screenshots.filter((s) => s.locked).length);

	const selection = new Selection(() => shown, (s) => String(s.id), { param: 'item' });
	let current = $derived(selection.current);

	// Up and down move a row of the grid, as many tiles as fit across.
	let grid = $state<HTMLElement>();
	function across(): number {
		if (!grid) return 1;
		return Math.max(1, getComputedStyle(grid).gridTemplateColumns.split(' ').filter(Boolean).length);
	}

	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let noticeFor = $state<number | undefined>();
	$effect(() => {
		if (current?.id !== noticeFor) {
			error = '';
			notice = '';
		}
	});

	async function change(s: Screenshot, c: Changes, done: string) {
		if (saving || readOnly) return false;
		saving = true;
		error = '';
		notice = '';
		noticeFor = s.id;
		try {
			list = await saveScreenshot(save, s.id, c);
			notice = done;
			return true;
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
			return false;
		} finally {
			saving = false;
		}
	}

	const lock = (s: Screenshot) => change(s, { locked: !s.locked }, s.locked ? 'OmniSight may write a new description when it takes this again.' : 'Kept as written: OmniSight won’t write a new description.');
	const retake = (s: Screenshot) => change(s, { needsRefresh: !s.needsRefresh }, s.needsRefresh ? 'No longer marked to be taken again.' : 'Marked to be taken again.');

	// Editing is a state of the address (?mode=edit), so Back cancels it.
	let mode = $derived<'read' | 'edit'>(param('mode') === 'edit' && current ? 'edit' : 'read');
	let draftName = $state('');
	let draftText = $state('');
	let draftLock = $state(false);
	$effect(() => {
		if (mode !== 'edit') return;
		untrack(() => {
			draftName = current?.name ?? '';
			draftText = current?.description ?? '';
			draftLock = true; // a description you wrote is one you want kept
			error = '';
			notice = '';
		});
	});
	function startEdit() {
		if (current && !readOnly) setParams({ item: String(current.id), mode: 'edit' });
	}
	async function submit() {
		const s = current;
		if (!s) return;
		const c: Changes = { description: draftText, locked: draftLock };
		if (draftName.trim() !== s.name) c.name = draftName;
		if (await change(s, c, draftLock ? 'Saved, and kept as written.' : 'Saved. OmniSight may write a new description when it takes this again.')) await leave('mode');
	}

	async function openFull(s: Screenshot) {
		try {
			window.open(await mediaURL(imagePath(save, s.id)), '_blank', 'noopener');
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		return [
			{ key: 'ArrowRight', label: 'Next', hidden: true, disabled: !reading, run: () => selection.move(1) },
			{ key: 'ArrowLeft', label: 'Previous', hidden: true, disabled: !reading, run: () => selection.move(-1) },
			{ key: 'ArrowDown', label: 'Down', hidden: true, disabled: !reading, run: () => selection.move(across()) },
			{ key: 'ArrowUp', label: 'Up', hidden: true, disabled: !reading, run: () => selection.move(-across()) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !reading || !current, disabled: readOnly, run: startEdit },
			{ key: 'k', shown: 'K', label: current?.locked ? 'Let OmniSight rewrite' : 'Keep as written', hidden: !reading || !current, disabled: readOnly || saving, run: () => current && lock(current) },
			{ key: 'Enter', label: 'Save', command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: 'Cancel', hidden: reading, run: () => leave('mode') }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Screenshots</h1>
	<span class="muted n" use:clipped>{plural(list.screenshots.length, 'screenshot', 'screenshots')}{locked ? ` · ${locked} kept as written` : ''}</span>
	{#if types.length > 1}
		<Segmented value={type} options={[{ value: '', label: 'All' }, ...types.map((t) => ({ value: t, label: typeLabel(t) }))]} label="Subject" onchange={(v) => setParams({ type: v || null })} />
	{/if}
	<input type="search" bind:value={query} placeholder="Search screenshots" aria-label="Search screenshots" />
</div>
<p class="muted lead">What OmniSight has seen: a picture of each place, and the description a vision model wrote of it, which SkyrimNet gives the AI when you’re there.</p>
{#if readOnly}<p class="notice warn">Skyrim is running, so screenshots can only be looked at.</p>{/if}

<div class="split">
	<div class="list">
		{#if !list.screenshots.length}
			<p class="muted">OmniSight hasn’t taken any screenshots in this playthrough. It takes them as you explore while it’s switched on.</p>
		{:else if !shown.length}
			<p class="muted">Nothing matches “{query}”.</p>
		{:else}
			<div class="grid" role="listbox" aria-label="Screenshots" bind:this={grid}>
				{#each shown as s (s.id)}
					<div class="tile" role="option" tabindex="-1" aria-selected={selection.is(s)} onclick={() => selection.select(s)} onkeydown={(e) => e.key === ' ' && selection.select(s)}>
						<span class="picture">
							{#if s.hasImage}<MediaImage path={imagePath(save, s.id, true)} alt="" lazy />{:else}<span class="none muted">No image</span>{/if}
						</span>
						<span class="caption" use:clipped>{s.name}</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Screenshot">
		{#if current}
			{@const s = current}
			{#key s.id}
				<div in:fly={{ y: 6, duration: 180 }}>
					{#if s.hasImage}
						<button type="button" class="full" class:editing={mode === 'edit'} onclick={() => openFull(s)} title="Open full size" style:aspect-ratio={s.width && s.height ? `${s.width} / ${s.height}` : '1'}>
							<MediaImage path={imagePath(save, s.id)} alt="What OmniSight saw at {s.name}" />
						</button>
					{:else}
						<p class="notice">Its image isn’t in <span class="id">{list.folder}</span>, so only its description is here.</p>
					{/if}
					<h2 class="caps">{s.name}</h2>
					<p class="muted small">
						{#if s.location && s.location !== s.name && s.location !== s.cell}{s.location}{' · '}{/if}<span class="id">{s.cell || s.subject}</span>{[s.indoor ? 'indoors' : '', s.weather, s.timeOfDay]
							.filter(Boolean)
							.map((w) => ` · ${w}`)
							.join('')}
					</p>
					<p class="small">
						{#if s.date}{s.date.time}, {s.date.day}{:else if s.gameTime}{s.gameTime}{/if}<span class="muted">{s.date || s.gameTime ? ' · ' : ''}{realTime(Date.parse(s.taken))}{#if s.manual} · taken with your hotkey{/if}</span>
					</p>
					{#if s.error}<p class="notice error">OmniSight couldn’t describe it: {s.error}</p>{/if}
					{#if error}<p class="notice error" role="alert">{error}</p>{/if}
					{#if notice}<p class="notice" role="status">{notice}</p>{/if}

					{#if mode === 'edit'}
						<div class="form">
							<label>
								<span class="caps">Name</span>
								<input type="text" bind:value={draftName} />
							</label>
							<label class="grow">
								<span class="caps">Description</span>
								<textarea bind:value={draftText} rows="12"></textarea>
							</label>
							<label class="check">
								<input type="checkbox" bind:checked={draftLock} />
								<span>Keep it as written: OmniSight won’t write a new one</span>
							</label>
							<div class="buttons">
								<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
								<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
							</div>
						</div>
					{:else}
						<div class="switches" role="group" aria-label="Switches">
							<button type="button" role="switch" aria-checked={s.locked} disabled={readOnly || saving} onclick={() => lock(s)} title="description_locked">
								<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
								<span class="caps">Keep as written</span>
							</button>
							<button type="button" role="switch" aria-checked={s.needsRefresh} disabled={readOnly || saving} onclick={() => retake(s)} title="needs_refresh">
								<span class="mark" aria-hidden="true"><svg viewBox="0 0 10 10" width="10" height="10"><path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg></span>
								<span class="caps">Take again</span>
							</button>
							<button type="button" class="button" onclick={startEdit} disabled={readOnly}>Edit</button>
						</div>
						{#if s.description}
							<p class="text">{s.description}</p>
							{#if s.model}<p class="muted small">Described by <span class="id">{s.model}</span>{s.locked ? ', and kept as written' : ''}.</p>{/if}
						{:else}
							<p class="muted">No description yet.</p>
						{/if}
					{/if}
				</div>
			{/key}
		{/if}
	</aside>
</div>

<style>
	.bar {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
		margin-right: calc(var(--pane) + var(--space-5));
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	.bar .n {
		white-space: nowrap;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		margin-right: auto;
	}
	.bar input {
		flex: 0 1 16rem;
		min-width: 9rem;
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
	.bar,
	.split {
		--pane: clamp(20rem, 34vw, 34rem);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) var(--pane);
		gap: var(--space-5);
		align-items: start;
	}
	/* The grid: square tiles, as many across as fit, each a picture with its
	   name beneath. */
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
		gap: var(--space-3);
		padding-bottom: var(--space-5);
	}
	.tile {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		min-width: 0;
		cursor: default;
	}
	.picture {
		display: block;
		aspect-ratio: 1;
		border: 1px solid var(--rule);
		border-radius: var(--radius);
		overflow: hidden;
		transition: border-color 120ms;
	}
	.tile:hover .picture {
		border-color: var(--rule-strong);
	}
	/* The selected tile: the accent's frame, and its name in ink. */
	.tile[aria-selected='true'] .picture {
		border-color: var(--accent);
		box-shadow: 0 0 0 1px var(--accent);
	}
	.caption {
		font-size: var(--step--1);
		color: var(--muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tile[aria-selected='true'] .caption {
		color: var(--ink);
	}
	.none {
		display: grid;
		place-items: center;
		height: 100%;
		font-size: var(--step--1);
		background: var(--glass);
	}
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	.full {
		display: block;
		width: 100%;
		padding: 0;
		margin-bottom: var(--space-4);
		border: 1px solid var(--rule);
		border-radius: var(--radius);
		overflow: hidden;
		background: none;
		cursor: zoom-in;
	}
	/* Capped, so the description is in view beneath it, and smaller still
	   while it's being edited. */
	.full {
		max-height: 46vh;
	}
	.full.editing {
		max-height: 24vh;
	}
	.full :global(img) {
		object-fit: contain;
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
		overflow-wrap: anywhere;
	}
	.pane .small {
		margin-bottom: var(--space-1);
		overflow-wrap: anywhere;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.text {
		white-space: pre-wrap;
		line-height: 1.6;
		margin: var(--space-4) 0 var(--space-2);
		max-width: 65ch;
	}
	.pane > div > .muted:last-child {
		padding-bottom: var(--space-5);
	}
	.switches {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2) var(--space-4);
		margin-top: var(--space-4);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.switches .button {
		margin-left: auto;
	}
	[role='switch'] {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		padding: 0;
		background: none;
		border: 0;
		cursor: pointer;
		font-size: var(--step--1);
	}
	[role='switch']:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.mark {
		display: grid;
		place-items: center;
		width: 14px;
		height: 14px;
		border: 1px solid var(--rule-strong);
		border-radius: 3px;
		color: transparent;
	}
	[role='switch'][aria-checked='true'] .mark {
		background: var(--accent);
		border-color: var(--accent);
		color: var(--on-accent);
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		margin-top: var(--space-4);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.form label:not(.check) {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.form label .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.form textarea {
		resize: vertical;
		line-height: 1.55;
	}
	.check {
		display: flex;
		gap: var(--space-2);
		align-items: baseline;
		font-size: var(--step--1);
	}
	.buttons {
		display: flex;
		gap: var(--space-2);
	}
	/* Save stays in reach however long the form is. */
	.form .buttons {
		position: sticky;
		bottom: 0;
		margin: 0;
		padding: var(--space-3) 0 var(--space-5);
		background: var(--panel);
		border-top: 1px solid var(--rule);
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
	}
</style>
