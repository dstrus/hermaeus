// Forms for records: each kind's fields, in blocks, over the record as an
// object. A field names its value by a dotted path into the object
// ("response.type", "parameterMapping.0.name"); a block of rows repeats its
// blocks for each item of a list. The form edits the object in place; the
// server merges it back into the file, so whatever the form leaves alone,
// and comments, stay as they were.

import { m } from '#lib/copy/messages.ts';

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

const f = m.records.forms;

// A choice's label is its message, named by its value in camelCase
// (player_thought is playerThought), or else the value itself.
const camel = (v: string) => v.toLowerCase().replace(/_(.)/g, (_, c: string) => c.toUpperCase());
const options = (values: string[], labels: Record<string, string> = {}): Choice[] => values.map((value) => ({ value, label: labels[camel(value)] ?? value }));

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
const responses = options(['player_thought', 'player_dialogue', 'npc_thought', 'direct_narration', 'persistent_generic', 'diary_entry', 'dynamic_bio_update', 'enable_virtual_npc', 'disable_virtual_npc', 'activate_voice_effect', 'deactivate_voice_effect'], f.triggers.response.type.choices);
const scoped = ['diary_entry', 'dynamic_bio_update', 'npc_thought', 'activate_voice_effect', 'deactivate_voice_effect'];
const audiences = options(['player', 'originator', 'target', 'originator_or_target', 'everyone', 'nearby_npcs', 'enabled_virtual_npcs'], f.triggers.audience.choices);
const scopes = options(['triggering_actor', 'player', 'all_pinned_actors', 'all_nearby_actors', 'all_enabled_virtual_npcs'], f.triggers.response.targetScope.choices);
const andOr = options(['AND', 'OR'], f.andOr);

function eventRule(base: string): Block[] {
	const t = f.triggers;
	return [
		{
			fields: [
				{ key: `${base}.eventType`, label: t.eventType.label, choices: options(triggerEvents), hint: t.eventType.hint },
				{ key: `${base}.logicalOperator`, label: t.logicalOperator.label, type: 'select', choices: andOr }
			]
		},
		{
			title: t.conditions.title,
			rows: `${base}.schemaConditions`,
			add: t.conditions.add,
			blank: () => ({ fieldPath: '', operator: 'equals', value: '' }),
			empty: t.conditions.empty,
			item: (c) => [
				{
					fields: [
						{ key: `${c}.fieldPath`, label: t.conditions.fieldPath.label, hint: t.conditions.fieldPath.hint },
						{ key: `${c}.operator`, label: t.conditions.operator.label, type: 'select', choices: options(operators) },
						{ key: `${c}.value`, label: t.conditions.value.label, type: 'any' },
						{ key: `${c}.caseSensitive`, label: t.conditions.caseSensitive.label, type: 'check' }
					]
				}
			]
		}
	];
}

const triggers: FormKind = {
	noun: f.triggers.noun,
	nameKey: 'name',
	blank: (name) => ({ name, description: '', eventCriteria: { eventType: '' }, response: { type: 'player_thought', content: '' }, audience: 'player', enabled: true }),
	blocks(d, ctx) {
		const t = f.triggers;
		const r = t.response;
		const rt = String(getAt(d, 'response.type') ?? '');
		const several = Array.isArray(d.eventCriteria);
		return [
			{
				fields: [
					{ key: 'name', label: t.name.label, readonly: true },
					{ key: 'enabled', label: t.enabled.label, type: 'check' },
					{ key: 'description', label: t.description.label, type: 'area', rows: 2, wide: true }
				]
			},
			...(several ? [{ title: t.whenAny.title, rows: 'eventCriteria', add: t.whenAny.add, blank: () => ({ eventType: '' }), item: (b: string) => eventRule(b) } as RowsBlock] : [{ title: t.when, fields: [] } as FieldBlock, ...eventRule('eventCriteria')]),
			{
				title: t.then,
				fields: [
					{ key: 'response.type', label: r.type.label, type: 'select', choices: responses },
					{ key: 'audience', label: t.audience.label, type: 'select', choices: audiences },
					scoped.includes(rt) && { key: 'response.targetScope', label: r.targetScope.label, type: 'select', choices: scopes },
					getAt(d, 'response.targetScope') === 'all_nearby_actors' && { key: 'response.nearbyRadius', label: r.nearbyRadius.label, type: 'num', hint: r.nearbyRadius.hint },
					/virtual_npc/.test(rt) && { key: 'response.targetName', label: r.targetName.label },
					rt === 'activate_voice_effect' && { key: 'response.effectId', label: r.effectId.label, choices: options(ctx.effects), hint: r.effectId.hint },
					!/virtual_npc|deactivate_voice/.test(rt) && { key: 'response.content', label: r.content.label, type: 'area', rows: 4, keep: true, wide: true, hint: r.content.hint }
				]
			},
			{
				title: t.timing,
				fields: [
					{ key: 'probability', label: t.probability.label, type: 'num', hint: t.probability.hint },
					{ key: 'cooldownSeconds', label: t.cooldownSeconds.label, type: 'num', hint: t.cooldownSeconds.hint },
					{ key: 'priority', label: t.priority.label, type: 'num' },
					{ key: 'interrupt', label: t.interrupt.label, type: 'check' }
				]
			}
		];
	}
};

const actions: FormKind = {
	noun: f.actions.noun,
	nameKey: 'name',
	blank: (name) => ({ name, description: '', enabled: true, questEditorId: '', scriptName: '', executionFunctionName: '', parameterMapping: [{ type: 'speaker', name: 'akActor', description: 'the actor doing it' }], eligibilityRules: [] }),
	blocks(d, ctx) {
		const a = f.actions;
		const noCall = !d.questEditorId && !d.scriptName && !d.executionFunctionName;
		return [
			{
				fields: [
					{ key: 'name', label: a.name.label, readonly: true },
					{ key: 'enabled', label: a.enabled.label, type: 'check' },
					{ key: 'description', label: a.description.label, type: 'area', rows: 3, wide: true, hint: a.description.hint },
					{ key: 'customCategory', label: a.customCategory.label, choices: options(ctx.categories), hint: a.customCategory.hint },
					{ key: 'defaultPriority', label: a.defaultPriority.label, type: 'num' },
					'eventString' in d && { key: 'eventString', label: a.eventString.label, type: 'area', rows: 2, wide: true, hint: a.eventString.hint },
					'useConfirmationWindow' in d && { key: 'useConfirmationWindow', label: a.useConfirmationWindow.label, type: 'check' },
					'isShortLived' in d && { key: 'isShortLived', label: a.isShortLived.label, type: 'check' }
				]
			},
			{
				title: a.call.title,
				note: noCall ? a.call.note : undefined,
				fields: [
					{ key: 'questEditorId', label: a.questEditorId.label },
					{ key: 'scriptName', label: a.scriptName.label },
					{ key: 'executionFunctionName', label: a.executionFunctionName.label }
				]
			},
			{
				title: a.parameters.title,
				rows: 'parameterMapping',
				add: a.parameters.add,
				blank: () => ({ type: 'dynamic', name: '', description: '' }),
				item: (p) => {
					const pm = a.parameters;
					const t = getAt(d, `${p}.type`);
					const dyn = t === 'dynamic' || t === 'dynamic_uuid';
					return [
						{
							fields: [
								{
									key: `${p}.type`,
									label: pm.type.label,
									type: 'select',
									choices: options(['speaker', 'dynamic', 'dynamic_uuid', 'static'], pm.type.choices)
								},
								{ key: `${p}.name`, label: pm.name.label },
								t === 'static' && { key: `${p}.value`, label: pm.value.label, type: 'any' },
								{ key: `${p}.description`, label: pm.description.label, type: 'area', rows: 2, wide: true, hint: dyn ? pm.description.hint : undefined },
								dyn && { key: `${p}.optional`, label: pm.optional.label, type: 'check' },
								dyn && 'default' in ((getAt(d, p) as Obj) ?? {}) && { key: `${p}.default`, label: pm.default.label, type: 'any' },
								dyn && { key: `${p}.useNearbyActors`, label: pm.useNearbyActors.label, type: 'check', on: true }
							]
						}
					];
				}
			},
			{
				title: a.rules.title,
				rows: 'eligibilityRules',
				add: a.rules.add,
				blank: () => ({ conditions: [{ decoratorName: '', arguments: ['currentActor'], comparisonOperator: '==', expectedValue: true }], logicalOperator: 'AND', required: true }),
				empty: a.rules.empty,
				item: (r) => [
					{
						note: getAt(d, `${r}.required`) === false ? a.rules.note : undefined,
						fields: [
							{ key: `${r}.logicalOperator`, label: a.rules.logicalOperator.label, type: 'select', choices: andOr },
							{ key: `${r}.required`, label: a.rules.required.label, type: 'check' }
						]
					},
					{
						title: a.conditions.title,
						rows: `${r}.conditions`,
						add: a.conditions.add,
						blank: () => ({ decoratorName: '', arguments: ['currentActor'], comparisonOperator: '==', expectedValue: true }),
						item: (c) => [
							{
								fields: [
									{ key: `${c}.decoratorName`, label: a.conditions.decoratorName.label, choices: options(ctx.decorators) },
									{ key: `${c}.arguments`, label: a.conditions.arguments.label, type: 'csv', hint: a.conditions.arguments.hint },
									{ key: `${c}.comparisonOperator`, label: a.conditions.comparisonOperator.label, type: 'select', choices: options(['==', '!=', '>', '<', '>=', '<=', 'contains', 'not_contains', 'is_null', 'is_not_null', 'is_empty', 'is_not_empty']) },
									{ key: `${c}.expectedValue`, label: a.conditions.expectedValue.label, type: 'any' }
								]
							}
						]
					}
				]
			}
		];
	}
};

const fk = f.filters.kind.choices;
const filterKinds: Record<string, string> = { actor: fk.actor, memory: fk.memory, dialogue_rule: fk.dialogueRule, tts_rule: fk.ttsRule };

const filters: FormKind = {
	noun: f.filters.noun,
	blank: (name) => ({ kind: 'actor', name, description: '' }),
	blocks(d) {
		const t = f.filters;
		const rule = d.kind === 'dialogue_rule' || d.kind === 'tts_rule';
		const kind: FieldSpec = { key: 'kind', label: t.kind.label, type: 'select', choices: options(Object.keys(filterKinds), t.kind.choices) };
		if (rule)
			return [
				{
					fields: [
						kind,
						{ key: 'enabled', label: t.enabled.label, type: 'check' },
						{ key: 'id', label: t.id.label, readonly: true },
						{ key: 'priority', label: t.priority.label, type: 'num', hint: t.priority.hint },
						{ key: 'pattern', label: t.pattern.label, type: 'area', rows: 2, wide: true, hint: t.pattern.hint },
						{ key: 'replacement', label: t.replacement.label, keep: true, wide: true, hint: d.kind === 'dialogue_rule' ? t.replacement.dialogueHint : t.replacement.hint }
					]
				}
			];
		// Each list's key, by its name in the file, with its labels.
		const lists: [string, string, string][] = [
			['Faction', t.factionWhitelist.label, t.factionBlacklist.label],
			['Race', t.raceWhitelist.label, t.raceBlacklist.label],
			['Gender', t.genderWhitelist.label, t.genderBlacklist.label]
		];
		return [
			{ fields: [kind, { key: 'enabled', label: t.enabled.label, type: 'check' }, { key: 'name', label: t.name.label }, { key: 'description', label: t.description.label, type: 'area', rows: 2, wide: true }] },
			{
				title: t.lists,
				fields: lists.flatMap(([x, white, black]): FieldSpec[] => [
					{ key: `${x}Whitelist`, label: white, type: 'list' },
					{ key: `${x}Blacklist`, label: black, type: 'list' }
				])
			}
		];
	}
};

const dialogueKinds: Record<string, string> = { lists: f.dialogueActions.kind.choices.lists, instruction: f.dialogueActions.kind.choices.instruction };

const dialogueActions: FormKind = {
	noun: f.dialogueActions.noun,
	blank: (name) => ({ kind: 'instruction', key: name, name: '', text: '' }),
	blocks(d) {
		const t = f.dialogueActions;
		const kind: FieldSpec = { key: 'kind', label: t.kind.label, type: 'select', choices: options(Object.keys(dialogueKinds), t.kind.choices) };
		if (d.kind === 'instruction')
			return [
				{
					fields: [
						kind,
						{ key: 'enabled', label: t.enabled.label, type: 'check' },
						{ key: 'key', label: t.key.label, readonly: true },
						{ key: 'name', label: t.name.label },
						{ key: 'category', label: t.category.label, type: 'select', choices: options(['quest', 'follower', 'merchant', 'trainer', 'carriage', 'innkeeper', 'bard', 'marriage', 'crime', 'other']) },
						{ key: 'text', label: t.text.label, type: 'area', rows: 3, wide: true, hint: t.text.hint }
					]
				}
			];
		return [
			{
				fields: [kind, { key: 'enabled', label: t.enabled.label, type: 'check' }, { key: 'name', label: t.name.label }, { key: 'description', label: t.description.label, type: 'area', rows: 2, wide: true }, { key: 'whitelist', label: t.whitelist.label, type: 'list' }, { key: 'blacklist', label: t.blacklist.label, type: 'list' }]
			}
		];
	}
};

const e = f.entities;

const entities: FormKind = {
	noun: e.noun,
	blank: (name) => ({ entityName: name, displayName: name, voiceId: '', ttsOverride: '', language: '', conversationMode: 'private', enabled: true }),
	blocks: () => [
		{
			fields: [
				{ key: 'entityName', label: e.entityName.label, hint: e.entityName.hint },
				{ key: 'enabled', label: e.enabled.label, type: 'check' },
				{ key: 'displayName', label: e.displayName.label },
				{ key: 'voiceId', label: e.voiceId.label, hint: e.voiceId.hint },
				{ key: 'ttsOverride', label: e.ttsOverride.label, keep: true },
				{ key: 'language', label: e.language.label, keep: true },
				{
					key: 'conversationMode',
					label: e.conversationMode.label,
					type: 'select',
					choices: [
						{ value: 'private', label: e.conversationMode.choices.private },
						{ value: 'public', label: e.conversationMode.choices.public }
					],
					hint: e.conversationMode.hint
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
