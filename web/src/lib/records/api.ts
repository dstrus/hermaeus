import { request } from '#lib/api/client.ts';
import type { ActionList, Origin, TriggerList } from '#lib/api/types/records.ts';

export function getActions(fetchFn?: typeof fetch): Promise<ActionList> {
	return request<ActionList>('GET', '/api/actions', undefined, fetchFn);
}

export function getTriggers(fetchFn?: typeof fetch): Promise<TriggerList> {
	return request<TriggerList>('GET', '/api/triggers', undefined, fetchFn);
}

// sourceLabel names where a record comes from, as the Plugins page does.
export function sourceLabel(o: Origin): string {
	return o.layerKind === 'overlay' ? 'Your edits' : o.layerKind === 'base' ? 'SkyrimNet’s defaults' : o.source;
}

// words turns an id such as direct_narration or nearby_npcs into words.
export function words(id: string): string {
	return id.replace(/_/g, ' ').replace(/([a-z])([A-Z])/g, '$1 $2').toLowerCase();
}
