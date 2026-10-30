<script lang="ts">
	import TriggersPage from '#lib/records/TriggersPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<TriggersPage list={data.list} />
