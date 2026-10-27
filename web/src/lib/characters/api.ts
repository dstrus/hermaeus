import { playthroughPath, request } from '#lib/api/client.ts';
import type { CharactersResponse, Overview, PinsResponse, Stats } from '#lib/api/types/characters.ts';

// What load functions depend on for the character list; invalidate it after
// changing anything the list shows.
export const charactersDependency = 'app:characters';

export function getCharacters(save: string, fetchFn?: typeof fetch): Promise<CharactersResponse> {
	return request<CharactersResponse>('GET', playthroughPath(save, 'characters'), undefined, fetchFn);
}

// setPinned pins or unpins an NPC, and returns the pins as they are afterward.
export function setPinned(save: string, uuid: string, pinned: boolean): Promise<PinsResponse> {
	return request<PinsResponse>(pinned ? 'PUT' : 'DELETE', playthroughPath(save, 'pins', uuid));
}

// characterName is how a character is shown, including ones SkyrimNet never
// named.
export function characterName(c: { name: string; uuid: string }): string {
	return c.name || `Unnamed (${c.uuid})`;
}

export function getOverview(save: string, uuid: string, fetchFn?: typeof fetch): Promise<Overview> {
	return request<Overview>('GET', playthroughPath(save, 'characters', uuid, 'overview'), undefined, fetchFn);
}

export function getStats(save: string, uuid: string, fetchFn?: typeof fetch): Promise<Stats> {
	return request<Stats>('GET', playthroughPath(save, 'characters', uuid, 'stats'), undefined, fetchFn);
}
