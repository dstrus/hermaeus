import { playthroughPath, request } from '#lib/api/client.ts';
import type { Summary } from '#lib/api/types/summary.ts';

export function getSummary(save: string, fetchFn?: typeof fetch): Promise<Summary> {
	return request<Summary>('GET', playthroughPath(save, 'summary'), undefined, fetchFn);
}

// A playthrough's sections, as the Overview's doorways list them.
export interface Doorway {
	id: string;
	label: string;
	href: string;
	n: number;
	unit: [string, string]; // what's counted, one and many
}

export function doorways(save: string, s: Summary): Doorway[] {
	const base = `/${encodeURIComponent(save)}`;
	const c = s.counts;
	return [
		{ id: 'characters', label: 'Characters', href: `${base}/characters`, n: c.characters, unit: ['with something stored', 'with something stored'] },
		{ id: 'events', label: 'Events', href: `${base}/history/events`, n: c.events, unit: ['event', 'events'] },
		{ id: 'memories', label: 'Memories', href: `${base}/history/memories`, n: c.memories, unit: ['memory', 'memories'] },
		{ id: 'diary', label: 'Diary', href: `${base}/history/diary`, n: c.diary, unit: ['entry', 'entries'] },
		{ id: 'knowledge', label: 'Knowledge', href: `${base}/history/knowledge`, n: c.knowledge, unit: ['entry', 'entries'] },
		{ id: 'screenshots', label: 'Screenshots', href: `${base}/screenshots`, n: c.screenshots, unit: ['screenshot', 'screenshots'] },
		{ id: 'songs', label: 'Bard songs', href: `${base}/songs`, n: c.songs, unit: ['song', 'songs'] }
	];
}
