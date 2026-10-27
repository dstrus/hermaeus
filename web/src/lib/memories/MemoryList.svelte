<script lang="ts">
	import type { Memory } from '#lib/api/types/memories.ts';
	import { dayLabel, groupByDay, timeLabel } from './dates.ts';

	let { memories }: { memories: Memory[] } = $props(); // newest first

	let query = $state('');
	let type = $state('');
	let emotion = $state('');
	let sort = $state<'newest' | 'oldest' | 'importance'>('newest');

	let types = $derived([...new Set(memories.map((m) => m.type).filter(Boolean))].sort());
	let emotions = $derived([...new Set(memories.map((m) => m.emotion).filter(Boolean))].sort());

	let shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const list = memories.filter(
			(m) =>
				(!type || m.type === type) &&
				(!emotion || m.emotion === emotion) &&
				(!q || [m.content, m.title, m.location, m.emotion, ...m.tags].some((v) => v.toLowerCase().includes(q)))
		);
		if (sort === 'oldest') return list.toReversed();
		if (sort === 'importance') return list.toSorted((a, b) => b.importance - a.importance || b.id - a.id);
		return list;
	});
	// Sorted by time, memories go under day headings; by importance, each shows its day.
	let groups = $derived(sort === 'importance' ? [{ day: '', items: shown }] : groupByDay(shown));
</script>

<div role="search">
	<input type="search" bind:value={query} placeholder="Search memories" aria-label="Search memories" />
	<select bind:value={type} aria-label="Type">
		<option value="">All types</option>
		{#each types as t (t)}<option value={t}>{t}</option>{/each}
	</select>
	<select bind:value={emotion} aria-label="Emotion">
		<option value="">All emotions</option>
		{#each emotions as e (e)}<option value={e}>{e}</option>{/each}
	</select>
	<select bind:value={sort} aria-label="Order">
		<option value="newest">Newest first</option>
		<option value="oldest">Oldest first</option>
		<option value="importance">Most important first</option>
	</select>
</div>

{#if !memories.length}
	<p>No memories yet.</p>
{:else if !shown.length}
	<p>No memories match.</p>
{/if}

{#each groups as group (group.items[0]?.id)}
	{#if group.day}<h2>{group.day}</h2>{/if}
	{#each group.items as m (m.id)}
		<article class:inactive={!m.active}>
			<p>
				<span>{timeLabel(m)}{#if !group.day}, {dayLabel(m)}{/if}</span>
				· {m.type}
				{#if m.emotion}· {m.emotion}{/if}
				· importance {m.importance.toFixed(2)}
			</p>
			<ul class="flags">
				{#if m.retired !== undefined}
					<li title="SkyrimNet doesn’t appear to recall memories stored under an earlier identity.">
						stored under “{m.retired || 'an earlier identity'}”
					</li>
				{/if}
				{#if !m.active}<li>inactive</li>{/if}
				{#if m.alwaysInject}<li>always included</li>{/if}
				{#if m.condition}<li title="Only used while this is true: {m.condition}">conditional</li>{/if}
				{#if m.diaryEntry !== undefined}<li>copy of diary entry {m.diaryEntry}</li>{/if}
				{#if !m.embedded}<li title="SkyrimNet rebuilds missing search vectors when it starts">search vector pending</li>{/if}
			</ul>
			{#if m.title}<h3>{m.title}</h3>{/if}
			<p>{m.content}</p>
			{#if m.location}<p><small>{m.location}</small></p>{/if}
		</article>
	{/each}
{/each}

<style>
	.inactive {
		opacity: 0.6;
	}
	.flags {
		display: flex;
		gap: 0.75rem;
		list-style: none;
		padding: 0;
	}
</style>
