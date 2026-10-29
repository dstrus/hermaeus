<script lang="ts">
	import TranslatorPage from '#lib/translator/TranslatorPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<TranslatorPage save={data.playthrough.id} translator={data.translator} characters={data.characters} catalog={data.catalog} />
