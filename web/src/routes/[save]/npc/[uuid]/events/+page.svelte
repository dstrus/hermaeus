<script lang="ts">
	import { characterName } from '#lib/characters/api.ts';
	import EventTable from '#lib/events/EventTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<EventTable save={data.playthrough.id} name={characterName(data.character)} events={data.events} characters={data.characters} />
