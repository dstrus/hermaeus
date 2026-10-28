<script lang="ts">
	import PromptEditor from '#lib/prompts/PromptEditor.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data.path}
	<PromptEditor save={data.playthrough.id} path={data.path} doc={data.doc} />
{/key}
