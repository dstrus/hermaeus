<script lang="ts">
	import MemoryList from '#lib/memories/MemoryList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<MemoryList
	save={data.playthrough.id}
	uuid={data.character.uuid}
	memories={data.memories}
	readOnly={data.status.gameRunning}
/>
