<!-- What an engine's pages share in their title bar: whether SkyrimNet uses
     it, and a menu to go to another engine's same tab. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { href } from '#lib/nav.ts';
	import { engines, type Engine } from './voice.ts';

	let { save, engine, inUse, tab = '' }: { save: string; engine: Engine; inUse?: Engine; tab?: '' | 'clones' } = $props();

	// Another engine opens on the same tab when it has one.
	const hasClones = (id: string) => id === 'inworld' || id === 'elevenlabs';
	function go(id: string) {
		goto(href(save, 'engines', `/${id}${tab === 'clones' && hasClones(id) ? '/clones' : ''}`));
	}
</script>

{#if inUse?.id === engine.id}
	<span class="muted">In use</span>
{:else}
	<span class="muted">Not in use · <a href="{href(save, 'speech')}?item=tts.engine">Choose the engine</a></span>
{/if}
<select aria-label="Engine" value={engine.id} onchange={(e) => go(e.currentTarget.value)}>
	{#each engines as e (e.id)}<option value={e.id}>{e.label}{inUse?.id === e.id ? ' · in use' : ''}</option>{/each}
</select>
