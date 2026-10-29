import { redirect } from '@sveltejs/kit';
import { settingsPages } from '#lib/settings/section.ts';
import type { PageLoad } from './$types';

// There's no page of all settings; the address opens the first one.
export const load: PageLoad = ({ params }) => {
	redirect(307, `/${encodeURIComponent(params.save)}/settings/${settingsPages[0]!.path}`);
};
