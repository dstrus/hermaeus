import { error } from '@sveltejs/kit';
import { m } from '#lib/copy/messages.ts';
import type { PageLoad } from './$types';

// Only the engines that clone voices into an account have this tab.
export const load: PageLoad = async ({ parent }) => {
	const { clones, engine } = await parent();
	if (!clones) error(404, m.features.engine.noClones({ engine: engine.label }));
	return { cache: clones };
};
