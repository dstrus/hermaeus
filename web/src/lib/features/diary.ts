import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.diaries;

// Diaries: when SkyrimNet writes an NPC's diary entry, and from what.
// Diary.yaml, in the dashboard's sections.
export const diary: Feature = {
	id: 'diaries',
	file: 'Diary',
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.general.title,
			fields: [
				{ key: 'enabled', label: f.general.enabled.label, what: f.general.enabled.what, kind: 'switch' },
				{ key: 'respect_day_boundary', label: f.general.respectDayBoundary.label, what: f.general.respectDayBoundary.what, kind: 'switch' },
				{ key: 'min_events_for_generation', label: f.general.minEventsForGeneration.label, what: f.general.minEventsForGeneration.what, kind: 'count', min: 0, max: 200, unit: f.general.minEventsForGeneration.unit }
			]
		},
		{
			title: f.content.title,
			fields: [
				{ key: 'max_recent_events', label: f.content.maxRecentEvents.label, what: f.content.maxRecentEvents.what, kind: 'count', min: 0, max: 1000, unit: f.content.maxRecentEvents.unit },
				{ key: 'max_recent_memories', label: f.content.maxRecentMemories.label, what: f.content.maxRecentMemories.what, kind: 'count', min: 0, max: 20, unit: f.content.maxRecentMemories.unit },
				{ key: 'min_importance_score', label: f.content.minImportanceScore.label, what: f.content.minImportanceScore.what, kind: 'percent' },
				{ key: 'target_entry_length', label: f.content.targetEntryLength.label, what: f.content.targetEntryLength.what, kind: 'count', min: 50, max: 10000, step: 50, unit: f.content.targetEntryLength.unit }
			]
		},
		{
			title: f.recall.title,
			fields: [
				{ key: 'exclude_from_recall', label: f.recall.excludeFromRecall.label, what: f.recall.excludeFromRecall.what, kind: 'switch' }
			]
		}
	]
};
