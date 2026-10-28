import { ApiError } from '#lib/api/client.ts';
import { getOverview } from '#lib/characters/api.ts';
import { getSpeech } from '#lib/settings/api.ts';
import type { SpeechRule } from '#lib/api/types/settings.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, parent }) => {
	const { character } = await parent();
	// The speech pattern comes from Universal Translator's settings file; a
	// problem reading it shouldn't keep the overview from opening.
	let speech: SpeechRule | null = null;
	let speechError = '';
	const [overview] = await Promise.all([
		getOverview(params.save, params.uuid, fetch),
		character.bio
			? getSpeech(character.bio, fetch)
					.then((r) => (speech = r))
					.catch((e) => (speechError = e instanceof ApiError ? e.message : String(e)))
			: undefined
	]);
	return { overview, speech, speechError };
};
