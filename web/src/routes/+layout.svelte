<script lang="ts">
	import { fitToViewport } from '#lib/ui/fit.ts';
	import { groups } from '#lib/nav.ts';
	import { pluginSettingsHref } from '#lib/features/plugins.ts';
	import '#lib/ui/theme.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { live, watchGame } from '#lib/api/live.svelte.ts';
	import FolderForm from '#lib/folders/FolderForm.svelte';
	import Mark from '#lib/ui/Mark.svelte';
	import ToggleBar from '#lib/settings/ToggleBar.svelte';
	import HintBar from '#lib/ui/HintBar.svelte';
	import ThemeSwitch from '#lib/ui/ThemeSwitch.svelte';
	import { handleKey } from '#lib/ui/shortcuts.svelte.ts';
	import type { Playthrough } from '#lib/api/types/store.ts';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	$effect(() => watchGame(data.status.gameRunning));

	let save = $derived(page.params.save);
	let base = $derived(save ? `/${encodeURIComponent(save)}` : '');

	// A playthrough is named by its player, with its save ID only when
	// another playthrough has the same player.
	function playthroughLabel(p: Playthrough): string {
		const name = p.player || 'No player yet';
		const shared = data.status.playthroughs.filter((o) => (o.player || 'No player yet') === name).length > 1;
		return shared ? `${name} · ${p.id}` : name;
	}
	let charactersActive = $derived(!!save && (page.url.pathname === base || page.url.pathname.startsWith(`${base}/npc/`) || page.url.pathname.startsWith(`${base}/bios`)));
	let historyActive = $derived(!!save && page.url.pathname.startsWith(`${base}/history`));
	let screenshotsActive = $derived(!!save && page.url.pathname.startsWith(`${base}/screenshots`));
	let songsActive = $derived(!!save && page.url.pathname.startsWith(`${base}/songs`));
	// A group's page is current at its URL or below it.
	const here = (path: string) => page.url.pathname === `${base}/${path}` || page.url.pathname.startsWith(`${base}/${path}/`);
</script>

<svelte:window onkeydown={handleKey} />

<div class="app" class:withsections={!!save && !!data.status.skyrimNetDir}>
	<header class="top">
		<a class="name caps" href="/"><Mark size="1.35em" /> Hermaeus</a>
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
		{#if data.status.skyrimNetDir}
			<div class="quick"><ToggleBar toggles={data.toggles} error={data.togglesError} /></div>
		{/if}
	</header>

	{#if !data.status.skyrimNetDir}
		<main class="solo"><FolderForm /></main>
	{:else if save}
		<nav class="sections" aria-label="Sections" use:fitToViewport>
			<ul aria-label="This playthrough">
				<li><a class="caps" href={base} aria-current={charactersActive ? 'page' : undefined}>Characters</a></li>
				<li><a class="caps" href="{base}/history/events" aria-current={historyActive ? 'page' : undefined}>History</a></li>
				<li><a class="caps" href="{base}/screenshots" aria-current={screenshotsActive ? 'page' : undefined}>Screenshots</a></li>
				<li><a class="caps" href="{base}/songs" aria-current={songsActive ? 'page' : undefined}>Bard songs</a></li>
			</ul>
			{#each groups as g (g.path)}
				<p class="group caps">{g.label}</p>
				<ul aria-label={g.label}>
					{#each g.pages as p (p.path)}
						<li><a class="caps" href="{base}/{g.path}/{p.path}" aria-current={here(`${g.path}/${p.path}`) ? 'page' : undefined}>{p.label}</a></li>
					{/each}
				</ul>
			{/each}
			{#if data.pluginSettings.length}
				<p class="group caps">Plugin settings</p>
				<ul aria-label="Plugin settings">
					{#each data.pluginSettings as p (p.folder)}
						<li><a class="caps" href={pluginSettingsHref(save, p.folder)} aria-current={here(`plugin-settings/${encodeURIComponent(p.folder)}`) ? 'page' : undefined}>{p.title}</a></li>
					{/each}
				</ul>
			{/if}
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
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-6);
		padding: var(--space-3) var(--space-6);
		border-bottom: 1px solid var(--rule);
	}
	/* The quick toggles take a line of their own under the name and state. */
	.quick {
		flex-basis: 100%;
		margin-top: calc(-1 * var(--space-3));
		padding-top: var(--space-2);
		border-top: 1px solid var(--rule);
	}
	.name {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
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
	/* The sidebar stays put as the page scrolls, fitted above the hints,
	   and scrolls on its own when its groups run longer than the window. */
	.sections {
		position: sticky;
		top: 0;
		align-self: start;
		overflow-y: auto;
		overscroll-behavior: contain;
		padding: var(--space-5) var(--space-4) var(--space-5) var(--space-6);
		border-right: 1px solid var(--rule);
	}
	.sections ul {
		list-style: none;
		margin: 0 0 var(--space-5);
		padding: 0;
	}
	.sections a {
		display: block;
		padding: var(--space-1) 0;
		font-size: var(--step-1);
		text-decoration: none;
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
		.group {
			display: none;
		}
		.sections {
			position: static;
			overflow: visible;
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
