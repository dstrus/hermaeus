<!-- Actions: what NPCs can do when the model picks it, from your plugins
     and your own edits. -->
<script lang="ts">
	import type { ActionList } from '#lib/api/types/records.ts';
	import { m } from '#lib/copy/messages.ts';
	import Inja from '#lib/ui/Inja.svelte';
	import RecordsPage from './RecordsPage.svelte';
	import { words } from './api.ts';

	let { list }: { list: ActionList } = $props();
	const msg = m.records.actions;
</script>

<RecordsPage
	title={msg.title}
	copy={msg}
	lead={msg.lead}
	records={list.actions}
	columns={[{ label: msg.category, class: 'category' }, { label: msg.priority, class: 'priority' }]}
>
	{#snippet cells(a)}
		<td class="category muted">{a.category ? words(a.category) : ''}</td>
		<td class="priority muted">{a.priority ?? ''}</td>
	{/snippet}
	{#snippet pane(a)}
		<p class="text"><Inja text={a.description} /></p>
		<dl>
			{#if a.category}<dt class="caps">{msg.category}</dt><dd>{words(a.category)}</dd>{/if}
			{#if a.priority !== undefined}<dt class="caps">{msg.priority}</dt><dd>{a.priority}</dd>{/if}
			<dt class="caps">{msg.whoMay}</dt><dd>{a.rules ? msg.rules({ count: a.rules }) : msg.anyone}</dd>
			{#if a.function}<dt class="caps">{msg.runs}</dt><dd class="id">{a.script}.{a.function}</dd>{/if}
			{#if a.quest}<dt class="caps">{msg.quest}</dt><dd class="id">{a.quest}</dd>{/if}
			{#if a.event}<dt class="caps">{msg.eventText}</dt><dd class="id">{a.event}</dd>{/if}
			{#if a.confirm}<dt class="caps">{msg.confirm}</dt><dd>{msg.asksFirst}</dd>{/if}
		</dl>
		{#if a.params.length}
			<h3 class="caps">{msg.fillsIn}</h3>
			<ul class="params">
				{#each a.params as p (p.name)}
					<li><span class="id">{p.name}</span><span class="muted kind">· {p.optional ? msg.optional({ kind: words(p.type) }) : words(p.type)}</span>{#if p.description}<br /><span class="small">{p.description}</span>{/if}</li>
				{/each}
			</ul>
		{/if}
	{/snippet}
	{#snippet note()}
		{msg.note}
	{/snippet}
</RecordsPage>

<style>
	:global(td.category),
	:global(th.category) {
		width: 11rem;
	}
	:global(td.priority),
	:global(th.priority) {
		width: 5.5rem;
		text-align: right;
	}
	.text {
		white-space: pre-wrap;
		line-height: 1.5;
		margin: var(--space-3) 0;
	}
	dl {
		display: grid;
		grid-template-columns: 5.5rem 1fr;
		gap: var(--space-1) var(--space-3);
		margin: var(--space-4) 0;
	}
	dt {
		color: var(--muted);
		font-size: var(--step--1);
		padding-top: 0.15em;
	}
	dd {
		margin: 0;
		min-width: 0;
		overflow-wrap: anywhere;
	}
	h3 {
		font-size: var(--step--1);
		color: var(--accent);
		margin: var(--space-4) 0 var(--space-2);
	}
	.params {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	/* The parameter's kind, a step away from its monospace name. */
	.kind {
		margin-left: var(--space-2);
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
</style>
