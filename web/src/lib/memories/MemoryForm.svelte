<script lang="ts">
	import { untrack } from 'svelte';
	import type { Memory, MemoryChanges } from '#lib/api/types/memories.ts';
	import { m } from '#lib/copy/messages.ts';
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
	<label class="field"><span class="caps">{m.memories.form.text}</span><textarea bind:value={f.content} required rows="7"></textarea></label>
	<label class="field"><span class="caps">{m.memories.form.title}</span><input bind:value={f.title} placeholder={m.memories.form.titlePlaceholder} /></label>
	<div class="pair">
		<label class="field">
			<span class="caps">{m.memories.type}</span>
			<select bind:value={f.type}>
				{#each types as t (t)}<option value={t}>{t}</option>{/each}
			</select>
		</label>
		<label class="field"><span class="caps">{m.memories.mood}</span><input bind:value={f.emotion} list="memory-emotions" /></label>
	</div>
	<datalist id="memory-emotions">
		{#each commonEmotions as e (e)}<option value={e}></option>{/each}
	</datalist>
	<div class="pair">
		<label class="field"><span class="caps">{m.memories.place}</span><input bind:value={f.location} /></label>
		<label class="field"><span class="caps">{m.memories.importance}</span><input type="number" bind:value={f.importance} min="0" max="1" step="0.05" required /></label>
	</div>
	<label class="field"><span class="caps">{m.memories.tags}</span><input bind:value={f.tags} placeholder={m.memories.form.tagsPlaceholder} /></label>
	<label class="field">
		<span class="caps">{m.memories.onlyWhile}</span>
		<input bind:value={f.condition} spellcheck="false" placeholder={m.memories.form.conditionPlaceholder} />
	</label>
	<label class="check"><input type="checkbox" bind:checked={f.active} /> {m.memories.form.active}</label>
	<label class="check"><input type="checkbox" bind:checked={f.alwaysInject} /> {m.memories.form.alwaysInject}</label>
	<div class="buttons">
		<button class="button primary" disabled={saving}>{memory ? m.common.save : m.memories.form.addMemory}</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>{m.common.cancel}</button>
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
	.check {
		display: flex;
		gap: var(--space-2);
		align-items: center;
	}
	/* Save stays in reach however long the form is. */
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0 var(--space-5);
		background: var(--panel);
		border-top: 1px solid var(--rule);
		margin-top: var(--space-2);
	}
</style>
