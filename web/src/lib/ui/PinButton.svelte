<!-- The pin: the one control for pinning anywhere. In a table row, a
     pinned one shows always and an unpinned one when the row is under the
     pointer or holds focus, so a row can be pinned without opening it. In
     a pane or a header it always shows. The P shortcut does the same for
     the selected row. -->
<script lang="ts">
	let {
		pinned,
		name,
		disabled = false,
		reason,
		always = false,
		size = 14,
		onclick
	}: {
		pinned: boolean;
		name: string; // what's pinned, for assistive technology
		disabled?: boolean;
		reason?: string; // why it's disabled, for the title
		always?: boolean; // shown even when unpinned and not under the pointer
		size?: number; // the icon's size in pixels
		onclick: () => void;
	} = $props();
</script>

<button
	type="button"
	class="pin"
	class:pinned
	class:always
	aria-pressed={pinned}
	aria-label={pinned ? `Unpin ${name}` : `Pin ${name}`}
	title={disabled && reason ? reason : pinned ? 'Unpin' : 'Pin'}
	style:--size="{size}px"
	{disabled}
	onclick={(e) => {
		e.stopPropagation(); // not a row selection, nor a double-click's first half
		onclick();
	}}
	ondblclick={(e) => e.stopPropagation()}
>
	<svg viewBox="0 0 16 16" width={size} height={size} aria-hidden="true">
		<path class="head" d="M6 2h4l-.6 4.4L12 8.6V10H4V8.6l2.6-2.2Z" />
		<path d="M8 10v4.5" />
	</svg>
</button>

<style>
	.pin {
		display: inline-grid;
		place-items: center;
		width: calc(var(--size) + 10px);
		height: calc(var(--size) + 10px);
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
	.pin.always,
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
	.pin.always:disabled {
		opacity: 0.4;
	}
	/* On its own, the pin has an edge, as a button does. */
	.pin.always {
		border: 1px solid var(--rule-strong);
	}
	.pin.always:hover:not(:disabled) {
		border-color: var(--accent);
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
