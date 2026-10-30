import type { Feature } from './schema.ts';

// The memory filter: which characters form memories at all. MemoryFilter.yaml
// holds the mode; the lists are filters/ files in the content library.
export const memoryFilter: Feature = {
	id: 'memory-filter',
	file: 'MemoryFilter',
	title: 'Memory filter',
	what: 'Which characters form memories at all: a few switches here, and lists of factions, races, and genders below.',
	lists: 'memory',
	sections: [
		{
			title: 'Mode',
			fields: [
				{ key: 'EnableMemoryFiltering', label: 'Filter memories', what: 'Check each character against the switches and lists here before they form a memory.', kind: 'switch' },
				{
					key: 'FilterMode',
					label: 'Mode',
					what: 'With a blacklist, everyone forms memories except those the blacklists name. With a whitelist, only those the whitelists name do.',
					kind: 'choice',
					choices: [
						{ value: 'BlacklistExclusion', label: 'Blacklist', what: 'Everyone but those listed' },
						{ value: 'WhitelistOnly', label: 'Whitelist', what: 'Only those listed' }
					]
				},
				{ key: 'AllowOnlyActorsCapableOfSpeech', label: 'Only those who can speak', what: 'Characters who can’t speak form no memories, unless a whitelist names them.', kind: 'switch' },
				{ key: 'AllowPlayerMemories', label: 'Your character too', what: 'Your own character forms memories.', kind: 'switch' }
			]
		}
	]
};
