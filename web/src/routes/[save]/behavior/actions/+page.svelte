<script lang="ts">
	import ActionsPage from '#lib/records/ActionsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<ActionsPage list={data.list} />
