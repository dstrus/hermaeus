import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// History opens on its events.
export const load: PageLoad = ({ params }) => {
	redirect(307, `/${encodeURIComponent(params.save)}/history/events`);
};
