import { request } from '#lib/api/client.ts';
import { href } from '#lib/nav.ts';
import type { CloneCache, Edit, File, FileList, FileValues, PinsResponse, SecretValue, Setting, SpeechEdit, SpeechRule, ToggleList, ToggleRequest } from '#lib/api/types/settings.ts';

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

// setFilePinned pins or unpins a settings file, and returns the pins as
// they are afterward.
export function setFilePinned(name: string, pinned: boolean): Promise<PinsResponse> {
	return request<PinsResponse>(pinned ? 'PUT' : 'DELETE', `/api/settings/${encodeURIComponent(name)}/pin`);
}

// The page listing everything the player changed, across files.
export const changedPage = 'changed';

// settingsHref is the Other page's list of settings files, or one file.
export function settingsHref(save: string, name?: string): string {
	return href(save, 'other', name ? `/${encodeURIComponent(name)}` : '');
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

export type Kind = 'switch' | 'number' | 'hotkey' | 'text' | 'list' | 'group' | 'secret' | 'none';

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
	const spaced = part
		.replace(/_/g, ' ')
		.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
		.toLowerCase();
	return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}

// An Entry is one row of a map being edited: its key, the kind of value it
// holds, and that value as the row's field keeps it.
export interface Entry {
	key: string;
	kind: 'switch' | 'number' | 'text' | 'json';
	text: string; // for number, text, and json
	on: boolean; // for switch
}

export function toEntries(obj: Record<string, unknown>): Entry[] {
	return Object.entries(obj).map(([key, v]) => {
		if (typeof v === 'boolean') return { key, kind: 'switch', text: '', on: v };
		if (typeof v === 'number') return { key, kind: 'number', text: String(v), on: false };
		if (typeof v === 'string') return { key, kind: 'text', text: v, on: false };
		return { key, kind: 'json', text: v === null ? '' : JSON.stringify(v), on: false };
	});
}

// fromEntries builds the map back, refusing an empty or repeated key, a
// number that isn't one, or JSON that doesn't parse. An empty number or
// JSON field is no value.
export function fromEntries(entries: Entry[]): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const e of entries) {
		const key = e.key.trim();
		if (!key) throw new Error('Every row needs a key.');
		if (key in out) throw new Error(`${key} is listed twice.`);
		switch (e.kind) {
			case 'switch':
				out[key] = e.on;
				break;
			case 'number': {
				const t = e.text.trim();
				if (t === '') out[key] = null;
				else if (Number.isNaN(Number(t))) throw new Error(`${key} holds a number; “${e.text}” isn’t one.`);
				else out[key] = Number(t);
				break;
			}
			case 'json': {
				const t = e.text.trim();
				if (t === '') out[key] = null;
				else {
					try {
						out[key] = JSON.parse(t);
					} catch {
						throw new Error(`${key} isn’t valid JSON.`);
					}
				}
				break;
			}
			default:
				out[key] = e.text;
		}
	}
	return out;
}

// parseJSONText is the object or list a text setting holds as JSON, or
// undefined when it holds something else.
export function parseJSONText(v: unknown): Record<string, unknown> | unknown[] | undefined {
	if (typeof v !== 'string' || !/^\s*[[{]/.test(v)) return undefined;
	try {
		const parsed: unknown = JSON.parse(v);
		return typeof parsed === 'object' && parsed !== null ? (parsed as Record<string, unknown> | unknown[]) : undefined;
	} catch {
		return undefined;
	}
}

// An NPC's Universal Translator speech pattern, by bio template.
export function getSpeech(template: string, fetchFn?: typeof fetch): Promise<SpeechRule> {
	return request<SpeechRule>('GET', `/api/speech/${encodeURIComponent(template)}`, undefined, fetchFn);
}

// setSpeech sets the pattern; an empty one removes the NPC's rule.
export function setSpeech(template: string, edit: SpeechEdit): Promise<SpeechRule> {
	return request<SpeechRule>('PUT', `/api/speech/${encodeURIComponent(template)}`, edit);
}

// getClones lists the voices a cloud voice engine has cloned into the
// player's account: voice type, language, and when each was last used.
export function getClones(engine: string, fetchFn?: typeof fetch): Promise<CloneCache> {
	return request<CloneCache>('GET', `/api/voice/clones/${encodeURIComponent(engine)}`, undefined, fetchFn);
}
