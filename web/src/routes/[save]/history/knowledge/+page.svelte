<script lang="ts">
	import KnowledgeTable from '#lib/memories/KnowledgeTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<KnowledgeTable save={data.playthrough.id} knowledge={data.knowledge} />
