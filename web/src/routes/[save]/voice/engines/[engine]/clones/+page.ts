import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// Only the engines that clone voices into an account have this tab.
export const load: PageLoad = async ({ parent }) => {
	const { clones, engine } = await parent();
	if (!clones) error(404, `${engine.label} keeps no cloned voices.`);
	return { cache: clones };
};
