<script lang="ts">
	import '#lib/ui/theme.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { live, watchGame } from '#lib/api/live.svelte.ts';
	import FolderForm from '#lib/folders/FolderForm.svelte';
	import HintBar from '#lib/ui/HintBar.svelte';
	import ThemeSwitch from '#lib/ui/ThemeSwitch.svelte';
	import { handleKey } from '#lib/ui/shortcuts.svelte.ts';
	import type { Playthrough } from '#lib/api/types/store.ts';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	$effect(() => watchGame(data.status.gameRunning));

	let save = $derived(page.params.save);
	let base = $derived(save ? `/${encodeURIComponent(save)}` : '');

	// The sections of a playthrough. Those not built yet are listed so the
	// shape of Hermaeus is clear, but can't be opened.
	const sections = [
		{ label: 'Characters', built: true },
		{ label: 'Events' },
		{ label: 'Knowledge' },
		{ label: 'Screenshots' },
		{ label: 'Bard songs' }
	];
	const configuration = [{ label: 'Prompts' }, { label: 'Settings' }, { label: 'Plugins' }, { label: 'Content' }, { label: 'Overrides' }];
	// A playthrough is named by its player, with its save ID only when
	// another playthrough has the same player.
	function playthroughLabel(p: Playthrough): string {
		const name = p.player || 'No player yet';
		const shared = data.status.playthroughs.filter((o) => (o.player || 'No player yet') === name).length > 1;
		return shared ? `${name} · ${p.id}` : name;
	}
	let charactersActive = $derived(!!save && (page.url.pathname === base || page.url.pathname.startsWith(`${base}/npc/`)));
</script>

<svelte:window onkeydown={handleKey} />

<div class="app" class:withsections={!!save && !!data.status.skyrimNetDir}>
	<header class="top">
		<a class="name caps" href="/">Hermaeus</a>
		<div class="state">
			{#if save && data.status.playthroughs.length}
				<label class="playthrough">
					<span class="caps muted">Playthrough</span>
					<select value={save} onchange={(e) => goto(`/${encodeURIComponent(e.currentTarget.value)}`)}>
						{#each data.status.playthroughs as p (p.id)}
							<option value={p.id} title={p.id}>{playthroughLabel(p)}</option>
						{/each}
					</select>
				</label>
			{/if}
			<span class="game caps" class:running={live.gameRunning} role="status">
				{live.gameRunning ? 'Skyrim running · read-only' : 'Skyrim closed'}
			</span>
			<ThemeSwitch theme={data.status.theme} />
		</div>
	</header>

	{#if !data.status.skyrimNetDir}
		<main class="solo"><FolderForm /></main>
	{:else if save}
		<nav class="sections" aria-label="Sections">
			<ul aria-label="This playthrough">
				<li><a class="caps" href={base} aria-current={charactersActive ? 'page' : undefined}>Characters</a></li>
			</ul>
			<p class="later caps">Not built yet</p>
			<ul class="soon" aria-label="This playthrough, not built yet">
				{#each sections.filter((s) => !s.built) as s (s.label)}<li class="caps">{s.label}</li>{/each}
			</ul>
			<p class="group caps">Configuration</p>
			<p class="later caps">Not built yet</p>
			<ul class="soon" aria-label="Configuration, not built yet">
				{#each configuration as s (s.label)}<li class="caps">{s.label}</li>{/each}
			</ul>
		</nav>
		<main>{@render children()}</main>
	{:else}
		<main class="solo">{@render children()}</main>
	{/if}

	<HintBar />
</div>

<style>
	.app {
		display: grid;
		grid-template-columns: 1fr;
		grid-template-rows: auto 1fr;
		min-height: 100vh;
	}
	.app.withsections {
		grid-template-columns: 13rem 1fr;
	}
	.top {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		gap: var(--space-6);
		padding: var(--space-3) var(--space-6);
		border-bottom: 1px solid var(--rule);
	}
	.name {
		font-size: var(--step-2);
		font-weight: 600;
		text-decoration: none;
	}
	.state {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: var(--space-5);
	}
	.playthrough {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.playthrough span {
		font-size: var(--step--1);
	}
	.playthrough select {
		padding: var(--space-1) var(--space-2);
	}
	.game {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.game.running {
		color: var(--warn);
	}
	.sections {
		padding: var(--space-5) var(--space-4) var(--space-5) var(--space-6);
		border-right: 1px solid var(--rule);
	}
	.sections ul {
		list-style: none;
		margin: 0 0 var(--space-5);
		padding: 0;
	}
	.sections a,
	.sections .soon li {
		display: block;
		padding: var(--space-1) 0;
		font-size: var(--step-1);
		text-decoration: none;
	}
	.sections a {
		color: var(--muted);
	}
	.sections a:hover {
		color: var(--ink);
	}
	.sections a[aria-current='page'] {
		color: var(--accent);
		font-weight: 600;
	}
	/* A group's label, never as bright as the current section. */
	.group {
		margin-top: var(--space-4);
		font-size: var(--step--1);
		color: var(--muted);
	}
	.later {
		font-size: var(--step--1);
		color: var(--muted);
		padding-bottom: var(--space-1);
		margin-bottom: var(--space-1);
		border-bottom: 1px solid var(--rule);
	}
	.soon {
		color: var(--muted);
	}
	main {
		padding: var(--space-5) var(--space-6) 5rem;
		min-width: 0;
	}
	main.solo {
		max-width: 64rem;
	}
	@media (max-width: 60rem) {
		.app.withsections {
			grid-template-columns: 1fr;
		}
		.top {
			flex-wrap: wrap;
			gap: var(--space-2) var(--space-5);
		}
		.state {
			flex-wrap: wrap;
			gap: var(--space-2) var(--space-4);
		}
		.later,
		.group,
		.soon {
			display: none;
		}
		.sections {
			display: flex;
			gap: var(--space-5);
			border-right: 0;
			border-bottom: 1px solid var(--rule);
			padding: var(--space-2) var(--space-6);
			overflow-x: auto;
		}
		.sections ul {
			display: flex;
			gap: var(--space-5);
			margin: 0;
		}
	}
</style>
