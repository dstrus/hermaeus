<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { setFolder } from '#lib/api/status.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';

	const t = m.folders;

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
	<h1 class="caps">{t.title}</h1>
	<p>
		<Fill text={t.notFound.text}>{#snippet data()}<code>data</code>{/snippet}{#snippet config()}<code>config</code>{/snippet}{#snippet logs()}<code>logs</code>{/snippet}{#snippet overwrite()}<code>overwrite\SKSE\Plugins\SkyrimNet</code>{/snippet}</Fill>
	</p>
	<label>
		<span class="caps">{t.folder}</span>
		<input bind:value={path} required autocomplete="off" spellcheck="false" />
	</label>
	<label>
		<span class="caps">{t.modFolder}</span>
		<input bind:value={modPath} autocomplete="off" spellcheck="false" placeholder={t.modPlaceholder} />
	</label>
	<p class="muted small">
		<Fill text={t.modHint.text}>{#snippet overwrite()}<code>overwrite</code>{/snippet}</Fill>
	</p>
	<div><button class="button primary" disabled={saving}>{t.use}</button></div>
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
