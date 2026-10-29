<script lang="ts">
	import FeaturePage from '#lib/features/FeaturePage.svelte';
	import SettingsTable from '#lib/settings/SettingsTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data.title}
	{#if data.feature}
		<FeaturePage save={data.playthrough.id} feature={data.feature} file={data.file} lists={data.lists} />
	{:else}
		<SettingsTable save={data.playthrough.id} title={data.title} file={data.file} rows={data.rows} byFile={data.byFile} />
	{/if}
{/key}
