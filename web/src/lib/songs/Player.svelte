<!-- A song's player: play or pause, where it is, and a bar to seek with.
     The audio is fetched with the session's token (see mediaURL) when the
     player first plays. -->
<script lang="ts">
	import { ApiError, mediaURL } from '#lib/api/client.ts';
	import { m } from '#lib/copy/messages.ts';
	import { clock } from './api.ts';

	let { path, label }: { path: string; label: string } = $props();

	let audio = $state<HTMLAudioElement>();
	let src = $state('');
	let playing = $state(false);
	let loading = $state(false);
	let time = $state(0);
	let duration = $state(NaN);
	let error = $state('');

	async function toggle() {
		if (!audio) return;
		if (playing) {
			audio.pause();
			return;
		}
		if (!src) {
			loading = true;
			error = '';
			try {
				src = await mediaURL(path);
			} catch (e) {
				error = e instanceof ApiError ? e.message : String(e);
				return;
			} finally {
				loading = false;
			}
			await new Promise((r) => setTimeout(r)); // let the element take the source
		}
		audio.play().catch((e) => (error = String(e)));
	}

	function seek(e: Event) {
		if (audio && isFinite(duration)) audio.currentTime = Number((e.currentTarget as HTMLInputElement).value);
	}
</script>

<div class="player">
	<button type="button" class="play" onclick={toggle} aria-label={playing ? m.songs.pause({ label }) : m.songs.play({ label })} disabled={loading}>
		{#if playing}
			<svg viewBox="0 0 12 12" width="12" height="12" aria-hidden="true"><path d="M3 2h2v8H3zM7 2h2v8H7z" fill="currentColor" /></svg>
		{:else}
			<svg viewBox="0 0 12 12" width="12" height="12" aria-hidden="true"><path d="M3.5 1.8v8.4L10 6z" fill="currentColor" /></svg>
		{/if}
	</button>
	<input type="range" min="0" max={isFinite(duration) ? duration : 0} step="0.1" value={time} oninput={seek} disabled={!src || !isFinite(duration)} aria-label={m.songs.position({ label })} style:--done="{isFinite(duration) && duration ? (time / duration) * 100 : 0}%" />
	<span class="time muted">{clock(time)} / {src ? clock(duration) : '–:––'}</span>
	<audio bind:this={audio} {src} preload="metadata" bind:currentTime={time} bind:duration onplay={() => (playing = true)} onpause={() => (playing = false)} onended={() => (playing = false)}></audio>
</div>
{#if error}<p class="notice error">{error}</p>{/if}

<style>
	.player {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-2) var(--space-3);
		background: var(--glass-strong);
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
	}
	.play {
		display: grid;
		place-items: center;
		flex: none;
		width: 2rem;
		height: 2rem;
		border-radius: 50%;
		border: 1px solid var(--accent);
		background: var(--accent);
		color: var(--on-accent);
		cursor: pointer;
	}
	.play:disabled {
		opacity: 0.6;
		cursor: progress;
	}
	/* The bar: a thin rule, filled with the accent as far as it's played. */
	input[type='range'] {
		flex: 1;
		min-width: 0;
		height: 4px;
		padding: 0;
		border: 0;
		border-radius: 2px;
		appearance: none;
		background: linear-gradient(to right, var(--accent) var(--done), var(--rule-strong) var(--done));
		cursor: pointer;
	}
	input[type='range']:disabled {
		cursor: default;
	}
	input[type='range']::-webkit-slider-thumb {
		appearance: none;
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: var(--ink);
	}
	input[type='range']::-moz-range-thumb {
		width: 12px;
		height: 12px;
		border: 0;
		border-radius: 50%;
		background: var(--ink);
	}
	.time {
		flex: none;
		font-size: var(--step--1);
		font-variant-numeric: tabular-nums;
	}
	audio {
		display: none;
	}
	.notice {
		margin-top: var(--space-2);
	}
</style>
