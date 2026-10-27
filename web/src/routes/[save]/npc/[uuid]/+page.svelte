<script lang="ts">
	import { characterName } from '#lib/characters/api.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let c = $derived(data.character);
</script>

<h1>{characterName(c)}</h1>
<dl>
	{#if c.formerNames.length}
		<dt>Previously known as</dt>
		<dd>{c.formerNames.join(', ')}</dd>
	{/if}
	<dt>Memories</dt>
	<dd>{c.memories}</dd>
	<dt>Diary entries</dt>
	<dd>{c.diaries}</dd>
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
