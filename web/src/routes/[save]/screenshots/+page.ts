import { getScreenshots } from '#lib/screenshots/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, params }) => ({ list: await getScreenshots(params.save, fetch) });
