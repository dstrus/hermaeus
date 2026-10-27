<script lang="ts">
	import { page } from '$app/state';
	import CharacterList from '#lib/characters/CharacterList.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
</script>

<div class="playthrough">
	<aside>
		<CharacterList
			save={data.playthrough.id}
			characters={data.characters}
			current={page.params.uuid}
			readOnly={data.status.gameRunning}
		/>
	</aside>
	<section>
		{@render children()}
	</section>
</div>

<style>
	.playthrough {
		display: grid;
		grid-template-columns: minmax(14rem, 20rem) 1fr;
		gap: 2rem;
	}
	@media (max-width: 40rem) {
		.playthrough {
			grid-template-columns: 1fr;
		}
	}
</style>
