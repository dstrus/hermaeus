<script lang="ts">
	import { m } from '#lib/copy/messages.ts';
	import EngineHead from '#lib/features/EngineHead.svelte';
	import FeaturePage from '#lib/features/FeaturePage.svelte';
	import { href } from '#lib/nav.ts';
	import Tabs from '#lib/ui/Tabs.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let save = $derived(data.playthrough.id);
	let base = $derived(href(save, 'engines', `/${data.engine.id}`));
</script>

{#key data.engine.id}
	<FeaturePage {save} feature={data.feature} file={data.file}>
		{#snippet extra()}
			<EngineHead {save} engine={data.engine} inUse={data.inUse} />
		{/snippet}
		{#snippet tabs()}
			{#if data.clones}
				<Tabs
					label={data.engine.label}
					tabs={[
						{ href: base, label: m.features.engine.settings },
						{ href: `${base}/clones`, label: m.features.engine.clones, count: data.clones.voices.length + data.clones.unnamed }
					]}
				/>
			{/if}
		{/snippet}
	</FeaturePage>
{/key}
