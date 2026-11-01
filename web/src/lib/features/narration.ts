import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.narration;

// Narration: a narrator describing what happens. Part of SkyrimNet.yaml.
export const narration: Feature = {
	id: 'narration',
	file: 'game',
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.narration.title,
			fields: [
				{ key: 'narration.enabled', label: f.narration.narrationEnabled.label, what: f.narration.narrationEnabled.what, kind: 'switch' },
				{ key: 'narration.subtitlesOnly', label: f.narration.narrationSubtitlesOnly.label, what: f.narration.narrationSubtitlesOnly.what, kind: 'switch' },
				{ key: 'narration.disableDuringCombat', label: f.narration.narrationDisableDuringCombat.label, what: f.narration.narrationDisableDuringCombat.what, kind: 'switch' },
				{
					key: 'narration.disableFactions',
					label: f.narration.narrationDisableFactions.label,
					what: f.narration.narrationDisableFactions.what,
					kind: 'text',
					placeholder: 'DA10PilgrimsTrustFaction, ThievesGuildFaction'
				}
			]
		}
	]
};
