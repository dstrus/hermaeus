<!-- The voices a cloud voice engine has cloned into the player's account,
     most recently used first. SkyrimNet keeps the list; this only reads it. -->
<script lang="ts">
	import type { CloneCache } from '#lib/api/types/settings.ts';
	import { plural, realTime } from '#lib/format.ts';

	let { cache, label }: { cache: CloneCache; label: string } = $props();

	let over = $derived(cache.maxVoices > 0 && cache.voices.length + cache.unnamed > cache.maxVoices);
</script>

<section class="clones">
	{#if !cache.voices.length && !cache.unnamed}
		<p class="muted small">None yet. SkyrimNet’s dashboard can clone voice types into your {label} account all at once with its bulk clone.</p>
	{:else}
		<p class="muted small">
			{plural(cache.voices.length + cache.unnamed, 'voice', 'voices')} in your {label} account{cache.maxVoices > 0 ? `, of up to ${cache.maxVoices.toLocaleString()} kept` : ''}{cache.maxVoices > 0 ? '. Before cloning past the limit, SkyrimNet deletes the voice used longest ago' : ''}.
		</p>
		{#if over}<p class="notice warn small">More than the limit. Each time SkyrimNet clones another voice, it first deletes the one used longest ago, but never one cloned outside SkyrimNet.</p>{/if}
		<table>
			<thead>
				<tr>
					<th class="caps">Voice type</th>
					<th class="caps lang">Language</th>
					<th class="caps when">Last used</th>
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
		{#if cache.unnamed}<p class="muted small">{plural(cache.unnamed, 'other voice', 'other voices')} with a name Hermaeus can’t read.</p>{/if}
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
