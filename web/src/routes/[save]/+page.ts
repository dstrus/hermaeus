import { getModels } from '#lib/models/api.ts';
import { getSettingsFile } from '#lib/settings/api.ts';
import { getSummary } from '#lib/summary/api.ts';
import type { PageLoad } from './$types';

// The playthrough's Overview. Models and the voice engine come from
// SkyrimNet's settings files; a problem reading those shouldn't keep the
// Overview from opening.
export const load: PageLoad = async ({ params, fetch }) => {
	const [summary, models, game] = await Promise.all([
		getSummary(params.save, fetch),
		getModels(fetch).catch(() => null),
		getSettingsFile('game', fetch).catch(() => null)
	]);
	const engine = game?.values.find((v) => v.key === 'tts.engine');
	return { summary, models, engine: engine ? String(engine.value ?? engine.default ?? '') : null };
};
