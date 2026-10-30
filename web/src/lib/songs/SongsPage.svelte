<!-- The playthrough's bard songs: each one a bard composed and sang, newest
     first, with the selected one's player and lyrics in the pane. A song's
     words can be edited, and a song deleted so bards stop singing it. -->
<script lang="ts">
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import type { Changes, List, Song } from '#lib/api/types/songs.ts';
	import { plural, realTime } from '#lib/format.ts';
	import { leave, param, setParams } from '#lib/ui/address.svelte.ts';
	import { clipped } from '#lib/ui/clipped.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { deleteShortcuts, useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { audioPath, createdMs, deleteSong, parts, saveSong } from './api.ts';
	import Player from './Player.svelte';
	import { href } from '#lib/nav.ts';

	let { list: loaded }: { list: List } = $props();

	let list = $derived(loaded);
	let save = $derived(page.params.save ?? '');
	let readOnly = $derived(live.gameRunning);
	const npcHref = (uuid: string) => `/${encodeURIComponent(save)}/npc/${uuid}`;

	let query = $state('');
	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return list.songs.filter((s) => !q || [s.title, s.composer, s.style, s.lyrics].some((v) => v.toLowerCase().includes(q)));
	});
	let composers = $derived(new Set(list.songs.map((s) => s.composer).filter(Boolean)).size);

	const selection = new Selection(() => shown, (s) => String(s.id), { param: 'item' });
	let current = $derived(selection.current);

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

	// Editing is a state of the address (?mode=edit), so Back cancels it.
	let mode = $derived<'read' | 'edit'>(param('mode') === 'edit' && current ? 'edit' : 'read');
	let draft = $state<Required<Changes>>({ title: '', lyrics: '', style: '' });
	$effect(() => {
		if (mode !== 'edit') return;
		untrack(() => {
			draft = { title: current?.title ?? '', lyrics: current?.lyrics ?? '', style: current?.style ?? '' };
			error = '';
			notice = '';
		});
	});
	function startEdit() {
		if (current && !readOnly) setParams({ item: String(current.id), mode: 'edit' });
	}

	async function run(s: Song, change: () => Promise<List>, done: string): Promise<boolean> {
		if (saving || readOnly) return false;
		saving = true;
		error = '';
		notice = '';
		noticeFor = s.id;
		try {
			list = await change();
			notice = done;
			return true;
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
			return false;
		} finally {
			saving = false;
		}
	}

	async function submit() {
		const s = current;
		if (s && (await run(s, () => saveSong(save, s.id, draft), 'Saved. The recording is as it was; bards sing it as before.'))) await leave('mode');
	}

	async function remove() {
		const s = current;
		if (!s || !confirm(`Delete “${s.title}”? Bards stop singing it. The database is backed up first, and the audio file stays in the bard songs folder.`)) return;
		const at = shown.indexOf(s);
		if (await run(s, () => deleteSong(save, s.id), `Deleted “${s.title}”.`)) {
			const next = shown[Math.min(at, shown.length - 1)];
			noticeFor = next?.id;
			setParams({ item: next ? String(next.id) : null }, 'replace');
		}
	}

	useShortcuts(() => {
		const reading = mode === 'read';
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, disabled: !reading, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, disabled: !reading, run: () => selection.move(-1) },
			{ key: 'e', shown: 'E', label: 'Edit', hidden: !reading || !current, disabled: readOnly, run: startEdit },
			...deleteShortcuts({ label: 'Delete song', hidden: !reading || !current, disabled: readOnly || saving, run: remove }),
			{ key: 'Enter', label: 'Save', command: true, hidden: reading, disabled: saving, run: submit },
			{ key: 'Escape', shown: 'Esc', label: 'Cancel', hidden: reading, run: () => leave('mode') }
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Bard songs</h1>
	<span class="muted n" use:clipped>{plural(list.songs.length, 'song', 'songs')}{composers ? ` by ${plural(composers, 'bard', 'bards')}` : ''}</span>
	<input type="search" bind:value={query} placeholder="Search songs" aria-label="Search songs" />
</div>
<p class="muted lead">Songs the bards of this playthrough composed and sang, from what was happening around them. A bard can sing any of them again. How bards compose, and how often, is set under <a href={href(save, 'bard-singing')}>Bard singing</a>.</p>
{#if readOnly}<p class="notice warn">Skyrim is running, so songs can only be listened to.</p>{/if}

<div class="split">
	<div class="list">
		{#if !list.songs.length}
			<p class="muted">No bard has composed a song in this playthrough yet. Ask one to, with bard singing switched on.</p>
		{:else if !shown.length}
			<p class="muted">Nothing matches “{query}”.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<th class="caps">Song</th>
						<th class="caps by">Composed by</th>
						<th class="caps when">Composed</th>
						<th class="caps plays">Plays</th>
					</tr>
				</thead>
				<tbody>
					{#each shown as s (s.id)}
						<tr aria-selected={selection.is(s)} onclick={() => selection.select(s)}>
							<td class="mark">{#if selection.is(s)}<Diamond />{/if}</td>
							<td use:clipped>{s.title}{#if s.instrumental}<span class="flag">instrumental</span>{/if}{#if !s.hasAudio}<span class="flag">no audio</span>{/if}</td>
							<td class="by muted" use:clipped>{s.composer || 'Unknown'}</td>
							<td class="when muted">{new Date(createdMs(s.created)).toLocaleDateString(undefined, { dateStyle: 'medium' })}</td>
							<td class="plays muted">{s.plays}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Song">
		{#if current}
			{@const s = current}
			{#key s.id}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2>{s.title}</h2>
					<p class="muted small">
						{#if s.composer}By {#if s.composerUUID}<a href={npcHref(s.composerUUID)}>{s.composer}</a>{:else}{s.composer}{/if}{' · '}{/if}{realTime(createdMs(s.created))}{' · '}{plural(s.plays, 'play', 'plays')}{s.voice ? ` · ${s.voice} voice` : ''}
					</p>
					{#if s.hasAudio}
						<div class="player"><Player path={audioPath(save, s.id)} label={s.title} /></div>
					{:else}
						<p class="notice"><span class="id">{s.file}</span> isn’t in <span class="id">{list.folder}</span>, so it can’t be played.</p>
					{/if}
					{#if error}<p class="notice error" role="alert">{error}</p>{/if}
					{#if notice}<p class="notice" role="status">{notice}</p>{/if}

					{#if mode === 'edit'}
						<div class="form">
							<p class="muted small">Changing the words doesn’t change the recording.</p>
							<label>
								<span class="caps">Title</span>
								<input type="text" bind:value={draft.title} />
							</label>
							<label>
								<span class="caps">Style</span>
								<textarea bind:value={draft.style} rows="3"></textarea>
							</label>
							<label>
								<span class="caps">Lyrics</span>
								<textarea class="lyrics-edit" bind:value={draft.lyrics} rows="18"></textarea>
							</label>
							<div class="buttons">
								<button type="button" class="button primary" onclick={submit} disabled={saving}>Save</button>
								<button type="button" class="button" onclick={() => leave('mode')} disabled={saving}>Cancel</button>
							</div>
						</div>
					{:else}
						<div class="buttons">
							<button type="button" class="button" onclick={startEdit} disabled={readOnly}>Edit</button>
							<button type="button" class="button" onclick={remove} disabled={readOnly || saving}>Delete song</button>
						</div>
						{#if s.style}
							<p class="small style"><span class="caps muted">Style</span> {s.style}</p>
						{/if}
						<div class="lyrics">
							{#each parts(s.lyrics) as p, i (i)}
								<section>
									{#if p.name}<h3 class="caps">{p.name}</h3>{/if}
									<p>{p.text}</p>
								</section>
							{/each}
						</div>
					{/if}
				</div>
			{/key}
		{/if}
	</aside>
</div>

<style>
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
		margin-right: calc(var(--pane) + var(--space-5));
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
	.bar input {
		margin-left: auto;
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
		overflow-wrap: anywhere;
	}
	.bar,
	.split {
		--pane: clamp(20rem, 32vw, 30rem);
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) var(--pane);
		gap: var(--space-5);
		align-items: start;
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
	th.by {
		width: 30%;
	}
	th.when {
		width: 8rem;
	}
	th.plays,
	td.plays {
		width: 4.5rem;
		text-align: right;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
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
	/* A flag is a word of metadata, never a button. */
	.flag {
		margin-left: var(--space-2);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.flag::before {
		content: '· ';
	}
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5) var(--space-5) 0;
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	/* The title in the reading face: it's the song's name, not a label. */
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		line-height: 1.2;
		margin-bottom: var(--space-1);
		overflow-wrap: anywhere;
	}
	.player {
		margin: var(--space-4) 0;
	}
	.pane .notice {
		margin: var(--space-3) 0;
	}
	.buttons {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin: var(--space-3) 0;
	}
	.style {
		margin: var(--space-3) 0;
		padding-bottom: var(--space-3);
		border-bottom: 1px solid var(--rule);
	}
	.style .caps {
		margin-right: var(--space-2);
	}
	/* Lyrics read as a song sheet: each part named in the margin's voice,
	   its lines kept as written. */
	.lyrics {
		padding-bottom: var(--space-6);
	}
	.lyrics section + section {
		margin-top: var(--space-4);
	}
	.lyrics h3 {
		font-size: var(--step--1);
		color: var(--accent);
		margin-bottom: var(--space-1);
	}
	.lyrics p {
		white-space: pre-wrap;
		line-height: 1.65;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		margin-top: var(--space-3);
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.form label {
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
