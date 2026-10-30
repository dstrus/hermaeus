import { getSongs } from '#lib/songs/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, params }) => ({ list: await getSongs(params.save, fetch) });
