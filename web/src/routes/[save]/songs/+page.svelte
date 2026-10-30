<script lang="ts">
	import SongsPage from '#lib/songs/SongsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<SongsPage list={data.list} />
