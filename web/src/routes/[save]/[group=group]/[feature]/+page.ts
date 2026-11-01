import { error } from '@sveltejs/kit';
import { m } from '#lib/copy/messages.ts';
import { filtersDependency, getFilterLists } from '#lib/features/filters.ts';
import { featureById } from '#lib/features/index.ts';
import { featureGroup } from '#lib/nav.ts';
import { getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// A feature's page: one settings file in the feature's own terms.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	const feature = featureGroup(params.group, params.feature) && featureById(params.feature);
	if (!feature) error(404, m.features.page.noPage({ feature: params.feature, group: params.group }));
	depends(settingsDependency);
	if (feature.lists) depends(filtersDependency);
	const [file, lists] = await Promise.all([getSettingsFile(feature.file, fetch), feature.lists ? getFilterLists(feature.lists, fetch) : undefined]);
	return { feature, file, lists };
};
