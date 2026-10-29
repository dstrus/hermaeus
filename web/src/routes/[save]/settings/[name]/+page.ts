import { error } from '@sveltejs/kit';
import { filtersDependency, getFilterLists } from '#lib/features/filters.ts';
import { featureById } from '#lib/features/index.ts';
import { getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// A feature's page: one settings file in the feature's own terms.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	const feature = featureById(params.name);
	if (!feature) error(404, `There’s no settings page “${params.name}”.`);
	depends(settingsDependency);
	if (feature.lists) depends(filtersDependency);
	const [file, lists] = await Promise.all([getSettingsFile(feature.file, fetch), feature.lists ? getFilterLists(feature.lists, fetch) : undefined]);
	return { feature, file, lists };
};
