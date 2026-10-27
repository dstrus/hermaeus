<script lang="ts">
	import type { PageProps } from './$types.ts';

	let { data }: PageProps = $props();
</script>

<h1>Playthroughs</h1>
{#if data.status.playthroughs.length === 0}
	<p>There are no SkyrimNet databases in <code>{data.status.skyrimNetDir}</code> yet. SkyrimNet makes one the first time you load a save with it.</p>
{:else}
	<ul>
		{#each data.status.playthroughs as p (p.id)}
			<li>
				<a href="/{encodeURIComponent(p.id)}">{p.player || p.id}</a>
				— last played {new Date(p.modified).toLocaleString()}
			</li>
		{/each}
	</ul>
{/if}
