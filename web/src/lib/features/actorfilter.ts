import type { Feature } from './schema.ts';

// The actor filter: which characters SkyrimNet involves at all, to speak,
// react, or be spoken to. ActorFilter.yaml holds the mode; the lists are
// filters/ files in the content library.
export const actorFilter: Feature = {
	id: 'actor-filter',
	file: 'ActorFilter',
	title: 'Actor filter',
	what: 'Which characters SkyrimNet involves at all: a few switches here, and lists of factions, races, and genders below. A whitelist entry also lets a character who can’t speak take part, the way Paarthurnax does.',
	lists: 'actor',
	sections: [
		{
			title: 'Mode',
			fields: [
				{
					key: 'FilterMode',
					label: 'Mode',
					what: 'With a blacklist, every character takes part except those the blacklists name. With a whitelist, only those the whitelists name do.',
					kind: 'choice',
					choices: [
						{ value: 'BlacklistExclusion', label: 'Blacklist', what: 'Everyone but those listed' },
						{ value: 'WhitelistOnly', label: 'Whitelist', what: 'Only those listed' }
					]
				},
				{ key: 'AllowOnlyActorsCapableOfSpeech', label: 'Only those who can speak', what: 'Characters who can’t speak are left out, unless a whitelist names them.', kind: 'switch' },
				{ key: 'AllowSummonedOrReanimated', label: 'Summoned and raised', what: 'Summoned and reanimated characters always take part, whatever the lists say.', kind: 'switch' }
			]
		}
	]
};
