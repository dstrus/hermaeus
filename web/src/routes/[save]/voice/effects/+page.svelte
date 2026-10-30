<script lang="ts">
	import VoiceEffectsPage from '#lib/records/VoiceEffectsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<VoiceEffectsPage list={data.list} />
