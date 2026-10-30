<!-- Triggers: how SkyrimNet responds when an event happens, from your
     plugins and your own edits. -->
<script lang="ts">
	import type { TriggerList } from '#lib/api/types/records.ts';
	import Inja from '#lib/ui/Inja.svelte';
	import RecordsPage from './RecordsPage.svelte';
	import { words } from './api.ts';
	import { page } from '$app/state';
	import { href } from '#lib/nav.ts';

	let { list }: { list: TriggerList } = $props();
	const seconds = (s: number) => (s >= 60 && s % 60 === 0 ? `${s / 60} min` : `${s} s`);
</script>

<RecordsPage
	title="Triggers"
	noun={['trigger', 'triggers']}
	lead="How SkyrimNet responds when something happens in the game: a kind of event, the conditions it must meet, and what follows, such as narration or a diary entry."
	records={list.triggers}
	columns={[{ label: 'When', class: 'when' }, { label: 'Then', class: 'then' }, { label: 'Cooldown', class: 'cooldown' }]}
>
	{#snippet cells(t)}
		<td class="when muted">{words(t.eventType)}</td>
		<td class="then muted">{t.responseType ? words(t.responseType) : ''}</td>
		<td class="cooldown muted">{t.cooldown !== undefined ? seconds(t.cooldown) : ''}</td>
	{/snippet}
	{#snippet pane(t)}
		<p class="text">{t.description}</p>
		<dl>
			<dt class="caps">When</dt><dd><span class="id">{t.eventType}</span></dd>
			{#if t.conditions.length}
				<dt class="caps">And</dt>
				<dd>
					<ul class="conditions">
						{#each t.conditions as c, i (i)}<li><span class="id">{c.field}</span> {words(c.operator)} <span class="id">{c.value === '' ? '""' : c.value}</span></li>{/each}
					</ul>
				</dd>
			{/if}
			{#if t.probability !== undefined}<dt class="caps">Chance</dt><dd>{Math.round(t.probability * 100)}%</dd>{/if}
			{#if t.responseType}<dt class="caps">Then</dt><dd>{words(t.responseType)}{t.targetScope ? `, for ${words(t.targetScope)}` : ''}</dd>{/if}
			{#if t.responseType === 'activate_voice_effect' || t.responseType === 'deactivate_voice_effect'}
				<dt class="caps">Effect</dt>
				<dd>{#if t.effectId}<a href={href(page.params.save ?? '', 'effects', `?item=${encodeURIComponent(t.effectId)}`)}>{t.effectId}</a>{:else}Whichever the target has{/if}</dd>
			{/if}
			{#if t.audience}<dt class="caps">Who hears</dt><dd>{words(t.audience)}</dd>{/if}
			{#if t.cooldown !== undefined}<dt class="caps">Cooldown</dt><dd>{seconds(t.cooldown)}</dd>{/if}
			{#if t.priority !== undefined}<dt class="caps">Priority</dt><dd>{t.priority}</dd>{/if}
			{#if t.interrupt}<dt class="caps">Interrupts</dt><dd>Whoever is speaking</dd>{/if}
		</dl>
		{#if t.content}
			<h3 class="caps">What it says</h3>
			<p class="text content"><Inja text={t.content} /></p>
		{/if}
	{/snippet}
	{#snippet note()}
		A trigger’s file can be edited on the Content page; making new ones is done in SkyrimNet’s dashboard for now.
	{/snippet}
</RecordsPage>

<style>
	:global(td.when),
	:global(th.when) {
		width: 12rem;
	}
	:global(td.then),
	:global(th.then) {
		width: 10rem;
	}
	:global(td.cooldown),
	:global(th.cooldown) {
		width: 6.5rem;
		text-align: right;
	}
	.text {
		white-space: pre-wrap;
		line-height: 1.5;
		margin: var(--space-3) 0;
	}
	.content {
		font-size: var(--step--1);
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
	.conditions {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	h3 {
		font-size: var(--step--1);
		color: var(--accent);
		margin: var(--space-4) 0 var(--space-2);
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
</style>
