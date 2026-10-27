<script lang="ts">
	import PinnedNPCs from '#lib/characters/PinnedNPCs.svelte';
	import type { PageProps } from './$types.ts';

	let { data }: PageProps = $props();
</script>

<h1>{data.playthrough.player || data.playthrough.id}</h1>
<PinnedNPCs save={data.playthrough.id} pins={data.pins.pins} />
