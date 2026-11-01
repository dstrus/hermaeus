import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.actorFilter;

// The actor filter: which characters SkyrimNet involves at all, to speak,
// react, or be spoken to. ActorFilter.yaml holds the mode; the lists are
// filters/ files in the content library.
export const actorFilter: Feature = {
	id: 'actor-filter',
	file: 'ActorFilter',
	title: f.title,
	what: f.what,
	lists: 'actor',
	sections: [
		{
			title: f.mode.title,
			fields: [
				{
					key: 'FilterMode',
					label: f.mode.filterMode.label,
					what: f.mode.filterMode.what,
					kind: 'choice',
					choices: [
						{ value: 'BlacklistExclusion', label: f.mode.filterMode.choices.blacklistExclusion.label, what: f.mode.filterMode.choices.blacklistExclusion.what },
						{ value: 'WhitelistOnly', label: f.mode.filterMode.choices.whitelistOnly.label, what: f.mode.filterMode.choices.whitelistOnly.what }
					]
				},
				{ key: 'AllowOnlyActorsCapableOfSpeech', label: f.mode.allowOnlyActorsCapableOfSpeech.label, what: f.mode.allowOnlyActorsCapableOfSpeech.what, kind: 'switch' },
				{ key: 'AllowSummonedOrReanimated', label: f.mode.allowSummonedOrReanimated.label, what: f.mode.allowSummonedOrReanimated.what, kind: 'switch' }
			]
		}
	]
};
