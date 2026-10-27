<script lang="ts">
	import { page } from '$app/state';
	import { characterName } from '#lib/characters/api.ts';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let c = $derived(data.character);
	let base = $derived(`/${encodeURIComponent(data.playthrough.id)}/npc/${c.uuid}`);
	let tabs = $derived([
		{ href: base, label: 'Overview' },
		{ href: `${base}/memories`, label: `Memories (${c.memories})` },
		{ href: `${base}/diary`, label: `Diary (${c.diaries})` }
	]);
</script>

<header>
	<h1>{characterName(c)}</h1>
	{#if c.formerNames.length}
		<p>Previously known as {c.formerNames.map((n) => n || 'an unnamed identity').join(', ')}</p>
	{/if}
	<nav aria-label="{characterName(c)}’s records">
		{#each tabs as tab (tab.href)}
			<a href={tab.href} aria-current={page.url.pathname === tab.href ? 'page' : undefined}>{tab.label}</a>
		{/each}
	</nav>
</header>

{@render children()}

<style>
	nav {
		display: flex;
		gap: 1rem;
	}
	[aria-current='page'] {
		font-weight: bold;
	}
</style>
