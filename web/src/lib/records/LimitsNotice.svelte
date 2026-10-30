<!-- A warning where Hermaeus edits something SkyrimNet limits in ways that
     aren't published: Hermaeus can't show those limits or check them. -->
<script lang="ts">
	let { title, children }: { title: string; children: import('svelte').Snippet } = $props();
</script>

<aside class="limits" role="note">
	<p class="caps title">{title}</p>
	<p class="text">{@render children()}</p>
</aside>

<style>
	.limits {
		max-width: 65ch;
		margin: var(--space-3) 0 var(--space-4);
		padding: var(--space-3) var(--space-4);
		border-left: 3px solid var(--warn);
		background: color-mix(in srgb, var(--warn) 10%, transparent);
		border-radius: 0 var(--radius) var(--radius) 0;
	}
	.title {
		font-size: var(--step--1);
		color: var(--warn);
		margin-bottom: var(--space-1);
	}
	.text {
		font-size: var(--step--1);
		line-height: 1.5;
	}
</style>
