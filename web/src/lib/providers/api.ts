import { request } from '#lib/api/client.ts';
import { href } from '#lib/nav.ts';
import type { KeyEdit, ProviderDetails, ProviderEdit, ProvidersResponse } from '#lib/api/types/settings.ts';

export const providersDependency = 'app:providers';

export function getProviders(fetchFn?: typeof fetch): Promise<ProvidersResponse> {
	return request<ProvidersResponse>('GET', '/api/providers', undefined, fetchFn);
}

export function saveProvider(id: string, edit: ProviderEdit): Promise<ProvidersResponse> {
	return request<ProvidersResponse>('PUT', `/api/providers/${encodeURIComponent(id)}`, edit);
}

export function addProvider(edit: ProviderEdit): Promise<ProvidersResponse> {
	return request<ProvidersResponse>('POST', '/api/providers', edit);
}

export function removeProvider(id: string, modified: number): Promise<ProvidersResponse> {
	return request<ProvidersResponse>('DELETE', `/api/providers/${encodeURIComponent(id)}`, { modified, key: '' } satisfies KeyEdit);
}

// setKey sets the key, or clears it when empty. The key goes to Hermaeus
// and into Providers.yaml; nothing ever sends it back.
export function setKey(id: string, edit: KeyEdit): Promise<ProvidersResponse> {
	return request<ProvidersResponse>('PUT', `/api/providers/${encodeURIComponent(id)}/key`, edit);
}

export function providersHref(save: string, id?: string): string {
	const base = href(save, 'providers');
	return id ? `${base}?item=${encodeURIComponent(id)}` : base;
}

export const apiStyles = [
	{ value: 'chat', label: 'Chat', what: 'OpenAI-style chat completions: what every task but Decisions uses.' },
	{ value: 'decisions', label: 'Decisions', what: 'OpenRouter’s decisions API, for the Decisions task only.' },
	{ value: 'systemone', label: 'System One', what: 'A System One decision endpoint, for the Decisions task only.' }
];

export function styleLabel(style: string): string {
	return apiStyles.find((s) => s.value === style)?.label ?? style;
}

export const sortings = [
	{ value: 'default', label: 'OpenRouter’s default' },
	{ value: 'price', label: 'Cheapest first' },
	{ value: 'throughput', label: 'Fastest throughput first' },
	{ value: 'latency', label: 'Lowest latency first' }
];

// keyState says whether a provider can be reached: its own key, a borrowed
// one, or none.
export function keyState(p: ProviderDetails, all: ProviderDetails[]): 'own' | 'borrowed' | 'none' {
	if (p.hasKey) return 'own';
	if (p.keyFrom && all.find((o) => o.id === p.keyFrom)?.hasKey) return 'borrowed';
	return 'none';
}

export function editOf(p: ProviderDetails, modified: number): ProviderEdit {
	return {
		modified,
		name: p.name,
		endpoint: p.endpoint,
		style: p.style,
		streaming: p.streaming,
		structuredOutputs: p.structuredOutputs,
		reasoning: p.reasoning,
		settings: p.settings,
		sorting: p.sorting,
		cacheTTL: p.cacheTTL
	};
}

export function blankProvider(modified: number): ProviderEdit {
	return { modified, name: '', endpoint: '', style: 'chat', streaming: true, structuredOutputs: false, reasoning: false, settings: '', sorting: '', cacheTTL: '' };
}
