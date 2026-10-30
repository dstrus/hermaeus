// Route parameter matchers, by name: [group=group] only matches a sidebar
// group's path, so /<save>/<group>/<feature> doesn't swallow other URLs.
import { defineParams } from '@sveltejs/kit/params';
import { groups } from '#lib/nav.ts';

export const params = defineParams({
	group: (param) => (groups.some((g) => g.path === param) ? param : undefined)
});
