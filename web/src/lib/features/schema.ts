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
	| 'hotkey'; // a key code, -1 for none

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
	if (v === undefined || v === null) return '—';
	switch (f.kind) {
		case 'switch':
			return v === true ? 'On' : 'Off';
		case 'percent':
			return typeof v === 'number' ? `${Number((v * 100).toFixed(1))}%` : String(v);
		case 'minutes':
			return typeof v === 'number' ? `${v} min` : String(v);
		case 'count':
			return typeof v === 'number' ? `${v}${f.unit ? ` ${f.unit}` : ''}` : String(v);
		case 'choice': {
			const c = f.choices?.find((c) => c.value === String(v));
			return c ? c.label : String(v);
		}
		case 'hotkey':
			return typeof v === 'number' ? keyName(v) : String(v);
		case 'set':
			return Array.isArray(v) ? `${v.length} of ${f.choices?.length ?? v.length}` : String(v);
		case 'text':
			return String(v).trim() === '' ? 'None' : String(v);
		default:
			return String(v);
	}
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

// draftOf is a value as the form holds it: text, percentages as whole numbers.
export function draftOf(f: Field, v: unknown): string {
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
