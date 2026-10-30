import { getProviders, providersDependency } from '#lib/providers/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, depends }) => {
	depends(providersDependency);
	return { providers: await getProviders(fetch) };
};
