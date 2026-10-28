<script lang="ts">
	import BioCatalog from '#lib/bios/BioCatalog.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<BioCatalog save={data.playthrough.id} catalog={data.catalog} saveBios={data.saveBios} characters={data.characters} />
