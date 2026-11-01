<!-- What an engine's pages share in their title bar: whether SkyrimNet uses
     it, and a menu to go to another engine's same tab. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { m } from '#lib/copy/messages.ts';
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
	<span class="muted">{m.features.engine.inUse}</span>
{:else}
	<span class="muted">{m.features.engine.notInUse} · <a href="{href(save, 'speech')}?item=tts.engine">{m.features.engine.choose}</a></span>
{/if}
<select aria-label={m.features.engine.menu} value={engine.id} onchange={(e) => go(e.currentTarget.value)}>
	{#each engines as e (e.id)}<option value={e.id}>{inUse?.id === e.id ? m.features.engine.optionInUse({ engine: e.label }) : e.label}</option>{/each}
</select>
