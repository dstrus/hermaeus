<script lang="ts">
	import Overview from '#lib/summary/Overview.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Overview save={data.playthrough.id} player={data.playthrough.player} summary={data.summary} models={data.models} engine={data.engine} toggles={data.toggles} />
