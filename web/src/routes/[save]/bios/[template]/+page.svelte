<script lang="ts">
	import { goto } from '$app/navigation';
	import BioEditor from '#lib/bios/BioEditor.svelte';
	import { catalogHref, templateName } from '#lib/bios/api.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import Up from '#lib/ui/Up.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let name = $derived(templateName(data.template));

	// Esc goes back to the list when the editor has nothing of its own to cancel.
	useShortcuts(() => [{ key: 'Escape', shown: 'Esc', label: m.characters.which.all, run: () => goto(catalogHref(data.playthrough.id)) }]);
</script>

<header>
	<Up href={catalogHref(data.playthrough.id)} label={m.characters.which.all} />
	<h1 class="caps">{name}</h1>
	<p class="muted"><Fill text={m.bios.notMet.text}>{#snippet template()}<span class="id">{data.template}</span>{/snippet}</Fill></p>
</header>

{#key data.template}
	<BioEditor save={data.playthrough.id} template={data.template} {name} versions={data.versions} />
{/key}

<style>
	header {
		margin-bottom: var(--space-4);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		margin-bottom: var(--space-2);
	}
	.id {
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
</style>
