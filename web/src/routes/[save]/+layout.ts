import { error } from '@sveltejs/kit';
import { charactersDependency, getCharacters } from '#lib/characters/api.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ params, parent, fetch, depends }) => {
	const { status } = await parent();
	const playthrough = status.playthroughs.find((p) => p.id === params.save);
	if (!playthrough) error(404, `There’s no playthrough “${params.save}” in this SkyrimNet folder.`);
	depends(charactersDependency);
	const { characters } = await getCharacters(params.save, fetch);
	return { playthrough, characters };
};
