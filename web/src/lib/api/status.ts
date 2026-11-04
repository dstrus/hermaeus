import { request } from './client.ts';
import type { FolderRequest, Status } from './types/api.ts';
import type { Listing } from './types/install.ts';

export function getStatus(fetchFn?: typeof fetch): Promise<Status> {
	return request<Status>('GET', '/api/status', undefined, fetchFn);
}

export function setFolder(path: string, modPath?: string): Promise<Status> {
	return request<Status>('PUT', '/api/folder', { path, modPath } satisfies FolderRequest);
}

// browseFolders lists a folder's subfolders for the folder picker, marked
// for the kind of folder it's after. With no path, it lists the folder
// Hermaeus runs from.
export function browseFolders(kind: 'skyrimnet' | 'mod', path = ''): Promise<Listing> {
	return request<Listing>('GET', `/api/folders?for=${kind}&path=${encodeURIComponent(path)}`);
}
