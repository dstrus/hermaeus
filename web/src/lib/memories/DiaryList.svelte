<script lang="ts">
	import type { DiaryEntry } from '#lib/api/types/memories.ts';
	import { groupByDay, timeLabel } from './dates.ts';

	let { entries }: { entries: DiaryEntry[] } = $props(); // newest first

	let query = $state('');
	let oldestFirst = $state(false);

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = entries.filter(
			(d) => !q || [d.content, d.location, d.emotion, ...d.tags].some((v) => v.toLowerCase().includes(q))
		);
		return oldestFirst ? list.toReversed() : list;
	});
	let groups = $derived(groupByDay(shown));
</script>

<div role="search">
	<input type="search" bind:value={query} placeholder="Search the diary" aria-label="Search the diary" />
	<label><input type="checkbox" bind:checked={oldestFirst} /> Oldest first</label>
</div>

{#if !entries.length}
	<p>No diary entries yet.</p>
{:else if !shown.length}
	<p>No diary entries match.</p>
{/if}

{#each groups as group (group.items[0]?.id)}
	<h2>{group.day}</h2>
	{#each group.items as d (d.id)}
		<article>
			<p>
				<span>{timeLabel(d)}</span>
				{#if d.emotion}· {d.emotion}{/if}
				· importance {d.importance.toFixed(2)}
			</p>
			<ul class="flags">
				{#if d.retired !== undefined}
					<li title="SkyrimNet doesn’t appear to recall entries stored under an earlier identity.">
						stored under “{d.retired || 'an earlier identity'}”
					</li>
				{/if}
				{#if d.memory === undefined}<li title="SkyrimNet keeps a copy of each entry as a memory, and none was found for this one">no memory copy</li>{/if}
			</ul>
			<p class="entry">{d.content}</p>
			{#if d.location}<p><small>{d.location}</small></p>{/if}
		</article>
	{/each}
{/each}

<style>
	.flags {
		display: flex;
		gap: 0.75rem;
		list-style: none;
		padding: 0;
	}
	.entry {
		white-space: pre-line;
	}
</style>
