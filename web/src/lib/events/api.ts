import { playthroughPath, request } from '#lib/api/client.ts';
import type { EventsResponse } from '#lib/api/types/characters.ts';

export function getEvents(save: string, uuid: string, fetchFn?: typeof fetch): Promise<EventsResponse> {
	return request<EventsResponse>('GET', playthroughPath(save, 'characters', uuid, 'events'), undefined, fetchFn);
}

// fieldLabel turns a detail's key such as book_name or isPlayerInitiated
// into words.
export function fieldLabel(key: string): string {
	const spaced = key.replace(/_/g, ' ').replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase();
	return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}
