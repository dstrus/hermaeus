<!-- Two versions of a bio read as one document, with what the later one
     added or dropped marked in place. Unchanged sections stay, quieter,
     so the changes are read in their context. -->
<script lang="ts">
	import type { Section } from '#lib/api/types/bios.ts';
	import { sectionLabel } from './api.ts';
	import { diffSections } from './diff.ts';

	let { from, to, fromLabel, toLabel }: { from: Section[]; to: Section[]; fromLabel: string; toLabel: string } = $props();

	let sections = $derived(diffSections(from, to));
	let changed = $derived(sections.filter((s) => s.status !== 'same').length);
	const tag = (status: string) =>
		status === 'added' ? `only in ${toLabel.toLowerCase()}` : status === 'removed' ? `only in ${fromLabel.toLowerCase()}` : status === 'rewritten' ? 'rewritten' : status === 'changed' ? 'changed' : 'unchanged';
</script>

<p class="muted key">
	{#if changed}
		{changed} of {sections.length} sections differ. <del>Dropped</del> text is struck through; <ins>added</ins> text is marked.
	{:else}
		The two are the same.
	{/if}
</p>
<article class="document">
	{#each sections as s (s.name)}
		<section class:same={s.status === 'same'}>
			<h2 class="caps">
				{sectionLabel(s.name)}
				<span class="tag" class:changed={s.status !== 'same'}>{tag(s.status)}</span>
			</h2>
			<div class="text">
				{#if s.status === 'rewritten'}
					<p class="caps muted which">{fromLabel}</p>
					<p><del>{s.ops[0]?.text}</del></p>
					<p class="caps muted which">{toLabel}</p>
					<p><ins>{s.ops[1]?.text}</ins></p>
				{:else}
					<p>
						{#each s.ops as op, i (i)}
							{#if op.kind === 'added'}<ins>{op.text}</ins>{:else if op.kind === 'removed'}<del>{op.text}</del>{:else}{op.text}{/if}
						{/each}
					</p>
				{/if}
			</div>
		</section>
	{/each}
</article>

<style>
	.key {
		margin-bottom: var(--space-4);
		font-size: var(--step--1);
	}
	/* Each section: its name and status in a margin column, its text beside. */
	section {
		display: grid;
		grid-template-columns: 7rem minmax(0, 68ch);
		gap: var(--space-3);
		padding: var(--space-3) 0 var(--space-4);
		border-top: 1px solid var(--rule);
	}
	section.same .text {
		color: var(--muted);
	}
	h2 {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		font-size: var(--step--1);
		color: var(--muted);
		padding-top: 0.2em;
	}
	.tag {
		font-family: var(--font-text);
		font-weight: 400;
		letter-spacing: 0;
		text-transform: none;
	}
	.tag.changed {
		color: var(--accent);
	}
	p {
		white-space: pre-wrap;
		line-height: 1.6;
	}
	/* A rewritten section reads as the old text, then the new, each named. */
	.which {
		font-size: var(--step--1);
		margin-bottom: var(--space-1);
	}
	p + .which {
		margin-top: var(--space-3);
	}
	/* Added text carries a wash of the accent; dropped text is struck, in a
	   wash of the error color, since it's gone. Neither is a fill of its own.
	   A rewritten section's two texts need no marks: their names say which. */
	ins {
		text-decoration: none;
		background: color-mix(in srgb, var(--accent) 22%, transparent);
		border-radius: 1px;
	}
	del {
		color: var(--muted);
		text-decoration: line-through;
		text-decoration-color: var(--danger);
		background: color-mix(in srgb, var(--danger) 12%, transparent);
		border-radius: 1px;
	}
	.which + p del,
	.which + p ins {
		text-decoration: none;
		background: none;
	}
	.which + p del {
		color: var(--muted);
	}
	@media (max-width: 60rem) {
		section {
			grid-template-columns: 1fr;
		}
	}
</style>
