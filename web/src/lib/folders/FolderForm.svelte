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
	<h1>Where is SkyrimNet?</h1>
	<p>
		Hermaeus didn’t find SkyrimNet’s folder, the one containing <code>data</code>, <code>config</code>, and
		<code>logs</code>. With Mod Organizer 2, it’s usually <code>overwrite\SKSE\Plugins\SkyrimNet</code> in your MO2
		instance.
	</p>
	<label>
		SkyrimNet folder
		<input bind:value={path} required autocomplete="off" spellcheck="false" />
	</label>
	<button disabled={saving}>Use this folder</button>
	{#if error}
		<p role="alert">{error}</p>
	{/if}
</form>
