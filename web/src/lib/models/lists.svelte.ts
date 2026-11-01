// Providers' model lists, fetched over the Internet when asked and kept for
// the session, so a form can offer a provider's models by name.
import { ApiError, request } from '#lib/api/client.ts';
import type { ModelInfo, ModelList } from '#lib/api/types/settings.ts';
import { m } from '#lib/copy/messages.ts';

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
		return this.lists[provider]?.models.find((x) => x.id === model);
	}
}

export const modelLists = new ModelLists();

// describe is a model's context and prices as one line, when known.
export function describe(info: ModelInfo): string {
	const t = m.models.lists;
	const parts: string[] = [];
	if (info.name && info.name !== info.id) parts.push(info.name);
	if (info.context) parts.push(t.context({ tokens: tokens(info.context) }));
	if (info.prompt !== undefined && info.completion !== undefined) {
		parts.push(info.prompt === 0 && info.completion === 0 ? t.free : t.prices({ input: price(info.prompt), output: price(info.completion) }));
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
