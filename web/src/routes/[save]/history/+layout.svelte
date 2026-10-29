<!-- History: everything SkyrimNet recorded in this playthrough, across every
     character: events, memories, and diary entries. -->
<script lang="ts">
	import { page } from '$app/state';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let base = $derived(`/${encodeURIComponent(data.playthrough.id)}/history`);
	let tabs = $derived([
		{ href: `${base}/events`, label: 'Events' },
		{ href: `${base}/memories`, label: 'Memories' },
		{ href: `${base}/diary`, label: 'Diary' }
	]);
</script>

<header>
	<h1 class="caps">History</h1>
	<p class="muted">Everything SkyrimNet recorded in this playthrough, everyone's together.</p>
	<nav aria-label="History">
		{#each tabs as tab (tab.href)}
			<a class="caps" href={tab.href} aria-current={page.url.pathname === tab.href ? 'page' : undefined}>{tab.label}</a>
		{/each}
	</nav>
</header>

{@render children()}

<style>
	header {
		margin-bottom: var(--space-5);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		margin-bottom: var(--space-2);
	}
	nav {
		display: flex;
		gap: var(--space-5);
		margin-top: var(--space-4);
		border-bottom: 1px solid var(--rule);
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
	}
</style>
