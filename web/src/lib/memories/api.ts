import { playthroughPath, request } from '#lib/api/client.ts';
import type { DiaryResponse, MemoriesResponse } from '#lib/api/types/memories.ts';

export function getMemories(save: string, uuid: string, fetchFn?: typeof fetch): Promise<MemoriesResponse> {
	return request<MemoriesResponse>('GET', playthroughPath(save, 'characters', uuid, 'memories'), undefined, fetchFn);
}

export function getDiary(save: string, uuid: string, fetchFn?: typeof fetch): Promise<DiaryResponse> {
	return request<DiaryResponse>('GET', playthroughPath(save, 'characters', uuid, 'diary'), undefined, fetchFn);
}
