<script lang="ts">
	import PluginsPage from '#lib/plugins/PluginsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<PluginsPage list={data.list} />
