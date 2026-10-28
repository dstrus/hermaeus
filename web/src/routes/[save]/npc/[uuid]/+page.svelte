<script lang="ts">
	import { groupByDay, timeLabel } from '#lib/memories/dates.ts';
	import { realTime } from '#lib/format.ts';
	import SpeechPattern from '#lib/characters/SpeechPattern.svelte';
	import Diamond from '#lib/ui/Diamond.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let o = $derived(data.overview);
	let milestones = $derived(groupByDay(o.milestones));
</script>

<div class="overview">
	<section>
		<h2 class="caps">Identity</h2>
		<dl>
			{#if o.firstEvent && o.lastEvent}
				<dt class="caps">Active</dt>
				<dd>{o.firstEvent.day} to {o.lastEvent.day}</dd>
			{/if}
			{#if o.formId}
				<dt class="caps">Form ID</dt>
				<dd class="id">{o.formId}</dd>
			{/if}
			{#if o.modFile}
				<dt class="caps">Plugin</dt>
				<dd>{o.modFile}{#if o.localFormId}<span class="muted">{' · local ID '}<span class="id">{o.localFormId}</span></span>{/if}</dd>
			{/if}
			{#if o.bio}
				<dt class="caps">Bio template</dt>
				<dd class="id">{o.bio}</dd>
			{/if}
			<dt class="caps">SkyrimNet ID</dt>
			<dd class="id">{o.uuid}</dd>
			{#if o.firstSeen}
				<dt class="caps">First seen</dt>
				<dd>{realTime(o.firstSeen)}</dd>
			{/if}
			{#if o.updated}
				<dt class="caps">Record updated</dt>
				<dd>{realTime(o.updated)}</dd>
			{/if}
			{#each o.retired as r (r.uuid)}
				<dt class="caps">Earlier identity</dt>
				<dd>{r.name || 'unnamed'}{#if r.since}<span class="muted">{' · re-identified '}{realTime(r.since)}</span>{/if}</dd>
			{/each}
		</dl>
	</section>

	{#if data.character.bio}
		<SpeechPattern template={data.character.bio} rule={data.speech} loadError={data.speechError} />
	{/if}

	<section class="milestones">
		<h2 class="caps">Milestones</h2>
		{#if milestones.length}
			<ol class="timeline">
				{#each milestones as group (group.items[0]?.eventId)}
					<li class="day caps">{group.day}</li>
					{#each group.items as m (m.eventId)}
						<li class="event"><Diamond /><span class="muted">{timeLabel(m)}</span><span class="summary" title={m.summary}>{m.summary}</span></li>
					{/each}
				{/each}
			</ol>
		{:else}
			<p class="muted">Nothing that marks a turn in their story yet, such as joining you, a trade or a gift.</p>
		{/if}
	</section>
</div>

<style>
	/* Identity and the speech pattern stack on the left; milestones run
	   down the right. */
	.overview {
		display: grid;
		grid-template-columns: minmax(0, 34rem) minmax(0, 1fr);
		grid-template-rows: auto 1fr; /* the first row is Identity's height; the pattern sits right under it */
		grid-auto-flow: dense;
		gap: var(--space-5) var(--space-7);
		max-width: 80rem;
	}
	.milestones {
		grid-column: 2;
		grid-row: 1 / span 2;
	}
	@media (max-width: 60rem) {
		.overview {
			grid-template-columns: 1fr;
		}
		.milestones {
			grid-column: 1;
			grid-row: auto;
		}
	}
	h2 {
		font-size: var(--step-0);
		color: var(--accent);
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule-strong);
		margin-bottom: var(--space-3);
	}
	dl {
		display: grid;
		grid-template-columns: 9rem 1fr;
		gap: var(--space-2) var(--space-4);
		margin: 0;
	}
	dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.2em;
	}
	dd {
		margin: 0;
	}
	.id {
		font-variant-numeric: tabular-nums;
		letter-spacing: 0.02em;
	}
	.timeline {
		list-style: none;
		margin: 0;
		padding: 0 0 0 var(--space-3);
		border-left: 1px solid var(--rule-strong);
	}
	.timeline .day {
		font-size: var(--step--1);
		color: var(--muted);
		margin: var(--space-4) 0 var(--space-1);
	}
	.timeline .day:first-child {
		margin-top: 0;
	}
	.event {
		display: flex;
		align-items: baseline;
		gap: var(--space-3);
		padding: var(--space-1) 0;
		margin-left: calc(-1 * var(--space-3) - 5px);
	}
	.summary {
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		overflow: hidden;
	}
	@media (max-width: 70rem) {
		.overview {
			grid-template-columns: 1fr;
		}
	}
</style>
