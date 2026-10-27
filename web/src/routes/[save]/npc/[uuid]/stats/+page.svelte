<script lang="ts">
	import CountList from '#lib/characters/CountList.svelte';
	import { characterName } from '#lib/characters/api.ts';
	import type { Count } from '#lib/api/types/characters.ts';
	import { count, plural } from '#lib/format.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let s = $derived(data.stats);
	let name = $derived(characterName(data.character));
	let mem = $derived(s.memories);

	// People link to their page when they're in the character list.
	let listed = $derived(new Set(data.characters.map((c) => c.uuid)));
	let person = $derived((c: Count) =>
		c.uuid && listed.has(c.uuid) ? `/${encodeURIComponent(data.playthrough.id)}/npc/${c.uuid}/stats` : undefined
	);
	let nothing = $derived(!s.events.total && !mem.total && !s.diaries);
	let summary = $derived.by(() => {
		let events = plural(s.events.total, 'event', 'events');
		if (s.events.total) {
			const detail = [`${count(s.events.started)} they started`];
			if (s.events.kills) detail.push(plural(s.events.kills, 'kill', 'kills'));
			events += ` (${detail.join(', ')})`;
		}
		return [events, plural(mem.total, 'memory', 'memories'), plural(s.diaries, 'diary entry', 'diary entries')].join(', ');
	});
</script>

<p>{summary}</p>

{#if nothing}
	<p>SkyrimNet hasn’t recorded anything about {name} in this playthrough yet.</p>
{/if}

{#if s.events.total}
	<h2>Spends time with</h2>
	<CountList items={s.company} linkTo={person} />
	<p><small>Events they were both part of.</small></p>
{/if}

{#if s.talksTo.length}
	<h2>Talks to</h2>
	<CountList items={s.talksTo} linkTo={person} />
	<p><small>Lines {name} spoke to them.</small></p>
{/if}

{#if mem.total}
	<h2>Memories</h2>
	<ul>
		{#each mem.types as t (t.type)}
			<li>{t.type || 'No type'} — {count(t.n)} (average importance {t.importance.toFixed(2)})</li>
		{/each}
	</ul>
	<ul class="flags">
		{#if mem.inactive}<li>{count(mem.inactive)} inactive</li>{/if}
		{#if mem.alwaysInject}<li>{count(mem.alwaysInject)} always included</li>{/if}
		{#if mem.pendingEmbedding}
			<li title="SkyrimNet builds their search vectors the next time it starts; until then, it can’t find them by meaning">
				{count(mem.pendingEmbedding)} not yet searchable
			</li>
		{/if}
		{#if mem.unembeddable}
			<li title="SkyrimNet couldn’t build search vectors for these, so it can’t find them by meaning">
				{count(mem.unembeddable)} unsearchable
			</li>
		{/if}
	</ul>
	<h3>Mood</h3>
	<CountList items={mem.emotions} blank="no emotion" />
	<h3>Places</h3>
	<CountList items={mem.places} blank="no location" />
{/if}

{#if s.mentions.length || s.unknownMentions}
	<h2>Mentioned in memories</h2>
	<CountList items={s.mentions} linkTo={person} />
	{#if s.unknownMentions}
		<p><small>Plus {plural(s.unknownMentions, 'mention', 'mentions')} of actors SkyrimNet has no name for, such as enemies that died.</small></p>
	{/if}
{/if}

{#if s.events.total}
	<h2>Events by type</h2>
	<CountList items={s.events.types} />
{/if}

<style>
	.flags {
		display: flex;
		gap: 0.75rem;
		list-style: none;
		padding: 0;
	}
</style>
