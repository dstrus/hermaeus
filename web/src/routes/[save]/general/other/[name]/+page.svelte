<script lang="ts">
	import SettingsTable from '#lib/settings/SettingsTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data.title}
	<SettingsTable save={data.playthrough.id} title={data.title} file={data.file} rows={data.rows} byFile={data.byFile} />
{/key}
