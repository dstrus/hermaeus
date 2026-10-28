<script lang="ts">
	import { characterName } from '#lib/characters/api.ts';
	import DiaryTable from '#lib/memories/DiaryTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<DiaryTable save={data.playthrough.id} uuid={data.character.uuid} name={characterName(data.character)} entries={data.entries} />
