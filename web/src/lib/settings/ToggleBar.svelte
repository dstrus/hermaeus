<!-- SkyrimNet's quick toggles as one row: each switches a feature on or
     off in its settings file. -->
<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { live } from '#lib/api/live.svelte.ts';
	import { setToggle, settingsDependency } from '#lib/settings/api.ts';
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
			await invalidate(settingsDependency); // a toggle is a setting; a settings page may be open
		} catch (e) {
			current = before;
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			busy = '';
		}
	}
</script>

<div class="toggles" role="group" aria-label="Quick toggles">
	<span class="group caps muted" aria-hidden="true">Quick toggles</span>
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
			<!-- A check box, as the game's own option menus draw them. -->
			<span class="mark" aria-hidden="true">
				<svg viewBox="0 0 10 10" width="10" height="10">
					<path d="M2 5.3 4.1 7.4 8 3.1" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
				</svg>
			</span>
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
	.group {
		font-size: var(--step--1);
		padding-right: var(--space-2);
	}
	/* The master switch stands apart: everything after it depends on it. */
	button.master {
		padding-right: var(--space-4);
		border-right: 1px solid var(--rule-strong);
	}
	.mark {
		display: inline-grid;
		place-items: center;
		width: 0.9rem;
		height: 0.9rem;
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		color: var(--on-accent);
	}
	.mark svg {
		opacity: 0;
	}
	button:hover:not(:disabled) .mark {
		border-color: var(--ink);
	}
	button[aria-checked='true'] .mark {
		background: var(--accent);
		border-color: var(--accent);
	}
	button[aria-checked='true'] .mark svg {
		opacity: 1;
	}
	.notice {
		font-size: var(--step--1);
	}
</style>
