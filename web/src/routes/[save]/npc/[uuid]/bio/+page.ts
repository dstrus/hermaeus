import { bioDependency, getBio } from '#lib/bios/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, parent, fetch, depends }) => {
	const { character } = await parent();
	depends(bioDependency);
	const versions = character.bio ? (await getBio(params.save, character.bio, fetch)).versions : [];
	return { versions };
};
