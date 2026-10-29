import type { Feature } from './schema.ts';

// Diaries: when SkyrimNet writes an NPC's diary entry, and from what.
// Diary.yaml, in the dashboard's sections.
export const diary: Feature = {
	id: 'diaries',
	file: 'Diary',
	title: 'Diaries',
	what: 'When SkyrimNet writes an entry in an NPC’s diary, what it draws on, and whether those entries come back as memories.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'enabled', label: 'Write diaries', what: 'Let NPCs keep diaries at all.', kind: 'switch' },
				{ key: 'respect_day_boundary', label: 'One day at a time', what: 'Only events since midnight, game time, go into an entry.', kind: 'switch' },
				{ key: 'min_events_for_generation', label: 'Events before an entry', what: 'How many events not yet written up an NPC needs before an entry is written.', kind: 'count', min: 0, max: 200, unit: 'events' }
			]
		},
		{
			title: 'Content',
			fields: [
				{ key: 'max_recent_events', label: 'Recent events considered', what: 'How many of the latest events an entry can draw on.', kind: 'count', min: 0, max: 1000, unit: 'events' },
				{ key: 'max_recent_memories', label: 'Memories considered', what: 'How many memories an entry can draw on.', kind: 'count', min: 0, max: 20, unit: 'memories' },
				{ key: 'min_importance_score', label: 'Memories above', what: 'The importance a memory needs to be drawn on.', kind: 'percent' },
				{ key: 'target_entry_length', label: 'Entry length', what: 'The length an entry aims for, in words.', kind: 'count', min: 50, max: 10000, step: 50, unit: 'words' }
			]
		},
		{
			title: 'Recall',
			fields: [
				{ key: 'exclude_from_recall', label: 'Keep diaries out of recall', what: 'Skip the memories made from diary entries when an NPC recalls memories in conversation. The entries are still indexed, so searching them by hand still works.', kind: 'switch' }
			]
		}
	]
};
