import { request } from '#lib/api/client.ts';
import type { ToggleList, ToggleRequest } from '#lib/api/types/settings.ts';

// What load functions depend on for the quick toggles; invalidate it after
// changing one elsewhere.
export const togglesDependency = 'app:toggles';

export function getToggles(fetchFn?: typeof fetch): Promise<ToggleList> {
	return request<ToggleList>('GET', '/api/toggles', undefined, fetchFn);
}

export function setToggle(id: string, on: boolean): Promise<ToggleList> {
	return request<ToggleList>('PUT', `/api/toggles/${encodeURIComponent(id)}`, { on } satisfies ToggleRequest);
}
