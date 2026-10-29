import type { Feature } from './schema.ts';

// The bio's sections SkyrimNet may rewrite, by name as in the bio file.
const blocks = ['summary', 'interject_summary', 'background', 'personality', 'appearance', 'aspirations', 'relationships', 'occupation', 'skills', 'speech_style'].map((value) => ({
	value,
	label: value.replace(/_/g, ' ')
}));

// Dynamic bios: how SkyrimNet rewrites characters' bios as you play.
// DynamicBio.yaml, in the dashboard's sections.
export const dynamicBio: Feature = {
	id: 'dynamic-bios',
	file: 'DynamicBio',
	title: 'Dynamic bios',
	what: 'How SkyrimNet rewrites a character’s bio as you play: which sections it may change, and what it draws on.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'enabled', label: 'Rewrite bios', what: 'Let SkyrimNet update a character’s dynamic bio from what happens.', kind: 'switch' },
				{ key: 'preserve_core_personality', label: 'Keep the core personality', what: 'A rewrite keeps the character’s defining traits, whatever has happened.', kind: 'switch' }
			]
		},
		{
			title: 'Sections',
			fields: [{ key: 'updatable_blocks', label: 'Sections it may rewrite', what: 'Which sections of a bio a rewrite may change. The rest stay as written.', kind: 'set', choices: blocks }]
		},
		{
			title: 'Context',
			fields: [
				{ key: 'max_recent_events', label: 'Recent events considered', what: 'How many of the latest events a rewrite can draw on.', kind: 'count', min: 10, max: 500, unit: 'events' },
				{ key: 'max_memories', label: 'Memories considered', what: 'How many memories a rewrite can draw on.', kind: 'count', min: 0, max: 100, unit: 'memories' },
				{ key: 'min_importance_score', label: 'Memories above', what: 'The importance a memory needs to be drawn on.', kind: 'percent' }
			]
		}
	]
};
