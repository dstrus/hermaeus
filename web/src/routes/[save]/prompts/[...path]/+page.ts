import { getPrompt, promptDependency } from '#lib/prompts/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(promptDependency);
	return { path: params.path, doc: await getPrompt(params.path, fetch) };
};
