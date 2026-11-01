import { request } from '#lib/api/client.ts';
import type { Created, ListEdit, ListsResponse } from '#lib/api/types/filters.ts';
import { m } from '#lib/copy/messages.ts';

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
const t = m.features.filters.fields;
export const filterFields = [
	{ key: 'FactionWhitelist', label: t.factionWhitelist.label, noun: t.factionWhitelist.noun },
	{ key: 'FactionBlacklist', label: t.factionBlacklist.label, noun: t.factionBlacklist.noun },
	{ key: 'RaceWhitelist', label: t.raceWhitelist.label, noun: t.raceWhitelist.noun },
	{ key: 'RaceBlacklist', label: t.raceBlacklist.label, noun: t.raceBlacklist.noun },
	{ key: 'GenderWhitelist', label: t.genderWhitelist.label, noun: t.genderWhitelist.noun, placeholder: t.genderWhitelist.placeholder },
	{ key: 'GenderBlacklist', label: t.genderBlacklist.label, noun: t.genderBlacklist.noun, placeholder: t.genderBlacklist.placeholder }
];
