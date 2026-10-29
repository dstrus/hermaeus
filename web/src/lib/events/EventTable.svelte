<!-- Everything SkyrimNet recorded a character taking part in, newest first
     under the day it happened, with the selected event's details in the
     pane. Events are SkyrimNet's record of play; they're read, not edited. -->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import type { Character, Event, EventsResponse } from '#lib/api/types/characters.ts';
	import { count, realTime } from '#lib/format.ts';
	import { dayLabel, timeLabel } from '#lib/memories/dates.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { clipped } from '#lib/ui/clipped.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { characterName } from '#lib/characters/api.ts';
	import { fieldLabel } from './api.ts';

	let {
		save,
		uuid,
		name = 'They',
		events: response,
		characters
	}: {
		save: string;
		uuid?: string; // whose events these are; everyone's when left out
		name?: string;
		events: EventsResponse;
		characters: Character[];
	} = $props();

	// One character's own events, where they act or are acted on, read in
	// full ink; those they were merely present for are dimmed, so theirs
	// surface by contrast. Likewise their name among the labels. Everyone's
	// events are all in full ink.
	const theirs = (e: Event) => !uuid || e.role !== 'present';
	const self = (who: string | undefined) => !!uuid && who === uuid;
	// Everyone's events can be narrowed to one character's doings.
	let actor = $state('');

	// Who can be opened: the characters with something stored, by UUID.
	let known = $derived(new Set(characters.map((c) => c.uuid)));
	const pageOf = (uuid: string | undefined) => (uuid && known.has(uuid) ? `/${encodeURIComponent(save)}/npc/${uuid}` : undefined);

	let query = $state('');
	let type = $state('');
	let role = $state(''); // '', 'part' (acting or acted on), or 'present'

	let matching = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return response.events.filter(
			(e) =>
				(!type || e.type === type) &&
				(!role || (role === 'present' ? e.role === 'present' : e.role !== 'present')) &&
				(!actor || e.fromUuid === actor || e.toUuid === actor) &&
				(!q || [e.summary, e.label, e.location ?? '', e.text ?? ''].some((v) => v.toLowerCase().includes(q)))
		);
	});
	// A character can have thousands of events; the list grows as asked.
	const page = 300;
	let limit = $state(page);
	$effect(() => {
		void matching;
		limit = page;
	});
	let shown = $derived(matching.slice(0, limit));

	const selection = new Selection(() => shown, (e) => String(e.id), { param: 'item' });
	let current = $derived(selection.current);

	// The details shown in the pane: every field but the ones the summary
	// already says, as words.
	let fields = $derived.by(() => {
		if (!current?.data) return [];
		return Object.entries(current.data)
			.filter(([, v]) => v !== null && v !== '' && v !== false)
			.map(([k, v]) => ({ key: k, label: fieldLabel(k), value: typeof v === 'string' ? v : JSON.stringify(v) }));
	});

	useShortcuts(() => [
		{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) }
	]);
</script>

<!-- A name as a label on the words, opening the character when they're
     known; the character whose events these are stands out among them. -->
{#snippet who(name: string, href: string | undefined, own: boolean)}
	{#if href}<a class="name" class:own {href}>{name}</a>{:else}<span class="name" class:own>{name}</span>{/if}
{/snippet}

<div class="tools">
	<span class="muted">{count(response.events.length)} {response.events.length === 1 ? 'event' : 'events'}</span>
	<select bind:value={type} aria-label="Type">
		<option value="">Every type</option>
		{#each response.types as t (t.type)}
			<option value={t.type}>{t.label} · {count(t.count)}</option>
		{/each}
	</select>
	{#if uuid}
		<select bind:value={role} aria-label="Their part">
			<option value="">Any part</option>
			<option value="part">Acting or acted on</option>
			<option value="present">Only present</option>
		</select>
	{:else}
		<select bind:value={actor} aria-label="Character">
			<option value="">Everyone</option>
			{#each characters as c (c.uuid)}
				<option value={c.uuid}>{characterName(c)}</option>
			{/each}
		</select>
	{/if}
	<input type="search" bind:value={query} placeholder="Filter events" aria-label="Filter events" />
</div>

<div class="split" class:everyone={!uuid}>
	<div class="list">
		{#if !response.events.length}
			<p class="muted">SkyrimNet hasn’t recorded {name} taking part in anything yet.</p>
		{:else if !shown.length}
			<p class="muted">No events match.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th class="mark" aria-label="Selected"></th>
						<th class="caps when">When</th>
						{#if !uuid}<th class="caps who">Who</th>{/if}
						<th class="caps text">Event</th>
						<th class="caps type">Type</th>
					</tr>
				</thead>
				<tbody>
					{#each shown as e, i (e.id)}
						{#if dayLabel(e) !== (shown[i - 1] ? dayLabel(shown[i - 1]!) : '')}
							<tr class="day"><th colspan={uuid ? 4 : 5} scope="rowgroup" class="caps sticky-day">{dayLabel(e)}</th></tr>
						{/if}
						<tr aria-selected={selection.is(e)} class:bystander={!theirs(e)} onclick={() => selection.select(e)}>
							<td class="mark">{#if selection.is(e)}<Diamond />{/if}</td>
							<td class="when">{timeLabel(e)}</td>
							{#if !uuid}
								<td class="who" use:clipped>
									{#if e.from}{@render who(e.from, pageOf(e.fromUuid), false)}{/if}{#if e.from && e.to}{' › '}{/if}{#if e.to}{@render who(e.to, pageOf(e.toUuid), false)}{/if}
								</td>
							{/if}
							<td class="text">
								<span class="line">
									{#if e.speaker && e.line}
										{@render who(e.speaker, pageOf(e.speakerUuid), self(e.speakerUuid))}{#if e.listener}{' to '}{@render who(e.listener, pageOf(e.listenerUuid), self(e.listenerUuid))}{/if}{': '}“{e.line}”
									{:else}
										{e.summary}
									{/if}
								</span>
							</td>
							<td class="type caps" use:clipped>{e.label}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if matching.length > shown.length}
				<p class="more">
					<button type="button" class="textbutton" onclick={() => (limit += page)}>Show {count(Math.min(page, matching.length - shown.length))} more</button>
					<span class="muted">of {count(matching.length - shown.length)} further</span>
				</p>
			{/if}
		{/if}
	</div>

	<aside class="pane" use:fitToViewport aria-label="Event">
		{#if current}
			{#key current.id}
				<div in:fly={{ y: 6, duration: 180 }}>
					<h2 class="caps">{current.label}</h2>
					<p class="muted">{current.date ? `${current.date.time}, ${current.date.day}` : dayLabel(current)}</p>
					<p class="content">
						{#if current.speaker && current.line}
							{@render who(current.speaker, pageOf(current.speakerUuid), self(current.speakerUuid))}{#if current.listener}{' to '}{@render who(current.listener, pageOf(current.listenerUuid), self(current.listenerUuid))}{/if}{': '}“{current.line}”
						{:else}
							{current.text || current.summary}
						{/if}
					</p>
					<dl>
						{#if current.location}<dt class="caps">Place</dt><dd>{current.location}</dd>{/if}
						{#if current.role}<dt class="caps">{name}’s part</dt><dd>{current.role}</dd>{/if}
						{#if !uuid && current.from}<dt class="caps">From</dt><dd>{@render who(current.from, pageOf(current.fromUuid), false)}</dd>{/if}
						{#if !uuid && current.to}<dt class="caps">To</dt><dd>{@render who(current.to, pageOf(current.toUuid), false)}</dd>{/if}
						{#if current.present}<dt class="caps">Present</dt><dd>{count(current.present)} {current.present === 1 ? 'character' : 'characters'}</dd>{/if}
						{#if current.written}<dt class="caps">Recorded</dt><dd>{realTime(current.written)}</dd>{/if}
						<dt class="caps">Event</dt><dd>#{current.id}</dd>
					</dl>
					{#if fields.length}
						<h3 class="caps">Details</h3>
						<dl>
							{#each fields as f (f.key)}
								<dt class="caps">{f.label}</dt><dd>{f.value}</dd>
							{/each}
						</dl>
					{/if}
				</div>
			{/key}
		{/if}
	</aside>
</div>

<style>
	.tools {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-3);
	}
	.tools input {
		width: 16rem;
		max-width: 100%;
	}
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(20rem, 26rem);
		gap: var(--space-5);
		align-items: start;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		vertical-align: top;
		white-space: nowrap;
	}
	tbody tr {
		cursor: default;
	}
	tbody tr[aria-selected]:not([aria-selected='true']):hover td {
		background: var(--glass);
	}
	tr[aria-selected='true'] td {
		background: var(--select);
	}
	tr.bystander td {
		color: var(--muted);
	}
	th {
		text-align: left;
		padding: var(--space-2) var(--space-3);
		font-size: var(--step--1);
		color: var(--muted);
		font-weight: inherit;
		border-bottom: 1px solid var(--rule-strong);
		white-space: nowrap;
	}
	.mark {
		width: 1.5rem;
		padding-right: 0;
	}
	tbody .mark {
		border-bottom-color: var(--rule);
	}
	.when {
		color: var(--muted);
	}
	.text {
		width: 100%;
		max-width: 0;
		white-space: normal;
	}
	/* Who spoke and to whom are labels on the words, not part of them. */
	.name {
		font-family: var(--font-label);
		font-weight: 500;
		font-size: var(--step--1);
		text-transform: uppercase;
		letter-spacing: var(--track);
		color: var(--muted);
	}
	.name.own {
		color: var(--ink);
	}
	tr.bystander .name.own {
		color: inherit;
	}
	/* A name that opens a character is underlined like any link here: in
	   the hairline at rest, the accent under the pointer. */
	a.name {
		text-decoration: underline;
		text-decoration-color: var(--rule-strong);
		text-underline-offset: 0.25em;
	}
	a.name:hover {
		color: var(--ink);
		text-decoration-color: var(--accent);
	}
	/* Two lines of the summary at most in the list; the pane has it all. */
	.line {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	td.who {
		max-width: 9rem;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	/* Everyone's events carry a Who column, so the type gives way sooner;
	   the type filter and the pane still say it. */
	@media (max-width: 90rem) {
		.split.everyone :global(.type) {
			display: none;
		}
	}
	td.type {
		max-width: 11rem;
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: var(--step--1);
		color: var(--muted);
	}
	.more {
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3);
	}
	.day th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	.pane {
		position: sticky;
		top: 0;
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
		overflow: auto;
	}
	.pane h2 {
		font-size: var(--step-2);
		font-weight: 600;
		margin-bottom: var(--space-1);
	}
	.pane h3 {
		font-size: var(--step--1);
		color: var(--muted);
		margin: var(--space-4) 0 var(--space-2);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule);
	}
	.content {
		margin: var(--space-3) 0;
		white-space: pre-wrap;
		line-height: 1.6;
	}
	dl {
		display: grid;
		grid-template-columns: 7rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: 0;
	}
	dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
		overflow-wrap: anywhere;
	}
	dd {
		margin: 0;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	@media (max-width: 60rem) {
		.split {
			grid-template-columns: 1fr;
		}
		.pane {
			position: static;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
	}
</style>
