<script lang="ts">
	import OverridesPage from '#lib/records/OverridesPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<OverridesPage list={data.list} />
