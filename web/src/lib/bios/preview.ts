// A glimpse of a character's bio for a pane: the summary section of the
// bio in use, fetched when the character is selected and kept for as long
// as the page is open.
import { getBio, kinds } from './api.ts';

export interface Preview {
	kind: string; // which bio it is, as kinds names it
	text: string;
}

export class BioPreviews {
	private cache = new Map<string, Promise<Preview | null>>();

	constructor(private save: () => string) {} // read when fetching, so a changed playthrough counts

	// of resolves to the preview for a template, or null when there's no bio.
	of(template: string): Promise<Preview | null> {
		let p = this.cache.get(template);
		if (!p) {
			p = getBio(this.save(), template)
				.then(({ versions }) => {
					const v = versions[0];
					if (!v) return null;
					const section = v.sections.find((s) => s.name === 'summary') ?? v.sections[0];
					return section ? { kind: kinds[v.kind]?.label.toLowerCase() ?? v.kind, text: section.text } : null;
				})
				.catch(() => null);
			this.cache.set(template, p);
		}
		return p;
	}
}
