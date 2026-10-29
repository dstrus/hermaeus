// Providers' model lists, fetched over the Internet when asked and kept for
// the session, so a form can offer a provider's models by name.
import { ApiError, request } from '#lib/api/client.ts';
import type { ModelInfo, ModelList } from '#lib/api/types/settings.ts';

class ModelLists {
	lists = $state<Record<string, ModelList>>({});
	loading = $state<Record<string, boolean>>({});
	errors = $state<Record<string, string>>({});

	// fetch asks the provider, or asks again when refresh is set.
	async fetch(provider: string, refresh = false) {
		this.loading[provider] = true;
		this.errors[provider] = '';
		try {
			this.lists[provider] = await request<ModelList>('GET', `/api/models/providers/${encodeURIComponent(provider)}${refresh ? '?refresh' : ''}`);
		} catch (e) {
			this.errors[provider] = e instanceof ApiError ? e.message : String(e);
		} finally {
			this.loading[provider] = false;
		}
	}

	find(provider: string, model: string): ModelInfo | undefined {
		return this.lists[provider]?.models.find((m) => m.id === model);
	}
}

export const modelLists = new ModelLists();

// describe is a model's context and prices as one line, when known.
export function describe(m: ModelInfo): string {
	const parts: string[] = [];
	if (m.name && m.name !== m.id) parts.push(m.name);
	if (m.context) parts.push(`${tokens(m.context)} context`);
	if (m.prompt !== undefined && m.completion !== undefined) {
		parts.push(m.prompt === 0 && m.completion === 0 ? 'free' : `$${price(m.prompt)} in, $${price(m.completion)} out per million tokens`);
	}
	return parts.join(' · ');
}

// tokens reads 128000 as 128k and 1048576 as 1M.
function tokens(n: number): string {
	if (n >= 1e6) return `${(n / 1e6).toFixed(n % 1e6 > 5e4 ? 1 : 0)}M`;
	if (n >= 1000) return `${Math.round(n / 1000)}k`;
	return String(n);
}

function price(n: number): string {
	return n >= 10 ? n.toFixed(0) : n >= 1 ? n.toFixed(2).replace(/0$/, '') : n.toFixed(3).replace(/0+$/, '');
}
