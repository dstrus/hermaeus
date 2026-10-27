<script lang="ts">
	import { ApiError } from '#lib/api/client.ts';
	import type { PinnedNPC } from '#lib/api/types/characters.ts';
	import { setPinned } from './pins.ts';

	let { save, pins: loaded }: { save: string; pins: PinnedNPC[] } = $props();

	// Starts as the loaded pins, and follows them when the page loads new
	// ones; unpinning replaces it with the server's reply in between.
	let pins = $derived(loaded);
	let error = $state('');

	async function unpin(npc: PinnedNPC) {
		error = '';
		try {
			pins = (await setPinned(save, npc.uuid, false)).pins;
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}
</script>

<section>
	<h2>Pinned NPCs</h2>
	{#if pins.length === 0}
		<p>No one is pinned. Pins you make in SkyrimNet’s dashboard show here too.</p>
	{:else}
		<ul>
			{#each pins as npc (npc.uuid)}
				<li>
					{npc.name || `Unnamed NPC ${npc.uuid}`}
					<button onclick={() => unpin(npc)}>Unpin</button>
				</li>
			{/each}
		</ul>
	{/if}
	{#if error}
		<p role="alert">{error}</p>
	{/if}
</section>
