<script lang="ts">
	import { m } from '#lib/copy/messages.ts';
	import CloneList from '#lib/features/CloneList.svelte';
	import EngineHead from '#lib/features/EngineHead.svelte';
	import { href } from '#lib/nav.ts';
	import Tabs from '#lib/ui/Tabs.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let save = $derived(data.playthrough.id);
	let base = $derived(href(save, 'engines', `/${data.engine.id}`));
</script>

<div class="bar">
	<h1 class="caps">{data.engine.label}</h1>
	<span class="extra"><EngineHead {save} engine={data.engine} inUse={data.inUse} tab="clones" /></span>
</div>
<p class="muted lead">{data.engine.what}</p>
<Tabs
	label={data.engine.label}
	tabs={[
		{ href: base, label: m.features.engine.settings },
		{ href: `${base}/clones`, label: m.features.engine.clones, count: data.cache.voices.length + data.cache.unnamed }
	]}
/>
<CloneList cache={data.cache} label={data.engine.label} />

<style>
	/* The same title bar as the Settings tab's. */
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
	.extra {
		margin-left: auto;
		display: flex;
		align-items: baseline;
		gap: var(--space-3);
	}
	.lead {
		max-width: 65ch;
		margin-bottom: var(--space-2);
	}
	@media (max-width: 60rem) {
		.bar {
			margin-right: 0;
			flex-wrap: wrap;
		}
	}
</style>
