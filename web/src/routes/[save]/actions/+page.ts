import { getActions } from '#lib/records/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => ({ list: await getActions(fetch) });
