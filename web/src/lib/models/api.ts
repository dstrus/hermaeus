import { request } from '#lib/api/client.ts';
import { m } from '#lib/copy/messages.ts';
import { href } from '#lib/nav.ts';
import type { Models, ModelTest, ModelTestResult, Provider, Slot, Task, TaskEdit } from '#lib/api/types/settings.ts';

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
	const base = href(save, 'models');
	return task ? `${base}?item=${encodeURIComponent(task)}` : base;
}

// copyKey is a task's id as en.yaml keys it: gamemaster_evaluation as
// gamemasterEvaluation, UniversalTranslator as universalTranslator.
const copyKey = (id: string) => (id.charAt(0).toLowerCase() + id.slice(1)).replace(/_(.)/g, (_, c: string) => c.toUpperCase());

function ownTask(id: string): { label: string; what: string } | undefined {
	const tasks: Record<string, { label: string; what: string }> = m.models.tasks;
	const key = copyKey(id);
	return Object.hasOwn(tasks, key) ? tasks[key] : undefined;
}

// taskLabel names a task: one of SkyrimNet's own by its name here, and a
// plugin's from its id, so sever_background is Sever background.
export function taskLabel(id: string): string {
	const own = ownTask(id);
	if (own) return own.label;
	const words = id.replace(/_/g, ' ');
	return words.charAt(0).toUpperCase() + words.slice(1);
}

// taskWhat says what one of SkyrimNet's own tasks does; a plugin's says nothing.
export function taskWhat(id: string): string {
	return ownTask(id)?.what ?? '';
}

// sample is the prompt a task's test starts from: shaped like its job for
// SkyrimNet's own tasks, and a check for a reply for a plugin's.
export function sample(id: string): [system: string, prompt: string] {
	const samples: Record<string, unknown> = m.models.samples;
	const s = samples[copyKey(id)];
	if (Object.hasOwn(samples, copyKey(id)) && typeof s === 'object' && s !== null) {
		const { system, prompt } = s as { system: string; prompt: string };
		return [system, prompt];
	}
	return ['', m.models.samples.pluginTask];
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

const efforts = m.models.efforts;
export const reasoningEfforts = [
	{ value: '', label: efforts.inherited },
	{ value: 'off', label: efforts.off },
	{ value: 'default', label: efforts.default },
	{ value: 'minimal', label: efforts.minimal },
	{ value: 'low', label: efforts.low },
	{ value: 'medium', label: efforts.medium },
	{ value: 'high', label: efforts.high }
];

// testModel sends one prompt to one of a task's models, through the
// provider SkyrimNet would use, and returns the reply and what it cost.
export function testModel(task: string, t: ModelTest): Promise<ModelTestResult> {
	return request<ModelTestResult>('POST', `/api/models/${encodeURIComponent(task)}/test`, t);
}
