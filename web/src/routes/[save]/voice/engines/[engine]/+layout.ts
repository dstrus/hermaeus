import { error } from '@sveltejs/kit';
import { engineById, engineByValue, engineFeature } from '#lib/features/voice.ts';
import { valueOf } from '#lib/features/schema.ts';
import { getClones, getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import type { LayoutLoad } from './$types';

// One engine's settings, its cloned voices when it keeps them, and which
// engine SkyrimNet uses: for its Settings and Cloned voices tabs.
export const load: LayoutLoad = async ({ params, fetch, depends }) => {
	const engine = engineById(params.engine);
	if (!engine) error(404, `There’s no voice engine “${params.engine}”.`);
	depends(settingsDependency);
	const cloud = engine.id === 'inworld' || engine.id === 'elevenlabs';
	const [file, game, clones] = await Promise.all([getSettingsFile(engine.file, fetch), getSettingsFile('game', fetch).catch(() => undefined), cloud ? getClones(engine.file, fetch).catch(() => undefined) : undefined]);
	return { engine, feature: engineFeature(engine), file, clones, inUse: engineByValue(valueOf(game?.values.find((s) => s.key === 'tts.engine'))) };
};
