import { playthroughPath, request } from '#lib/api/client.ts';
import type { PinsResponse } from '#lib/api/types/characters.ts';

export function getPins(save: string, fetchFn?: typeof fetch): Promise<PinsResponse> {
	return request<PinsResponse>('GET', playthroughPath(save, 'pins'), undefined, fetchFn);
}

// setPinned pins or unpins an NPC, and returns the pins as they are afterward.
export function setPinned(save: string, uuid: string, pinned: boolean): Promise<PinsResponse> {
	return request<PinsResponse>(pinned ? 'PUT' : 'DELETE', playthroughPath(save, 'pins', uuid));
}
