<!-- A playthrough at a glance, in the Playthroughs list's pane: where its
     story stands, what's stored, and its newest diary entry. -->
<script lang="ts">
	import { ApiError } from '#lib/api/client.ts';
	import type { Playthrough } from '#lib/api/types/store.ts';
	import type { Summary } from '#lib/api/types/summary.ts';
	import { count, plural } from '#lib/format.ts';
	import { dayLabel } from '#lib/memories/dates.ts';
	import { doorways, getSummary } from './api.ts';

	let { playthrough: p }: { playthrough: Playthrough } = $props();

	// Each playthrough is read once while the list is open.
	const seen = new Map<string, Summary>();
	let summary = $state<Summary | null>(null);
	let error = $state('');
	$effect(() => {
		const id = p.id;
		const known = seen.get(id) ?? null;
		error = '';
		summary = known;
		if (known) return;
		getSummary(id).then(
			(s) => {
				seen.set(id, s);
				if (p.id === id) summary = s;
			},
			(e) => {
				if (p.id === id) error = e instanceof ApiError ? e.message : String(e);
			}
		);
	});

	let base = $derived(`/${encodeURIComponent(p.id)}`);
	let doors = $derived(summary ? doorways(p.id, summary) : []);
	let most = $derived(Math.max(1, ...doors.map((d) => d.n)));
	let latest = $derived(summary?.diary[0]);
</script>

<div class="title">
	<h2 class="caps">{p.player || 'No player yet'}</h2>
	<a class="button primary" href={base}>Open</a>
</div>
<p class="muted id">{p.id}</p>
{#if error}
	<p class="notice error">{error}</p>
{:else if summary}
	{#if summary.now}
		<p class="now">{summary.now.time}, {summary.now.day}</p>
		<p class="muted small">{plural(summary.days, 'game day', 'game days')} recorded since {summary.began?.day}</p>
	{/if}
	<ul class="doors" aria-label="Stored">
		{#each doors as d (d.id)}
			<li>
				<a class="caps" href={d.href}>{d.label}</a>
				<span class="n">{count(d.n)}</span>
				<span class="bar" style:width="{(d.n / most) * 100}%"></span>
			</li>
		{/each}
	</ul>
	{#if summary.company.length}
		<p class="small"><span class="caps label">Company</span> {summary.company.map((c) => c.name || 'Unnamed').join(', ')}</p>
	{/if}
	{#if latest}
		<section class="latest">
			<h3 class="caps">Latest diary · {dayLabel(latest)}</h3>
			<p>{latest.content.replace(/\n\s*\n/g, '\n')}</p>
		</section>
	{/if}
{/if}

<style>
	h2 {
		font-size: var(--step-2);
		font-weight: 600;
	}
	.id {
		font-size: var(--step--1);
		margin-bottom: var(--space-4);
	}
	.small {
		font-size: var(--step--1);
	}
	.now + .small {
		margin-bottom: var(--space-4);
	}
	ul {
		list-style: none;
		margin: 0 0 var(--space-4);
		padding: 0;
	}
	/* What's stored, as a stat list whose names open their section. */
	.doors li {
		position: relative;
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding: var(--space-1) 0 calc(var(--space-1) + 3px);
		border-bottom: 1px solid var(--rule);
	}
	.doors a {
		text-decoration: none;
	}
	.doors a:hover {
		text-decoration: underline;
	}
	.doors .n {
		color: var(--muted);
	}
	.doors .bar {
		position: absolute;
		left: 0;
		bottom: 2px;
		height: 2px;
		background: var(--accent);
		opacity: 0.55;
	}
	.label {
		color: var(--muted);
		margin-right: var(--space-2);
	}
	/* The newest entry takes what height is left, fading out at the bottom. */
	.latest {
		flex: 1;
		min-height: 4rem;
		margin-top: var(--space-4);
		overflow: hidden;
		mask-image: linear-gradient(to bottom, #000 calc(100% - 3rem), transparent);
	}
	.latest h3 {
		font-size: var(--step--1);
		color: var(--muted);
		margin-bottom: var(--space-2);
	}
	.latest p {
		font-size: var(--step--1);
		line-height: 1.6;
		white-space: pre-line;
	}
	/* The name, with Open beside it so it never falls below the pane. */
	.title {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
	}
	.title a {
		text-decoration: none;
	}
</style>
