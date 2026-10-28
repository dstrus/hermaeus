<!-- A pin in a table row: pinned shows always, unpinned shows when the row
     is under the pointer or holds focus, so a row can be pinned without
     opening it. The P shortcut does the same for the selected row. -->
<script lang="ts">
	let {
		pinned,
		name,
		disabled = false,
		onclick
	}: {
		pinned: boolean;
		name: string; // what's pinned, for assistive technology
		disabled?: boolean;
		onclick: () => void;
	} = $props();
</script>

<button
	type="button"
	class="pin"
	class:pinned
	aria-pressed={pinned}
	aria-label={pinned ? `Unpin ${name}` : `Pin ${name}`}
	title={pinned ? 'Unpin' : 'Pin'}
	{disabled}
	onclick={(e) => {
		e.stopPropagation(); // not a row selection, nor a double-click's first half
		onclick();
	}}
	ondblclick={(e) => e.stopPropagation()}
>
	<svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
		<path class="head" d="M6 2h4l-.6 4.4L12 8.6V10H4V8.6l2.6-2.2Z" />
		<path d="M8 10v4.5" />
	</svg>
</button>

<style>
	.pin {
		display: inline-grid;
		place-items: center;
		width: 1.5rem;
		height: 1.5rem;
		padding: 0;
		background: none;
		border: 0;
		border-radius: var(--radius);
		color: var(--muted);
		cursor: pointer;
		opacity: 0;
	}
	/* Shown for a pinned row, and for any row under the pointer or with focus. */
	.pin.pinned,
	.pin:focus-visible,
	:global(tr:hover) .pin,
	:global(tr:focus-within) .pin {
		opacity: 1;
	}
	.pin.pinned {
		color: var(--ink);
	}
	.pin:hover:not(:disabled) {
		color: var(--ink);
		background: var(--glass);
	}
	.pin:disabled {
		cursor: not-allowed;
		opacity: 0;
	}
	svg {
		fill: none;
		stroke: currentColor;
		stroke-width: 1.4;
		stroke-linejoin: round;
		stroke-linecap: round;
	}
	.pinned .head {
		fill: currentColor;
	}
</style>
