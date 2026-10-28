import { bioDependency, getCatalog, getSaveBios } from '#lib/bios/api.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(bioDependency);
	const [catalog, saveBios] = await Promise.all([getCatalog(fetch), getSaveBios(params.save, fetch)]);
	return { catalog, saveBios: saveBios.bios };
};
