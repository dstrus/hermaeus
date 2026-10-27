import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, parent }) => {
	const { characters } = await parent();
	const character = characters.find((c) => c.uuid === params.uuid);
	if (!character) error(404, 'That character has nothing stored in this playthrough.');
	return { character };
};
