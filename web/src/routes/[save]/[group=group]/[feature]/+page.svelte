<script lang="ts">
	import FeaturePage from '#lib/features/FeaturePage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data.feature.id}
	<FeaturePage save={data.playthrough.id} feature={data.feature} file={data.file} lists={data.lists} />
{/key}
