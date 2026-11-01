import { playthroughPath, request } from '#lib/api/client.ts';
import type { Summary } from '#lib/api/types/summary.ts';
import { m } from '#lib/copy/messages.ts';

export function getSummary(save: string, fetchFn?: typeof fetch): Promise<Summary> {
	return request<Summary>('GET', playthroughPath(save, 'summary'), undefined, fetchFn);
}

// A playthrough's sections, as the Overview's doorways list them.
export interface Doorway {
	id: string;
	label: string;
	href: string;
	n: number;
	unit: (p: { count: number }) => string; // what's counted, by how many
}

export function doorways(save: string, s: Summary): Doorway[] {
	const base = `/${encodeURIComponent(save)}`;
	const c = s.counts;
	return [
		{ id: 'characters', ...m.overview.doors.characters, href: `${base}/characters`, n: c.characters },
		{ id: 'events', ...m.overview.doors.events, href: `${base}/history/events`, n: c.events },
		{ id: 'memories', ...m.overview.doors.memories, href: `${base}/history/memories`, n: c.memories },
		{ id: 'diary', ...m.overview.doors.diary, href: `${base}/history/diary`, n: c.diary },
		{ id: 'knowledge', ...m.overview.doors.knowledge, href: `${base}/history/knowledge`, n: c.knowledge },
		{ id: 'screenshots', ...m.overview.doors.screenshots, href: `${base}/screenshots`, n: c.screenshots },
		{ id: 'songs', ...m.overview.doors.songs, href: `${base}/songs`, n: c.songs }
	];
}
