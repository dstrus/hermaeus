<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { page } from '$app/state';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import { characterName, charactersDependency, setPinned } from '#lib/characters/api.ts';
	import { count } from '#lib/format.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import Up from '#lib/ui/Up.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let c = $derived(data.character);
	let list = $derived(`/${encodeURIComponent(data.playthrough.id)}`);
	let base = $derived(`${list}/npc/${c.uuid}`);
	let tabs = $derived([
		{ href: base, label: 'Overview' },
		{ href: `${base}/stats`, label: 'Stats' },
		{ href: `${base}/memories`, label: 'Memories', n: c.memories },
		{ href: `${base}/diary`, label: 'Diary', n: c.diaries },
		...(c.bio ? [{ href: `${base}/bio`, label: 'Bio' }] : [])
	]);
	let error = $state('');

	// Esc goes back to the character table when the page has nothing of its own to cancel.
	useShortcuts(() => [{ key: 'Escape', shown: 'Esc', label: 'Characters', run: () => goto(list) }]);

	async function togglePin() {
		error = '';
		try {
			await setPinned(data.playthrough.id, c.uuid, !c.pinned);
			await invalidate(charactersDependency);
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}
</script>

<header>
	<Up href={list} label="Characters" />
	<div class="title">
		<h1 class="caps">{characterName(c)}</h1>
		{#if c.group !== 'player'}
			<button
				type="button"
				class="button"
				aria-pressed={c.pinned}
				disabled={live.gameRunning}
				title={live.gameRunning ? 'Close Skyrim to change pins' : undefined}
				onclick={togglePin}>{c.pinned ? 'Unpin' : 'Pin'}</button
			>
		{/if}
	</div>
	<p class="muted">
		{[c.formerNames.length ? `Previously ${c.formerNames.map((n) => n || 'unnamed').join(', ')}` : '', c.group === 'player' ? 'You' : c.pinned ? 'Pinned' : '', c.modFile]
			.filter(Boolean)
			.join(' · ')}
	</p>
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<nav aria-label="{characterName(c)}’s records">
		{#each tabs as tab (tab.href)}
			<a class="caps" href={tab.href} aria-current={page.url.pathname === tab.href ? 'page' : undefined}>
				{tab.label}{#if tab.n !== undefined}<span class="n">{count(tab.n)}</span>{/if}
			</a>
		{/each}
	</nav>
</header>

{@render children()}

<style>
	header {
		margin-bottom: var(--space-5);
	}
	.title {
		display: flex;
		align-items: center;
		gap: var(--space-4);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1.05;
	}
	header > p {
		margin-top: var(--space-1);
	}
	nav {
		display: flex;
		gap: var(--space-6);
		margin-top: var(--space-4);
		border-bottom: 1px solid var(--rule);
		overflow-x: auto; /* scrolls on its own on narrow screens */
	}
	nav a {
		white-space: nowrap;
	}
	nav a {
		padding: var(--space-2) 0;
		font-size: var(--step-1);
		color: var(--muted);
		text-decoration: none;
		margin-bottom: -1px;
		border-bottom: 1px solid transparent;
	}
	nav a:hover {
		color: var(--ink);
	}
	nav a[aria-current='page'] {
		color: var(--ink);
		font-weight: 600;
		border-bottom: 2px solid var(--accent);
		margin-bottom: -1px;
	}
	.n {
		margin-left: var(--space-2);
		color: var(--muted);
		font-family: var(--font-text);
		font-size: var(--step--1);
		letter-spacing: 0;
	}
</style>
