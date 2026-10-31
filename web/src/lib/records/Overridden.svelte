<!-- A record switched in content-registry.json against what its file says,
     and the way back to following the file. -->
<script lang="ts">
	let { override, disabled, onfollow }: { override?: string; disabled: boolean; onfollow: () => void } = $props();
</script>

{#if override === 'enabled' || override === 'disabled'}
	<p class="muted small">
		Switched {override === 'enabled' ? 'on' : 'off'} in content-registry.json, where SkyrimNet’s dashboard keeps its switches. Its file says {override === 'enabled' ? 'off' : 'on'}.
		<button type="button" class="textbutton" onclick={onfollow} {disabled}>Follow its file</button>
	</p>
{/if}

<style>
	p {
		font-size: var(--step--1);
	}
</style>
