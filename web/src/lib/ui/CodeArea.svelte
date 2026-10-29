<!-- A text area for a template, with its markup set apart as it's typed:
     the highlighted text sits beneath a text area whose own text is
     invisible, the two sharing one font, one wrap, and one scroll. -->
<script lang="ts">
	import Inja from './Inja.svelte';

	let { value = $bindable(), label, rows }: { value: string; label: string; rows?: number } = $props();

	let mirror = $state<HTMLElement>();
	function sync(e: Event) {
		if (!mirror) return;
		const t = e.currentTarget as HTMLTextAreaElement;
		mirror.scrollTop = t.scrollTop;
		mirror.scrollLeft = t.scrollLeft;
	}
</script>

<div class="code">
	<pre bind:this={mirror} aria-hidden="true"><Inja text={value} />{'\n'}</pre>
	<textarea bind:value spellcheck="false" autocapitalize="off" autocomplete="off" aria-label={label} {rows} onscroll={sync}></textarea>
</div>

<style>
	/* The field's material goes on the box, beneath the highlighted text;
	   the text area over it is clear, so it never dims the marks. */
	.code {
		position: relative;
		display: flex;
		flex-direction: column;
		min-height: 0;
		background: var(--glass-strong);
		border-radius: var(--radius);
	}
	/* The same box, font, and wrapping for both, so the letters line up. */
	pre,
	textarea {
		margin: 0;
		padding: var(--space-2) var(--space-3);
		font: 400 var(--step--1) / 1.55 var(--font-mono);
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		tab-size: 4;
	}
	pre {
		position: absolute;
		inset: 0;
		overflow: hidden;
		border: 1px solid transparent;
		pointer-events: none;
	}
	textarea {
		position: relative;
		flex: 1;
		width: 100%;
		min-height: 8rem;
		resize: none;
		color: transparent;
		caret-color: var(--ink);
		background: transparent;
	}
	textarea::selection {
		color: transparent;
	}
</style>
