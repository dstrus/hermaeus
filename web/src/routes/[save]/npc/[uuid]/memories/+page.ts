import { getMemories } from '#lib/memories/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	const { memories } = await getMemories(params.save, params.uuid, fetch);
	return { memories };
};
