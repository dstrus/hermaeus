import { getCatalog } from '#lib/bios/api.ts';
import { getTranslator, translatorDependency } from '#lib/translator/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(translatorDependency);
	// The catalog names every NPC with a bio, for the rules and the form.
	const [translator, catalog] = await Promise.all([getTranslator(fetch), getCatalog(fetch)]);
	return { translator, catalog };
};
