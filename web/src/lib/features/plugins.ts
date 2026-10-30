// A plugin's settings page, built from the schema the plugin ships: its
// categories become sections, and each field's type a kind of value.
import { request } from '#lib/api/client.ts';
import type { PluginSchema, SchemaField } from '#lib/api/types/settings.ts';
import type { Feature, Field, Section } from './schema.ts';

export function getPluginSchemas(fetchFn?: typeof fetch): Promise<PluginSchema[]> {
	return request<PluginSchema[]>('GET', '/api/plugin-settings', undefined, fetchFn);
}

export function pluginSettingsHref(save: string, folder: string): string {
	return `/${encodeURIComponent(save)}/plugin-settings/${encodeURIComponent(folder)}`;
}

function fieldOf(f: SchemaField): Field {
	const base = { key: f.path, label: f.name || f.path, what: f.description ?? '' };
	switch (f.type) {
		case 'bool':
			return { ...base, kind: 'switch' };
		case 'int':
			return { ...base, kind: 'count', min: f.min, max: f.max };
		case 'float':
			return { ...base, kind: 'number', min: f.min, max: f.max };
		case 'select':
			return { ...base, kind: 'choice', choices: (f.options ?? []).map((o) => ({ value: o, label: o })) };
		case 'hotkey':
			return { ...base, kind: 'hotkey', pluginHotkey: true };
		default:
			return { ...base, kind: f.secret ? 'secret' : 'text' };
	}
}

export function featureOf(ps: PluginSchema): Feature {
	const sections: Section[] = [];
	for (const f of ps.fields) {
		const title = f.category || 'Settings';
		let s = sections.find((x) => x.title === title);
		if (!s) sections.push((s = { title, fields: [] }));
		s.fields.push(fieldOf(f));
	}
	return {
		id: `plugin/${ps.folder}`,
		file: ps.name,
		title: ps.title,
		what: [ps.description, 'These are the plugin’s own settings, as it describes them.'].filter(Boolean).join(' '),
		sections
	};
}
