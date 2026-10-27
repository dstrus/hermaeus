<script lang="ts">
	import MemoryList from '#lib/memories/MemoryList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<MemoryList memories={data.memories} />
