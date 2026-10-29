import { getAllMemories, memoriesDependency } from '#lib/memories/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(memoriesDependency);
	const { memories } = await getAllMemories(params.save, fetch);
	return { memories };
};
