<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import { BusyMemories, type Character, type Group } from '#lib/api/types/characters.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import SortHeader from '#lib/ui/SortHeader.svelte';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { sorted, type Sort } from '#lib/ui/sort.ts';
	import { count } from '#lib/format.ts';
	import { characterName, charactersDependency, setPinned } from './api.ts';

	let { save, characters }: { save: string; characters: Character[] } = $props(); // characters in group order

	let query = $state('');
	let sort = $state<Sort>({ key: 'group', descending: false });
	let error = $state('');

	const groupNames: Record<Group, string> = { player: 'You', pinned: 'Pinned', busy: `${BusyMemories}+ memories`, other: 'Everyone else' };

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = characters.filter((c) => !q || characterName(c).toLowerCase().includes(q) || c.bio.toLowerCase().includes(q));
		return sorted(list, sort, {
			name: (c) => (c.name ? c.name : `￿${c.uuid}`),
			memories: (c) => c.memories,
			diaries: (c) => c.diaries
		});
	});
	// In their natural order, characters sit under their group's heading.
	let grouped = $derived(sort.key === 'group');

	const selection = new Selection(() => shown, (c) => c.uuid);
	let current = $derived(selection.current);
	const href = (c: Character) => `/${encodeURIComponent(save)}/npc/${c.uuid}`;

	async function togglePin(c: Character) {
		error = '';
		try {
			await setPinned(save, c.uuid, !c.pinned);
			await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}

	useShortcuts(() => {
		const c = selection.current;
		return [
			{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
			{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
			{ key: 'Enter', shown: 'Enter', label: 'Open', disabled: !c, run: () => c && goto(href(c)) },
			{
				key: 'p',
				shown: 'P',
				label: c?.pinned ? 'Unpin' : 'Pin',
				disabled: !c || c.group === 'player' || live.gameRunning,
				run: () => c && togglePin(c)
			}
		];
	});
</script>

<div class="bar">
	<h1 class="caps">Characters</h1>
	<span class="muted">{count(characters.length)} with something stored</span>
	<input type="search" bind:value={query} placeholder="Filter by name or bio" aria-label="Filter characters" />
</div>
{#if error}<p class="notice error" role="alert">{error}</p>{/if}

{#if !characters.length}
	<p class="muted">No one has anything stored in this playthrough yet.</p>
{:else if !shown.length}
	<p class="muted">No one matches “{query}”.</p>
{:else}
	<div class="split">
	<table>
		<thead>
			<tr>
				<th class="mark" aria-label="Selected"></th>
				<SortHeader label="Name" key="name" bind:sort />
				<SortHeader label="Memories" key="memories" bind:sort descendingFirst align="end" />
				<SortHeader label="Diary" key="diaries" bind:sort descendingFirst align="end" />
				<th class="plugin caps">Plugin</th>
			</tr>
		</thead>
		<tbody>
			{#each shown as c, i (c.uuid)}
				{#if grouped && c.group !== shown[i - 1]?.group}
					<tr class="group"><th colspan="5" class="caps" scope="rowgroup">{groupNames[c.group]}</th></tr>
				{/if}
				<tr aria-selected={selection.is(c)} onclick={() => selection.select(c)} ondblclick={() => goto(href(c))}>
					<td class="mark">{#if selection.is(c)}<Diamond />{/if}</td>
					<td class="name">
						<a href={href(c)}>{characterName(c)}</a>
						{#if c.formerNames.length}<span class="muted"> · was {c.formerNames.map((n) => n || 'unnamed').join(', ')}</span>{/if}
					</td>
					<td class="n">{count(c.memories)}</td>
					<td class="n">{c.diaries ? count(c.diaries) : ''}</td>
					<td class="plugin muted">{c.modFile}</td>
				</tr>
			{/each}
		</tbody>
	</table>
	<aside class="pane" aria-label="Character">
		{#if current}
			<h2 class="caps">{characterName(current)}</h2>
			<p class="muted">{groupNames[current.group]}{#if current.formerNames.length}{' · previously '}{current.formerNames.map((n) => n || 'unnamed').join(', ')}{/if}</p>
			<dl>
				<dt class="caps">Memories</dt><dd>{count(current.memories)}</dd>
				<dt class="caps">Diary</dt><dd>{count(current.diaries)} {current.diaries === 1 ? 'entry' : 'entries'}</dd>
				{#if current.modFile}<dt class="caps">Plugin</dt><dd>{current.modFile}</dd>{/if}
				{#if current.bio}<dt class="caps">Bio template</dt><dd>{current.bio}</dd>{/if}
				<dt class="caps">SkyrimNet ID</dt><dd>{current.uuid}</dd>
			</dl>
			<div class="actions">
				<a class="button primary" href={href(current)}>Open</a>
				{#if current.group !== 'player'}
					<button
						type="button"
						class="button"
						aria-pressed={current.pinned}
						disabled={live.gameRunning}
						onclick={() => togglePin(current!)}>{current.pinned ? 'Unpin' : 'Pin'}</button
					>
				{/if}
			</div>
			{#if live.gameRunning && current.group !== 'player'}<p class="notice warn">Close Skyrim to change pins.</p>{/if}
		{/if}
	</aside>
	</div>
{/if}

<style>
	/* Keeps the filter over the table, not the pane. */
	.bar {
		display: flex;
		align-items: baseline;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
		margin-right: calc(clamp(18rem, 26vw, 24rem) + var(--space-5));
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	.bar input {
		margin-left: auto;
		width: 18rem;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	/* The item card: a full-height zone of the menu beside the list. */
	.pane {
		position: sticky;
		top: 0;
		min-height: calc(100vh - 7rem);
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.pane dl {
		display: grid;
		grid-template-columns: 7.5rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: var(--space-4) 0;
	}
	.pane dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	.pane dd {
		margin: 0;
		overflow-wrap: anywhere;
	}
	.actions {
		display: flex;
		gap: var(--space-2);
	}
	.actions a {
		text-decoration: none;
	}
	.pane .notice {
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
			min-height: 0;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
		.split :global(.plugin) {
			display: none;
		}
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
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
	.name a {
		text-decoration: none;
	}
	.name a:hover {
		text-decoration: underline;
	}
	.n {
		text-align: right;
		width: 7rem;
	}
	.plugin {
		font-size: var(--step--1);
		text-align: left;
		padding: var(--space-2) var(--space-3);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
		font-weight: inherit;
	}
	tbody .plugin {
		border-bottom-color: var(--rule);
	}
	.group th {
		text-align: left;
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
		font-size: var(--step--1);
		border-bottom: 1px solid var(--rule-strong);
	}
</style>
