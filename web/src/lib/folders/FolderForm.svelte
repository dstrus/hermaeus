<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { setFolder } from '#lib/api/status.ts';

	let path = $state('');
	let error = $state('');
	let saving = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			await setFolder(path);
			await invalidateAll(); // reload the status, which now has the folder
		} catch (e) {
			error = e instanceof ApiError ? e.message : String(e);
		} finally {
			saving = false;
		}
	}
</script>

<form onsubmit={submit}>
	<h1 class="caps">Where is SkyrimNet?</h1>
	<p>
		Hermaeus didn’t find SkyrimNet’s folder, the one containing <code>data</code>, <code>config</code>, and
		<code>logs</code>. With Mod Organizer 2, it’s usually <code>overwrite\SKSE\Plugins\SkyrimNet</code> in your MO2
		instance.
	</p>
	<label>
		<span class="caps">SkyrimNet folder</span>
		<input bind:value={path} required autocomplete="off" spellcheck="false" />
	</label>
	<div><button class="button primary" disabled={saving}>Use this folder</button></div>
	{#if error}
		<p class="notice error" role="alert">{error}</p>
	{/if}
</form>

<style>
	form {
		display: grid;
		gap: var(--space-4);
		max-width: 44rem;
	}
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
	}
	p {
		max-width: 65ch;
	}
	label {
		display: grid;
		gap: var(--space-1);
	}
	label span {
		font-size: var(--step--1);
		color: var(--muted);
	}
</style>
