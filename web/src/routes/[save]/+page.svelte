<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<h1>{data.playthrough.player || data.playthrough.id}</h1>
<p>Choose a character to see what SkyrimNet has stored about them.</p>
