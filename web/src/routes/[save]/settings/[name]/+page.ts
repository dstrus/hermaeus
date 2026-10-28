import { error } from '@sveltejs/kit';
import { ApiError } from '#lib/api/client.ts';
import { changedPage, fileLabel, getSettings, getSettingsFile, settingsDependency, type Row } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// One settings file, or, for the changed page, every setting the player
// changed from its default across all the files.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(settingsDependency);
	if (params.name === changedPage) {
		const { files } = await getSettings(fetch);
		const rows: Row[] = files.flatMap((f) => f.values.filter((s) => s.changed).map((setting) => ({ file: f.file, setting })));
		return { title: 'Changed settings', file: null, rows, byFile: true };
	}
	try {
		const f = await getSettingsFile(params.name, fetch);
		const rows: Row[] = f.values.map((setting) => ({ file: f.file, setting }));
		return { title: fileLabel(f.file), file: f.file, rows, byFile: false };
	} catch (e) {
		if (e instanceof ApiError && (e.status === 400 || e.status === 404)) error(404, e.message);
		throw e;
	}
};
