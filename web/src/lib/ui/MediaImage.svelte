<!-- An image from the API, fetched with the session's token (see mediaURL).
     A lazy one waits until it's near the window, so a long grid only loads
     what's seen. -->
<script lang="ts">
	import { mediaURL } from '#lib/api/client.ts';
	import { m } from '#lib/copy/messages.ts';

	let { path, alt, lazy = false }: { path: string; alt: string; lazy?: boolean } = $props();

	let box = $state<HTMLElement>();
	let near = $state(false);
	let src = $state('');
	let failed = $state(false);

	$effect(() => {
		if (!lazy || !box) {
			near = true;
			return;
		}
		const seen = new IntersectionObserver((entries) => {
			if (entries.some((e) => e.isIntersecting)) {
				near = true;
				seen.disconnect();
			}
		}, { rootMargin: '400px' });
		seen.observe(box);
		return () => seen.disconnect();
	});

	$effect(() => {
		if (!near) return;
		const want = path;
		let gone = false;
		failed = false;
		mediaURL(want).then(
			(url) => {
				if (!gone) src = url;
			},
			() => {
				if (!gone) failed = true;
			}
		);
		return () => {
			gone = true;
		};
	});
</script>

<span class="media" bind:this={box} class:failed>
	{#if src}<img {src} {alt} draggable="false" />{:else if failed}<span class="muted">{m.ui.noImage}</span>{/if}
</span>

<style>
	.media {
		display: grid;
		place-items: center;
		width: 100%;
		height: 100%;
		background: var(--glass);
	}
	img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.failed {
		font-size: var(--step--1);
	}
</style>
