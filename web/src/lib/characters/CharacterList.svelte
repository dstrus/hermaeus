<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { BusyMemories, type Character, type Group } from '#lib/api/types/characters.ts';
	import { characterName, charactersDependency, setPinned } from './api.ts';

	let {
		save,
		characters,
		current,
		readOnly
	}: {
		save: string;
		characters: Character[]; // already in group order
		current?: string; // the UUID of the character being shown
		readOnly: boolean; // Skyrim is running, so pins can't change
	} = $props();

	let query = $state('');
	let error = $state('');

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		if (!q) return characters;
		return characters.filter((c) => characterName(c).toLowerCase().includes(q) || c.bio.toLowerCase().includes(q));
	});
	let hasTop = $derived(characters.some((c) => c.group === 'pinned' || c.group === 'busy'));

	const headings: Record<Group, string | null> = {
		player: null,
		pinned: 'Pinned',
		busy: `${BusyMemories}+ memories`,
		other: 'Everyone else'
	};
	const groups: Group[] = ['player', 'pinned', 'busy', 'other'];

	function counts(c: Character): string {
		const parts = [`${c.memories} ${c.memories === 1 ? 'memory' : 'memories'}`];
		if (c.diaries) parts.push(`${c.diaries} ${c.diaries === 1 ? 'diary entry' : 'diary entries'}`);
		return parts.join(', ');
	}

	async function togglePin(c: Character) {
		error = '';
		try {
			await setPinned(save, c.uuid, !c.pinned);
			await invalidate(charactersDependency); // regroup with the new pins
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		}
	}
</script>

<nav aria-label="Characters">
	<input type="search" bind:value={query} placeholder="Find a character" aria-label="Find a character" />
	{#if error}
		<p role="alert">{error}</p>
	{/if}
	{#if !characters.length}
		<p>No one has anything stored in this playthrough yet.</p>
	{:else if !shown.length}
		<p>No one matches “{query}”.</p>
	{/if}
	{#each groups as group (group)}
		{@const members = shown.filter((c) => c.group === group)}
		{#if members.length}
			{#if headings[group]}
				<h2>{group === 'other' && !hasTop ? 'NPCs' : headings[group]}</h2>
			{/if}
			<ul>
				{#each members as c (c.uuid)}
					<li>
						<a href="/{encodeURIComponent(save)}/npc/{c.uuid}" aria-current={c.uuid === current ? 'page' : undefined}>
							{characterName(c)}
						</a>
						<small>{counts(c)}</small>
						{#if c.group !== 'player'}
							<button
								onclick={() => togglePin(c)}
								aria-pressed={c.pinned}
								disabled={readOnly}
								title={readOnly ? 'Close Skyrim to change pins' : undefined}
							>
								{c.pinned ? 'Unpin' : 'Pin'}
							</button>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	{/each}
</nav>

<style>
	ul {
		list-style: none;
		padding: 0;
	}
	[aria-current='page'] {
		font-weight: bold;
	}
</style>
