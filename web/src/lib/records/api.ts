import { request } from '#lib/api/client.ts';
import type { Layer } from '#lib/api/types/layers.ts';
import type { ActionList, ContentList, Origin, OverrideList, RecordEdit, Revert, Switch, TriggerList, Version } from '#lib/api/types/records.ts';

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
	return id
		.replace(/_/g, ' ')
		.replace(/([a-z])([A-Z])/g, '$1 $2')
		.toLowerCase();
}

// What the record loaders depend on; invalidate it after a record changes.
export const recordsDependency = 'app:records';

export function getContent(fetchFn?: typeof fetch): Promise<ContentList> {
	return request<ContentList>('GET', '/api/content', undefined, fetchFn);
}

export function getOverrides(fetchFn?: typeof fetch): Promise<OverrideList> {
	return request<OverrideList>('GET', '/api/overrides', undefined, fetchFn);
}

// filePath is the API path for a file, whose own path has slashes.
function filePath(what: 'file' | 'enabled', path: string): string {
	return `/api/content/${what}/` + path.split('/').map(encodeURIComponent).join('/');
}

// getVersions is every layer's copy of a file, highest priority first.
export function getVersions(path: string): Promise<Version[]> {
	return request<Version[]>('GET', filePath('file', path));
}

export function saveRecord(path: string, edit: RecordEdit): Promise<Version[]> {
	return request<Version[]>('PUT', filePath('file', path), edit);
}

export function removeRecordCopy(path: string, r: Revert): Promise<Version[]> {
	return request<Version[]>('DELETE', filePath('file', path), r);
}

// switchRecord switches a record on or off in your copy of it, making that
// copy if there isn't one. It reads the record first, for your copy's time.
export async function switchRecord(path: string, enabled: boolean): Promise<Version[]> {
	const mine = (await getVersions(path)).find((v) => v.kind === 'overlay');
	return request<Version[]>('PUT', filePath('enabled', path), {
		modified: mine?.modified ?? 0,
		enabled
	} satisfies Switch);
}

// layerName names a content layer, as sourceLabel does for a record.
export function layerName(l: Layer | undefined): string {
	if (!l) return 'an unknown layer';
	return l.kind === 'overlay' ? 'Your edits' : l.kind === 'base' ? 'SkyrimNet’s defaults' : l.source;
}
