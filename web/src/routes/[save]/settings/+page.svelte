<script lang="ts">
	import SettingsFiles from '#lib/settings/SettingsFiles.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<SettingsFiles save={data.playthrough.id} files={data.list.files} elsewhere={data.list.elsewhere} />
