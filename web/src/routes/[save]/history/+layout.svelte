<!-- History: everything SkyrimNet recorded in this playthrough, across every
     character: events, memories, and diary entries. -->
<script lang="ts">
	import { m } from '#lib/copy/messages.ts';
	import Tabs from '#lib/ui/Tabs.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let base = $derived(`/${encodeURIComponent(data.playthrough.id)}/history`);
	let tabs = $derived([
		{ href: `${base}/events`, label: m.history.events },
		{ href: `${base}/memories`, label: m.history.memories },
		{ href: `${base}/diary`, label: m.history.diary },
		{ href: `${base}/knowledge`, label: m.history.knowledge }
	]);
</script>

<header>
	<h1 class="caps">{m.history.title}</h1>
	<p class="muted">{m.history.lead}</p>
	<Tabs {tabs} label={m.history.title} />
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
