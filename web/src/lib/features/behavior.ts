// How characters take turns to speak, think, and react: the settings files
// for conversation, NPC thoughts, and the player's own reactions.
import type { Feature } from './schema.ts';

const wait = (key: string, label: string, what: string) => ({ key: `ttl_seconds.${key}`, label, what: `${what} At 0, it’s never dropped.`, kind: 'count' as const, min: 0, max: 3600, unit: 'seconds', zero: 'Never dropped' });

export const conversation: Feature = {
	id: 'conversation',
	file: 'Conversation',
	title: 'Conversation',
	what: 'How characters take turns to speak: how far one line leads to another, which reasons to speak win out, and how long each kind may wait its turn before it’s dropped.',
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
	what: 'Thoughts a character has but doesn’t say. A thought comes only when a trigger or a mod asks for one; none come on their own.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'enabled', label: 'NPC thoughts', what: 'Let characters have thoughts at all. Off, requests for one are ignored.', kind: 'switch' },
				{ key: 'perNPCCooldownSeconds', label: 'Between thoughts', what: 'The least time between two thoughts of the same character, so one character doesn’t take every turn.', kind: 'count', min: 0, max: 3600, unit: 'seconds' },
				{ key: 'eventHistoryCount', label: 'Recent events considered', what: 'How many of the latest events a thought can draw on.', kind: 'count', min: 0, max: 500, unit: 'events' }
			]
		},
		{
			title: 'In dialogue',
			fields: [{ key: 'allowInlineInDialogue', label: 'Thoughts while speaking', what: 'Let a character have a thought in the middle of what they say. It isn’t spoken, and it’s kept as a thought of its own.', kind: 'switch' }]
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
