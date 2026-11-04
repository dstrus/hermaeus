<script lang="ts">
	import { goto } from '$app/navigation';
	import FolderForm from '#lib/folders/FolderForm.svelte';
	import { m } from '#lib/copy/messages.ts';
	import { useShortcuts } from '#lib/ui/shortcuts.svelte.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// Cancelling goes back to the playthrough. Saving goes to the list of
	// playthroughs, since the new folder may not have this one.
	const close = (saved: boolean) => goto(saved ? '/' : `/${encodeURIComponent(data.playthrough.id)}`);

	useShortcuts(() => [{ key: 'Escape', shown: 'Esc', label: m.common.cancel, run: () => close(false) }]);
</script>

<FolderForm current={{ path: data.status.skyrimNetDir, modPath: data.status.modDirDetected ? '' : data.status.modDir }} {close} />
