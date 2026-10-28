<script lang="ts">
	import PromptTable from '#lib/prompts/PromptTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<PromptTable save={data.playthrough.id} list={data.list} />
