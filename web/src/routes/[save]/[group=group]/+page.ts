import { redirect } from '@sveltejs/kit';
import { groups } from '#lib/nav.ts';
import type { PageLoad } from './$types';

// A group has no page of its own; its URL opens its first page.
export const load: PageLoad = ({ params }) => {
	const g = groups.find((g) => g.path === params.group)!;
	redirect(307, `/${encodeURIComponent(params.save)}/${g.path}/${g.pages[0]!.path}`);
};
