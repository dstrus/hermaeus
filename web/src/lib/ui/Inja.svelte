<!-- A template's text with its Inja markup set apart: statements in the
     accent (the structure that acts), expressions in a heavier ink (what
     gets filled in), comments muted, prose as it is. Put it inside a pre
     or a block that keeps white space. -->
<script lang="ts">
	import { tokens } from './inja.ts';

	let { text }: { text: string } = $props();
	let parts = $derived(tokens(text));
</script>

{#each parts as t, i (i)}{#if t.kind === 'text'}{t.text}{:else}<span class={t.kind}>{t.text}</span>{/if}{/each}

<style>
	.delim {
		color: var(--muted);
	}
	.keyword {
		color: var(--accent);
		font-weight: 500;
	}
	.name,
	.punct,
	.number {
		font-weight: 500;
	}
	.string {
		color: var(--muted);
	}
	.comment {
		color: var(--muted);
		font-style: italic;
	}
</style>
