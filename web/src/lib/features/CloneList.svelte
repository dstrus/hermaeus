<!-- The voices a cloud voice engine has cloned into the player's account,
     most recently used first. SkyrimNet keeps the list; this only reads it. -->
<script lang="ts">
	import type { CloneCache } from '#lib/api/types/settings.ts';
	import { m } from '#lib/copy/messages.ts';
	import { realTime } from '#lib/format.ts';

	const t = m.features.clones;

	let { cache, label }: { cache: CloneCache; label: string } = $props();

	let over = $derived(cache.maxVoices > 0 && cache.voices.length + cache.unnamed > cache.maxVoices);
</script>

<section class="clones">
	{#if !cache.voices.length && !cache.unnamed}
		<p class="muted small">{t.none({ account: label })}</p>
	{:else}
		<p class="muted small">
			{cache.maxVoices > 0 ? t.countKept({ count: cache.voices.length + cache.unnamed, account: label, max: cache.maxVoices.toLocaleString() }) : t.count({ count: cache.voices.length + cache.unnamed, account: label })}
		</p>
		{#if over}<p class="notice warn small">{t.over}</p>{/if}
		<table>
			<thead>
				<tr>
					<th class="caps">{t.voiceType}</th>
					<th class="caps lang">{t.language}</th>
					<th class="caps when">{t.lastUsed}</th>
				</tr>
			</thead>
			<tbody>
				{#each cache.voices as v, i (i)}
					<tr>
						<td class="id">{v.voiceType}</td>
						<td class="lang muted">{v.language ?? ''}</td>
						<td class="when muted">{v.lastUsed ? realTime(v.lastUsed) : '—'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		{#if cache.unnamed}<p class="muted small">{t.unnamed({ count: cache.unnamed })}</p>{/if}
	{/if}
</section>

<style>
	.clones {
		margin-top: var(--space-4);
		max-width: 64rem;
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
		max-width: 65ch;
		margin-bottom: var(--space-3);
	}
	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
		margin-bottom: var(--space-3);
	}
	th {
		padding: var(--space-2) var(--space-3);
		text-align: left;
		font-size: var(--step--1);
		color: var(--muted);
		border-bottom: 1px solid var(--rule-strong);
	}
	td {
		padding: var(--space-1) var(--space-3);
		border-bottom: 1px solid var(--rule);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.lang {
		width: 7rem;
	}
	.when {
		width: 13rem;
	}
	.id {
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
</style>
