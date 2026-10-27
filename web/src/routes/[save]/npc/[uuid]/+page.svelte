<script lang="ts">
	import { groupByDay, timeLabel } from '#lib/memories/dates.ts';
	import { realTime } from '#lib/format.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let o = $derived(data.overview);
	let milestones = $derived(groupByDay(o.milestones));
</script>

<dl>
	{#if o.formId}
		<dt>Form ID</dt>
		<dd><code>{o.formId}</code></dd>
	{/if}
	{#if o.modFile}
		<dt>Plugin</dt>
		<dd>{o.modFile}{#if o.localFormId}{' '}(local ID <code>{o.localFormId}</code>){/if}</dd>
	{/if}
	{#if o.bio}
		<dt>Bio template</dt>
		<dd><code>{o.bio}</code></dd>
	{/if}
	{#if o.firstEvent && o.lastEvent}
		<dt>Active</dt>
		<dd>{o.firstEvent.day} to {o.lastEvent.day}</dd>
	{/if}
	<dt>SkyrimNet ID</dt>
	<dd><code>{o.uuid}</code></dd>
	{#if o.firstSeen}
		<dt>First seen</dt>
		<dd>{realTime(o.firstSeen)}</dd>
	{/if}
	{#if o.updated}
		<dt>Record updated</dt>
		<dd>{realTime(o.updated)}</dd>
	{/if}
	{#each o.retired as r (r.uuid)}
		<dt>Earlier identity</dt>
		<dd>{r.name || 'unnamed'}{#if r.since} (re-identified {realTime(r.since)}){/if}</dd>
	{/each}
</dl>

{#if milestones.length}
	<h2>Milestones</h2>
	{#each milestones as group (group.items[0]?.eventId)}
		<h3>{group.day}</h3>
		<ul>
			{#each group.items as m (m.eventId)}
				<li>{timeLabel(m)} {m.summary}</li>
			{/each}
		</ul>
	{/each}
{/if}
