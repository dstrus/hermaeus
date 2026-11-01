import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.memoryFilter;

// The memory filter: which characters form memories at all. MemoryFilter.yaml
// holds the mode; the lists are filters/ files in the content library.
export const memoryFilter: Feature = {
	id: 'memory-filter',
	file: 'MemoryFilter',
	title: f.title,
	what: f.what,
	lists: 'memory',
	sections: [
		{
			title: f.mode.title,
			fields: [
				{ key: 'EnableMemoryFiltering', label: f.mode.enableMemoryFiltering.label, what: f.mode.enableMemoryFiltering.what, kind: 'switch' },
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
				{ key: 'AllowPlayerMemories', label: f.mode.allowPlayerMemories.label, what: f.mode.allowPlayerMemories.what, kind: 'switch' }
			]
		}
	]
};
