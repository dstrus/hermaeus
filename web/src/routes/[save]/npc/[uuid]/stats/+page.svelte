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
	// Memories SkyrimNet can't use as it should, worth knowing about.
	let health = $derived(
		[
			mem.inactive && `${count(mem.inactive)} inactive`,
			mem.alwaysInject && `${count(mem.alwaysInject)} always included`,
			mem.pendingEmbedding && `${count(mem.pendingEmbedding)} not yet searchable (SkyrimNet builds the vectors when it starts)`,
			mem.unembeddable && `${count(mem.unembeddable)} unsearchable`
		].filter((x): x is string => !!x)
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

<p class="summary">{summary}</p>

{#if nothing}
	<p class="muted">SkyrimNet hasn’t recorded anything about {name} in this playthrough yet.</p>
{/if}

<div class="grid">
	{#if s.events.total}
		<CountList title="Spends time with" items={s.company} linkTo={person} note="Events they were both part of." />
	{/if}
	{#if s.talksTo.length}
		<CountList title="Talks to" items={s.talksTo} linkTo={person} note="Lines {name} spoke to them." />
	{/if}
	{#if s.mentions.length || s.unknownMentions}
		<CountList
			title="Mentioned in memories"
			items={s.mentions}
			linkTo={person}
			note={s.unknownMentions ? `Plus ${plural(s.unknownMentions, 'mention', 'mentions')} of actors SkyrimNet has no name for, such as enemies that died.` : undefined}
		/>
	{/if}
	{#if mem.total}
		<section>
			<h2 class="caps">Memories by type</h2>
			<table>
				<thead><tr><th class="caps">Type</th><th class="caps n">Count</th><th class="caps n">Avg. importance</th></tr></thead>
				<tbody>
					{#each mem.types as t (t.type)}
						<tr><td>{t.type || 'No type'}</td><td class="n">{count(t.n)}</td><td class="n">{t.importance.toFixed(2)}</td></tr>
					{/each}
				</tbody>
			</table>
			{#if health.length}<p class="note">{health.join(' · ')}</p>{/if}
		</section>
		<CountList title="Mood" items={mem.emotions} blank="no emotion" />
		<CountList title="Places" items={mem.places} blank="no location" />
	{/if}
	{#if s.events.total}
		<CountList title="Events by type" items={s.events.types} />
	{/if}
</div>

<style>
	.summary {
		font-size: var(--step-1);
		margin-bottom: var(--space-5);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(19rem, 1fr));
		gap: var(--space-6) var(--space-7);
		max-width: 80rem;
	}
	h2 {
		font-size: var(--step-0);
		color: var(--gold);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		padding: var(--space-1) 0;
		border-bottom: 1px solid var(--rule);
		font-weight: inherit;
	}
	td {
		padding: var(--space-1) 0;
		border-bottom: 1px solid var(--rule);
	}
	.n {
		text-align: right;
		padding-left: var(--space-4);
	}
	.note {
		font-size: var(--step--1);
		margin-top: var(--space-2);
		color: var(--warn);
	}
</style>
