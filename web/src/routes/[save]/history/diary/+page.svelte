<script lang="ts">
	import DiaryTable from '#lib/memories/DiaryTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<DiaryTable save={data.playthrough.id} entries={data.entries} characters={data.characters} />
