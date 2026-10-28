import { request } from '#lib/api/client.ts';
import type { Edit, File, FileList, FileValues, SecretValue, Setting, ToggleList, ToggleRequest } from '#lib/api/types/settings.ts';

// What load functions depend on for the quick toggles and the settings
// files; invalidate them after changing either, since the toggles are
// settings too.
export const togglesDependency = 'app:toggles';
export const settingsDependency = 'app:settings';

export function getToggles(fetchFn?: typeof fetch): Promise<ToggleList> {
	return request<ToggleList>('GET', '/api/toggles', undefined, fetchFn);
}

export function setToggle(id: string, on: boolean): Promise<ToggleList> {
	return request<ToggleList>('PUT', `/api/toggles/${encodeURIComponent(id)}`, { on } satisfies ToggleRequest);
}

export function getSettings(fetchFn?: typeof fetch): Promise<FileList> {
	return request<FileList>('GET', '/api/settings', undefined, fetchFn);
}

export function getSettingsFile(name: string, fetchFn?: typeof fetch): Promise<FileValues> {
	return request<FileValues>('GET', `/api/settings/${encodeURIComponent(name)}`, undefined, fetchFn);
}

export function saveSettings(name: string, edit: Edit): Promise<FileValues> {
	return request<FileValues>('PATCH', `/api/settings/${encodeURIComponent(name)}`, edit);
}

// The page listing everything the player changed, across files.
export const changedPage = 'changed';

export function settingsHref(save: string, name?: string): string {
	const base = `/${encodeURIComponent(save)}/settings`;
	return name ? `${base}/${encodeURIComponent(name)}` : base;
}

// fileLabel is a settings file as people know it: by its file name, with a
// plugin's named after the plugin.
export function fileLabel(f: File): string {
	const plugin = f.name.match(/^Plugin_(.+)$/);
	if (plugin) return `${plugin[1]} (plugin)`;
	return f.path.replace(/^.*\//, '').replace(/\.yaml$/i, '');
}

// A row of a settings table: a setting with the file it belongs to, since
// the changed page mixes files.
export interface Row {
	file: File;
	setting: Setting;
}

export function isSecret(v: unknown): v is SecretValue {
	return typeof v === 'object' && v !== null && (v as SecretValue).secret === true;
}

export type Kind = 'switch' | 'number' | 'text' | 'list' | 'group' | 'secret' | 'none';

// kindOf is what sort of value a setting holds, from its value or, when the
// file doesn't set it, its default.
export function kindOf(s: Setting): Kind {
	const v = s.set ? s.value : s.default;
	if (s.secret || isSecret(v)) return 'secret';
	if (v === undefined || v === null) return 'none';
	if (typeof v === 'boolean') return 'switch';
	if (typeof v === 'number') return 'number';
	if (typeof v === 'string') return 'text';
	if (Array.isArray(v)) return 'list';
	return 'group';
}

// show is a value as a table cell reads it.
export function show(v: unknown): string {
	if (v === undefined) return '';
	if (isSecret(v)) return v.filled ? 'Filled' : 'Empty';
	if (v === null) return '~';
	if (typeof v === 'string') return v;
	if (typeof v === 'boolean' || typeof v === 'number') return String(v);
	return JSON.stringify(v);
}

// editable is the text an edit form starts from.
export function editable(v: unknown): string {
	if (v === undefined || v === null) return '';
	if (typeof v === 'string') return v;
	if (typeof v === 'boolean' || typeof v === 'number') return String(v);
	return JSON.stringify(v, null, 2);
}

// words turns a key's part such as locationAnalysis or bard_singing into
// "Location analysis", for a group heading.
export function words(part: string): string {
	const spaced = part.replace(/_/g, ' ').replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase();
	return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}
