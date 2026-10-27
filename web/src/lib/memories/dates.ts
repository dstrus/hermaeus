// When a memory or diary entry happened: its in-game date, or else when
// SkyrimNet wrote it. The server works out in-game dates; nothing here
// computes one.
import type { GameDate } from '#lib/api/types/calendar.ts';

export interface Dated {
	date?: GameDate;
	written?: number; // Unix ms
}

export function dayLabel(x: Dated): string {
	if (x.date) return x.date.day;
	if (x.written) return `Written ${new Date(x.written).toLocaleDateString(undefined, { dateStyle: 'medium' })}`;
	return 'No date';
}

export function timeLabel(x: Dated): string {
	if (x.date) return x.date.time;
	if (x.written) return new Date(x.written).toLocaleTimeString(undefined, { timeStyle: 'short' });
	return '';
}

// groupByDay splits a list into runs that share a day, keeping its order.
export function groupByDay<T extends Dated>(items: T[]): { day: string; items: T[] }[] {
	const groups: { day: string; items: T[] }[] = [];
	for (const item of items) {
		const day = dayLabel(item);
		const last = groups.at(-1);
		if (last?.day === day) last.items.push(item);
		else groups.push({ day, items: [item] });
	}
	return groups;
}
