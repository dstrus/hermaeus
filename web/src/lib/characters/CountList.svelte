<script lang="ts">
	import type { Count } from '#lib/api/types/characters.ts';
	import { count } from '#lib/format.ts';

	let {
		items,
		linkTo,
		empty = 'None yet.',
		blank = 'none'
	}: {
		items: Count[];
		linkTo?: (c: Count) => string | undefined; // where an item links, if anywhere
		empty?: string;
		blank?: string; // what to call an item with no name
	} = $props();
</script>

{#if items.length}
	<ul>
		{#each items as c, i (c.uuid ?? c.key ?? `${c.name}:${i}`)}
			{@const href = linkTo?.(c)}
			<li>
				{#if href}<a {href}>{c.name}</a>{:else if c.name}{c.name}{:else}<em>{blank}</em>{/if}
				— {count(c.n)}
			</li>
		{/each}
	</ul>
{:else}
	<p>{empty}</p>
{/if}
