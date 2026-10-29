import { getPlugins, pluginsDependency } from '#lib/plugins/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(pluginsDependency);
	return { list: await getPlugins(fetch) };
};
