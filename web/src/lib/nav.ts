// The sections of Hermaeus beyond a playthrough's own (characters, history,
// screenshots, songs), as the sidebar lists them and as their URLs nest:
// /<save>/<group>/<page>. A page's URL is built here and nowhere else.
export interface NavPage {
	path: string; // under its group
	label: string;
	feature?: boolean; // a settings file shown by FeaturePage, served by the group's [feature] route
}

export interface NavGroup {
	path: string;
	label: string;
	pages: NavPage[];
}

// A feature page, by its feature's id (see #lib/features).
const feature = (path: string, label: string): NavPage => ({ path, label, feature: true });

export const groups: NavGroup[] = [
	{
		path: 'behavior',
		label: 'Behavior',
		pages: [
			{ path: 'prompts', label: 'Prompts' },
			{ path: 'actions', label: 'Actions' },
			{ path: 'triggers', label: 'Triggers' },
			{ path: 'translator', label: 'Translator' },
			feature('gamemaster', 'GameMaster'),
			feature('narration', 'Narration'),
			feature('conversation', 'Conversation'),
			feature('npc-thoughts', 'NPC thoughts'),
			feature('player-reactions', 'Player reactions'),
			feature('bard-singing', 'Bard singing'),
			feature('actor-filter', 'Actor filter')
		]
	},
	{ path: 'memory', label: 'Memory', pages: [feature('memories', 'Memories'), feature('memory-filter', 'Memory filter'), feature('diaries', 'Diaries'), feature('dynamic-bios', 'Dynamic bios')] },
	{ path: 'voice', label: 'Voice', pages: [feature('speech', 'Speech'), { path: 'engines', label: 'Engines' }, { path: 'effects', label: 'Voice effects' }, feature('samples', 'Voice samples'), feature('speech-to-text', 'Speech to text')] },
	{
		path: 'ai',
		label: 'AI',
		pages: [
			{ path: 'models', label: 'Models' },
			{ path: 'providers', label: 'Providers' }
		]
	},
	{
		path: 'library',
		label: 'Library',
		pages: [
			{ path: 'plugins', label: 'Plugins' },
			{ path: 'content', label: 'Content' },
			{ path: 'overrides', label: 'Overrides' }
		]
	},
	{ path: 'general', label: 'General', pages: [feature('hotkeys', 'Hotkeys'), { path: 'other', label: 'Other' }] }
];

// href is a page's URL in a playthrough, by its page path; it finds the
// group the page belongs to.
export function href(save: string, page: string, rest = ''): string {
	const group = groups.find((g) => g.pages.some((p) => p.path === page));
	if (!group) throw new Error(`No page ${page}`);
	return `/${encodeURIComponent(save)}/${group.path}/${page}${rest}`;
}

// featureGroup is the group whose [feature] route shows a feature, if any.
export function featureGroup(group: string, id: string): NavGroup | undefined {
	const g = groups.find((g) => g.path === group);
	return g?.pages.some((p) => p.feature && p.path === id) ? g : undefined;
}
