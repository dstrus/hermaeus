import { getStats } from '#lib/characters/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	return { stats: await getStats(params.save, params.uuid, fetch) };
};
