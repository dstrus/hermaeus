// How characters take turns to speak, think, and react: the settings files
// for conversation, NPC thoughts, and the player's own reactions.
import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const waiting = m.features.conversation.waiting;
const wait = (key: string, t: { label: string; what: string }) => ({ key: `ttl_seconds.${key}`, label: t.label, what: waiting.dropped({ what: t.what }), kind: 'count' as const, min: 0, max: 120, unit: waiting.unit, zero: waiting.zero });

const c = m.features.conversation;

export const conversation: Feature = {
	id: 'conversation',
	file: 'Conversation',
	title: c.title,
	what: c.what,
	sections: [
		{
			title: c.turnTaking.title,
			fields: [
				{ key: 'max_chain_depth', label: c.turnTaking.maxChainDepth.label, what: c.turnTaking.maxChainDepth.what, kind: 'count', min: 0, max: 20, unit: c.turnTaking.maxChainDepth.unit },
				{ key: 'takeover_before_audio', label: c.turnTaking.takeoverBeforeAudio.label, what: c.turnTaking.takeoverBeforeAudio.what, kind: 'switch' },
				{ key: 'lead_seconds', label: c.turnTaking.leadSeconds.label, what: c.turnTaking.leadSeconds.what, kind: 'number', min: 0, max: 10, zero: c.turnTaking.leadSeconds.zero },
				{ key: 'generation_timeout_seconds', label: c.turnTaking.generationTimeoutSeconds.label, what: c.turnTaking.generationTimeoutSeconds.what, kind: 'count', min: 10, max: 300, unit: c.turnTaking.generationTimeoutSeconds.unit }
			]
		},
		{
			title: waiting.title,
			what: waiting.what,
			fields: [wait('player', waiting.ttlSecondsPlayer), wait('directed', waiting.ttlSecondsDirected), wait('urgent', waiting.ttlSecondsUrgent), wait('reaction', waiting.ttlSecondsReaction), wait('continuation', waiting.ttlSecondsContinuation)]
		}
	]
};

const nt = m.features.npcThoughts;

export const npcThoughts: Feature = {
	id: 'npc-thoughts',
	file: 'NpcThoughts',
	title: nt.title,
	what: nt.what,
	sections: [
		{
			title: nt.general.title,
			fields: [
				{ key: 'enabled', label: nt.general.enabled.label, what: nt.general.enabled.what, kind: 'switch' },
				{ key: 'perNPCCooldownSeconds', label: nt.general.perNPCCooldownSeconds.label, what: nt.general.perNPCCooldownSeconds.what, kind: 'count', min: 0, max: 600, unit: nt.general.perNPCCooldownSeconds.unit },
				{ key: 'eventHistoryCount', label: nt.general.eventHistoryCount.label, what: nt.general.eventHistoryCount.what, kind: 'count', min: 5, max: 200, unit: nt.general.eventHistoryCount.unit }
			]
		},
		{
			title: nt.inDialogue.title,
			fields: [{ key: 'allowInlineInDialogue', label: nt.inDialogue.allowInlineInDialogue.label, what: nt.inDialogue.allowInlineInDialogue.what, kind: 'switch' }]
		}
	]
};

const pr = m.features.playerReactions;
const events = pr.events.eventConfigs;

export const playerReactions: Feature = {
	id: 'player-reactions',
	file: 'PlayerDialogue',
	title: pr.title,
	what: pr.what,
	sections: [
		{
			title: pr.general.title,
			fields: [{ key: 'enabled', label: pr.general.enabled.label, what: pr.general.enabled.what, kind: 'switch' }]
		},
		{
			title: pr.events.title,
			fields: [
				{
					key: 'eventConfigs',
					label: events.label,
					what: events.what,
					kind: 'rows',
					noun: events.noun,
					add: events.add,
					cols: [
						{ key: 'eventType', label: events.cols.eventType.label, kind: 'text', placeholder: 'combat' },
						{ key: 'enabled', label: events.cols.enabled.label, kind: 'switch' },
						{
							key: 'responseType',
							label: events.cols.responseType.label,
							kind: 'choice',
							choices: [
								{ value: 'thought', label: events.cols.responseType.choices.thought.label },
								{ value: 'audible', label: events.cols.responseType.choices.audible.label }
							]
						},
						{ key: 'probability', label: events.cols.probability.label, kind: 'percent' },
						{ key: 'cooldownSeconds', label: events.cols.cooldownSeconds.label, kind: 'count', unit: events.cols.cooldownSeconds.unit, min: 0, max: 3600 }
					]
				}
			]
		}
	]
};

const bs = m.features.bardSinging;
const providers = bs.providers;

const musicProviders = [
	{ value: 'openrouter', label: providers.openrouter.label, what: providers.openrouter.what },
	{ value: 'suno', label: providers.suno.label, what: providers.suno.what },
	{ value: 'nano_gpt', label: providers.nanoGpt.label, what: providers.nanoGpt.what },
	{ value: 'minimax', label: providers.minimax.label, what: providers.minimax.what },
	{ value: 'acestep_cloud', label: providers.acestepCloud.label, what: providers.acestepCloud.what },
	{ value: 'acestep_local', label: providers.acestepLocal.label, what: providers.acestepLocal.what }
];

// service is the music service as the key's and model's sentences name it,
// and where is where its key comes from.
const providerKey = (id: string, service: string, where: string) => ({ key: `bard_singing.${id}.api_key`, label: bs.providerKey.label, what: bs.providerKey.what({ service, where }), kind: 'secret' as const });
const providerModel = (id: string, service: string, example: string) => ({ key: `bard_singing.${id}.model`, label: bs.providerModel.label, what: bs.providerModel.what({ service }), kind: 'text' as const, placeholder: example });

const general = bs.general;
const composing = bs.composing;

export const bardSinging: Feature = {
	id: 'bard-singing',
	file: 'BardSinging',
	title: bs.title,
	what: bs.what,
	see: { text: bs.see.text, label: bs.seeLabel, path: 'songs' },
	sections: [
		{
			title: general.title,
			fields: [
				{ key: 'bard_singing.enabled', label: general.bardSingingEnabled.label, what: general.bardSingingEnabled.what, kind: 'switch' },
				{ key: 'bard_singing.generation.compose_enabled', label: general.bardSingingGenerationComposeEnabled.label, what: general.bardSingingGenerationComposeEnabled.what, kind: 'switch' },
				{ key: 'bard_singing.playback.volume', label: general.bardSingingPlaybackVolume.label, what: general.bardSingingPlaybackVolume.what, kind: 'percent' }
			]
		},
		{
			title: composing.title,
			fields: [
				{ key: 'bard_singing.provider', label: composing.bardSingingProvider.label, what: composing.bardSingingProvider.what, kind: 'choice', custom: true, choices: musicProviders },
				{
					key: 'bard_singing.backup_provider',
					label: composing.bardSingingBackupProvider.label,
					what: composing.bardSingingBackupProvider.what,
					kind: 'choice',
					custom: true,
					choices: [...musicProviders, { value: 'none', label: composing.bardSingingBackupProvider.choices.none.label, what: composing.bardSingingBackupProvider.choices.none.what }]
				},
				{ key: 'bard_singing.generation.interval_hours', label: composing.bardSingingGenerationIntervalHours.label, what: composing.bardSingingGenerationIntervalHours.what, kind: 'number', min: 1, max: 168, unit: composing.bardSingingGenerationIntervalHours.unit },
				{ key: 'bard_singing.generation.max_cached_songs', label: composing.bardSingingGenerationMaxCachedSongs.label, what: composing.bardSingingGenerationMaxCachedSongs.what, kind: 'count', min: 1, max: 100, unit: composing.bardSingingGenerationMaxCachedSongs.unit },
				{ key: 'bard_singing.generation.default_style', label: composing.bardSingingGenerationDefaultStyle.label, what: composing.bardSingingGenerationDefaultStyle.what, kind: 'text', placeholder: composing.bardSingingGenerationDefaultStyle.placeholder },
				{ key: 'bard_singing.generation.negative_tags', label: composing.bardSingingGenerationNegativeTags.label, what: composing.bardSingingGenerationNegativeTags.what, kind: 'text', placeholder: composing.bardSingingGenerationNegativeTags.placeholder },
				{ key: 'bard_singing.generation.language', label: composing.bardSingingGenerationLanguage.label, what: composing.bardSingingGenerationLanguage.what, kind: 'text', placeholder: 'en' }
			]
		},
		{ title: providers.openrouter.label, fields: [providerKey('openrouter', providers.openrouter.service, providers.openrouter.where), providerModel('openrouter', providers.openrouter.service, 'google/lyria-3-pro-preview')] },
		{
			title: providers.suno.label,
			fields: [
				providerKey('suno', providers.suno.service, providers.suno.where),
				{ key: 'bard_singing.suno.base_url', label: providers.suno.bardSingingSunoBaseUrl.label, what: providers.suno.bardSingingSunoBaseUrl.what, kind: 'text' },
				providerModel('suno', providers.suno.modelService, 'V5')
			]
		},
		{ title: providers.nanoGpt.label, fields: [providerKey('nano_gpt', providers.nanoGpt.service, providers.nanoGpt.where), providerModel('nano_gpt', providers.nanoGpt.service, 'google/lyria-3-pro/music')] },
		{ title: providers.minimax.label, fields: [providerKey('minimax', providers.minimax.service, providers.minimax.where), providerModel('minimax', providers.minimax.service, 'music-3.0-free')] },
		{ title: providers.acestepCloud.label, fields: [providerKey('acestep_cloud', providers.acestepCloud.service, providers.acestepCloud.where), providerModel('acestep_cloud', providers.acestepCloud.service, 'acemusic/acestep-v1.5-turbo')] },
		{
			title: providers.acestepLocal.label,
			fields: [
				{ key: 'bard_singing.acestep_local.base_url', label: providers.acestepLocal.bardSingingAcestepLocalBaseUrl.label, what: providers.acestepLocal.bardSingingAcestepLocalBaseUrl.what, kind: 'text', placeholder: 'http://localhost:8002/v1/chat/completions' },
				providerModel('acestep_local', providers.acestepLocal.service, 'acemusic/acestep-v1.5-turbo')
			]
		}
	]
};
