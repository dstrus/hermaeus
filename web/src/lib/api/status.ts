import { request } from './client.ts';
import type { FolderRequest, Status } from './types/api.ts';

export function getStatus(fetchFn?: typeof fetch): Promise<Status> {
	return request<Status>('GET', '/api/status', undefined, fetchFn);
}

export function setFolder(path: string, modPath?: string): Promise<Status> {
	return request<Status>('PUT', '/api/folder', { path, modPath } satisfies FolderRequest);
}
