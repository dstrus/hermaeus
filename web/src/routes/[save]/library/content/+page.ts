import { getContent, recordsDependency } from '#lib/records/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(recordsDependency);
	return { list: await getContent(fetch) };
};
