<!-- A list of records (a pronunciation dictionary, model parameters, the
     events list) edited as a stack of small forms: one block per item, a
     field per key, remove and add. Each field keeps the kind of value the
     list already uses for that key. -->
<script lang="ts">
	type Scalar = string | number | boolean | null;
	type Field = 'switch' | 'number' | 'text';

	let { items = $bindable() }: { items: Record<string, Scalar>[] } = $props();

	// Every key the items use, in order of first appearance, with the kind of
	// value each holds (from the first item that sets it).
	let fields = $derived.by(() => {
		const out: { key: string; kind: Field }[] = [];
		const seen = new Map<string, Field | undefined>();
		for (const item of items) {
			for (const [key, v] of Object.entries(item)) {
				if (!seen.has(key)) seen.set(key, undefined);
				if (seen.get(key) === undefined && v !== null) seen.set(key, typeof v === 'boolean' ? 'switch' : typeof v === 'number' ? 'number' : 'text');
			}
		}
		for (const [key, kind] of seen) out.push({ key, kind: kind ?? 'text' });
		return out;
	});

	function add() {
		const blank: Record<string, Scalar> = {};
		for (const f of fields) blank[f.key] = f.kind === 'switch' ? false : f.kind === 'number' ? 0 : '';
		items = [...items, blank];
	}
	function remove(i: number) {
		items = items.filter((_, j) => j !== i);
	}
	// A number field holds its text while typed; the list keeps numbers.
	function setNumber(item: Record<string, Scalar>, key: string, text: string) {
		const n = Number(text.trim());
		item[key] = text.trim() === '' || Number.isNaN(n) ? text : n;
	}
</script>

<ol class="items">
	{#each items as item, i (i)}
		<li>
			<div class="head">
				<span class="caps muted">{i + 1}</span>
				<button type="button" class="textbutton" onclick={() => remove(i)}>Remove</button>
			</div>
			{#each fields as f (f.key)}
				<label>
					<span class="key">{f.key}</span>
					{#if f.kind === 'switch'}
						<input type="checkbox" checked={item[f.key] === true} onchange={(e) => (item[f.key] = e.currentTarget.checked)} />
					{:else if f.kind === 'number'}
						<input type="text" inputmode="decimal" value={item[f.key] ?? ''} oninput={(e) => setNumber(item, f.key, e.currentTarget.value)} />
					{:else}
						<input type="text" value={item[f.key] ?? ''} oninput={(e) => (item[f.key] = e.currentTarget.value)} />
					{/if}
				</label>
			{/each}
		</li>
	{/each}
</ol>
<button type="button" class="button" onclick={add}>Add {fields.length ? 'another' : 'an item'}</button>

<style>
	.items {
		list-style: none;
		margin: 0 0 var(--space-3);
		padding: 0;
	}
	li {
		padding: var(--space-2) 0 var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		font-size: var(--step--1);
		margin-bottom: var(--space-1);
	}
	.head .textbutton {
		font-size: var(--step--1);
	}
	label {
		display: grid;
		grid-template-columns: minmax(6rem, 2fr) 3fr;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-1) 0;
	}
	.key {
		font-family: var(--font-mono);
		font-size: var(--step--1);
		color: var(--muted);
		overflow-wrap: anywhere;
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
</style>
