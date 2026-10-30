<!-- History: everything SkyrimNet recorded in this playthrough, across every
     character: events, memories, and diary entries. -->
<script lang="ts">
	import Tabs from '#lib/ui/Tabs.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let base = $derived(`/${encodeURIComponent(data.playthrough.id)}/history`);
	let tabs = $derived([
		{ href: `${base}/events`, label: 'Events' },
		{ href: `${base}/memories`, label: 'Memories' },
		{ href: `${base}/diary`, label: 'Diary' },
		{ href: `${base}/knowledge`, label: 'Knowledge' }
	]);
</script>

<header>
	<h1 class="caps">History</h1>
	<p class="muted">Everything SkyrimNet recorded in this playthrough, everyone’s together, and the knowledge any NPC can draw on.</p>
	<Tabs {tabs} label="History" />
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
</style>
