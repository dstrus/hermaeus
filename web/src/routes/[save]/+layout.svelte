<script lang="ts">
	import QuickFind from '#lib/characters/QuickFind.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
</script>

<QuickFind save={data.playthrough.id} characters={data.characters} />
{@render children()}
