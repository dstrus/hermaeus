<!-- One choice among a few, shown as a row of segments. -->
<script lang="ts">
	let {
		value = $bindable(),
		options,
		label
	}: {
		value: string;
		options: { value: string; label: string; title?: string }[];
		label: string; // for assistive technology
	} = $props();
</script>

<div class="segments caps" role="radiogroup" aria-label={label}>
	{#each options as o (o.value)}
		<button type="button" role="radio" aria-checked={value === o.value} title={o.title} onclick={() => (value = o.value)}>
			{o.label}
		</button>
	{/each}
</div>

<style>
	.segments {
		display: inline-flex;
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		overflow: hidden;
	}
	button {
		background: none;
		border: 0;
		padding: var(--space-1) var(--space-3);
		color: var(--muted);
		font-size: var(--step-0);
		letter-spacing: var(--track);
		cursor: pointer;
	}
	button + button {
		border-left: 1px solid var(--rule-strong);
	}
	button[aria-checked='true'] {
		background: var(--select);
		color: var(--ink);
	}
	button:hover {
		color: var(--ink);
	}
</style>
