<!-- Where SkyrimNet is: asked when Hermaeus finds no folder, or opened from
     the Playthroughs page to change the one in use. -->
<script lang="ts">
	import { untrack } from 'svelte';
	import { invalidateAll } from '$app/navigation';
	import { ApiError } from '#lib/api/client.ts';
	import { setFolder } from '#lib/api/status.ts';
	import Fill from '#lib/copy/Fill.svelte';
	import { m } from '#lib/copy/messages.ts';

	let {
		current,
		close
	}: {
		current?: { path: string; modPath: string }; // the folders in use, when changing them
		close?: () => Promise<void>; // ends a change, saved or not
	} = $props();

	const t = m.folders;

	let path = $state(untrack(() => current?.path ?? ''));
	let modPath = $state(untrack(() => current?.modPath ?? ''));
	let error = $state('');
	let saving = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			// When changing, the mod folder field starts out filled in, so
			// emptying it means find it again. Asked afresh, it starts empty,
			// so an empty one leaves a remembered mod folder alone.
			await setFolder(path, current ? modPath.trim() : modPath.trim() || undefined);
			await close?.(); // settles before the reload, which it would cancel
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
		{#if current}
			<Fill text={t.change.text}>{#snippet flag()}<code>-dir</code>{/snippet}</Fill>
		{:else}
			<Fill text={t.notFound.text}>{#snippet data()}<code>data</code>{/snippet}{#snippet config()}<code>config</code>{/snippet}{#snippet logs()}<code>logs</code>{/snippet}{#snippet overwrite()}<code>overwrite\SKSE\Plugins\SkyrimNet</code>{/snippet}</Fill>
		{/if}
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
	<div class="buttons">
		<button class="button primary" disabled={saving}>{t.use}</button>
		{#if close}<button type="button" class="button" onclick={close} disabled={saving}>{m.common.cancel}</button>{/if}
	</div>
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
	.buttons {
		display: flex;
		gap: var(--space-3);
	}
</style>
