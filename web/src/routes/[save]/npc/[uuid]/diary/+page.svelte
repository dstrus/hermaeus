<script lang="ts">
	import { characterName } from '#lib/characters/api.ts';
	import DiaryList from '#lib/memories/DiaryList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<DiaryList
	save={data.playthrough.id}
	uuid={data.character.uuid}
	name={characterName(data.character)}
	entries={data.entries}
	readOnly={data.status.gameRunning}
/>
