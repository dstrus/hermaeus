import { diaryDependency, getDiary } from '#lib/memories/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(diaryDependency);
	const { entries } = await getDiary(params.save, params.uuid, fetch);
	return { entries };
};
