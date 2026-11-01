import { request } from '#lib/api/client.ts';
import type { List, Order, Plugin, Switch } from '#lib/api/types/plugins.ts';
import { m } from '#lib/copy/messages.ts';

export const pluginsDependency = 'app:plugins';

export function getPlugins(fetchFn?: typeof fetch): Promise<List> {
	return request<List>('GET', '/api/plugins', undefined, fetchFn);
}

// setOrder puts the installed plugins in order, highest priority first.
export function setOrder(order: Order): Promise<List> {
	return request<List>('PUT', '/api/plugins/order', order);
}

export function setEnabled(id: string, s: Switch): Promise<List> {
	return request<List>('PUT', `/api/plugins/${encodeURIComponent(id)}/enabled`, s);
}

// priorityOf is the installed plugins as the registry keeps them, highest
// priority first, from the list as shown, lowest first.
export function priorityOf(shown: Plugin[]): string[] {
	return shown.filter((p) => p.kind === 'plugin').map((p) => p.id).reverse();
}

// countLine is one folder's count: "9 prompts", "1 filter".
export function countLine(folder: string, files: number): string {
	const key = folder.replace(/_(.)/g, (_, c: string) => c.toUpperCase());
	const line = (m.plugins.contents as Record<string, ((p: { count: number }) => string) | undefined>)[key];
	return line ? line({ count: files }) : m.plugins.otherFolder({ count: files, folder: folder.replace(/_/g, ' ') });
}

// contentsLine is what a plugin holds, by folder: "9 prompts · 1 filter".
export function contentsLine(p: Plugin): string {
	return p.contents.map((c) => countLine(c.folder, c.files)).join(' · ');
}
