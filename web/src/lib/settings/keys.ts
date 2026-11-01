// SkyrimNet's hotkeys are Windows virtual-key codes (F1 is 112, A is 65),
// with -1 for no key. These turn a code into the key's name, and a key
// pressed in the browser into its code, through the key's physical position
// (KeyboardEvent.code), so a Mac keyboard gives the same code the game sees.
import { m } from '#lib/copy/messages.ts';

export const noKey = -1;

const k = m.settings.keys;

const names: Record<number, string> = {
	1: k.leftMouse,
	2: k.rightMouse,
	4: k.middleMouse,
	5: k.mouse4,
	6: k.mouse5,
	8: k.backspace,
	9: k.tab,
	13: k.enter,
	16: k.shift,
	17: k.ctrl,
	18: k.alt,
	19: k.pause,
	20: k.capsLock,
	27: k.esc,
	32: k.space,
	33: k.pageUp,
	34: k.pageDown,
	35: k.end,
	36: k.home,
	37: k.left,
	38: k.up,
	39: k.right,
	40: k.down,
	44: k.printScreen,
	45: k.insert,
	46: k.delete,
	91: k.leftWindows,
	92: k.rightWindows,
	93: k.menu,
	106: k.numpad({ key: '*' }),
	107: k.numpad({ key: '+' }),
	109: k.numpad({ key: '-' }),
	110: k.numpad({ key: '.' }),
	111: k.numpad({ key: '/' }),
	144: k.numLock,
	145: k.scrollLock,
	160: k.leftShift,
	161: k.rightShift,
	162: k.leftCtrl,
	163: k.rightCtrl,
	164: k.leftAlt,
	165: k.rightAlt,
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
	names[96 + i] = k.numpad({ key: i });
}
for (let i = 0; i < 26; i++) names[65 + i] = String.fromCharCode(65 + i);
for (let i = 1; i <= 24; i++) names[111 + i] = `F${i}`;

// keyName is a hotkey code as the key is labelled, or "Not set".
export function keyName(code: number): string {
	if (code === noKey) return k.notSet;
	return names[code] ?? k.unknown({ code });
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
