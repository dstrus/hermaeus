import { playthroughPath, request } from '#lib/api/client.ts';
import type { BioResponse, Edit, HistoryResponse, Restore, Version } from '#lib/api/types/bios.ts';

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
