import { playthroughPath, request } from '#lib/api/client.ts';
import type { BioResponse, Catalog, Edit, HistoryResponse, Restore, SaveBios, Version } from '#lib/api/types/bios.ts';

// What load functions depend on for a character's bio; invalidate it after
// changing one.
export const bioDependency = 'app:bio';

export function getBio(save: string, template: string, fetchFn?: typeof fetch): Promise<BioResponse> {
	return request<BioResponse>('GET', playthroughPath(save, 'bios', template), undefined, fetchFn);
}

export function saveBio(save: string, template: string, kind: string, edit: Edit): Promise<Version> {
	return request<Version>('PATCH', playthroughPath(save, 'bios', template, kind), edit);
}

// createBio gives the character a baseline bio in this playthrough, copied
// from the bio they use now.
export function createBio(save: string, template: string): Promise<Version> {
	return request<Version>('POST', playthroughPath(save, 'bios', template));
}

export function getHistory(save: string, template: string, kind: string): Promise<HistoryResponse> {
	return request<HistoryResponse>('GET', playthroughPath(save, 'bios', template, kind, 'history'));
}

export function restoreBio(save: string, template: string, kind: string, r: Restore): Promise<Version> {
	return request<Version>('POST', playthroughPath(save, 'bios', template, kind, 'restore'), r);
}

// kindLabel is what each kind of bio is called, and what it's for.
export const kinds: Record<string, { label: string; what: string }> = {
	dynamic: { label: 'Dynamic', what: 'The bio NPCs use now. SkyrimNet keeps rewriting it as you play.' },
	baseline: { label: 'Baseline', what: 'This playthrough’s fixed bio, which SkyrimNet starts from when it rewrites the dynamic one.' },
	default: { label: 'Default', what: 'The bio from SkyrimNet’s content, used until this playthrough has its own.' }
};

// sectionLabel turns a section name like interject_summary into words.
export function sectionLabel(name: string): string {
	return name.replace(/_/g, ' ');
}

// getCatalog lists every character SkyrimNet or a plugin ships a bio for.
export function getCatalog(fetchFn?: typeof fetch): Promise<Catalog> {
	return request<Catalog>('GET', '/api/bios', undefined, fetchFn);
}

// getSaveBios says which kinds of save bio each template has in a playthrough.
export function getSaveBios(save: string, fetchFn?: typeof fetch): Promise<SaveBios> {
	return request<SaveBios>('GET', playthroughPath(save, 'bios'), undefined, fetchFn);
}

// templateName turns a bio template such as laurina_cedalius_819 or
// bandit_generic into "Laurina Cedalius" or "Bandit (generic)".
export function templateName(template: string): string {
	const m = template.match(/^(.*?)_(?:([0-9A-Fa-f]{1,6})|(generic|special))$/);
	const words = (m ? m[1]! : template).replace(/_/g, ' ').replace(/(^|\s)\S/g, (c) => c.toUpperCase());
	return m?.[3] ? `${words} (${m[3]})` : words;
}

export function catalogHref(save: string, template?: string): string {
	const base = `/${encodeURIComponent(save)}/bios`;
	return template ? `${base}/${encodeURIComponent(template)}` : base;
}
