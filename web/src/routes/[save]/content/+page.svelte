<script lang="ts">
	import ContentPage from '#lib/records/ContentPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<ContentPage list={data.list} />
