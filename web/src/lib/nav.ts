// The sections of Hermaeus beyond a playthrough's own (characters, history,
// screenshots, songs), as the sidebar lists them and as their URLs nest:
// /<save>/<group>/<page>. A page's URL is built here and nowhere else.
import { m } from '#lib/copy/messages.ts';

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

const n = m.nav;

export const groups: NavGroup[] = [
	{
		path: 'behavior',
		label: n.behavior.label,
		pages: [
			{ path: 'prompts', label: n.behavior.prompts },
			{ path: 'actions', label: n.behavior.actions },
			{ path: 'triggers', label: n.behavior.triggers },
			{ path: 'translator', label: n.behavior.translator },
			feature('gamemaster', n.behavior.gamemaster),
			feature('narration', n.behavior.narration),
			feature('conversation', n.behavior.conversation),
			feature('npc-thoughts', n.behavior.npcThoughts),
			feature('player-reactions', n.behavior.playerReactions),
			feature('bard-singing', n.behavior.bardSinging),
			feature('actor-filter', n.behavior.actorFilter)
		]
	},
	{ path: 'memory', label: n.memory.label, pages: [feature('memories', n.memory.memories), feature('memory-filter', n.memory.memoryFilter), feature('diaries', n.memory.diaries), feature('dynamic-bios', n.memory.dynamicBios)] },
	{ path: 'voice', label: n.voice.label, pages: [feature('speech', n.voice.speech), { path: 'engines', label: n.voice.engines }, { path: 'effects', label: n.voice.effects }, feature('samples', n.voice.samples), feature('speech-to-text', n.voice.speechToText)] },
	{
		path: 'ai',
		label: n.ai.label,
		pages: [
			{ path: 'models', label: n.ai.models },
			{ path: 'providers', label: n.ai.providers }
		]
	},
	{
		path: 'library',
		label: n.library.label,
		pages: [
			{ path: 'plugins', label: n.library.plugins },
			{ path: 'content', label: n.library.content },
			{ path: 'overrides', label: n.library.overrides }
		]
	},
	{ path: 'general', label: n.general.label, pages: [feature('hotkeys', n.general.hotkeys), { path: 'other', label: n.general.other }] }
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
