<!-- Triggers: how SkyrimNet responds when an event happens, from your
     plugins and your own edits. -->
<script lang="ts">
	import type { TriggerList } from '#lib/api/types/records.ts';
	import { m } from '#lib/copy/messages.ts';
	import Inja from '#lib/ui/Inja.svelte';
	import RecordsPage from './RecordsPage.svelte';
	import { words } from './api.ts';
	import { page } from '$app/state';
	import { href } from '#lib/nav.ts';

	let { list }: { list: TriggerList } = $props();
	const msg = m.records.triggers;
	const seconds = (s: number) => (s >= 60 && s % 60 === 0 ? msg.minutes({ value: s / 60 }) : msg.seconds({ value: s }));
</script>

<RecordsPage
	title={msg.title}
	copy={msg}
	lead={msg.lead}
	records={list.triggers}
	columns={[{ label: msg.when, class: 'when' }, { label: msg.then, class: 'then' }, { label: msg.cooldown, class: 'cooldown' }]}
>
	{#snippet cells(t)}
		<td class="when muted">{words(t.eventType)}</td>
		<td class="then muted">{t.responseType ? words(t.responseType) : ''}</td>
		<td class="cooldown muted">{t.cooldown !== undefined ? seconds(t.cooldown) : ''}</td>
	{/snippet}
	{#snippet pane(t)}
		<p class="text">{t.description}</p>
		<dl>
			<dt class="caps">{msg.when}</dt><dd><span class="id">{t.eventType}</span></dd>
			{#if t.conditions.length}
				<dt class="caps">{msg.and}</dt>
				<dd>
					<ul class="conditions">
						{#each t.conditions as c, i (i)}<li><span class="id">{c.field}</span> {words(c.operator)} <span class="id">{c.value === '' ? '""' : c.value}</span></li>{/each}
					</ul>
				</dd>
			{/if}
			{#if t.probability !== undefined}<dt class="caps">{msg.chance}</dt><dd>{Math.round(t.probability * 100)}%</dd>{/if}
			{#if t.responseType}<dt class="caps">{msg.then}</dt><dd>{t.targetScope ? msg.forScope({ response: words(t.responseType), scope: words(t.targetScope) }) : words(t.responseType)}</dd>{/if}
			{#if t.responseType === 'activate_voice_effect' || t.responseType === 'deactivate_voice_effect'}
				<dt class="caps">{msg.effect}</dt>
				<dd>{#if t.effectId}<a href={href(page.params.save ?? '', 'effects', `?item=${encodeURIComponent(t.effectId)}`)}>{t.effectId}</a>{:else}{msg.whicheverTarget}{/if}</dd>
			{/if}
			{#if t.audience}<dt class="caps">{msg.whoHears}</dt><dd>{words(t.audience)}</dd>{/if}
			{#if t.cooldown !== undefined}<dt class="caps">{msg.cooldown}</dt><dd>{seconds(t.cooldown)}</dd>{/if}
			{#if t.priority !== undefined}<dt class="caps">{msg.priority}</dt><dd>{t.priority}</dd>{/if}
			{#if t.interrupt}<dt class="caps">{msg.interrupts}</dt><dd>{msg.whoeverSpeaking}</dd>{/if}
		</dl>
		{#if t.content}
			<h3 class="caps">{msg.whatItSays}</h3>
			<p class="text content"><Inja text={t.content} /></p>
		{/if}
	{/snippet}
	{#snippet note()}
		{msg.note}
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
