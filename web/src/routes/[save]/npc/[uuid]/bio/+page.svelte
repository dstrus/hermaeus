<script lang="ts">
	import BioEditor from '#lib/bios/BioEditor.svelte';
	import { characterName } from '#lib/characters/api.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if data.character.bio}
	<BioEditor save={data.playthrough.id} template={data.character.bio} name={characterName(data.character)} versions={data.versions} />
{:else}
	<p class="muted">SkyrimNet has no bio template for {characterName(data.character)}, so there is no bio to show.</p>
{/if}
