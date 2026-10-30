<!-- A record form's blocks: groups of fields, and lists whose items repeat
     blocks of their own (a trigger's conditions, an action's parameters).
     Every field reads and writes the record object it's given, in place. -->
<script lang="ts">
	import FormBlocks from './FormBlocks.svelte';
	import { getAt, setAt, shown, stored, type Block, type FieldSpec, type Obj } from './forms.ts';

	let { blocks, draft, depth = 0 }: { blocks: Block[]; draft: Obj; depth?: number } = $props();

	const listId = (key: string) => `form-${key.replace(/[^a-zA-Z0-9]/g, '-')}`;
	const fieldsOf = (b: Block) => ('fields' in b ? b.fields.filter((f): f is FieldSpec => !!f) : []);

	function edit(f: FieldSpec, text: string) {
		setAt(draft, f.key, stored(f, text, getAt(draft, f.key)));
	}
	function listOf(path: string): unknown[] {
		const v = getAt(draft, path);
		return Array.isArray(v) ? v : [];
	}
	function add(path: string, item: unknown) {
		const v = getAt(draft, path);
		if (Array.isArray(v)) v.push(item);
		else setAt(draft, path, [item]);
	}
	function move(path: string, i: number, by: number) {
		const v = listOf(path);
		const j = i + by;
		if (j < 0 || j >= v.length) return;
		[v[i], v[j]] = [v[j], v[i]];
	}
	function remove(path: string, i: number) {
		listOf(path).splice(i, 1);
	}
</script>

{#each blocks as b, bi (bi)}
	{#if 'rows' in b}
		{@const items = listOf(b.rows)}
		<section class="rows" class:nested={depth > 0}>
			<h3 class="caps">{b.title}</h3>
			{#if !items.length}<p class="muted small">{b.empty ?? 'None.'}</p>{/if}
			<ol>
				{#each items as _, i (i)}
					<li>
						<div class="item">
							<FormBlocks blocks={b.item(`${b.rows}.${i}`)} {draft} depth={depth + 1} />
						</div>
						<span class="tools">
							<button type="button" class="textbutton" onclick={() => move(b.rows, i, -1)} disabled={i === 0} aria-label="Move {b.add} {i + 1} up">Up</button>
							<button type="button" class="textbutton" onclick={() => move(b.rows, i, 1)} disabled={i === items.length - 1} aria-label="Move {b.add} {i + 1} down">Down</button>
							<button type="button" class="textbutton" onclick={() => remove(b.rows, i)} aria-label="Remove {b.add} {i + 1}">Remove</button>
						</span>
					</li>
				{/each}
			</ol>
			<button type="button" class="textbutton add" onclick={() => add(b.rows, b.blank())}>Add {b.add}</button>
		</section>
	{:else}
		{@const fields = fieldsOf(b)}
		<section class="fields" class:titled={!!b.title}>
			{#if b.title}<h3 class="caps">{b.title}</h3>{/if}
			{#if b.note}<p class="notice warn small">{b.note}</p>{/if}
			{#if fields.length}
				<div class="grid">
					{#each fields as f (f.key)}
						{@const v = getAt(draft, f.key)}
						{#if f.type === 'check'}
							<label class="check">
								<input type="checkbox" checked={v === true} onchange={(e) => setAt(draft, f.key, e.currentTarget.checked)} />
								<span>{f.label}</span>
							</label>
						{:else}
							<label class:wide={f.wide || f.type === 'area' || f.type === 'list'}>
								<span class="caps">{f.label}</span>
								{#if f.type === 'select'}
									<select value={v === undefined || v === null ? '' : String(v)} onchange={(e) => edit(f, e.currentTarget.value)}>
										{#if !f.choices?.some((c) => c.value === String(v ?? ''))}<option value={String(v ?? '')}>{v === undefined || v === null || v === '' ? '—' : String(v)}</option>{/if}
										{#each f.choices ?? [] as c (c.value)}<option value={c.value}>{c.label}</option>{/each}
									</select>
								{:else if f.type === 'area' || f.type === 'list'}
									<textarea rows={f.rows ?? (f.type === 'list' ? 4 : 3)} value={shown(f, v)} oninput={(e) => edit(f, e.currentTarget.value)} spellcheck={f.type === 'area'} placeholder={f.type === 'list' ? 'One per line' : undefined}></textarea>
								{:else}
									<input
										type="text"
										inputmode={f.type === 'num' ? 'decimal' : undefined}
										value={shown(f, v)}
										oninput={(e) => edit(f, e.currentTarget.value)}
										disabled={f.readonly}
										list={f.choices?.length ? listId(f.key) : undefined}
										autocomplete="off"
										spellcheck="false"
									/>
									{#if f.choices?.length}
										<datalist id={listId(f.key)}>
											{#each f.choices as c (c.value)}<option value={c.value}>{c.label === c.value ? '' : c.label}</option>{/each}
										</datalist>
									{/if}
								{/if}
								{#if f.readonly}<span class="muted small">It names the file, so it can’t change here.</span>{:else if f.hint}<span class="muted small">{f.hint}</span>{/if}
							</label>
						{/if}
					{/each}
				</div>
			{/if}
		</section>
	{/if}
{/each}

<style>
	section {
		margin-bottom: var(--space-4);
	}
	h3 {
		font-size: var(--step--1);
		color: var(--accent);
		margin-bottom: var(--space-2);
	}
	.titled,
	.rows:not(.nested) {
		padding-top: var(--space-3);
		border-top: 1px solid var(--rule);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(13rem, 1fr));
		gap: var(--space-3) var(--space-4);
		align-items: start;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		min-width: 0;
	}
	label.wide {
		grid-column: 1 / -1;
	}
	label > .caps {
		font-size: var(--step--1);
		color: var(--muted);
	}
	input[type='text'],
	select,
	textarea {
		width: 100%;
		min-width: 0;
	}
	textarea {
		resize: vertical;
		line-height: 1.5;
	}
	input:disabled {
		color: var(--muted);
	}
	.check {
		flex-direction: row;
		align-items: center;
		gap: var(--space-2);
		align-self: end;
		padding-bottom: var(--space-2);
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	/* A list's items: each a numbered block of its own, with its tools. */
	ol {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	li {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: var(--space-3);
		padding: var(--space-3) 0 var(--space-1) var(--space-3);
		margin-bottom: var(--space-2);
		border-left: 2px solid var(--rule-strong);
	}
	.nested li {
		border-left-color: var(--rule);
	}
	.item :global(section:last-child) {
		margin-bottom: 0;
	}
	.tools {
		display: flex;
		gap: var(--space-3);
		font-size: var(--step--1);
	}
	.tools .textbutton:disabled {
		color: var(--muted);
		text-decoration: none;
		cursor: default;
	}
	.add {
		font-size: var(--step--1);
	}
</style>
