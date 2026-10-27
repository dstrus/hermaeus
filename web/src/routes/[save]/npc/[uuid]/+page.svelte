<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let c = $derived(data.character);
</script>

<dl>
	{#if c.bio}
		<dt>Bio template</dt>
		<dd><code>{c.bio}</code></dd>
	{/if}
	{#if c.modFile}
		<dt>Plugin</dt>
		<dd>{c.modFile}</dd>
	{/if}
	<dt>Actor UUID</dt>
	<dd><code>{c.uuid}</code></dd>
</dl>
