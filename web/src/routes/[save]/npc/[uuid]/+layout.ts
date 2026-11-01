import { error } from '@sveltejs/kit';
import { m } from '#lib/copy/messages.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ params, parent }) => {
	const { characters } = await parent();
	const character = characters.find((c) => c.uuid === params.uuid);
	if (!character) error(404, m.characters.page.notFound);
	return { character };
};
