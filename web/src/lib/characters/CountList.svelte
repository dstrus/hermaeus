<!-- A ranked list: what comes up most, with how often, and a bar for scale. -->
<script lang="ts">
	import type { Count } from '#lib/api/types/characters.ts';
	import { count } from '#lib/format.ts';

	let {
		title,
		items,
		linkTo,
		note,
		empty = 'None yet.',
		blank = 'none'
	}: {
		title: string;
		items: Count[];
		linkTo?: (c: Count) => string | undefined; // where an item links, if anywhere
		note?: string; // what the counts are
		empty?: string;
		blank?: string; // what to call an item with no name
	} = $props();

	let most = $derived(Math.max(1, ...items.map((c) => c.n)));
</script>

<section>
	<h2 class="caps">{title}</h2>
	{#if items.length}
		<table>
			<tbody>
				{#each items as c, i (c.uuid ?? c.key ?? `${c.name}:${i}`)}
					{@const href = linkTo?.(c)}
					<tr>
						<td class="name">
							{#if href}<a {href}>{c.name}</a>{:else if c.name}{c.name}{:else}<span class="muted">{blank}</span>{/if}
							<span class="bar" style:width="{(c.n / most) * 100}%"></span>
						</td>
						<td class="n">{count(c.n)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p class="muted">{empty}</p>
	{/if}
	{#if note}<p class="note muted">{note}</p>{/if}
</section>

<style>
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
	td {
		padding: var(--space-1) 0;
		border-bottom: 1px solid var(--rule);
		vertical-align: top;
	}
	.name {
		position: relative;
		padding-bottom: calc(var(--space-1) + 3px);
	}
	.name a {
		text-decoration: none;
	}
	.name a:hover {
		text-decoration: underline;
	}
	.bar {
		position: absolute;
		left: 0;
		bottom: 2px;
		height: 2px;
		background: var(--accent);
		opacity: 0.55;
	}
	.n {
		text-align: right;
		padding-left: var(--space-4);
		color: var(--muted);
		width: 1%;
		white-space: nowrap;
	}
	.note {
		font-size: var(--step--1);
		margin-top: var(--space-2);
	}
</style>
