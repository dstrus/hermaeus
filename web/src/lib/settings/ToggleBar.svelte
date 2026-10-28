<!-- SkyrimNet's quick toggles as one row: each switches a feature on or
     off in its settings file. -->
<script lang="ts">
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import { setToggle } from '#lib/settings/api.ts';
	import type { Toggle } from '#lib/api/types/settings.ts';

	let { toggles, error: loadError = '' }: { toggles: Toggle[]; error?: string } = $props();

	// Follows what was loaded, and changes at once when switched.
	let current = $derived(toggles);
	let error = $derived(loadError);
	let busy = $state('');

	function title(t: Toggle): string {
		if (!t.available) return `${t.what} Not switchable yet: SkyrimNet creates its settings file the next time it starts.`;
		if (live.gameRunning) return `${t.what} Skyrim is running, so nothing can be changed.`;
		return t.what;
	}

	async function flip(t: Toggle) {
		const before = current;
		const on = !t.on;
		current = current.map((c) => (c.id === t.id ? { ...c, on } : c));
		busy = t.id;
		error = '';
		try {
			current = (await setToggle(t.id, on)).toggles;
		} catch (e) {
			current = before;
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			busy = '';
		}
	}
</script>

<div class="toggles" role="group" aria-label="Quick toggles">
	{#each current as t (t.id)}
		<button
			type="button"
			role="switch"
			aria-checked={t.on}
			class:master={t.master}
			disabled={!t.available || live.gameRunning || busy === t.id}
			title={title(t)}
			onclick={() => flip(t)}
		>
			<span class="mark" aria-hidden="true"></span>
			<span class="caps">{t.label}</span>
		</button>
	{/each}
	{#if error}
		<span class="notice error" role="alert">{error}</span>
	{/if}
</div>

<style>
	.toggles {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1) var(--space-4);
	}
	button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		background: none;
		border: 0;
		padding: var(--space-1) 0;
		color: var(--muted);
		font-size: var(--step--1);
		cursor: pointer;
	}
	button[aria-checked='true'] {
		color: var(--ink);
	}
	button:hover:not(:disabled) {
		color: var(--ink);
	}
	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	/* The master switch stands apart: everything after it depends on it. */
	button.master {
		padding-right: var(--space-4);
		border-right: 1px solid var(--rule-strong);
	}
	.mark {
		width: 0.55rem;
		height: 0.55rem;
		border: 1px solid var(--rule-strong);
		border-radius: 1px;
	}
	button[aria-checked='true'] .mark {
		background: var(--accent);
		border-color: var(--accent);
	}
	.notice {
		font-size: var(--step--1);
	}
</style>
