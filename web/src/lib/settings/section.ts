// The Settings pages, as the sidebar lists them: each shows one settings
// file in its own terms, and Other has every file without a page of its
// own, as written. The Go side's own-page list (internal/settings) names the
// same files, so Other leaves them out.
export interface SettingsPage {
	path: string; // under settings/
	label: string;
	file?: string; // the settings file the page shows, for its count of changes
}

export const settingsPages: SettingsPage[] = [
	{ path: 'models', label: 'Models', file: 'OpenRouter' },
	{ path: 'providers', label: 'Providers', file: 'Providers' },
	{ path: 'translator', label: 'Translator', file: 'UniversalTranslator' },
	{ path: 'gamemaster', label: 'GameMaster' }, // part of SkyrimNet.yaml, which stays under Other
	{ path: 'narration', label: 'Narration' },
	{ path: 'memories', label: 'Memories', file: 'Memory' },
	{ path: 'memory-filter', label: 'Memory filter', file: 'MemoryFilter' },
	{ path: 'diaries', label: 'Diaries', file: 'Diary' },
	{ path: 'dynamic-bios', label: 'Dynamic bios', file: 'DynamicBio' },
	{ path: 'hotkeys', label: 'Hotkeys', file: 'Hotkey' },
	{ path: 'other', label: 'Other' }
];

export function settingsPageHref(save: string, path: string): string {
	return `/${encodeURIComponent(save)}/settings/${path}`;
}
