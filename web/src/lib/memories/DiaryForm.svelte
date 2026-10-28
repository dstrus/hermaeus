<script lang="ts">
	import { untrack } from 'svelte';
	import type { DiaryChanges, DiaryEntry } from '#lib/api/types/memories.ts';
	import { commonEmotions, parseTags } from './api.ts';

	let {
		entry,
		onsave,
		oncancel
	}: {
		entry?: DiaryEntry; // the entry to edit, or none for a new one
		onsave: (changes: DiaryChanges) => Promise<void>;
		oncancel: () => void;
	} = $props();

	// The form edits a snapshot of the entry as it was when the form opened
	// (untrack says so); the entry itself changes only once it's saved.
	const start = untrack(() => ({
		content: entry?.content ?? '',
		location: entry?.location ?? '',
		emotion: entry?.emotion ?? '',
		importance: entry?.importance ?? 0.5,
		tags: (entry?.tags ?? []).join(', ')
	}));
	const isNew = untrack(() => !entry);
	let f = $state({ ...start });
	let saving = $state(false);

	// changes holds only what differs from the entry; a new entry sends everything.
	function changes(): DiaryChanges {
		const c: DiaryChanges = {};
		const differs = <K extends keyof typeof start>(k: K) => isNew || f[k] !== start[k];
		if (differs('content')) c.content = f.content;
		if (differs('location')) c.location = f.location;
		if (differs('emotion')) c.emotion = f.emotion;
		if (differs('importance')) c.importance = f.importance;
		if (differs('tags')) c.tags = parseTags(f.tags);
		return c;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			await onsave(changes());
		} finally {
			saving = false;
		}
	}
</script>

<form onsubmit={submit}>
	<label class="field"><span class="caps">Entry</span><textarea bind:value={f.content} required rows="14"></textarea></label>
	<div class="pair">
		<label class="field"><span class="caps">Mood</span><input bind:value={f.emotion} list="diary-emotions" /></label>
		<label class="field"><span class="caps">Importance</span><input type="number" bind:value={f.importance} min="0" max="1" step="0.05" required /></label>
	</div>
	<datalist id="diary-emotions">
		{#each commonEmotions as e (e)}<option value={e}></option>{/each}
	</datalist>
	<label class="field"><span class="caps">Place</span><input bind:value={f.location} /></label>
	<label class="field"><span class="caps">Tags</span><input bind:value={f.tags} placeholder="Separated by commas" /></label>
	<p class="muted small">Its memory copy is kept in step with the text, mood, place, and importance.</p>
	<div class="buttons">
		<button class="button primary" disabled={saving}>{entry ? 'Save' : 'Add entry'}</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>Cancel</button>
	</div>
</form>

<style>
	form {
		display: grid;
		gap: var(--space-3);
		margin-top: var(--space-4);
	}
	.field {
		display: grid;
		gap: var(--space-1);
	}
	.field span {
		font-size: var(--step--1);
		color: var(--muted);
	}
	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-3);
	}
	textarea {
		resize: vertical;
		line-height: 1.55;
	}
	.small {
		font-size: var(--step--1);
	}
	/* Save stays in reach however long the form is. */
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0 var(--space-5);
		background: var(--pane);
		border-top: 1px solid var(--rule);
	}
</style>
