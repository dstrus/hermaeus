<script lang="ts">
	import CountList from '#lib/characters/CountList.svelte';
	import { characterName } from '#lib/characters/api.ts';
	import type { Count } from '#lib/api/types/characters.ts';
	import { m } from '#lib/copy/messages.ts';
	import { count } from '#lib/format.ts';
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
			mem.inactive && m.characters.stats.inactive({ count: mem.inactive }),
			mem.alwaysInject && m.characters.stats.alwaysIncluded({ count: mem.alwaysInject }),
			mem.pendingEmbedding && m.characters.stats.pending({ count: mem.pendingEmbedding }),
			mem.unembeddable && m.characters.stats.unsearchable({ count: mem.unembeddable })
		].filter((x): x is string => !!x)
	);
	let nothing = $derived(!s.events.total && !mem.total && !s.diaries);
	let summary = $derived.by(() => {
		let events = m.characters.stats.events({ count: s.events.total });
		if (s.events.total) {
			const detail = [m.characters.stats.started({ count: s.events.started })];
			if (s.events.kills) detail.push(m.characters.stats.kills({ count: s.events.kills }));
			events = m.characters.stats.eventsDetail({ events, detail: detail.join(', ') });
		}
		return [events, m.characters.stats.memories({ count: mem.total }), m.characters.stats.diaryEntries({ count: s.diaries })].join(', ');
	});
</script>

<p class="summary">{summary}</p>

{#if nothing}
	<p class="muted">{m.characters.stats.nothing({ name })}</p>
{/if}

<div class="grid">
	{#if s.events.total}
		<CountList title={m.characters.stats.company} items={s.company} linkTo={person} note={m.characters.stats.companyNote} />
	{/if}
	{#if s.talksTo.length}
		<CountList title={m.characters.stats.talksTo} items={s.talksTo} linkTo={person} note={m.characters.stats.talksToNote({ name })} />
	{/if}
	{#if s.mentions.length || s.unknownMentions}
		<CountList
			title={m.characters.stats.mentions}
			items={s.mentions}
			linkTo={person}
			note={s.unknownMentions ? m.characters.stats.unknownMentions({ count: s.unknownMentions }) : undefined}
		/>
	{/if}
	{#if mem.total}
		<section>
			<h2 class="caps">{m.characters.stats.byType}</h2>
			<table>
				<thead><tr><th class="caps">{m.characters.stats.type}</th><th class="caps n">{m.characters.stats.count}</th><th class="caps n">{m.characters.stats.avgImportance}</th></tr></thead>
				<tbody>
					{#each mem.types as t (t.type)}
						<tr><td>{t.type || m.characters.stats.noType}</td><td class="n">{count(t.n)}</td><td class="n">{t.importance.toFixed(2)}</td></tr>
					{/each}
				</tbody>
			</table>
			{#if health.length}<p class="note">{health.join(' · ')}</p>{/if}
		</section>
		<CountList title={m.characters.stats.mood} items={mem.emotions} blank={m.characters.stats.noEmotion} />
		<CountList title={m.characters.stats.places} items={mem.places} blank={m.characters.stats.noLocation} />
	{/if}
	{#if s.events.total}
		<CountList title={m.characters.stats.eventsByType} items={s.events.types} />
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
		color: var(--accent);
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
