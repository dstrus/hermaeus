import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.dynamicBios;

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
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.general.title,
			fields: [
				{ key: 'enabled', label: f.general.enabled.label, what: f.general.enabled.what, kind: 'switch' },
				{ key: 'preserve_core_personality', label: f.general.preserveCorePersonality.label, what: f.general.preserveCorePersonality.what, kind: 'switch' }
			]
		},
		{
			title: f.sections.title,
			fields: [{ key: 'updatable_blocks', label: f.sections.updatableBlocks.label, what: f.sections.updatableBlocks.what, kind: 'set', choices: blocks }]
		},
		{
			title: f.context.title,
			fields: [
				{ key: 'max_recent_events', label: f.context.maxRecentEvents.label, what: f.context.maxRecentEvents.what, kind: 'count', min: 10, max: 500, unit: f.context.maxRecentEvents.unit },
				{ key: 'max_memories', label: f.context.maxMemories.label, what: f.context.maxMemories.what, kind: 'count', min: 0, max: 100, unit: f.context.maxMemories.unit },
				{ key: 'min_importance_score', label: f.context.minImportanceScore.label, what: f.context.minImportanceScore.what, kind: 'percent' }
			]
		}
	]
};
