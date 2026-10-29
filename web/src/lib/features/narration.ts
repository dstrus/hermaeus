import type { Feature } from './schema.ts';

// Narration: a narrator describing what happens. Part of SkyrimNet.yaml.
export const narration: Feature = {
	id: 'narration',
	file: 'game',
	title: 'Narration',
	what: 'A narrator describing what happens around you, spoken or as subtitles.',
	sections: [
		{
			title: 'Narration',
			fields: [
				{ key: 'narration.enabled', label: 'Narrate', what: 'Have a narrator describe what happens.', kind: 'switch' },
				{ key: 'narration.subtitlesOnly', label: 'Subtitles only', what: 'Show the narration as subtitles without voicing it. It still goes into prompts and the event history.', kind: 'switch' },
				{ key: 'narration.disableDuringCombat', label: 'Quiet in combat', what: 'No narration while you’re fighting.', kind: 'switch' },
				{
					key: 'narration.disableFactions',
					label: 'Quiet in these factions',
					what: 'No narration while you belong to any of these factions, by editor ID, separated by commas.',
					kind: 'text',
					placeholder: 'DA10PilgrimsTrustFaction, ThievesGuildFaction'
				}
			]
		}
	]
};
