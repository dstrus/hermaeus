import { error } from '@sveltejs/kit';
import { m } from '#lib/copy/messages.ts';
import { charactersDependency, getCharacters } from '#lib/characters/api.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ params, parent, fetch, depends }) => {
	const { status } = await parent();
	const playthrough = status.playthroughs.find((p) => p.id === params.save);
	if (!playthrough) error(404, m.shell.noPlaythrough({ save: params.save }));
	depends(charactersDependency);
	const { characters } = await getCharacters(params.save, fetch);
	return { playthrough, characters };
};
