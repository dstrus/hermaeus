// How characters take turns to speak, think, and react: the settings files
// for conversation, NPC thoughts, and the player's own reactions.
import type { Feature } from './schema.ts';

const wait = (key: string, label: string, what: string) => ({ key: `ttl_seconds.${key}`, label, what: `${what} At 0, it’s never dropped.`, kind: 'count' as const, min: 0, max: 3600, unit: 'seconds', zero: 'Never dropped' });

export const conversation: Feature = {
	id: 'conversation',
	file: 'Conversation',
	title: 'Conversation',
	what: 'How characters take turns to speak: how many times they can reply to one another, which reasons to speak are prioritized, and how long each kind may wait its turn before it’s dropped.',
	sections: [
		{
			title: 'Turn taking',
			fields: [
				{ key: 'max_chain_depth', label: 'Replies in a row', what: 'How many characters may answer one another after an opening line before the exchange ends.', kind: 'count', min: 0, max: 20, unit: 'replies' },
				{ key: 'takeover_before_audio', label: 'Urgent lines go first', what: 'A more urgent reason to speak, such as you or a line said to someone, replaces a line still being made, before any of it is heard.', kind: 'switch' },
				{ key: 'lead_seconds', label: 'Head start', what: 'How long before the current line ends the next one starts being made. At 0, SkyrimNet judges it from recent lines.', kind: 'number', min: 0, max: 60, zero: 'Auto' },
				{ key: 'generation_timeout_seconds', label: 'Give up after', what: 'A line still being made after this long is abandoned, and the next may go.', kind: 'count', min: 1, max: 600, unit: 'seconds' }
			]
		},
		{
			title: 'Waiting',
			what: 'How long each kind of reason to speak may wait for its turn.',
			fields: [wait('player', 'Your speech', 'Characters answering you.'), wait('directed', 'Said to someone', 'Lines meant for a particular character.'), wait('urgent', 'Urgent events', 'Lines about something that just happened, such as an attack.'), wait('reaction', 'Reactions', 'Characters reacting to what’s around them.'), wait('continuation', 'Replies', 'Characters answering one another.')]
		}
	]
};

export const npcThoughts: Feature = {
	id: 'npc-thoughts',
	file: 'NpcThoughts',
	title: 'NPC thoughts',
	what: 'Unspoken thoughts a character has, coming only when a trigger or a mod asks for one.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'enabled', label: 'NPC thoughts', what: 'Let characters have thoughts at all. When off, requests for thoughts are ignored.', kind: 'switch' },
				{ key: 'perNPCCooldownSeconds', label: 'Between thoughts', what: 'A minimum wait time between two thoughts for a single character, so one character doesn’t take every turn.', kind: 'count', min: 0, max: 3600, unit: 'seconds' },
				{ key: 'eventHistoryCount', label: 'Recent events considered', what: 'How many of the latest events a thought can draw on.', kind: 'count', min: 0, max: 500, unit: 'events' }
			]
		},
		{
			title: 'In dialogue',
			fields: [{ key: 'allowInlineInDialogue', label: 'Thoughts while speaking', what: 'Let a character record a separate, unspoken thought while speaking.', kind: 'switch' }]
		}
	]
};

export const playerReactions: Feature = {
	id: 'player-reactions',
	file: 'PlayerDialogue',
	title: 'Player reactions',
	what: 'Your character reacting to what happens without you doing anything: a thought, or a line said aloud, when certain events happen.',
	sections: [
		{
			title: 'General',
			fields: [{ key: 'enabled', label: 'Player reactions', what: 'Let your character react on their own. It’s the same switch as Player reactions among the quick toggles.', kind: 'switch' }]
		},
		{
			title: 'Events',
			fields: [
				{
					key: 'eventConfigs',
					label: 'Reactions by event',
					what: 'Each kind of event your character reacts to: whether they do, how, how likely they are to, and how long before they do again.',
					kind: 'rows',
					noun: ['event', 'events'],
					cols: [
						{ key: 'eventType', label: 'Event', kind: 'text', placeholder: 'combat' },
						{ key: 'enabled', label: 'On', kind: 'switch' },
						{
							key: 'responseType',
							label: 'As',
							kind: 'choice',
							choices: [
								{ value: 'thought', label: 'a thought' },
								{ value: 'audible', label: 'aloud' }
							]
						},
						{ key: 'probability', label: 'Chance', kind: 'percent' },
						{ key: 'cooldownSeconds', label: 'Cooldown', kind: 'count', unit: 's', min: 0, max: 86400 }
					]
				}
			]
		}
	]
};

const musicProviders = [
	{ value: 'openrouter', label: 'OpenRouter', what: 'openrouter.ai, with an OpenRouter key' },
	{ value: 'suno', label: 'Suno', what: 'sunoapi.org, with a key from there' },
	{ value: 'nano_gpt', label: 'NanoGPT', what: 'nano-gpt.com, with a NanoGPT key' },
	{ value: 'minimax', label: 'MiniMax', what: 'api.minimax.io, with a MiniMax key' },
	{ value: 'acestep_cloud', label: 'AceMusic', what: 'acemusic.ai’s AceStep, with an AceMusic key' },
	{ value: 'acestep_local', label: 'AceStep server', what: 'an AceStep server you run' }
];

const providerKey = (id: string, label: string, where: string) => ({ key: `bard_singing.${id}.api_key`, label: 'Key', what: `Your ${label} key, from ${where}. It’s kept apart from the keys on the Providers page.`, kind: 'secret' as const });
const providerModel = (id: string, label: string, example: string) => ({ key: `bard_singing.${id}.model`, label: 'Model', what: `The music model, as ${label} names it.`, kind: 'text' as const, placeholder: example });

export const bardSinging: Feature = {
	id: 'bard-singing',
	file: 'BardSinging',
	title: 'Bard singing',
	what: 'Bards singing songs: new ones a music service composes from what’s been happening and what bards remember, and ones already made.',
	see: { before: 'The songs of each playthrough are under ', label: 'Bard songs', path: 'songs', after: '.' },
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'bard_singing.enabled', label: 'Bard singing', what: 'Let bards sing at all. It’s the same switch as Bards among the quick toggles.', kind: 'switch' },
				{ key: 'bard_singing.generation.compose_enabled', label: 'Compose new songs', what: 'Have the music service below compose new songs. Off, bards sing only the songs they already have.', kind: 'switch' },
				{ key: 'bard_singing.playback.volume', label: 'Volume', what: 'How loud bards sing, next to the game’s other sounds. With Skyrim’s own audio engine in use, its SkyrimNet Bards slider sets this instead.', kind: 'percent' }
			]
		},
		{
			title: 'Composing',
			fields: [
				{ key: 'bard_singing.provider', label: 'Music service', what: 'Who composes new songs. Each costs credits there.', kind: 'choice', custom: true, choices: musicProviders },
				{ key: 'bard_singing.backup_provider', label: 'Backup', what: 'Who composes when the music service fails.', kind: 'choice', custom: true, choices: [...musicProviders, { value: 'none', label: 'None', what: 'no backup' }] },
				{ key: 'bard_singing.generation.interval_hours', label: 'How often', what: 'How many hours of game time pass between tries at a new song.', kind: 'count', min: 1, max: 720, unit: 'game hours' },
				{ key: 'bard_singing.generation.max_cached_songs', label: 'Songs kept', what: 'How many composed songs each playthrough keeps. Past it, the oldest goes first; imported songs don’t count and are never removed.', kind: 'count', min: 1, max: 1000, unit: 'songs' },
				{ key: 'bard_singing.generation.default_style', label: 'Style', what: 'Words for the kind of music wanted, separated by commas.', kind: 'text', placeholder: 'tavern music, bardic, medieval folk' },
				{ key: 'bard_singing.generation.negative_tags', label: 'Avoid', what: 'Words for what the music shouldn’t be, separated by commas, such as modern genres.', kind: 'text', placeholder: 'electric guitar, synthesizer' },
				{ key: 'bard_singing.generation.language', label: 'Sung in', what: 'The language songs are sung in, as a two-letter code. Only AceStep uses it.', kind: 'text', placeholder: 'en' }
			]
		},
		{ title: 'OpenRouter', fields: [providerKey('openrouter', 'OpenRouter', 'openrouter.ai'), providerModel('openrouter', 'OpenRouter', 'google/lyria-3-pro-preview')] },
		{
			title: 'Suno',
			fields: [
				providerKey('suno', 'sunoapi.org', 'sunoapi.org'),
				{ key: 'bard_singing.suno.base_url', label: 'Address', what: 'Where Suno’s service is reached. Change it only to go through a proxy of your own.', kind: 'text', placeholder: 'https://api.sunoapi.org/api/v1' },
				providerModel('suno', 'Suno', 'V5')
			]
		},
		{ title: 'NanoGPT', fields: [providerKey('nano_gpt', 'NanoGPT', 'nano-gpt.com'), providerModel('nano_gpt', 'NanoGPT', 'google/lyria-3-pro/music')] },
		{ title: 'MiniMax', fields: [providerKey('minimax', 'MiniMax', 'MiniMax'), providerModel('minimax', 'MiniMax', 'music-3.0-free')] },
		{ title: 'AceMusic', fields: [providerKey('acestep_cloud', 'AceMusic', 'acemusic.ai'), providerModel('acestep_cloud', 'AceMusic', 'acemusic/acestep-v1.5-turbo')] },
		{
			title: 'AceStep server',
			fields: [
				{ key: 'bard_singing.acestep_local.base_url', label: 'Address', what: 'Your AceStep server’s full chat-completions address.', kind: 'text', placeholder: 'http://localhost:8002/v1/chat/completions' },
				providerModel('acestep_local', 'your server', 'acemusic/acestep-v1.5-turbo')
			]
		}
	]
};
