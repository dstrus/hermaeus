// The feature pages: settings files shown in their feature's own terms,
// under their sidebar group (#lib/nav.ts). Adding one is a definition here,
// a page in its group, and an entry in the Go side's own-page list, so the
// Other page leaves its file out.
import { actorFilter } from './actorfilter.ts';
import { diary } from './diary.ts';
import { gameMaster } from './gamemaster.ts';
import { narration } from './narration.ts';
import { hotkeys } from './hotkeys.ts';
import { dynamicBio } from './dynamicbio.ts';
import { memory } from './memory.ts';
import { memoryFilter } from './memoryfilter.ts';
import { speech } from './voice.ts';
import type { Feature } from './schema.ts';
import { href } from '#lib/nav.ts';

export const features: Feature[] = [gameMaster, narration, actorFilter, memory, memoryFilter, diary, dynamicBio, speech, hotkeys];

export function featureById(id: string): Feature | undefined {
	return features.find((f) => f.id === id);
}

export function featureHref(save: string, f: Feature, key?: string): string {
	const base = href(save, f.id);
	return key ? `${base}?item=${encodeURIComponent(key)}` : base;
}
