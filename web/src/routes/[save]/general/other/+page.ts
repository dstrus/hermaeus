import { getSettings, settingsDependency } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(settingsDependency);
	return { list: await getSettings(fetch) };
};
