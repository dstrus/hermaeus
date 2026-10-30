import { playthroughPath, request } from '#lib/api/client.ts';
import type { Changes, List } from '#lib/api/types/songs.ts';

export function getSongs(save: string, fetchFn?: typeof fetch): Promise<List> {
	return request<List>('GET', playthroughPath(save, 'songs'), undefined, fetchFn);
}

export function saveSong(save: string, id: number, c: Changes): Promise<List> {
	return request<List>('PATCH', playthroughPath(save, 'songs', String(id)), c);
}

export function deleteSong(save: string, id: number): Promise<List> {
	return request<List>('DELETE', playthroughPath(save, 'songs', String(id)));
}

export function audioPath(save: string, id: number): string {
	return playthroughPath(save, 'songs', String(id), 'audio');
}

// A part of a song's lyrics: a section's name, such as Verse, and its lines.
export interface Part {
	name: string;
	text: string;
}

// parts splits lyrics at their [Section] markers. Text before the first
// marker is a part with no name.
export function parts(lyrics: string): Part[] {
	const out: Part[] = [];
	let cur: Part = { name: '', text: '' };
	for (const line of lyrics.split('\n')) {
		const m = /^\s*\[([^\]]+)\]\s*$/.exec(line);
		if (m) {
			if (cur.name || cur.text.trim()) out.push(cur);
			cur = { name: m[1]!.trim(), text: '' };
		} else {
			cur.text += (cur.text ? '\n' : '') + line;
		}
	}
	if (cur.name || cur.text.trim()) out.push(cur);
	return out.map((p) => ({ ...p, text: p.text.trim() }));
}

// seconds reads as 3:05.
export function clock(s: number): string {
	if (!isFinite(s) || s < 0) return '–:––';
	const m = Math.floor(s / 60);
	return `${m}:${String(Math.floor(s % 60)).padStart(2, '0')}`;
}

// createdMs reads SQLite's UTC "2026-10-11 03:43:14".
export function createdMs(created: string): number {
	return Date.parse(created.replace(' ', 'T') + (/[zZ]|[+-]\d\d:?\d\d$/.test(created) ? '' : 'Z'));
}
