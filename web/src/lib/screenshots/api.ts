import { playthroughPath, request } from '#lib/api/client.ts';
import type { Changes, List } from '#lib/api/types/screenshots.ts';

export function getScreenshots(save: string, fetchFn?: typeof fetch): Promise<List> {
	return request<List>('GET', playthroughPath(save, 'screenshots'), undefined, fetchFn);
}

export function saveScreenshot(save: string, id: number, c: Changes): Promise<List> {
	return request<List>('PATCH', playthroughPath(save, 'screenshots', String(id)), c);
}

// imagePath is the API path for a screenshot's image, or its thumbnail.
export function imagePath(save: string, id: number, thumb = false): string {
	return playthroughPath(save, 'screenshots', String(id), 'image') + (thumb ? '?size=thumb' : '');
}
