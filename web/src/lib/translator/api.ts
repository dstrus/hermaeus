import { request } from '#lib/api/client.ts';
import { href } from '#lib/nav.ts';
import type { Catalog } from '#lib/api/types/bios.ts';
import type { Character } from '#lib/api/types/characters.ts';
import type { Import, ImportResult, Rule, Translator, TranslatorEdit } from '#lib/api/types/settings.ts';
import { templateName } from '#lib/bios/api.ts';
import { modelsHref } from '#lib/models/api.ts';

export const translatorDependency = 'app:translator';

export function getTranslator(fetchFn?: typeof fetch): Promise<Translator> {
	return request<Translator>('GET', '/api/translator', undefined, fetchFn);
}

// saveTranslator writes the whole configuration; the server keeps the
// file's formatting where the change allows.
export function saveTranslator(edit: TranslatorEdit): Promise<Translator> {
	return request<Translator>('PUT', '/api/translator', edit);
}

export function importRules(imp: Import): Promise<ImportResult> {
	return request<ImportResult>('POST', '/api/translator/import', imp);
}

export function translatorHref(save: string): string {
	return href(save, 'translator');
}

// Where the model that rewrites the lines is chosen: the translator's task
// on the Models page.
export function modelHref(save: string): string {
	return modelsHref(save, 'UniversalTranslator');
}

// editOf is the configuration as it is, ready to be changed and saved.
export function editOf(t: Translator): TranslatorEdit {
	return {
		modified: t.modified,
		enabled: t.enabled,
		rules: t.rules.map((r) => ({ ...r })),
		global: t.global,
		filterMode: t.filterMode,
		filterPlugins: [...t.filterPlugins]
	};
}

export const entityTypes: { value: Rule['entityType']; label: string }[] = [
	{ value: 'npc', label: 'NPC' },
	{ value: 'faction', label: 'Faction' },
	{ value: 'race', label: 'Race' }
];

export function entityLabel(type: string): string {
	return entityTypes.find((t) => t.value === type)?.label ?? type;
}

// entityId is how SkyrimNet keys a faction or race: its editor ID and the
// last three digits of its form ID, in lower case. An NPC is keyed by their
// bio template.
export function entityId(type: string, editorId: string, formId: string): string {
	if (type === 'npc') return editorId.trim().toLowerCase();
	const form = formId.trim().replace(/^0x/i, '');
	return `${editorId.trim()}_${form.slice(-3)}`.toLowerCase();
}

// blankRule is a rule for the form to fill in.
export function blankRule(): Rule {
	return { id: '', entityType: 'npc', entityId: '', name: '', pattern: '', enabled: true };
}

// An NPC a rule can name: by the name SkyrimNet recorded for them in
// this playthrough, or else the name their bio template spells.
export interface NPC {
	template: string;
	name: string;
	here: boolean; // in this playthrough
}

// npcChoices lists every NPC with a bio: those in this playthrough, then
// the rest of the catalog.
export function npcChoices(characters: Character[], catalog: Catalog): NPC[] {
	const seen = new Set<string>();
	const out: NPC[] = [];
	for (const c of characters) {
		if (!c.bio || seen.has(c.bio.toLowerCase())) continue;
		seen.add(c.bio.toLowerCase());
		out.push({ template: c.bio, name: c.name || templateName(c.bio), here: true });
	}
	for (const e of catalog.bios) {
		if (seen.has(e.template.toLowerCase())) continue;
		seen.add(e.template.toLowerCase());
		out.push({ template: e.template, name: templateName(e.template), here: false });
	}
	return out.sort((a, b) => a.name.localeCompare(b.name) || a.template.localeCompare(b.template));
}

// npcOf finds the NPC a rule is for, by the template it's keyed by.
export function npcOf(rule: Rule, npcs: NPC[]): NPC | undefined {
	return npcs.find((n) => n.template.toLowerCase() === rule.entityId);
}

// The plugins a filter can name: those the playthrough's characters come from.
export function pluginChoices(characters: Character[]): string[] {
	return [...new Set(characters.map((c) => c.modFile).filter(Boolean))].sort((a, b) => a.localeCompare(b));
}
