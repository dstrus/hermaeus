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
	<label>Entry <textarea bind:value={f.content} required rows="10"></textarea></label>
	<label>Emotion <input bind:value={f.emotion} list="diary-emotions" /></label>
	<datalist id="diary-emotions">
		{#each commonEmotions as e (e)}<option value={e}></option>{/each}
	</datalist>
	<label>Location <input bind:value={f.location} /></label>
	<label>Importance <input type="number" bind:value={f.importance} min="0" max="1" step="0.05" required /></label>
	<label>Tags <input bind:value={f.tags} placeholder="Separated by commas" /></label>
	<p><small>Its memory copy is kept in step with the text, emotion, location, and importance.</small></p>
	<button disabled={saving}>{entry ? 'Save' : 'Add entry'}</button>
	<button type="button" onclick={oncancel} disabled={saving}>Cancel</button>
</form>
