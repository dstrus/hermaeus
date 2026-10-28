<script lang="ts">
	import { page } from '$app/state';
</script>

<h1 class="caps">{page.status === 404 ? 'Not found' : 'Something went wrong'}</h1>
<p>{page.error?.message}</p>
<p><a href="/">Back to your playthroughs</a></p>

<style>
	h1 {
		font-size: var(--step-3);
		font-weight: 600;
		line-height: 1;
		margin-bottom: var(--space-4);
	}
	p {
		max-width: 65ch;
		margin-bottom: var(--space-3);
	}
</style>
