import { bioDependency, getBio } from '#lib/bios/api.ts';
import type { PageLoad } from './$types';

// A character's bio by template, for characters not met in this playthrough
// (met ones have a Bio tab of their own).
export const load: PageLoad = async ({ params, fetch, depends }) => {
	depends(bioDependency);
	const { versions } = await getBio(params.save, params.template, fetch);
	return { template: params.template, versions };
};
