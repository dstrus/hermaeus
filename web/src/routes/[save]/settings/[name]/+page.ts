import { error } from '@sveltejs/kit';
import { ApiError } from '#lib/api/client.ts';
import { featureById } from '#lib/features/index.ts';
import { changedPage, fileLabel, getSettings, getSettingsFile, settingsDependency, type Row } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// One settings file, or, for the changed page, every setting the player
// changed from its default across all the files; or a feature's page,
// which shows its file in the feature's own terms.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(settingsDependency);
	const feature = featureById(params.name);
	if (feature) {
		return { title: feature.title, feature, file: await getSettingsFile(feature.file, fetch), rows: [] as Row[], byFile: false };
	}
	if (params.name === changedPage) {
		const { files } = await getSettings(fetch);
		const rows: Row[] = files.flatMap((f) => f.values.filter((s) => s.changed).map((setting) => ({ file: f.file, setting })));
		return { title: 'Changed settings', feature: undefined, file: null, rows, byFile: true };
	}
	try {
		const f = await getSettingsFile(params.name, fetch);
		const rows: Row[] = f.values.map((setting) => ({ file: f.file, setting }));
		return { title: fileLabel(f.file), feature: undefined, file: f.file, rows, byFile: false };
	} catch (e) {
		if (e instanceof ApiError && (e.status === 400 || e.status === 404)) error(404, e.message);
		throw e;
	}
};
