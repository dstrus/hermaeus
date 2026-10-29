<!-- Actions: what NPCs can do when the model picks it, from your plugins
     and your own edits. -->
<script lang="ts">
	import type { ActionList } from '#lib/api/types/records.ts';
	import Inja from '#lib/ui/Inja.svelte';
	import RecordsPage from './RecordsPage.svelte';
	import { words } from './api.ts';

	let { list }: { list: ActionList } = $props();
</script>

<RecordsPage
	title="Actions"
	noun={['action', 'actions']}
	lead="What NPCs can do, beyond speaking: the model reads each action’s description and picks one when it fits. Each comes from a plugin or your own edits."
	records={list.actions}
	columns={[{ label: 'Category', class: 'category' }, { label: 'Priority', class: 'priority' }]}
>
	{#snippet cells(a)}
		<td class="category muted">{a.category ? words(a.category) : ''}</td>
		<td class="priority muted">{a.priority ?? ''}</td>
	{/snippet}
	{#snippet pane(a)}
		<p class="text"><Inja text={a.description} /></p>
		<dl>
			{#if a.category}<dt class="caps">Category</dt><dd>{words(a.category)}</dd>{/if}
			{#if a.priority !== undefined}<dt class="caps">Priority</dt><dd>{a.priority}</dd>{/if}
			<dt class="caps">Who may</dt><dd>{a.rules ? `${a.rules} eligibility ${a.rules === 1 ? 'rule' : 'rules'}` : 'Anyone, at any time'}</dd>
			{#if a.function}<dt class="caps">Runs</dt><dd class="id">{a.script}.{a.function}</dd>{/if}
			{#if a.quest}<dt class="caps">Quest</dt><dd class="id">{a.quest}</dd>{/if}
			{#if a.event}<dt class="caps">Sends</dt><dd class="id">{a.event}</dd>{/if}
			{#if a.confirm}<dt class="caps">Confirm</dt><dd>Asks you first</dd>{/if}
		</dl>
		{#if a.params.length}
			<h3 class="caps">The model fills in</h3>
			<ul class="params">
				{#each a.params as p (p.name)}
					<li><span class="id">{p.name}</span><span class="muted kind">· {words(p.type)}{p.optional ? ', optional' : ''}</span>{#if p.description}<br /><span class="small">{p.description}</span>{/if}</li>
				{/each}
			</ul>
		{/if}
	{/snippet}
	{#snippet note()}
		SkyrimNet also has a few actions built in, such as continuing a conversation, which aren’t files and aren’t listed. Switching an action off is done in SkyrimNet’s dashboard for now.
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
