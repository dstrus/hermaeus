<script lang="ts">
	import FolderForm from '#lib/folders/FolderForm.svelte';
	import type { LayoutProps } from './$types.ts';

	let { data, children }: LayoutProps = $props();
</script>

<header>
	<a href="/">Hermaeus</a>
	{#if data.status.gameRunning}
		<p role="status">Skyrim is running. You can browse, but saving is off until you close the game.</p>
	{/if}
</header>

<main>
	{#if data.status.skyrimNetDir}
		{@render children()}
	{:else}
		<FolderForm />
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: system-ui, sans-serif;
		line-height: 1.5;
	}
	header,
	main {
		max-width: 60rem;
		margin: 0 auto;
		padding: 1rem;
	}
</style>
