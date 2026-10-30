// A feature page is a settings file shown in the feature's own terms: its
// settings in named sections, each with a label, a sentence on what it
// does, and the kind of value it takes. The page reads and writes the
// file through the ordinary settings API, one setting at a time.
import type { Setting } from '#lib/api/types/settings.ts';
import { keyName } from '#lib/settings/keys.ts';

export type Kind =
	| 'switch' // true or false
	| 'count' // a whole number
	| 'number' // any number
	| 'percent' // stored 0–1, shown as a percentage
	| 'minutes' // a whole number of minutes
	| 'text'
	| 'choice' // one of a few values, or any text when custom is set
	| 'set' // any number of the choices, kept as a list
	| 'hotkey' // a key code, -1 for none
	| 'secret' // text that's written but never shown, like an API key
	| 'lines' // a list of words or short phrases, one per line
	| 'pairs' // a list of two-part entries, such as a word and how to say it
	| 'map'; // the settings under the key, each a name and a value, edited together

export interface Choice {
	value: string;
	label: string;
	what?: string;
}

export interface Field {
	key: string; // the setting's dotted key in the file
	label: string;
	what: string; // what it does, as a sentence
	kind: Kind;
	min?: number;
	max?: number;
	step?: number;
	choices?: Choice[];
	custom?: boolean; // a choice may also be typed in
	unit?: string; // after a count, such as "events"
	placeholder?: string; // an example, for a text field
	pluginHotkey?: boolean; // a plugin's hotkey: unset is empty or 0, and it may be stored as text
	columns?: [string, string]; // pairs and map: the two parts' names, as headings
	pairKeys?: [string, string]; // pairs: each entry's two keys in the file
	noun?: [string, string]; // lines, pairs, and map: one entry, and many, for counts
	zero?: string; // a count or number of 0 shown as what it means, such as "No limit"
}

export interface Section {
	title: string;
	what?: string;
	fields: Field[];
}

export interface Feature {
	id: string; // the page's path under settings/
	file: string; // the settings file's name, as the settings API knows it
	title: string;
	what: string; // the page's lead
	sections: Section[];
	lists?: 'memory' | 'actor'; // the filter lists of this kind, shown below the settings
}

// valueOf is a setting's value as the file has it, or its default.
export function valueOf(s: Setting | undefined): unknown {
	if (!s) return undefined;
	return s.set ? s.value : s.default;
}

// show formats a value for a field.
export function show(f: Field, v: unknown): string {
	if (f.kind === 'hotkey') {
		const n = hotkeyCode(v);
		return n <= 0 ? keyName(-1) : keyName(n);
	}
	if (v === undefined || v === null) return '—';
	if (f.zero && v === 0) return f.zero;
	switch (f.kind) {
		case 'switch':
			return v === true ? 'On' : 'Off';
		case 'percent':
			return typeof v === 'number' ? `${Number((v * 100).toFixed(1))}%` : String(v);
		case 'minutes':
			return typeof v === 'number' ? `${v} min` : String(v);
		case 'count': {
			// A unit is written plural; one of it is singular ("1 line").
			const unit = f.unit && v === 1 && /[^s]s$/.test(f.unit) ? f.unit.slice(0, -1) : f.unit;
			return typeof v === 'number' ? `${v}${unit ? ` ${unit}` : ''}` : String(v);
		}
		case 'choice': {
			const c = f.choices?.find((c) => c.value === String(v));
			return c ? c.label : String(v);
		}
		case 'secret':
			return typeof v === 'object' && v !== null && 'filled' in v ? ((v as { filled: boolean }).filled ? 'Set' : 'Not set') : String(v) ? 'Set' : 'Not set';
		case 'set':
			return Array.isArray(v) ? `${v.length} of ${f.choices?.length ?? v.length}` : String(v);
		case 'lines':
		case 'pairs': {
			const n = Array.isArray(v) ? v.length : 0;
			return n === 0 ? 'None' : `${n} ${n === 1 ? (f.noun?.[0] ?? 'entry') : (f.noun?.[1] ?? 'entries')}`;
		}
		case 'map': {
			const n = v && typeof v === 'object' ? Object.keys(v).length : 0;
			return n === 0 ? 'None' : `${n} ${n === 1 ? (f.noun?.[0] ?? 'entry') : (f.noun?.[1] ?? 'entries')}`;
		}
		case 'text':
			return String(v).trim() === '' ? 'None' : String(v);
		default:
			return String(v);
	}
}

// Pair is one entry of a pairs field, as the form edits it.
export interface Pair {
	a: string;
	b: string;
}

// pairsOf reads a pairs field's list, as the file has it, into the form's rows.
export function pairsOf(f: Field, v: unknown): Pair[] {
	const [ka, kb] = f.pairKeys ?? ['a', 'b'];
	if (!Array.isArray(v)) return [];
	return v.map((e) => (e && typeof e === 'object' ? { a: String((e as Record<string, unknown>)[ka] ?? ''), b: String((e as Record<string, unknown>)[kb] ?? '') } : { a: String(e), b: '' }));
}

// pairsValue is the form's rows as the file keeps them. A row with nothing
// in it is dropped; one with only its second part is a mistake.
export function pairsValue(f: Field, rows: Pair[]): Record<string, string>[] {
	const [ka, kb] = f.pairKeys ?? ['a', 'b'];
	const kept = rows.filter((r) => r.a !== '' || r.b !== '');
	const missing = kept.find((r) => r.a.trim() === '');
	if (missing) throw new Error(`Each entry needs its ${(f.columns?.[0] ?? 'first part').toLowerCase()}.`);
	return kept.map((r) => ({ [ka]: r.a, [kb]: r.b }));
}

// linesOf and linesValue convert a lines field between the file's list and
// the form's text, one entry per line; blank lines are dropped.
export function linesOf(v: unknown): string {
	return Array.isArray(v) ? v.map(String).join('\n') : '';
}

export function linesValue(draft: string): string[] {
	return draft
		.split('\n')
		.map((l) => l.trim())
		.filter(Boolean);
}

// pickedOf is a set's value as the list of choices it holds.
export function pickedOf(f: Field, v: unknown): string[] {
	const known = new Set((f.choices ?? []).map((c) => c.value));
	return Array.isArray(v) ? v.map(String).filter((x) => known.has(x)) : [];
}

// setValue is a set as it should be written: the values already in the
// file keep their order, and newly picked ones follow in the choices' order.
export function setValue(f: Field, before: unknown, picked: string[]): string[] {
	const keep = pickedOf(f, before).filter((x) => picked.includes(x));
	const added = (f.choices ?? []).map((c) => c.value).filter((x) => picked.includes(x) && !keep.includes(x));
	return [...keep, ...added];
}

// hotkeyCode reads a key code stored as a number or as text, as plugins do;
// -1 when there's none.
export function hotkeyCode(v: unknown): number {
	const n = typeof v === 'string' ? (v.trim() === '' ? -1 : Number(v)) : typeof v === 'number' ? v : -1;
	return Number.isNaN(n) ? -1 : n;
}

// draftOf is a value as the form holds it: text, percentages as whole numbers.
export function draftOf(f: Field, v: unknown): string {
	if (f.kind === 'secret') return ''; // never shown, only replaced
	if (f.kind === 'hotkey') {
		const n = hotkeyCode(v);
		return String(n <= 0 ? -1 : n);
	}
	if (v === undefined || v === null) return '';
	if (f.kind === 'percent' && typeof v === 'number') return String(Number((v * 100).toFixed(2)));
	return String(v);
}

// parse reads the form's text as the value the setting takes, or throws
// with what's wrong.
export function parse(f: Field, draft: string): unknown {
	const t = draft.trim();
	switch (f.kind) {
		case 'switch':
			return t === 'true';
		case 'text':
			return draft;
		case 'choice':
			if (!t) throw new Error(`Choose a value for ${f.label}.`);
			return t;
		case 'hotkey':
			return number(t, f.label);
		case 'secret':
			return t;
		case 'percent': {
			const n = number(t, f.label);
			if (n < 0 || n > 100) throw new Error(`${f.label} is a percentage, from 0 to 100.`);
			return Math.round(n * 100) / 10000;
		}
		case 'count':
		case 'minutes': {
			const n = number(t, f.label);
			if (!Number.isInteger(n)) throw new Error(`${f.label} is a whole number.`);
			return within(f, n);
		}
		default:
			return within(f, number(t, f.label));
	}
}

function number(t: string, label: string): number {
	if (t === '') throw new Error(`${label} needs a number.`);
	const n = Number(t);
	if (Number.isNaN(n)) throw new Error(`${label} must be a number.`);
	return n;
}

function within(f: Field, n: number): number {
	if (f.min !== undefined && n < f.min) throw new Error(`${f.label} is at least ${f.min}.`);
	if (f.max !== undefined && n > f.max) throw new Error(`${f.label} is at most ${f.max}.`);
	return n;
}
