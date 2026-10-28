<!-- / from anywhere in a playthrough: find a character by name and open them. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Character } from '#lib/api/types/characters.ts';
	import Diamond from '#lib/ui/Diamond.svelte';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import { characterName } from './api.ts';

	let { save, characters }: { save: string; characters: Character[] } = $props();

	let dialog: HTMLDialogElement;
	let input = $state<HTMLInputElement>();
	let query = $state('');
	let at = $state(0);

	let matches = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const all = q ? characters.filter((c) => characterName(c).toLowerCase().includes(q)) : characters;
		return all.slice(0, 12);
	});

	function open() {
		query = '';
		at = 0;
		dialog.showModal();
		input?.focus();
	}

	function go(c: Character | undefined) {
		if (!c) return;
		dialog.close();
		goto(`/${encodeURIComponent(save)}/npc/${c.uuid}`);
	}

	function keys(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') at = Math.min(matches.length - 1, at + 1);
		else if (e.key === 'ArrowUp') at = Math.max(0, at - 1);
		else if (e.key === 'Enter') go(matches[at]);
		else return;
		e.preventDefault();
	}

	useShortcuts(() => [{ key: '/', label: 'Find', run: open }]);
</script>

<dialog bind:this={dialog} aria-label="Find a character" onclick={(e) => e.target === dialog && dialog.close()}>
	<input
		bind:this={input}
		bind:value={query}
		oninput={() => (at = 0)}
		onkeydown={keys}
		placeholder="Find a character"
		aria-label="Character name"
		spellcheck="false"
		autocomplete="off"
	/>
	{#if matches.length}
		<ul role="listbox" aria-label="Matching characters">
			{#each matches as c, i (c.uuid)}
				<li role="option" aria-selected={i === at}>
					<button type="button" onclick={() => go(c)} onmouseenter={() => (at = i)}>
						<span class="mark">{#if i === at}<Diamond />{/if}</span>{characterName(c)}
					</button>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="muted">No one matches “{query}”.</p>
	{/if}
</dialog>

<style>
	dialog {
		width: min(28rem, 90vw);
		margin: 18vh auto auto;
		padding: var(--space-4);
		background: var(--smoke);
		color: var(--ink);
		border: 1px solid var(--rule-strong);
		border-radius: var(--radius);
		box-shadow: 0 1.5rem 3rem rgb(0 0 0 / 0.45);
	}
	dialog::backdrop {
		background: rgb(0 0 0 / 0.45);
	}
	input {
		width: 100%;
	}
	ul {
		list-style: none;
		margin: var(--space-3) 0 0;
		padding: 0;
	}
	button {
		display: flex;
		align-items: center;
		width: 100%;
		background: none;
		border: 0;
		padding: var(--space-2) var(--space-1);
		text-align: left;
		cursor: pointer;
	}
	[aria-selected='true'] button {
		background: var(--select);
	}
	.mark {
		width: 1.25rem;
		display: inline-flex;
	}
	p {
		margin-top: var(--space-3);
	}
</style>
