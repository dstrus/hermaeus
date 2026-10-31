<script lang="ts">
	import CharacterTable from '#lib/characters/CharacterTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<CharacterTable save={data.playthrough.id} characters={data.characters} />
