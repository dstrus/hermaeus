import { request } from '#lib/api/client.ts';
import { m } from '#lib/copy/messages.ts';
import { href } from '#lib/nav.ts';
import type { Layer } from '#lib/api/types/layers.ts';
import type { Document, Edit, Matches, PinsResponse, PromptList, Revert, Saved } from '#lib/api/types/prompts.ts';

// What load functions depend on for the prompt list and an open prompt;
// invalidate them after saving or removing a copy.
export const promptsDependency = 'app:prompts';
export const promptDependency = 'app:prompt';

export function getPrompts(fetchFn?: typeof fetch): Promise<PromptList> {
	return request<PromptList>('GET', '/api/prompts', undefined, fetchFn);
}

// searchPrompts finds the prompts whose text contains q.
export function searchPrompts(q: string): Promise<Matches> {
	return request<Matches>('GET', `/api/prompts/search?q=${encodeURIComponent(q)}`);
}

// filePath is the API path for a prompt, whose own path has slashes.
function filePath(path: string): string {
	return '/api/prompts/file/' + path.split('/').map(encodeURIComponent).join('/');
}

export function getPrompt(path: string, fetchFn?: typeof fetch): Promise<Document> {
	return request<Document>('GET', filePath(path), undefined, fetchFn);
}

export function savePrompt(path: string, edit: Edit): Promise<Saved> {
	return request<Saved>('PUT', filePath(path), edit);
}

export function removeCopy(path: string, r: Revert): Promise<Document> {
	return request<Document>('DELETE', filePath(path), r);
}

// setPromptPinned pins or unpins a prompt, and returns the pins as they are
// afterward.
export function setPromptPinned(path: string, pinned: boolean): Promise<PinsResponse> {
	return request<PinsResponse>(pinned ? 'PUT' : 'DELETE', '/api/prompts/pins/' + path.split('/').map(encodeURIComponent).join('/'));
}

// promptHref is the page for a prompt in a playthrough.
export function promptHref(save: string, path: string): string {
	return href(save, 'prompts', '/' + path.split('/').map(encodeURIComponent).join('/'));
}

// promptName is a prompt's file name without its extension.
export function promptName(path: string): string {
	return path.slice(path.lastIndexOf('/') + 1).replace(/\.prompt$/i, '');
}

// promptFolder is the folder a prompt sits in under prompts/, or "" at the top.
export function promptFolder(path: string): string {
	const i = path.lastIndexOf('/');
	return i < 0 ? '' : path.slice(0, i);
}

// layerLabel is a layer as the versions switch names it.
export function layerLabel(l: Layer): string {
	if (l.kind === 'overlay') return m.prompts.yours;
	if (l.kind === 'base') return m.prompts.default;
	return l.source;
}
