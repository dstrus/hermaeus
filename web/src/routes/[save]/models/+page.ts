import { getModels, modelsDependency } from '#lib/models/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(modelsDependency);
	return { models: await getModels(fetch) };
};
