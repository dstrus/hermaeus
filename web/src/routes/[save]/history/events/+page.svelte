<script lang="ts">
	import EventTable from '#lib/events/EventTable.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<EventTable save={data.playthrough.id} events={data.events} characters={data.characters} />
