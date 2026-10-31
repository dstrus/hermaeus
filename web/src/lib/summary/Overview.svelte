<!-- A playthrough's Overview, read like the game's Journal: what was written
     and done last on the left, and beside it what SkyrimNet has stored, as
     the doorways into each section, with who travels along and how
     SkyrimNet is tuned. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Event } from '#lib/api/types/characters.ts';
	import type { Models } from '#lib/api/types/settings.ts';
	import type { Summary } from '#lib/api/types/summary.ts';
	import type { Toggle } from '#lib/api/types/settings.ts';
	import { engineByValue, noVoice } from '#lib/features/voice.ts';
	import { count, plural, realTime } from '#lib/format.ts';
	import { dayLabel, groupByDay, timeLabel } from '#lib/memories/dates.ts';
	import { href as pageHref } from '#lib/nav.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { clipped } from '#lib/ui/clipped.ts';
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { Selection } from '#lib/ui/selection.svelte.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { doorways } from './api.ts';

	let {
		save,
		player,
		summary: s,
		models,
		engine,
		toggles
	}: {
		save: string;
		player: string;
		summary: Summary;
		models: Models | null;
		engine: string | null; // tts.engine, or null when SkyrimNet.yaml couldn't be read
		toggles: Toggle[];
	} = $props();

	let base = $derived(`/${encodeURIComponent(save)}`);
	let doors = $derived(doorways(save, s));
	let most = $derived(Math.max(1, ...doors.map((d) => d.n)));

	// The doorways are the page's list: ↑ and ↓ mark one, Enter opens it.
	const selection = new Selection(() => doors, (d) => d.id);
	let marked = $derived(selection.current);

	let days = $derived(groupByDay(s.events));
	// A place by its own name, without the hold or cell SkyrimNet adds.
	const place = (where: string | undefined) => (where ?? '').split(',')[0]!.trim();
	// An excerpt keeps the entry's words; only the blank lines between its
	// paragraphs close up, so more of it fits.
	const excerpt = (text: string) => text.replace(/\n\s*\n/g, '\n');
	// A day without its era and year, which the header already shows.
	const shortDay = (day: string) => day.replace(/, \d*E \d+$/, '');
	const npc = (uuid: string | undefined) => (uuid ? `${base}/npc/${uuid}` : undefined);
	const eventHref = (e: Event) => `${base}/history/events?item=${e.id}`;
	const diaryHref = (id: number) => `${base}/history/diary?item=${id}`;

	// Tuning, as far as SkyrimNet's settings files say.
	let dialogue = $derived(models?.tasks.find((t) => t.id === 'default'));
	let voice = $derived(engine === null ? '' : engine === noVoice ? 'None' : (engineByValue(engine)?.label ?? engine));
	let on = $derived(toggles.filter((t) => t.available && t.on));
	let off = $derived(toggles.filter((t) => t.available && !t.on));
	let master = $derived(toggles.find((t) => t.master));

	let nothing = $derived(!s.counts.events && !s.counts.memories && !s.counts.diary);

	useShortcuts(() => [
		{ key: 'ArrowDown', label: 'Next', hidden: true, run: () => selection.move(1) },
		{ key: 'ArrowUp', label: 'Previous', hidden: true, run: () => selection.move(-1) },
		{ key: 'Enter', label: marked ? `Open ${marked.label}` : 'Open', disabled: !marked, run: () => marked && goto(marked.href) }
	]);
</script>

<!-- A name as a label on the words, opening the character when known. -->
{#snippet who(name: string, uuid: string | undefined)}
	{@const to = npc(uuid)}
	{#if to}<a class="name" href={to}>{name}</a>{:else}<span class="name">{name}</span>{/if}
{/snippet}

<div class="split">
	<div class="story">
		<header>
			<h1 class="caps">{player || 'No player yet'}</h1>
			<p class="status caps">
				{#if s.now}<span>{s.now.time}, {s.now.day}</span>{/if}
				{#if s.days}<span>{plural(s.days, 'game day', 'game days')} recorded</span>{/if}
				{#if s.written}<span>last event recorded {realTime(s.written)}</span>{/if}
			</p>
		</header>

		{#if nothing}
			<p class="empty">SkyrimNet hasn’t recorded anything in this playthrough yet. It starts when you play with it: events as they happen, then memories and diary entries made from them.</p>
		{:else}
			<section class="diary">
				<h2 class="caps">The latest diary</h2>
				{#each s.diary as d (d.id)}
					<article>
						<p class="dateline">
							{#if d.actor}{@render who(d.actor, d.actorUuid)}{/if}
							<span class="caps when" use:clipped>{shortDay(dayLabel(d))}{#if d.location}{' · '}{place(d.location)}{/if}{#if d.emotion}{' · '}{d.emotion}{/if}</span>
							<a class="textbutton" href={diaryHref(d.id)}>Read it in History</a>
						</p>
						<p class="entry">{excerpt(d.content)}</p>
					</article>
				{:else}
					<p class="muted">No diary entries yet. SkyrimNet’s characters write them as the days go by.</p>
				{/each}
			</section>

			<section class="lately">
				<h2 class="caps">Lately</h2>
				{#if s.events.length}
					<table>
						<tbody>
							{#each days as g (g.day)}
								<tr class="day"><th colspan="3" scope="rowgroup" class="caps">{g.day}</th></tr>
								{#each g.items as e, i (e.id)}
									{@const where = place(e.location) !== place(g.items[i - 1]?.location) ? place(e.location) : ''}
									<tr>
										<td class="when"><a href={eventHref(e)}>{timeLabel(e)}</a></td>
										<td class="text" colspan={where ? 1 : 2}>
											<span class="line" use:clipped>
												{#if e.speaker && e.line}
													{@render who(e.speaker, e.speakerUuid)}{#if e.listener}{' to '}{@render who(e.listener, e.listenerUuid)}{/if}{': '}“{e.line}”
												{:else}
													{e.summary}
												{/if}
											</span>
										</td>
										{#if where}<td class="where" use:clipped>{where}</td>{/if}
									</tr>
								{/each}
							{/each}
						</tbody>
					</table>
					<a class="textbutton more" href="{base}/history/events">Every event, in History</a>
				{:else}
					<p class="muted">No events yet.</p>
				{/if}
			</section>
		{/if}
	</div>

	<aside class="pane" aria-label="This playthrough" use:fitToViewport>
		<h2 class="caps title">This playthrough</h2>
		<p class="muted id">{save}</p>

		<section>
			<h3 class="caps">Stored</h3>
			<ul class="doors" aria-label="Sections">
				{#each doors as d (d.id)}
					<li aria-current={selection.is(d) ? 'true' : undefined}>
						<span class="mark">{#if selection.is(d)}<Diamond />{/if}</span>
						<a class="caps" href={d.href} onmouseenter={() => (selection.key = d.id)} onfocus={() => (selection.key = d.id)}>{d.label}</a>
						<span class="n">{count(d.n)}</span>
						<span class="bar" style:width="{(d.n / most) * 100}%"></span>
					</li>
				{/each}
			</ul>
		</section>

		{#if s.company.length}
			<section>
				<h3 class="caps">Company</h3>
				<p class="company">
					{#each s.company as c, i (c.uuid)}{#if i}{', '}{/if}<a href={npc(c.uuid)}>{c.name || 'Unnamed'}</a>{#if c.pinned}{' '}<span class="muted">(pinned)</span>{/if}{/each}
				</p>
				{#if s.company.some((c) => c.traveling)}<p class="muted note">{plural(s.company.filter((c) => c.traveling).length, 'travels', 'travel')} with you.</p>{/if}
			</section>
		{/if}

		<section>
			<h3 class="caps">Tuning</h3>
			<dl>
				{#if dialogue}
					<dt class="caps">Dialogue</dt>
					<dd>
						{dialogue.slots[0]?.model ?? 'No model yet'}
						{#if dialogue.slots.length > 1}<span class="muted">and {plural(dialogue.slots.length - 1, 'other', 'others')} in turn</span>{/if}
					</dd>
				{/if}
				{#if voice}<dt class="caps">Voice</dt><dd>{voice}</dd>{/if}
				{#if toggles.length}
					<dt class="caps">Toggles</dt>
					<dd>
						{#if master && !master.on}
							<span class="off">AI off: SkyrimNet makes no AI calls</span>
						{:else}
							{count(on.length)} of {count(on.length + off.length)} on
							{#if off.length}<span class="muted">Off: {off.map((t) => t.label).join(', ')}</span>{/if}
						{/if}
					</dd>
				{/if}
			</dl>
			<p class="links">
				<a class="textbutton" href={pageHref(save, 'models')}>Models</a>
				<a class="textbutton" href={pageHref(save, 'engines')}>Engines</a>
				<a class="textbutton" href={pageHref(save, 'prompts')}>Prompts</a>
			</p>
		</section>
	</aside>
</div>

<style>
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) clamp(18rem, 26vw, 24rem);
		gap: var(--space-6);
		align-items: start;
	}
	header {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		margin-bottom: var(--space-5);
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	.status {
		display: flex;
		flex-wrap: wrap;
		gap: 0 var(--space-4);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.empty {
		max-width: 60ch;
	}
	section + section {
		margin-top: var(--space-6);
	}
	.story h2 {
		font-size: var(--step-0);
		color: var(--accent);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-4);
	}
	article + article {
		margin-top: var(--space-5);
	}
	/* Whose entry, when and where, and the way to it, on one line across
	   the column. */
	.dateline {
		display: flex;
		align-items: baseline;
		gap: var(--space-3);
		font-size: var(--step--1);
		margin-bottom: var(--space-1);
	}
	.dateline .when {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--muted);
	}
	.dateline .textbutton {
		white-space: nowrap;
	}
	/* The diary reads as prose, at a reading measure. */
	.entry {
		max-width: 65ch;
		line-height: 1.6;
		white-space: pre-line;
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	/* Who spoke and to whom are labels on the words, as in History. */
	.name {
		font-family: var(--font-label);
		font-weight: 500;
		font-size: var(--step--1);
		text-transform: uppercase;
		letter-spacing: var(--track);
		color: var(--muted);
	}
	a.name {
		text-decoration: underline;
		text-underline-offset: 0.25em;
	}
	a.name:hover {
		color: var(--ink);
		text-decoration-color: var(--accent);
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		font-weight: inherit;
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule-strong);
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--rule);
		vertical-align: top;
	}
	tbody tr:not(.day):hover td {
		background: var(--glass);
	}
	.day th {
		padding: var(--space-5) var(--space-3) var(--space-1);
		color: var(--accent);
	}
	/* The first day sits right under the heading. */
	tbody .day:first-child th {
		padding-top: 0;
	}
	/* On a shorter window, two lines of each entry, so the newest events
	   still start in view. */
	@media (max-height: 50rem) {
		.entry {
			-webkit-line-clamp: 2;
			line-clamp: 2;
		}
	}
	td.when {
		width: 6rem;
		white-space: nowrap;
	}
	td.when a {
		color: var(--muted);
	}
	td.when a:hover {
		color: var(--ink);
	}
	.text {
		max-width: 0;
		width: 100%;
	}
	.line {
		display: block;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.where {
		width: 1%;
		max-width: 12rem;
		font-size: var(--step--1);
		color: var(--muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.more {
		display: inline-block;
		margin-top: var(--space-3);
		font-size: var(--step--1);
	}

	/* The pane, as on every list screen. */
	.pane {
		position: sticky;
		top: 0;
		min-height: calc(100vh - 7rem); /* fitToViewport sets the height */
		overflow-y: auto;
		padding: var(--space-5);
		background: var(--panel);
		border-left: 1px solid var(--rule-strong);
	}
	.pane .title {
		font-size: var(--step-2);
		font-weight: 600;
	}
	.id {
		font-size: var(--step--1);
		margin-bottom: var(--space-4);
	}
	.pane section + section {
		margin-top: var(--space-5);
	}
	.pane h3 {
		font-size: var(--step--1);
		color: var(--accent);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-1);
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	/* The doorways: a stat list whose names open their section. */
	.doors li {
		position: relative;
		display: grid;
		grid-template-columns: 1.25rem 1fr auto;
		align-items: baseline;
		padding: 0.125rem 0 calc(0.125rem + 3px);
		border-bottom: 1px solid var(--rule);
	}
	.doors li[aria-current] {
		background: var(--select);
	}
	.doors a {
		font-size: var(--step-1);
		text-decoration: none;
		color: var(--muted);
	}
	.doors li[aria-current] a,
	.doors a:hover {
		color: var(--ink);
	}
	.doors .n {
		color: var(--muted);
		padding-right: var(--space-2);
	}
	.doors .bar {
		position: absolute;
		left: 1.25rem;
		bottom: 2px;
		max-width: calc(100% - 1.25rem);
		height: 2px;
		background: var(--accent);
		opacity: 0.55;
	}
	.mark {
		display: inline-flex;
		justify-content: center;
	}
	.company {
		margin-top: var(--space-2);
	}
	.note {
		font-size: var(--step--1);
		margin-top: var(--space-1);
	}
	.company a {
		text-decoration: none;
	}
	.company a:hover {
		text-decoration: underline;
	}
	dl {
		display: grid;
		grid-template-columns: 6rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: var(--space-2) 0 var(--space-3);
	}
	dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	dd {
		margin: 0;
		overflow-wrap: anywhere;
	}
	dd .muted {
		display: block;
		font-size: var(--step--1);
	}
	.off {
		color: var(--warn);
	}
	.links {
		display: flex;
		gap: var(--space-4);
	}
	@media (max-width: 60rem) {
		.split {
			grid-template-columns: minmax(0, 1fr);
		}
		.pane {
			position: static;
			min-height: 0;
			border-left: 0;
			border-top: 1px solid var(--rule-strong);
		}
		.where {
			display: none;
		}
		/* Stacked, a dateline wraps rather than losing its place and mood. */
		.dateline {
			flex-wrap: wrap;
			gap: 0 var(--space-3);
		}
		.dateline .when {
			flex-basis: 100%;
			order: 3;
			white-space: normal;
		}
	}
</style>
