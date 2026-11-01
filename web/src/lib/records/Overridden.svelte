<!-- A record switched in content-registry.json against what its file says,
     and the way back to following the file. -->
<script lang="ts">
	import { m } from '#lib/copy/messages.ts';

	let { override, disabled, onfollow }: { override?: string; disabled: boolean; onfollow: () => void } = $props();
</script>

{#if override === 'enabled' || override === 'disabled'}
	<p class="muted small">
		{override === 'enabled' ? m.records.overridden.switchedOn : m.records.overridden.switchedOff}
		<button type="button" class="textbutton" onclick={onfollow} {disabled}>{m.records.overridden.follow}</button>
	</p>
{/if}

<style>
	p {
		font-size: var(--step--1);
	}
</style>
