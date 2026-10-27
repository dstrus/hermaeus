<script lang="ts">
	import DiaryList from '#lib/memories/DiaryList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<DiaryList entries={data.entries} />
