<!-- One Universal Translator rule as a form: who it's for, and how they
     speak. The global pattern is the same form with only the pattern. -->
<script lang="ts">
	import { tick, untrack } from 'svelte';
	import type { Rule } from '#lib/api/types/settings.ts';
	import Segmented from '#lib/ui/Segmented.svelte';
	import { entityId, entityTypes, type NPC } from './api.ts';

	let {
		rule,
		global = false,
		npcs = [],
		saving = false,
		onsave,
		oncancel
	}: {
		rule: Rule;
		global?: boolean; // the pattern for everyone: no entity to choose
		npcs?: NPC[];
		saving?: boolean;
		onsave: (rule: Rule) => void;
		oncancel: () => void;
	} = $props();

	// The fields start from the rule as it was opened.
	const start = untrack(() => rule);
	let type = $state(start.entityType);
	// An NPC is named by bio template; a faction or race by editor ID and
	// the end of its form ID, which together make its ID.
	let template = $state(start.entityType === 'npc' ? start.name || start.entityId : '');
	let editorId = $state(start.entityType === 'npc' ? '' : start.editorId || start.entityId.replace(/_[0-9a-f]{3}$/i, ''));
	let formId = $state(start.entityType === 'npc' ? '' : (start.entityId.match(/_([0-9a-f]{3})$/i)?.[1] ?? ''));
	let name = $state(start.entityType === 'npc' ? '' : start.name);
	let pattern = $state(start.pattern);
	let enabled = $state(start.enabled);
	let first = $state<HTMLElement>(); // the first field, focused when the form opens
	let patternField = $state<HTMLTextAreaElement>();
	let error = $state('');

	// An NPC can be typed by name or by bio template; a known name resolves
	// to their template, which is what the rule is keyed by.
	let known = $derived.by(() => {
		const q = template.trim().toLowerCase();
		return q ? npcs.find((n) => n.template.toLowerCase() === q) ?? npcs.find((n) => n.name.toLowerCase() === q) : undefined;
	});
	let chosen = $derived(known?.template ?? template);
	let id = $derived(entityId(type, type === 'npc' ? chosen : editorId, formId));

	$effect(() => {
		tick().then(() => (global ? patternField : first)?.focus());
	});

	function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		if (!global) {
			if (type === 'npc' && !template.trim()) return void (error = 'Name the NPC by bio template.');
			if (type !== 'npc' && !editorId.trim()) return void (error = `Give the ${type}’s editor ID.`);
			if (type !== 'npc' && !/^[0-9a-f]{3,8}$/i.test(formId.trim().replace(/^0x/i, ''))) return void (error = 'The form ID is hexadecimal, such as 0001BDB1; its last three digits are enough.');
		}
		onsave({
			...rule,
			entityType: type,
			entityId: global ? rule.entityId : id,
			name: global ? rule.name : type === 'npc' ? chosen.trim() : name.trim() || editorId.trim(),
			editorId: global || type === 'npc' ? '' : editorId.trim(),
			pattern,
			enabled
		});
	}
</script>

<form onsubmit={submit}>
	{#if !global}
		<div class="field">
			<span class="caps label">For</span>
			<Segmented bind:value={type} options={entityTypes} label="What the rule is for" />
		</div>
		{#if type === 'npc'}
			<label class="field">
				<span class="caps label">NPC</span>
				<input bind:this={first} type="text" bind:value={template} list="translator-npcs" placeholder="Aya" autocomplete="off" spellcheck="false" />
				<datalist id="translator-npcs">
					{#each npcs as n (n.template)}<option value={n.name}>{n.template}</option>{/each}
				</datalist>
				<span class="muted small">
					{#if known}
						{known.name}{known.here ? ' in this playthrough' : ''}, by bio template <span class="id">{known.template}</span>.
					{:else}
						Any NPC with a bio, by name or by bio template.
					{/if}
				</span>
			</label>
		{:else}
			<label class="field">
				<span class="caps label">Editor ID</span>
				<input bind:this={first} type="text" bind:value={editorId} placeholder={type === 'faction' ? 'ThievesGuildFaction' : 'NordRace'} autocomplete="off" spellcheck="false" />
			</label>
			<label class="field">
				<span class="caps label">Form ID</span>
				<input type="text" bind:value={formId} placeholder="00029DA9" autocomplete="off" spellcheck="false" />
				<span class="muted small">From the Creation Kit or xEdit. SkyrimNet keys the rule as <span class="id">{id || '…'}</span>.</span>
			</label>
			<label class="field">
				<span class="caps label">Shown as</span>
				<input type="text" bind:value={name} placeholder={type === 'faction' ? 'Thieves Guild' : 'Nords'} autocomplete="off" />
			</label>
		{/if}
	{/if}
	<label class="field">
		<span class="caps label">Speech pattern</span>
		<textarea bind:this={patternField} bind:value={pattern} rows="6" placeholder={global ? 'How every NPC without a rule of their own speaks' : 'How they speak: tone, rhythm, turns of phrase'}></textarea>
		<span class="muted small">{global ? 'Empty, NPCs without a rule speak as written.' : 'Empty, they speak as written, whatever the global pattern says.'}</span>
	</label>
	{#if !global}
		<label class="check">
			<input type="checkbox" bind:checked={enabled} />
			<span>In force</span>
			<span class="muted small">Off, the rule is skipped and the rules below it can apply.</span>
		</label>
	{/if}
	{#if error}<p class="notice error" role="alert">{error}</p>{/if}
	<div class="buttons">
		<button type="submit" class="button primary" disabled={saving}>Save</button>
		<button type="button" class="button" onclick={oncancel} disabled={saving}>Cancel</button>
	</div>
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.label {
		font-size: var(--step--1);
		color: var(--muted);
	}
	input[type='text'],
	textarea {
		width: 100%;
	}
	textarea {
		resize: vertical;
		line-height: 1.5;
	}
	.small {
		font-size: var(--step--1);
		line-height: 1.4;
	}
	.id {
		font-family: var(--font-mono);
		font-size: 0.95em;
	}
	.check {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: var(--space-1) var(--space-2);
		align-items: center;
	}
	.check .small {
		grid-column: 2;
	}
	/* The buttons stay in reach while the form scrolls, on the pane's material. */
	.buttons {
		position: sticky;
		bottom: 0;
		display: flex;
		gap: var(--space-2);
		padding: var(--space-3) 0;
		background: var(--panel);
	}
</style>
