// Forms for records: each kind's fields, in blocks, over the record as an
// object. A field names its value by a dotted path into the object
// ("response.type", "parameterMapping.0.name"); a block of rows repeats its
// blocks for each item of a list. The form edits the object in place; the
// server merges it back into the file, so whatever the form leaves alone,
// and comments, stay as they were.

export type Obj = { [key: string]: unknown };

export interface Choice {
	value: string;
	label: string;
}

export interface FieldSpec {
	key: string; // a dotted path into the record
	label: string;
	type?: 'text' | 'area' | 'check' | 'num' | 'select' | 'list' | 'csv' | 'any';
	choices?: Choice[]; // select: the values; text: suggestions
	hint?: string;
	readonly?: boolean; // the field that names the file
	keep?: boolean; // text: empty stays as empty text, rather than removing the field
	on?: boolean; // check: on when the field is absent, and left out when on
	rows?: number; // area: its height
	wide?: boolean; // spans the block's width
}

export interface FieldBlock {
	title?: string;
	note?: string;
	fields: (FieldSpec | false | undefined)[];
}

export interface RowsBlock {
	title: string;
	rows: string; // the list's path
	add: string; // what the add button adds, such as "condition"
	blank: () => unknown;
	item: (base: string) => Block[]; // the blocks for one item, at base
	empty?: string;
}

export type Block = FieldBlock | RowsBlock;

export interface FormKind {
	noun: string;
	// The field that must match the file's name, if any.
	nameKey?: string;
	blank: (name: string) => Obj;
	blocks: (d: Obj, ctx: FormContext) => Block[];
}

// What a form knows beyond its record: names from other records, for
// suggestions.
export interface FormContext {
	effects: string[]; // voice effect ids
	categories: string[]; // action categories in use
	decorators: string[]; // eligibility decorators in use
}

// ---------- paths ----------

const parts = (path: string) => path.split('.').map((p) => (/^\d+$/.test(p) ? Number(p) : p));

export function getAt(obj: unknown, path: string): unknown {
	let cur = obj;
	for (const p of parts(path)) {
		if (cur === null || typeof cur !== 'object') return undefined;
		cur = (cur as Obj)[p as string];
	}
	return cur;
}

// setAt sets a value, making the objects and lists on the way; undefined
// removes the field.
export function setAt(obj: Obj, path: string, value: unknown) {
	const ps = parts(path);
	let cur: Obj | unknown[] = obj;
	for (const [i, p] of ps.slice(0, -1).entries()) {
		const next = (cur as Obj)[p as string];
		if (next === null || typeof next !== 'object') {
			const made = typeof ps[i + 1] === 'number' ? [] : {};
			(cur as Obj)[p as string] = made;
			cur = made;
		} else cur = next as Obj;
	}
	const last = ps[ps.length - 1]!;
	if (value === undefined && !Array.isArray(cur)) delete (cur as Obj)[last as string];
	else (cur as Obj)[last as string] = value;
}

// ---------- values ----------

// shown is a field's value as its input holds it.
export function shown(f: FieldSpec, v: unknown): string {
	if (v === undefined || v === null) return '';
	if (f.type === 'list') return Array.isArray(v) ? v.map(String).join('\n') : String(v);
	if (f.type === 'csv') return Array.isArray(v) ? v.map(String).join(', ') : String(v);
	return String(v);
}

// stored is what an input's text means for the record; undefined removes
// the field.
export function stored(f: FieldSpec, text: string, before: unknown): unknown {
	const t = text.trim();
	switch (f.type) {
		case 'num':
			return t === '' || Number.isNaN(Number(t)) ? undefined : Number(t);
		case 'list': {
			const items = text
				.split('\n')
				.map((x) => x.trim())
				.filter(Boolean);
			return items.length ? items : undefined;
		}
		case 'csv':
			return t
				.split(',')
				.map((x) => x.trim())
				.filter(Boolean);
		case 'any':
			// A value of whatever kind it was, or the kind its text reads as.
			if (t === '') return typeof before === 'string' ? '' : undefined;
			if (typeof before === 'boolean' || t === 'true' || t === 'false') return t === 'true' ? true : t === 'false' ? false : text;
			if (typeof before !== 'string' && t !== '' && !Number.isNaN(Number(t))) return Number(t);
			return text;
		default:
			return text === '' && !f.keep ? undefined : text;
	}
}

// ---------- the kinds ----------

const options = (values: string[], labels: Record<string, string> = {}): Choice[] => values.map((value) => ({ value, label: labels[value] ?? value }));

const triggerEvents = [
	'spell_cast',
	'active_effect',
	'hit',
	'combat',
	'death',
	'activation',
	'equip',
	'sleep_start',
	'sleep_stop',
	'book_read',
	'quest_stage',
	'quest_start_stop',
	'location_change',
	'container_changed',
	'enter_bleedout',
	'animation_event',
	'mod_event',
	'crime',
	'dragon_soul',
	'dialogue',
	'notification',
	'messagebox',
	'trade_start',
	'trade_complete',
	'lock_changed',
	'quest_objective_state',
	'cell_attach_detach',
	'scene',
	'scene_action',
	'scene_phase',
	'package_apply',
	'package_remove',
	'diary_entry_created',
	'custom',
	'direct_narration',
	'persistent_generic',
	'player_thoughts',
	'npc_thoughts',
	'dialogue_npc',
	'dialogue_player',
	'dialogue_player_stt',
	'dialogue_player_text',
	'dialogue_background',
	'dialogue_player_monologue',
	'dialogue_player_telepathy',
	'dialogue_npc_telepathy',
	'*'
];
const operators = ['equals', 'not_equals', 'contains', 'not_contains', 'starts_with', 'ends_with', 'regex', 'greater_than', 'less_than', 'greater_equal', 'less_equal', 'length_equals', 'length_greater', 'length_less', 'has_field', 'not_has_field'];
const responses = options(['player_thought', 'player_dialogue', 'npc_thought', 'direct_narration', 'persistent_generic', 'diary_entry', 'dynamic_bio_update', 'enable_virtual_npc', 'disable_virtual_npc', 'activate_voice_effect', 'deactivate_voice_effect'], {
	player_thought: 'Your character thinks',
	player_dialogue: 'Your character speaks',
	npc_thought: 'A character thinks',
	direct_narration: 'Narration',
	persistent_generic: 'Tell characters, unspoken',
	diary_entry: 'A diary entry',
	dynamic_bio_update: 'A bio update',
	enable_virtual_npc: 'Switch on a virtual NPC',
	disable_virtual_npc: 'Switch off a virtual NPC',
	activate_voice_effect: 'Switch on a voice effect',
	deactivate_voice_effect: 'Switch off a voice effect'
});
const scoped = ['diary_entry', 'dynamic_bio_update', 'npc_thought', 'activate_voice_effect', 'deactivate_voice_effect'];
const audiences = options(['player', 'originator', 'target', 'originator_or_target', 'everyone', 'nearby_npcs', 'enabled_virtual_npcs'], {
	player: 'You',
	originator: 'Whoever caused it',
	target: 'Whoever it happened to',
	originator_or_target: 'Either of them',
	everyone: 'Everyone',
	nearby_npcs: 'Characters nearby',
	enabled_virtual_npcs: 'Virtual NPCs that are on'
});
const scopes = options(['triggering_actor', 'player', 'all_pinned_actors', 'all_nearby_actors', 'all_enabled_virtual_npcs'], {
	triggering_actor: 'Whoever set it off',
	player: 'You',
	all_pinned_actors: 'Every pinned character',
	all_nearby_actors: 'Every character nearby',
	all_enabled_virtual_npcs: 'Every virtual NPC that’s on'
});
const andOr = options(['AND', 'OR'], { AND: 'All of them', OR: 'Any of them' });

function eventRule(base: string): Block[] {
	return [
		{
			fields: [
				{ key: `${base}.eventType`, label: 'Event', choices: options(triggerEvents), hint: 'An event SkyrimNet records, or * for every event.' },
				{ key: `${base}.logicalOperator`, label: 'Conditions that must match', type: 'select', choices: andOr }
			]
		},
		{
			title: 'Conditions',
			rows: `${base}.schemaConditions`,
			add: 'condition',
			blank: () => ({ fieldPath: '', operator: 'equals', value: '' }),
			empty: 'None: every event of this kind sets it off.',
			item: (c) => [
				{
					fields: [
						{ key: `${c}.fieldPath`, label: 'Field', hint: 'Of the event, such as spell or victim.' },
						{ key: `${c}.operator`, label: 'Test', type: 'select', choices: options(operators) },
						{ key: `${c}.value`, label: 'Value', type: 'any' },
						{ key: `${c}.caseSensitive`, label: 'Case matters', type: 'check' }
					]
				}
			]
		}
	];
}

const triggers: FormKind = {
	noun: 'trigger',
	nameKey: 'name',
	blank: (name) => ({ name, description: '', eventCriteria: { eventType: '' }, response: { type: 'player_thought', content: '' }, audience: 'player', enabled: true }),
	blocks(d, ctx) {
		const rt = String(getAt(d, 'response.type') ?? '');
		const several = Array.isArray(d.eventCriteria);
		return [
			{
				fields: [
					{ key: 'name', label: 'Name', readonly: true },
					{ key: 'enabled', label: 'On', type: 'check' },
					{ key: 'description', label: 'Description', type: 'area', rows: 2, wide: true }
				]
			},
			...(several ? [{ title: 'When any of these happens', rows: 'eventCriteria', add: 'event', blank: () => ({ eventType: '' }), item: (b: string) => eventRule(b) } as RowsBlock] : [{ title: 'When', fields: [] } as FieldBlock, ...eventRule('eventCriteria')]),
			{
				title: 'Then',
				fields: [
					{ key: 'response.type', label: 'Response', type: 'select', choices: responses },
					{ key: 'audience', label: 'Who hears it', type: 'select', choices: audiences },
					scoped.includes(rt) && { key: 'response.targetScope', label: 'For whom', type: 'select', choices: scopes },
					getAt(d, 'response.targetScope') === 'all_nearby_actors' && { key: 'response.nearbyRadius', label: 'Nearby within', type: 'num', hint: 'In game units.' },
					/virtual_npc/.test(rt) && { key: 'response.targetName', label: 'Virtual NPC’s name' },
					rt === 'activate_voice_effect' && { key: 'response.effectId', label: 'Voice effect', choices: options(ctx.effects), hint: 'The effect’s id.' },
					!/virtual_npc|deactivate_voice/.test(rt) && { key: 'response.content', label: 'What it says', type: 'area', rows: 4, keep: true, wide: true, hint: 'A template: {{ originator }}, {{ target }}, {{ event_json.field }}, {{ player_name }}…' }
				]
			},
			{
				title: 'Timing',
				fields: [
					{ key: 'probability', label: 'Chance', type: 'num', hint: 'From 0 to 1.' },
					{ key: 'cooldownSeconds', label: 'Cooldown', type: 'num', hint: 'In seconds.' },
					{ key: 'priority', label: 'Priority', type: 'num' },
					{ key: 'interrupt', label: 'Interrupts whoever is speaking', type: 'check' }
				]
			}
		];
	}
};

const actions: FormKind = {
	noun: 'action',
	nameKey: 'name',
	blank: (name) => ({ name, description: '', enabled: true, questEditorId: '', scriptName: '', executionFunctionName: '', parameterMapping: [{ type: 'speaker', name: 'akActor', description: 'the actor doing it' }], eligibilityRules: [] }),
	blocks(d, ctx) {
		const noCall = !d.questEditorId && !d.scriptName && !d.executionFunctionName;
		return [
			{
				fields: [
					{ key: 'name', label: 'Name', readonly: true },
					{ key: 'enabled', label: 'On', type: 'check' },
					{ key: 'description', label: 'Description', type: 'area', rows: 3, wide: true, hint: 'What the AI reads to decide when to use it.' },
					{ key: 'customCategory', label: 'Category', choices: options(ctx.categories), hint: 'Actions in a category are offered together.' },
					{ key: 'defaultPriority', label: 'Priority', type: 'num' },
					'eventString' in d && { key: 'eventString', label: 'Event text', type: 'area', rows: 2, wide: true, hint: 'Recorded when it runs, such as {{ actor.name }} gave {{ player.name }} a hug.' },
					'useConfirmationWindow' in d && { key: 'useConfirmationWindow', label: 'Ask before running it', type: 'check' },
					'isShortLived' in d && { key: 'isShortLived', label: 'Short-lived', type: 'check' }
				]
			},
			{
				title: 'Papyrus call',
				note: noCall ? 'No call: this action is a category. The AI picks it, then chooses one of the category’s actions.' : undefined,
				fields: [
					{ key: 'questEditorId', label: 'Quest' },
					{ key: 'scriptName', label: 'Script' },
					{ key: 'executionFunctionName', label: 'Function' }
				]
			},
			{
				title: 'Parameters, in the function’s order',
				rows: 'parameterMapping',
				add: 'parameter',
				blank: () => ({ type: 'dynamic', name: '', description: '' }),
				item: (p) => {
					const t = getAt(d, `${p}.type`);
					const dyn = t === 'dynamic' || t === 'dynamic_uuid';
					return [
						{
							fields: [
								{
									key: `${p}.type`,
									label: 'From',
									type: 'select',
									choices: options(['speaker', 'dynamic', 'dynamic_uuid', 'static'], { speaker: 'The speaker', dynamic: 'The AI fills it in', dynamic_uuid: 'The AI picks a character, passed as its UUID', static: 'A fixed value' })
								},
								{ key: `${p}.name`, label: 'Name' },
								t === 'static' && { key: `${p}.value`, label: 'Value', type: 'any' },
								{ key: `${p}.description`, label: 'Description', type: 'area', rows: 2, wide: true, hint: dyn ? 'Tells the AI what to put here.' : undefined },
								dyn && { key: `${p}.optional`, label: 'The AI may leave it out', type: 'check' },
								dyn && 'default' in ((getAt(d, p) as Obj) ?? {}) && { key: `${p}.default`, label: 'Default', type: 'any' },
								dyn && { key: `${p}.useNearbyActors`, label: 'Choose from characters nearby', type: 'check', on: true }
							]
						}
					];
				}
			},
			{
				title: 'Who can use it: every rule must pass',
				rows: 'eligibilityRules',
				add: 'rule',
				blank: () => ({ conditions: [{ decoratorName: '', arguments: ['currentActor'], comparisonOperator: '==', expectedValue: true }], logicalOperator: 'AND', required: true }),
				empty: 'None: anyone can.',
				item: (r) => [
					{
						note: getAt(d, `${r}.required`) === false ? 'SkyrimNet ignores a rule that isn’t required.' : undefined,
						fields: [
							{ key: `${r}.logicalOperator`, label: 'Conditions that must pass', type: 'select', choices: andOr },
							{ key: `${r}.required`, label: 'Required', type: 'check' }
						]
					},
					{
						title: 'Conditions',
						rows: `${r}.conditions`,
						add: 'condition',
						blank: () => ({ decoratorName: '', arguments: ['currentActor'], comparisonOperator: '==', expectedValue: true }),
						item: (c) => [
							{
								fields: [
									{ key: `${c}.decoratorName`, label: 'Decorator', choices: options(ctx.decorators) },
									{ key: `${c}.arguments`, label: 'Arguments', type: 'csv', hint: 'Separated by commas: currentActor (the speaker), player…' },
									{ key: `${c}.comparisonOperator`, label: 'Test', type: 'select', choices: options(['==', '!=', '>', '<', '>=', '<=', 'contains', 'not_contains', 'is_null', 'is_not_null', 'is_empty', 'is_not_empty']) },
									{ key: `${c}.expectedValue`, label: 'Expected', type: 'any' }
								]
							}
						]
					}
				]
			}
		];
	}
};

const filterKinds: Record<string, string> = { actor: 'Who takes part (actor)', memory: 'Who forms memories (memory)', dialogue_rule: 'Rewrite what’s said (dialogue_rule)', tts_rule: 'Rewrite what’s spoken aloud (tts_rule)' };

const filters: FormKind = {
	noun: 'filter',
	blank: (name) => ({ kind: 'actor', name, description: '' }),
	blocks(d) {
		const rule = d.kind === 'dialogue_rule' || d.kind === 'tts_rule';
		const kind: FieldSpec = { key: 'kind', label: 'Kind', type: 'select', choices: options(Object.keys(filterKinds), filterKinds) };
		if (rule)
			return [
				{
					fields: [
						kind,
						{ key: 'enabled', label: 'On', type: 'check' },
						{ key: 'id', label: 'Id', readonly: true },
						{ key: 'priority', label: 'Priority', type: 'num', hint: 'Lower runs first; 100 when empty.' },
						{ key: 'pattern', label: 'Pattern', type: 'area', rows: 2, wide: true, hint: 'A regular expression.' },
						{ key: 'replacement', label: 'Replacement', keep: true, wide: true, hint: 'Empty removes what matches.' }
					]
				}
			];
		return [
			{ fields: [kind, { key: 'enabled', label: 'On', type: 'check' }, { key: 'name', label: 'Name' }, { key: 'description', label: 'Description', type: 'area', rows: 2, wide: true }] },
			{
				title: 'Lists, added to your own filter lists',
				fields: ['Faction', 'Race', 'Gender'].flatMap((x): FieldSpec[] => [
					{ key: `${x}Whitelist`, label: `${x} whitelist`, type: 'list' },
					{ key: `${x}Blacklist`, label: `${x} blacklist`, type: 'list' }
				])
			}
		];
	}
};

const dialogueKinds: Record<string, string> = { lists: 'Lines allowed and blocked (lists)', instruction: 'When to choose a line (instruction)' };

const dialogueActions: FormKind = {
	noun: 'dialogue action',
	blank: (name) => ({ kind: 'instruction', key: name, name: '', text: '' }),
	blocks(d) {
		const kind: FieldSpec = { key: 'kind', label: 'Kind', type: 'select', choices: options(Object.keys(dialogueKinds), dialogueKinds) };
		if (d.kind === 'instruction')
			return [
				{
					fields: [
						kind,
						{ key: 'enabled', label: 'On', type: 'check' },
						{ key: 'key', label: 'Dialogue line', readonly: true },
						{ key: 'name', label: 'Name' },
						{ key: 'category', label: 'Category', type: 'select', choices: options(['quest', 'follower', 'merchant', 'trainer', 'carriage', 'innkeeper', 'bard', 'marriage', 'crime', 'other']) },
						{ key: 'text', label: 'Instruction', type: 'area', rows: 3, wide: true, hint: 'Tells the AI when to choose this line.' }
					]
				}
			];
		return [
			{
				fields: [kind, { key: 'enabled', label: 'On', type: 'check' }, { key: 'name', label: 'Name' }, { key: 'description', label: 'Description', type: 'area', rows: 2, wide: true }, { key: 'whitelist', label: 'Always allowed lines', type: 'list' }, { key: 'blacklist', label: 'Lines never said', type: 'list' }]
			}
		];
	}
};

const entities: FormKind = {
	noun: 'entity',
	blank: (name) => ({ entityName: name, displayName: name, voiceId: '', ttsOverride: '', language: '', conversationMode: 'private', enabled: true }),
	blocks: () => [
		{
			fields: [
				{ key: 'entityName', label: 'Name', hint: 'How SkyrimNet and triggers refer to it.' },
				{ key: 'enabled', label: 'On', type: 'check' },
				{ key: 'displayName', label: 'Display name' },
				{ key: 'voiceId', label: 'Voice', hint: 'A voice type from your voice setup.' },
				{ key: 'ttsOverride', label: 'Voice engine', keep: true },
				{ key: 'language', label: 'Language', keep: true },
				{
					key: 'conversationMode',
					label: 'Conversation mode',
					type: 'select',
					choices: [
						{ value: 'private', label: 'Private' },
						{ value: 'public', label: 'Public' }
					],
					hint: 'Private: only you hear it. Public: characters nearby hear it and react too.'
				}
			]
		}
	]
};

// The kinds with forms. A record whose kind has none, or whose layout the
// form doesn't cover, is edited as YAML.
export const formKinds: Record<string, FormKind> = { triggers, actions, filters, dialogue_actions: dialogueActions, entities };

// fits is whether a form can edit a record: a filter's or dialogue
// action's kind must be one the form knows.
export function fits(kind: string, d: Obj | null): boolean {
	if (!d || !formKinds[kind]) return false;
	if (kind === 'filters') return typeof d.kind === 'string' && d.kind in filterKinds;
	if (kind === 'dialogue_actions') return typeof d.kind === 'string' && d.kind in dialogueKinds;
	return true;
}
