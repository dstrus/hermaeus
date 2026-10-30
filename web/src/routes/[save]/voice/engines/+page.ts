import { redirect } from '@sveltejs/kit';
import { engineByValue, engines } from '#lib/features/voice.ts';
import { getSettingsFile, settingsDependency } from '#lib/settings/api.ts';
import { valueOf } from '#lib/features/schema.ts';
import type { PageLoad } from './$types';

// Engines opens on the engine in use, or the first when there's none.
export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(settingsDependency);
	const game = await getSettingsFile('game', fetch).catch(() => undefined);
	const inUse = engineByValue(valueOf(game?.values.find((s) => s.key === 'tts.engine')));
	redirect(307, `/${encodeURIComponent(params.save)}/voice/engines/${(inUse ?? engines[0]!).id}`);
};
