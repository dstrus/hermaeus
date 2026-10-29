import { request } from '#lib/api/client.ts';
import type { Models, Provider, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';

export const modelsDependency = 'app:models';

export function getModels(fetchFn?: typeof fetch): Promise<Models> {
	return request<Models>('GET', '/api/models', undefined, fetchFn);
}

// saveTask writes one task's rotation and settings; the server keeps the
// rest of the file as it was.
export function saveTask(id: string, edit: TaskEdit): Promise<Models> {
	return request<Models>('PUT', `/api/models/${encodeURIComponent(id)}`, edit);
}

export function modelsHref(save: string, task?: string): string {
	const base = `/${encodeURIComponent(save)}/settings/models`;
	return task ? `${base}?item=${encodeURIComponent(task)}` : base;
}

export function providerName(providers: Provider[], id: string): string {
	return providers.find((p) => p.id === id)?.name ?? id;
}

// modelChoices lists every model named anywhere in the file, for the form.
export function modelChoices(models: Models): string[] {
	return [...new Set(models.tasks.flatMap((t) => t.slots.map((s) => s.model)))].sort((a, b) => a.localeCompare(b));
}

// rotation is a task's models as one line.
export function rotation(task: Task): string {
	return task.slots.map((s) => s.model).join(', ');
}

export function blankSlot(provider: string): Slot {
	return { model: '', provider, enabled: true };
}

export const reasoningEfforts = [
	{ value: '', label: 'Provider’s default' },
	{ value: 'off', label: 'Off' },
	{ value: 'minimal', label: 'Minimal' },
	{ value: 'low', label: 'Low' },
	{ value: 'medium', label: 'Medium' },
	{ value: 'high', label: 'High' }
];
