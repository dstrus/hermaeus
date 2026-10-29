import { error } from '@sveltejs/kit';
import { featureOf, getPluginSchemas } from '#lib/features/plugins.ts';
import { getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// A plugin's settings, as its own schema describes them.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(settingsDependency);
	const schema = (await getPluginSchemas(fetch)).find((s) => s.folder.toLowerCase() === params.name.toLowerCase());
	if (!schema) error(404, `No plugin describes settings called “${params.name}”.`);
	return { feature: featureOf(schema), file: await getSettingsFile(schema.name, fetch) };
};
