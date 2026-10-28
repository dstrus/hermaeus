<script lang="ts">
	import MemoryTable from '#lib/memories/MemoryTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<MemoryTable save={data.playthrough.id} uuid={data.character.uuid} memories={data.memories} />
