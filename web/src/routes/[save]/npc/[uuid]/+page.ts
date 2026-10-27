import { getOverview } from '#lib/characters/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	return { overview: await getOverview(params.save, params.uuid, fetch) };
};
