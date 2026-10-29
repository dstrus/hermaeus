// Tokens of an Inja template, for showing one with its markup set apart
// from its prose: {{ expressions }}, {% statements %}, and {# comments #},
// with the words inside a tag told apart from its delimiters. Nothing here
// checks the template; the server does that when it's saved.

export type Kind = 'text' | 'delim' | 'keyword' | 'string' | 'number' | 'name' | 'punct' | 'comment';

export interface Token {
	kind: Kind;
	text: string;
}

const keywords = new Set([
	'if', 'else', 'elif', 'endif', 'for', 'endfor', 'in', 'block', 'endblock', 'set', 'include', 'extends', 'raw', 'endraw',
	'and', 'or', 'not', 'true', 'false', 'null', 'none'
]);

const inside = /\s+|"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|\d+(?:\.\d+)?|[A-Za-z_][\w.]*|./gy;

// tokens splits a template into text and the parts of its tags.
export function tokens(text: string): Token[] {
	const out: Token[] = [];
	let i = 0;
	while (i < text.length) {
		const open = text.indexOf('{', i);
		const next = open < 0 ? '' : text[open + 1];
		if (open < 0 || !(next === '{' || next === '%' || next === '#')) {
			// Not a tag: take everything up to the next candidate.
			const end = open < 0 ? text.length : open + 1;
			out.push({ kind: 'text', text: text.slice(i, end) });
			i = end;
			continue;
		}
		if (open > i) out.push({ kind: 'text', text: text.slice(i, open) });
		const closer = next === '{' ? '}}' : next === '%' ? '%}' : '#}';
		let close = text.indexOf(closer, open + 2);
		if (close < 0) close = text.length;
		const end = Math.min(text.length, close + closer.length);
		if (next === '#') {
			out.push({ kind: 'comment', text: text.slice(open, end) });
			i = end;
			continue;
		}
		// A tag: its delimiters (with any whitespace trim dashes), then its words.
		let a = open + 2;
		if (text[a] === '-') a++;
		let b = close;
		if (text[b - 1] === '-' && b - 1 >= a) b--;
		out.push({ kind: 'delim', text: text.slice(open, a) });
		inside.lastIndex = a;
		let m: RegExpExecArray | null;
		while (inside.lastIndex < b && (m = inside.exec(text)) && m.index < b) {
			const t = m[0].length > b - m.index ? m[0].slice(0, b - m.index) : m[0];
			out.push({ kind: kindOf(t), text: t });
			if (inside.lastIndex > b) break;
		}
		out.push({ kind: 'delim', text: text.slice(b, end) });
		i = end;
	}
	return out;
}

function kindOf(t: string): Kind {
	const c = t[0]!;
	if (/\s/.test(c)) return 'text';
	if (c === '"' || c === "'") return 'string';
	if (/\d/.test(c)) return 'number';
	if (/[A-Za-z_]/.test(c)) return keywords.has(t) ? 'keyword' : 'name';
	return 'punct';
}
