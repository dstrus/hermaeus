<script lang="ts">
	import { untrack } from 'svelte';
	import type { Memory, MemoryChanges } from '#lib/api/types/memories.ts';
	import { commonEmotions, memoryTypes, parseTags } from './api.ts';

	let {
		memory,
		onsave,
		oncancel
	}: {
		memory?: Memory; // the memory to edit, or none for a new one
		onsave: (changes: MemoryChanges) => Promise<void>;
		oncancel: () => void;
	} = $props();

	// The form edits a snapshot of the memory as it was when the form opened
	// (untrack says so); the memory itself changes only once it's saved.
	const start = untrack(() => ({
		type: memory?.type ?? 'EXPERIENCE',
		title: memory?.title ?? '',
		content: memory?.content ?? '',
		location: memory?.location ?? '',
		emotion: memory?.emotion ?? '',
		importance: memory?.importance ?? 0.5,
		tags: (memory?.tags ?? []).join(', '),
		active: memory?.active ?? true,
		alwaysInject: memory?.alwaysInject ?? false,
		condition: memory?.condition ?? ''
	}));
	const isNew = untrack(() => !memory);
	let f = $state({ ...start });
	let saving = $state(false);
	let types = $derived(memoryTypes.includes(f.type) ? memoryTypes : [f.type, ...memoryTypes]);

	// changes holds only what differs from the memory, so a save touches only
	// those fields; a new memory sends everything.
	function changes(): MemoryChanges {
		const c: MemoryChanges = {};
		const differs = <K extends keyof typeof start>(k: K) => isNew || f[k] !== start[k];
		if (differs('type')) c.type = f.type;
		if (differs('title')) c.title = f.title;
		if (differs('content')) c.content = f.content;
		if (differs('location')) c.location = f.location;
		if (differs('emotion')) c.emotion = f.emotion;
		if (differs('importance')) c.importance = f.importance;
		if (differs('tags')) c.tags = parseTags(f.tags);
		if (differs('active')) c.active = f.active;
		if (differs('alwaysInject')) c.alwaysInject = f.alwaysInject;
		if (differs('condition')) c.condition = f.condition;
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
	<label>Text <textarea bind:value={f.content} required rows="5"></textarea></label>
	<label>Title <input bind:value={f.title} placeholder="Optional; few memories have one" /></label>
	<label>
		Type
		<select bind:value={f.type}>
			{#each types as t (t)}<option value={t}>{t}</option>{/each}
		</select>
	</label>
	<label>Emotion <input bind:value={f.emotion} list="memory-emotions" /></label>
	<datalist id="memory-emotions">
		{#each commonEmotions as e (e)}<option value={e}></option>{/each}
	</datalist>
	<label>Location <input bind:value={f.location} /></label>
	<label>Importance <input type="number" bind:value={f.importance} min="0" max="1" step="0.05" required /></label>
	<label>Tags <input bind:value={f.tags} placeholder="Separated by commas" /></label>
	<label><input type="checkbox" bind:checked={f.active} /> Active</label>
	<label><input type="checkbox" bind:checked={f.alwaysInject} /> Always included, not only when relevant</label>
	<label>
		Condition
		<input bind:value={f.condition} spellcheck="false" placeholder="Optional; only used while this is true" />
	</label>
	<button disabled={saving}>{memory ? 'Save' : 'Add memory'}</button>
	<button type="button" onclick={oncancel} disabled={saving}>Cancel</button>
</form>
