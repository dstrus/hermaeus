import { request } from '#lib/api/client.ts';
import type { Created, ListEdit, ListsResponse } from '#lib/api/types/filters.ts';

export const filtersDependency = 'app:filters';

export type FilterKind = 'memory' | 'actor';

export function getFilterLists(kind: FilterKind, fetchFn?: typeof fetch): Promise<ListsResponse> {
	return request<ListsResponse>('GET', `/api/filters/${kind}`, undefined, fetchFn);
}

export function createList(kind: FilterKind, edit: ListEdit): Promise<Created> {
	return request<Created>('POST', `/api/filters/${kind}`, edit);
}

// saveList writes your list, or your copy of a built-in or plugin list.
export function saveList(kind: FilterKind, id: string, edit: ListEdit): Promise<ListsResponse> {
	return request<ListsResponse>('PUT', `/api/filters/${kind}/${encodeURIComponent(id)}`, edit);
}

export function removeList(kind: FilterKind, id: string, modified: number): Promise<ListsResponse> {
	return request<ListsResponse>('DELETE', `/api/filters/${kind}/${encodeURIComponent(id)}`, { modified });
}

// The six lists a filter list adds to, in SkyrimNet's order.
export const filterFields = [
	{ key: 'FactionWhitelist', label: 'Faction whitelist', noun: 'faction' },
	{ key: 'FactionBlacklist', label: 'Faction blacklist', noun: 'faction' },
	{ key: 'RaceWhitelist', label: 'Race whitelist', noun: 'race' },
	{ key: 'RaceBlacklist', label: 'Race blacklist', noun: 'race' },
	{ key: 'GenderWhitelist', label: 'Gender whitelist', noun: 'gender' },
	{ key: 'GenderBlacklist', label: 'Gender blacklist', noun: 'gender' }
];
