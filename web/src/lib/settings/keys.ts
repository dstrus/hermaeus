// SkyrimNet's hotkeys are Windows virtual-key codes (F1 is 112, A is 65),
// with -1 for no key. These turn a code into the key's name, and a key
// pressed in the browser into its code, through the key's physical position
// (KeyboardEvent.code), so a Mac keyboard gives the same code the game sees.

export const noKey = -1;

const names: Record<number, string> = {
	1: 'Left mouse button',
	2: 'Right mouse button',
	4: 'Middle mouse button',
	5: 'Mouse button 4',
	6: 'Mouse button 5',
	8: 'Backspace',
	9: 'Tab',
	13: 'Enter',
	16: 'Shift',
	17: 'Ctrl',
	18: 'Alt',
	19: 'Pause',
	20: 'Caps Lock',
	27: 'Esc',
	32: 'Space',
	33: 'Page Up',
	34: 'Page Down',
	35: 'End',
	36: 'Home',
	37: 'Left',
	38: 'Up',
	39: 'Right',
	40: 'Down',
	44: 'Print Screen',
	45: 'Insert',
	46: 'Delete',
	91: 'Left Windows',
	92: 'Right Windows',
	93: 'Menu',
	106: 'Numpad *',
	107: 'Numpad +',
	109: 'Numpad -',
	110: 'Numpad .',
	111: 'Numpad /',
	144: 'Num Lock',
	145: 'Scroll Lock',
	160: 'Left Shift',
	161: 'Right Shift',
	162: 'Left Ctrl',
	163: 'Right Ctrl',
	164: 'Left Alt',
	165: 'Right Alt',
	186: ';',
	187: '=',
	188: ',',
	189: '-',
	190: '.',
	191: '/',
	192: '`',
	219: '[',
	220: '\\',
	221: ']',
	222: "'"
};
for (let i = 0; i < 10; i++) {
	names[48 + i] = String(i);
	names[96 + i] = `Numpad ${i}`;
}
for (let i = 0; i < 26; i++) names[65 + i] = String.fromCharCode(65 + i);
for (let i = 1; i <= 24; i++) names[111 + i] = `F${i}`;

// keyName is a hotkey code as the key is labelled, or "Not set".
export function keyName(code: number): string {
	if (code === noKey) return 'Not set';
	return names[code] ?? `Key ${code}`;
}

// The code for each physical key, by KeyboardEvent.code.
const byPosition: Record<string, number> = {
	Backspace: 8,
	Tab: 9,
	Enter: 13,
	NumpadEnter: 13,
	ShiftLeft: 160,
	ShiftRight: 161,
	ControlLeft: 162,
	ControlRight: 163,
	AltLeft: 164,
	AltRight: 165,
	Pause: 19,
	CapsLock: 20,
	Escape: 27,
	Space: 32,
	PageUp: 33,
	PageDown: 34,
	End: 35,
	Home: 36,
	ArrowLeft: 37,
	ArrowUp: 38,
	ArrowRight: 39,
	ArrowDown: 40,
	PrintScreen: 44,
	Insert: 45,
	Delete: 46,
	MetaLeft: 91,
	MetaRight: 92,
	ContextMenu: 93,
	NumpadMultiply: 106,
	NumpadAdd: 107,
	NumpadSubtract: 109,
	NumpadDecimal: 110,
	NumpadDivide: 111,
	NumLock: 144,
	ScrollLock: 145,
	Semicolon: 186,
	Equal: 187,
	Comma: 188,
	Minus: 189,
	Period: 190,
	Slash: 191,
	Backquote: 192,
	BracketLeft: 219,
	Backslash: 220,
	BracketRight: 221,
	Quote: 222
};
for (let i = 0; i < 10; i++) {
	byPosition[`Digit${i}`] = 48 + i;
	byPosition[`Numpad${i}`] = 96 + i;
}
for (let i = 0; i < 26; i++) byPosition[`Key${String.fromCharCode(65 + i)}`] = 65 + i;
for (let i = 1; i <= 24; i++) byPosition[`F${i}`] = 111 + i;

// codeOf is the hotkey code for a key press, or undefined for a key the
// game has no code for. Shift, Ctrl, and Alt count as keys of their own.
export function codeOf(e: KeyboardEvent): number | undefined {
	return byPosition[e.code];
}

// mouseCodeOf is the hotkey code for a mouse button press.
export function mouseCodeOf(e: MouseEvent): number | undefined {
	return [1, 4, 2, 5, 6][e.button];
}
