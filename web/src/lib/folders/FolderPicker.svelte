<!-- Browses for a folder, one level at a time, starting where Hermaeus runs
     from. Folders that are what it's after, or lead to it, are marked, and
     one that leads to it opens straight onto it. -->
<script lang="ts">
	import { tick } from 'svelte';
	import { ApiError } from '#lib/api/client.ts';
	import { browseFolders } from '#lib/api/status.ts';
	import type { Folder, Listing } from '#lib/api/types/install.ts';
	import { m } from '#lib/copy/messages.ts';

	let {
		kind,
		onpick,
		oncancel
	}: {
		kind: 'skyrimnet' | 'mod';
		onpick: (path: string) => void;
		oncancel: () => void;
	} = $props();

	const t = m.folders.picker;

	let listing = $state<Listing>();
	let error = $state('');
	let loading = $state(false);
	let list = $state<HTMLElement>();

	async function go(path = '') {
		error = '';
		loading = true;
		try {
			listing = await browseFolders(kind, path);
			await tick();
			list?.querySelector('button')?.focus();
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			loading = false;
		}
	}
	go();

	const markLabel = (f: Folder) =>
		f.mark === 'here' ? (kind === 'mod' ? t.mod : t.skyrimNet) : f.mark === 'inside' ? t.skyrimNetInside : '';

	// Esc closes the picker, not the form around it.
	function keydown(e: KeyboardEvent) {
		if (e.key !== 'Escape') return;
		e.stopPropagation();
		oncancel();
	}
</script>

<!-- The handler only hears Esc from the controls inside, which keep their own keys. -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="picker" role="group" aria-label={t.label} onkeydown={keydown}>
	{#if listing}
		<div class="where">
			{#if listing.roots.length > 1}
				<select aria-label={t.drive} value={listing.roots.find((r) => listing!.path.startsWith(r)) ?? ''} onchange={(e) => go(e.currentTarget.value)}>
					{#each listing.roots as r (r)}<option value={r}>{r}</option>{/each}
				</select>
			{/if}
			<span class="path">{listing.path}</span>
		</div>
		{#if listing.mark === 'here'}<p class="found">{kind === 'mod' ? t.isMod : t.isSkyrimNet}</p>{/if}
		<ul bind:this={list} class:loading>
			{#if listing.parent}
				<li><button type="button" class="row up caps" onclick={() => go(listing!.parent)}>{t.up}</button></li>
			{/if}
			{#each listing.folders as f (f.path)}
				<li>
					<button type="button" class="row" onclick={() => go(f.inside ?? f.path)}>
						<span class="name">{f.name}</span>
						{#if f.mark}<span class="mark caps" class:here={f.mark === 'here'}>{markLabel(f)}</span>{/if}
					</button>
				</li>
			{:else}
				<li class="muted empty">{t.empty}</li>
			{/each}
		</ul>
	{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<div class="buttons">
		<button type="button" class={['button', listing?.mark === 'here' && 'primary']} disabled={!listing} onclick={() => listing && onpick(listing.path)}>{t.choose}</button>
		<button type="button" class="button" onclick={oncancel}>{m.common.cancel}</button>
	</div>
</div>

<style>
	.picker {
		display: grid;
		gap: var(--space-3);
		padding: var(--space-3) 0;
		border-top: 1px solid var(--rule);
		border-bottom: 1px solid var(--rule);
	}
	.where {
		display: flex;
		align-items: baseline;
		gap: var(--space-3);
	}
	.path {
		color: var(--ink);
		word-break: break-all;
	}
	.found {
		font-size: var(--step--1);
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		max-height: 18rem;
		overflow-y: auto;
		border-top: 1px solid var(--rule-strong);
	}
	ul.loading {
		opacity: 0.6;
	}
	.row {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-4);
		width: 100%;
		padding: var(--space-2) var(--space-3);
		border: 0;
		border-bottom: 1px solid var(--rule);
		background: none;
		color: var(--ink);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.row:hover {
		background: var(--glass);
	}
	.row:focus-visible {
		background: var(--select);
	}
	.up {
		color: var(--muted);
		font-size: var(--step--1);
	}
	.name {
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.mark {
		flex: none;
		font-size: var(--step--1);
		color: var(--muted);
	}
	.mark.here {
		color: var(--ink);
	}
	.empty {
		padding: var(--space-2) var(--space-3);
	}
	.buttons {
		display: flex;
		gap: var(--space-3);
	}
</style>
