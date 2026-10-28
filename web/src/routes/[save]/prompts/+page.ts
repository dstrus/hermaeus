import { getPrompts, promptsDependency } from '#lib/prompts/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(promptsDependency);
	return { list: await getPrompts(fetch) };
};
