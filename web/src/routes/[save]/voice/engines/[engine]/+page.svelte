<script lang="ts">
	import { goto } from '$app/navigation';
	import FeaturePage from '#lib/features/FeaturePage.svelte';
	import { engines } from '#lib/features/voice.ts';
	import { href } from '#lib/nav.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let save = $derived(data.playthrough.id);
</script>

{#key data.engine.id}
	<FeaturePage {save} feature={data.feature} file={data.file}>
		{#snippet extra()}
			{#if data.inUse?.id === data.engine.id}
				<span class="muted">In use</span>
			{:else}
				<span class="muted">Not in use · <a href="{href(save, 'speech')}?item=tts.engine">Choose the engine</a></span>
			{/if}
			<select aria-label="Engine" value={data.engine.id} onchange={(e) => goto(href(save, 'engines', `/${e.currentTarget.value}`))}>
				{#each engines as e (e.id)}<option value={e.id}>{e.label}{data.inUse?.id === e.id ? ' · in use' : ''}</option>{/each}
			</select>
		{/snippet}
	</FeaturePage>
{/key}
