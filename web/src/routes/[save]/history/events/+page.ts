import { getAllEvents } from '#lib/events/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	return { events: await getAllEvents(params.save, fetch) };
};
