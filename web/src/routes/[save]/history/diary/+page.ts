import { diaryDependency, getAllDiary } from '#lib/memories/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(diaryDependency);
	const { entries } = await getAllDiary(params.save, fetch);
	return { entries };
};
