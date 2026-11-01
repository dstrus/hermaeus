<script lang="ts">
	import BioEditor from '#lib/bios/BioEditor.svelte';
	import { characterName } from '#lib/characters/api.ts';
	import { m } from '#lib/copy/messages.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if data.character.bio}
	<BioEditor save={data.playthrough.id} template={data.character.bio} name={characterName(data.character)} versions={data.versions} />
{:else}
	<p class="muted">{m.bios.noTemplate({ name: characterName(data.character) })}</p>
{/if}
