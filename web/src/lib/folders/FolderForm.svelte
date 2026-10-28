<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { setFolder } from '#lib/api/status.ts';

	let path = $state('');
	let modPath = $state('');
	let error = $state('');
	let saving = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			await setFolder(path, modPath.trim() || undefined);
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
	<label>
		<span class="caps">SkyrimNet’s mod folder</span>
		<input bind:value={modPath} autocomplete="off" spellcheck="false" placeholder="Only with Mod Organizer 2, and usually found on its own" />
	</label>
	<p class="muted small">
		Under Mod Organizer 2, SkyrimNet’s own files (its default bios, for one) stay in its mod folder while what it writes lands in
		<code>overwrite</code>. Hermaeus looks for the mod folder next to <code>overwrite</code>; name it here if it isn’t found.
	</p>
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
	.small {
		font-size: var(--step--1);
	}
</style>
