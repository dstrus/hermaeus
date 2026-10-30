<script lang="ts">
	import ScreenshotsPage from '#lib/screenshots/ScreenshotsPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<ScreenshotsPage list={data.list} />
