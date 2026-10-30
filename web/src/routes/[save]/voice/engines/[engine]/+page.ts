import { error } from '@sveltejs/kit';
import { engineById, engineByValue, engineFeature } from '#lib/features/voice.ts';
import { valueOf } from '#lib/features/schema.ts';
import { getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import type { PageLoad } from './$types';

// One engine's settings, and which engine SkyrimNet uses.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	const engine = engineById(params.engine);
	if (!engine) error(404, `There’s no voice engine “${params.engine}”.`);
	depends(settingsDependency);
	const [file, game] = await Promise.all([getSettingsFile(engine.file, fetch), getSettingsFile('game', fetch).catch(() => undefined)]);
	return { engine, feature: engineFeature(engine), file, inUse: engineByValue(valueOf(game?.values.find((s) => s.key === 'tts.engine'))) };
};
