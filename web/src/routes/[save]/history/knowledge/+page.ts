import { getKnowledge, knowledgeDependency } from '#lib/memories/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(knowledgeDependency);
	return { knowledge: await getKnowledge(params.save, fetch) };
};
