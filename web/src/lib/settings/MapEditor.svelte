<!-- A map of keys to values edited as rows: a key, its value, remove, and
     add. Each value keeps the kind it had (a switch, a number, text); a
     value that is itself a list or map is edited as JSON in its row. -->
<script lang="ts">
	import type { Entry } from './api.ts';

	let { entries = $bindable() }: { entries: Entry[] } = $props();

	function add() {
		entries = [...entries, { key: '', kind: 'text', text: '', on: false }];
	}
	function remove(i: number) {
		entries = entries.filter((_, j) => j !== i);
	}
</script>

<div class="rows">
	{#each entries as e, i (i)}
		<div class="row">
			<input type="text" class="key" bind:value={e.key} placeholder="key" aria-label="Key" spellcheck="false" />
			{#if e.kind === 'switch'}
				<input type="checkbox" bind:checked={e.on} aria-label="Value of {e.key}" />
			{:else if e.kind === 'number'}
				<input type="text" inputmode="decimal" bind:value={e.text} aria-label="Value of {e.key}" />
			{:else if e.kind === 'json'}
				<input type="text" class="json" bind:value={e.text} aria-label="Value of {e.key}, as JSON" spellcheck="false" />
			{:else}
				<input type="text" bind:value={e.text} aria-label="Value of {e.key}" />
			{/if}
			<button type="button" class="textbutton" onclick={() => remove(i)} aria-label="Remove {e.key}">Remove</button>
		</div>
	{/each}
</div>
<button type="button" class="button" onclick={add}>Add a key</button>

<style>
	.rows {
		display: grid;
		gap: var(--space-1);
		margin-bottom: var(--space-3);
	}
	.row {
		display: grid;
		grid-template-columns: minmax(5rem, 2fr) 3fr auto;
		align-items: center;
		gap: var(--space-2);
	}
	input[type='text'] {
		width: 100%;
		min-width: 0;
		padding: var(--space-1) var(--space-2);
		font-family: var(--font-mono);
		font-size: var(--step--1);
	}
	input[type='checkbox'] {
		justify-self: start;
		margin: 0;
	}
	.textbutton {
		font-size: var(--step--1);
	}
</style>
