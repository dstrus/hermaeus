// What changed between two versions of a bio, section by section and word
// by word, for reading what SkyrimNet's rewrite did or what an earlier
// version had.
import type { Section } from '#lib/api/types/bios.ts';

export interface Op {
	kind: 'same' | 'added' | 'removed';
	text: string;
}

export interface SectionDiff {
	name: string;
	// rewritten: so little survives that marking words would be noise; the
	// section reads as the old text, then the new.
	status: 'same' | 'changed' | 'rewritten' | 'added' | 'removed';
	ops: Op[];
}

// Words and the whitespace between them, each its own token, so a diff
// keeps the original spacing.
function tokens(s: string): string[] {
	return s.split(/(\s+)/).filter((t) => t !== '');
}

// Past this many tokens a side, the quadratic comparison would stall, and
// such a section is shown as wholly replaced.
const most = 2500;

// diffWords is the edit from a to b, as runs of kept, added, and removed text.
export function diffWords(a: string, b: string): Op[] {
	if (a === b) return a ? [{ kind: 'same', text: a }] : [];
	const x = tokens(a);
	const y = tokens(b);
	if (x.length > most || y.length > most) {
		return [...(a ? [{ kind: 'removed', text: a } as Op] : []), ...(b ? [{ kind: 'added', text: b } as Op] : [])];
	}
	// Longest common subsequence, then walk it back.
	const w = y.length + 1;
	const L = new Uint16Array((x.length + 1) * w);
	for (let i = x.length - 1; i >= 0; i--) {
		for (let j = y.length - 1; j >= 0; j--) {
			L[i * w + j] = x[i] === y[j] ? L[(i + 1) * w + j + 1]! + 1 : Math.max(L[(i + 1) * w + j]!, L[i * w + j + 1]!);
		}
	}
	const out: Op[] = [];
	const push = (kind: Op['kind'], text: string) => {
		const last = out[out.length - 1];
		if (last && last.kind === kind) last.text += text;
		else out.push({ kind, text });
	};
	let i = 0;
	let j = 0;
	while (i < x.length && j < y.length) {
		if (x[i] === y[j]) {
			push('same', x[i]!);
			i++;
			j++;
		} else if (L[(i + 1) * w + j]! >= L[i * w + j + 1]!) {
			push('removed', x[i]!);
			i++;
		} else {
			push('added', y[j]!);
			j++;
		}
	}
	for (; i < x.length; i++) push('removed', x[i]!);
	for (; j < y.length; j++) push('added', y[j]!);
	return out;
}

// diffSections compares two bios section by section: the later one's
// sections in order, then any the earlier had that are gone.
export function diffSections(from: Section[], to: Section[]): SectionDiff[] {
	const was = new Map(from.map((s) => [s.name, s.text]));
	const out: SectionDiff[] = to.map((s) => {
		const before = was.get(s.name);
		if (before === undefined) return { name: s.name, status: 'added', ops: [{ kind: 'added', text: s.text }] };
		const ops = diffWords(before, s.text);
		if (ops.every((o) => o.kind === 'same')) return { name: s.name, status: 'same', ops };
		const kept = ops.filter((o) => o.kind === 'same').reduce((n, o) => n + o.text.trim().length, 0);
		if (kept < Math.max(before.length, s.text.length) / 2) {
			return { name: s.name, status: 'rewritten', ops: [{ kind: 'removed', text: before }, { kind: 'added', text: s.text }] };
		}
		return { name: s.name, status: 'changed', ops };
	});
	const kept = new Set(to.map((s) => s.name));
	for (const s of from) if (!kept.has(s.name)) out.push({ name: s.name, status: 'removed', ops: [{ kind: 'removed', text: s.text }] });
	return out;
}

const blockRe = /\{%-?\s*block\s+(\w+)\s*-?%\}([\s\S]*?)\{%-?\s*endblock\s*-?%\}/g;

// parseSections splits a bio's text into its sections, as the server does.
export function parseSections(raw: string): Section[] {
	return [...raw.matchAll(blockRe)].map((m) => ({ name: m[1]!, text: m[2]!.trim() }));
}
