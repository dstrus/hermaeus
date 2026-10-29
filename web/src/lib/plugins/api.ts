import { request } from '#lib/api/client.ts';
import type { List, Order, Plugin, Switch } from '#lib/api/types/plugins.ts';

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
	const name = folder.replace(/_/g, ' ');
	if (files !== 1) return `${files} ${name}`;
	return `1 ${name.endsWith('ies') ? `${name.slice(0, -3)}y` : name.endsWith('s') ? name.slice(0, -1) : name}`;
}

// contentsLine is what a plugin holds, by folder: "9 prompts · 1 filter".
export function contentsLine(p: Plugin): string {
	return p.contents.map((c) => countLine(c.folder, c.files)).join(' · ');
}
