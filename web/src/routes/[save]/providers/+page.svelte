<script lang="ts">
	import ProvidersPage from '#lib/providers/ProvidersPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<ProvidersPage save={data.playthrough.id} providers={data.providers} />
