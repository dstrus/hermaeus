import { error } from '@sveltejs/kit';
import { getPins } from '#lib/characters/pins.ts';
import type { PageLoad } from './$types.ts';

export const load: PageLoad = async ({ params, parent, fetch }) => {
	const { status } = await parent();
	const playthrough = status.playthroughs.find((p) => p.id === params.save);
	if (!playthrough) error(404, `There’s no playthrough “${params.save}” in this SkyrimNet folder.`);
	return { playthrough, pins: await getPins(params.save, fetch) };
};
