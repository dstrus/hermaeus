import { playthroughPath, request } from '#lib/api/client.ts';
import type { BulkChanges, Count, DeleteDiary, DeletedDiary, DiaryChanges, DiaryResponse, KnowledgeResponse, MemoriesResponse, MemoryChanges, NewEntry, PackChanges, SaveResult } from '#lib/api/types/memories.ts';

// What load functions depend on for memories and diary entries; invalidate
// them after changing either.
export const memoriesDependency = 'app:memories';
export const diaryDependency = 'app:diary';

// The memory types SkyrimNet uses. A memory can have another, which is kept.
export const memoryTypes = ['EXPERIENCE', 'RELATIONSHIP', 'KNOWLEDGE', 'JOY', 'TRAUMA', 'LOCATION', 'SKILL'];

// Emotions SkyrimNet often gives memories, offered as suggestions.
export const commonEmotions = [
	'neutral', 'content', 'joyful', 'concerned', 'angry', 'disgusted', 'surprised',
	'determined', 'troubled', 'excited', 'fearful', 'grateful', 'hopeful', 'reflective'
];

export function getMemories(save: string, uuid: string, fetchFn?: typeof fetch): Promise<MemoriesResponse> {
	return request<MemoriesResponse>('GET', playthroughPath(save, 'characters', uuid, 'memories'), undefined, fetchFn);
}

export function addMemory(save: string, uuid: string, m: MemoryChanges): Promise<SaveResult> {
	return request<SaveResult>('POST', playthroughPath(save, 'characters', uuid, 'memories'), m);
}

export function updateMemory(save: string, id: number, c: MemoryChanges): Promise<SaveResult> {
	return request<SaveResult>('PATCH', playthroughPath(save, 'memories', String(id)), c);
}

export function updateMemories(save: string, b: BulkChanges): Promise<Count> {
	return request<Count>('POST', playthroughPath(save, 'memories', 'update'), b);
}

export function deleteMemories(save: string, ids: number[]): Promise<Count> {
	return request<Count>('POST', playthroughPath(save, 'memories', 'delete'), { ids });
}

export function moveToCurrentIdentity(save: string, ids: number[]): Promise<Count> {
	return request<Count>('POST', playthroughPath(save, 'memories', 'move'), { ids });
}

export function getDiary(save: string, uuid: string, fetchFn?: typeof fetch): Promise<DiaryResponse> {
	return request<DiaryResponse>('GET', playthroughPath(save, 'characters', uuid, 'diary'), undefined, fetchFn);
}

export function addDiaryEntry(save: string, uuid: string, c: DiaryChanges): Promise<SaveResult> {
	return request<SaveResult>('POST', playthroughPath(save, 'characters', uuid, 'diary'), c);
}

export function updateDiary(save: string, id: number, c: DiaryChanges): Promise<SaveResult> {
	return request<SaveResult>('PATCH', playthroughPath(save, 'diary', String(id)), c);
}

export function deleteDiaryEntries(save: string, d: DeleteDiary): Promise<DeletedDiary> {
	return request<DeletedDiary>('POST', playthroughPath(save, 'diary', 'delete'), d);
}

// savedNote says what a save did besides the change itself, for the user.
export function savedNote(r: SaveResult, what: string): string {
	const notes = [`${what} saved.`];
	if (r.diary !== undefined) notes.push(`Diary entry ${r.diary} was updated to match.`);
	if (r.memory !== undefined) notes.push(`Its memory copy was updated to match.`);
	if (r.reembed) notes.push('SkyrimNet makes a new search vector in the background the next time a save is loaded.');
	return notes.join(' ');
}

// parseTags reads a comma-separated list of tags.
export function parseTags(s: string): string[] {
	return s.split(',').map((t) => t.trim()).filter(Boolean);
}

// Everyone's memories and diary entries in the playthrough, for History.
export function getAllMemories(save: string, fetchFn?: typeof fetch): Promise<MemoriesResponse> {
	return request<MemoriesResponse>('GET', playthroughPath(save, 'memories'), undefined, fetchFn);
}

export function getAllDiary(save: string, fetchFn?: typeof fetch): Promise<DiaryResponse> {
	return request<DiaryResponse>('GET', playthroughPath(save, 'diary'), undefined, fetchFn);
}

// World knowledge: packs and the entries any NPC can draw on.
export const knowledgeDependency = 'app:knowledge';

export function getKnowledge(save: string, fetchFn?: typeof fetch): Promise<KnowledgeResponse> {
	return request<KnowledgeResponse>('GET', playthroughPath(save, 'knowledge'), undefined, fetchFn);
}

export function addKnowledge(save: string, entry: NewEntry): Promise<SaveResult> {
	return request<SaveResult>('POST', playthroughPath(save, 'knowledge'), entry);
}

// setPackEnabled switches a pack on or off, and returns the knowledge as it is afterward.
export function setPackEnabled(save: string, id: number, enabled: boolean): Promise<KnowledgeResponse> {
	return request<KnowledgeResponse>('PUT', playthroughPath(save, 'knowledge', 'packs', String(id)), { enabled } satisfies PackChanges);
}
