// The feature pages: settings files shown in their feature's own terms,
// under settings/<id>. Adding one is a definition here and an entry in
// the Go side's own-page list, so the Settings page points to it.
import { memory } from './memory.ts';
import type { Feature } from './schema.ts';

export const features: Feature[] = [memory];

export function featureById(id: string): Feature | undefined {
	return features.find((f) => f.id === id);
}

export function featureHref(save: string, f: Feature, key?: string): string {
	const base = `/${encodeURIComponent(save)}/settings/${f.id}`;
	return key ? `${base}?item=${encodeURIComponent(key)}` : base;
}
