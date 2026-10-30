<script lang="ts">
	import ModelsPage from '#lib/models/ModelsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<ModelsPage save={data.playthrough.id} models={data.models} />
