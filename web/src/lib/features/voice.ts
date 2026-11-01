// How SkyrimNet gives characters voices: Speech, from SkyrimNet.yaml, picks
// the engine and covers what every engine shares (the narrator, silent
// NPCs, subtitles); each engine has a page of its own file's settings.
import { m } from '#lib/copy/messages.ts';
import type { Feature, Field, Section } from './schema.ts';

const eng = m.features.engines;

// An engine as SkyrimNet names it in tts.engine, and its settings file.
export interface Engine {
	id: string; // the page's path under voice/engines/
	value: string; // tts.engine's value
	file: string;
	label: string;
	what: string;
	local: boolean; // runs on the player's own machine, rather than as a service
}

export const engines: Engine[] = [
	{ id: 'pockettts', value: 'pocket_tts', file: 'PocketTTS', label: eng.pockettts.label, what: eng.pockettts.what, local: true },
	{ id: 'piper', value: 'piper', file: 'Piper', label: eng.piper.label, what: eng.piper.what, local: true },
	{ id: 'inworld', value: 'inworld', file: 'Inworld', label: eng.inworld.label, what: eng.inworld.what, local: false },
	{ id: 'elevenlabs', value: 'elevenlabs', file: 'ElevenLabs', label: eng.elevenlabs.label, what: eng.elevenlabs.what, local: false },
	{ id: 'fateless', value: 'fateless', file: 'FatelessTTS', label: eng.fateless.label, what: eng.fateless.what, local: false },
	{ id: 'xtts', value: 'xtts', file: 'XTTS', label: eng.xtts.label, what: eng.xtts.what, local: true },
	{ id: 'zonos', value: 'zonos', file: 'Zonos', label: eng.zonos.label, what: eng.zonos.what, local: true },
	{ id: 'chatterbox', value: 'chatterbox', file: 'Chatterbox', label: eng.chatterbox.label, what: eng.chatterbox.what, local: true },
	{ id: 'higgs', value: 'higgs', file: 'Higgs', label: eng.higgs.label, what: eng.higgs.what, local: true },
	{ id: 'openai-compatible', value: 'openai_compatible', file: 'OpenAICompatibleTTS', label: eng.openaiCompatible.label, what: eng.openaiCompatible.what, local: true }
];

export const noVoice = 'noop'; // tts.engine when SkyrimNet speaks with no engine

export function engineByValue(value: unknown): Engine | undefined {
	return engines.find((e) => e.value === value);
}

export function engineById(id: string): Engine | undefined {
	return engines.find((e) => e.id === id);
}

const sp = m.features.speech;

export const speech: Feature = {
	id: 'speech',
	file: 'game',
	title: sp.title,
	what: sp.what,
	sections: [
		{
			title: sp.engine.title,
			fields: [
				{
					key: 'tts.engine',
					label: sp.engine.ttsEngine.label,
					what: sp.engine.ttsEngine.what,
					kind: 'choice',
					choices: [
						...engines.map((x) => ({ value: x.value, label: x.label, what: x.local ? sp.engine.ttsEngine.local : sp.engine.ttsEngine.service })),
						{ value: noVoice, label: sp.engine.ttsEngine.choices.noop.label, what: sp.engine.ttsEngine.choices.noop.what }
					]
				},
				{ key: 'tts.maxSegmentWords', label: sp.engine.ttsMaxSegmentWords.label, what: sp.engine.ttsMaxSegmentWords.what, kind: 'count', unit: sp.engine.ttsMaxSegmentWords.unit, zero: sp.engine.ttsMaxSegmentWords.zero }
			]
		},
		{
			title: sp.narrator.title,
			fields: [
				{ key: 'tts.enableNarratorAudioEffect', label: sp.narrator.ttsEnableNarratorAudioEffect.label, what: sp.narrator.ttsEnableNarratorAudioEffect.what, kind: 'switch' },
				{ key: 'tts.enablePlayerThoughtAudioEffect', label: sp.narrator.ttsEnablePlayerThoughtAudioEffect.label, what: sp.narrator.ttsEnablePlayerThoughtAudioEffect.what, kind: 'switch' },
				{ key: 'tts.useKokoroForNarrator', label: sp.narrator.ttsUseKokoroForNarrator.label, what: sp.narrator.ttsUseKokoroForNarrator.what, kind: 'switch' },
				{ key: 'tts.kokoroNarratorVoice', label: sp.narrator.ttsKokoroNarratorVoice.label, what: sp.narrator.ttsKokoroNarratorVoice.what, kind: 'text', placeholder: 'af_bella' },
				{ key: 'tts.kokoroVolumeLevel', label: sp.narrator.ttsKokoroVolumeLevel.label, what: sp.narrator.ttsKokoroVolumeLevel.what, kind: 'percent' }
			]
		},
		{
			title: sp.gameDialogue.title,
			what: sp.gameDialogue.what,
			fields: [
				{ key: 'vanillaDialogue.voicePlayerChoices', was: 'dbvo.enabled', label: sp.gameDialogue.vanillaDialogueVoicePlayerChoices.label, what: sp.gameDialogue.vanillaDialogueVoicePlayerChoices.what, kind: 'switch' },
				{ key: 'vanillaDialogue.voiceSilentNPCs', was: 'dbvo.voiceSilentNPCs', label: sp.gameDialogue.vanillaDialogueVoiceSilentNPCs.label, what: sp.gameDialogue.vanillaDialogueVoiceSilentNPCs.what, kind: 'switch' },
				{ key: 'vanillaDialogue.fixSubtitleDuringTTS', was: 'dbvo.fixSubtitleDuringTTS', label: sp.gameDialogue.vanillaDialogueFixSubtitleDuringTTS.label, what: sp.gameDialogue.vanillaDialogueFixSubtitleDuringTTS.what, kind: 'switch' }
			]
		},
		{
			title: sp.subtitles.title,
			fields: [
				{ key: 'subtitles.enableNPCSubtitles', label: sp.subtitles.subtitlesEnableNPCSubtitles.label, what: sp.subtitles.subtitlesEnableNPCSubtitles.what, kind: 'switch' },
				{ key: 'subtitles.enablePlayerSubtitles', label: sp.subtitles.subtitlesEnablePlayerSubtitles.label, what: sp.subtitles.subtitlesEnablePlayerSubtitles.what, kind: 'switch' },
				{ key: 'subtitles.enableConsoleOutput', label: sp.subtitles.subtitlesEnableConsoleOutput.label, what: sp.subtitles.subtitlesEnableConsoleOutput.what, kind: 'switch' }
			]
		},
		{
			title: sp.volume.title,
			fields: [{ key: 'audio.universalMaxVolume', label: sp.volume.audioUniversalMaxVolume.label, what: sp.volume.audioUniversalMaxVolume.what, kind: 'number', min: 0, max: 1 }]
		}
	]
};

type Text = { label: string; what: string };
const seconds = (key: string, t: Text & { unit: string }): Field => ({ key, label: t.label, what: t.what, kind: 'count', unit: t.unit });
const tune = (key: string, t: Text, min?: number, max?: number): Field => ({ key, label: t.label, what: t.what, kind: 'number', min, max });

// The connection settings every engine file has.
function connection(e: Engine): Section {
	const c = eng.connection;
	const endpoint: Field =
		e.id === 'inworld'
			? { key: 'endpoint', label: c.endpoint.label, what: c.endpoint.whatInworld, kind: 'secret' }
			: { key: 'endpoint', label: c.endpoint.label, what: e.id === 'piper' || e.id === 'pockettts' ? c.endpoint.whatBundled : c.endpoint.what, kind: 'text', placeholder: 'http://localhost:7860' };
	return {
		title: c.title,
		fields: [
			// Fateless speaks through the gateway SkyrimNet's sign-in pairs with.
			...(e.id === 'fateless' ? [] : [endpoint]),
			{ key: 'language', label: c.language.label, what: c.language.what, kind: 'text', placeholder: 'en' },
			seconds('timeout', c.timeout),
			seconds('connect_timeout', c.connectTimeout),
			{ key: 'max_parallel_requests', label: c.maxParallelRequests.label, what: c.maxParallelRequests.what, kind: 'count', unit: c.maxParallelRequests.unit },
			{ key: 'max_retries', label: c.maxRetries.label, what: c.maxRetries.what, kind: 'count', zero: c.maxRetries.zero },
			{ key: 'stream_audio', label: c.streamAudio.label, what: c.streamAudio.what, kind: 'switch' }
		]
	};
}

// What each engine has of its own.
const pocket = eng.pockettts.voice;
const piper = eng.piper.voice;
const fateless = eng.fateless.voice;
const inworld = eng.inworld;
const eleven = eng.elevenlabs;
const zonos = eng.zonos;
const chatterbox = eng.chatterbox;
const higgs = eng.higgs.voice;
const openai = eng.openaiCompatible;
const own: Record<string, Section[]> = {
	pockettts: [
		{
			title: pocket.title,
			fields: [
				{ key: 'max_lsd_steps', label: pocket.maxLsdSteps.label, what: pocket.maxLsdSteps.what, kind: 'count' },
				tune('temperature', pocket.temperature),
				{
					key: 'model_precision',
					label: pocket.modelPrecision.label,
					what: pocket.modelPrecision.what,
					kind: 'choice',
					choices: [
						{ value: 'FP32', label: 'FP32', what: pocket.modelPrecision.choices.fp32.what },
						{ value: 'INT8', label: 'INT8', what: pocket.modelPrecision.choices.int8.what }
					]
				},
				{ key: 'device_index', label: pocket.deviceIndex.label, what: pocket.deviceIndex.what, kind: 'count' }
			]
		}
	],
	piper: [
		{
			title: piper.title,
			fields: [
				{ key: 'enable_preloading', label: piper.enablePreloading.label, what: piper.enablePreloading.what, kind: 'switch' },
				tune('piper_settings.noise_scale', piper.piperSettingsNoiseScale, 0, 5),
				tune('piper_settings.noise_w', piper.piperSettingsNoiseW, 0, 5),
				tune('piper_settings.length_scale', piper.piperSettingsLengthScale, 0.1, 5),
				{ key: 'piper_settings.num_threads', label: piper.piperSettingsNumThreads.label, what: piper.piperSettingsNumThreads.what, kind: 'count' },
				{ key: 'speaker_id', label: piper.speakerId.label, what: piper.speakerId.what, kind: 'count' }
			]
		}
	],
	fateless: [
		{
			title: fateless.title,
			fields: [
				{
					key: 'model',
					label: fateless.model.label,
					what: fateless.model.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'fateless/speech-plus', label: fateless.model.choices.fatelessSpeechPlus.label, what: fateless.model.choices.fatelessSpeechPlus.what },
						{ value: 'fateless/speech', label: fateless.model.choices.fatelessSpeech.label, what: fateless.model.choices.fatelessSpeech.what }
					]
				},
				tune('speed', fateless.speed, 0.5, 1.5),
				tune('temperature', fateless.temperature, 0, 2)
			]
		}
	],
	inworld: [
		{
			title: inworld.account.title,
			fields: [
				{ key: 'api_key_base64', label: inworld.account.apiKeyBase64.label, what: inworld.account.apiKeyBase64.what, kind: 'secret' },
				{ key: 'workspace_id', label: inworld.account.workspaceId.label, what: inworld.account.workspaceId.what, kind: 'secret' }
			]
		},
		{
			title: inworld.voice.title,
			fields: [
				{
					key: 'model_id',
					label: inworld.voice.modelId.label,
					what: inworld.voice.modelId.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'inworld-tts-2', label: 'inworld-tts-2', what: inworld.voice.modelId.choices.inworldTts2.what },
						{ value: 'inworld-tts-2-flash', label: 'inworld-tts-2-flash', what: inworld.voice.modelId.choices.inworldTts2Flash.what }
					]
				},
				tune('temperature', inworld.voice.temperature, 0, 2),
				tune('speaking_rate', inworld.voice.speakingRate),
				{ key: 'sample_rate_hz', label: inworld.voice.sampleRateHz.label, what: inworld.voice.sampleRateHz.what, kind: 'count', min: 8000, max: 48000, unit: inworld.voice.sampleRateHz.unit },
				{ key: 'clone_cache.max_voices', label: inworld.voice.cloneCacheMaxVoices.label, what: inworld.voice.cloneCacheMaxVoices.what, kind: 'count', unit: inworld.voice.cloneCacheMaxVoices.unit, zero: inworld.voice.cloneCacheMaxVoices.zero }
			]
		}
	],
	elevenlabs: [
		{
			title: eleven.account.title,
			fields: [
				{ key: 'api_key', label: eleven.account.apiKey.label, what: eleven.account.apiKey.what, kind: 'secret' },
				{ key: 'clone_cache.shared_account', label: eleven.account.cloneCacheSharedAccount.label, what: eleven.account.cloneCacheSharedAccount.what, kind: 'switch' },
				{ key: 'clone_cache.max_voices', label: eleven.account.cloneCacheMaxVoices.label, what: eleven.account.cloneCacheMaxVoices.what, kind: 'count', unit: eleven.account.cloneCacheMaxVoices.unit, zero: eleven.account.cloneCacheMaxVoices.zero }
			]
		},
		{
			title: eleven.voice.title,
			fields: [
				{
					key: 'model_id',
					label: eleven.voice.modelId.label,
					what: eleven.voice.modelId.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'eleven_v4_turbo', label: 'eleven_v4_turbo' },
						{ value: 'eleven_multilingual_v2', label: 'eleven_multilingual_v2' }
					]
				},
				{ key: 'voice_settings.stability', label: eleven.voice.voiceSettingsStability.label, what: eleven.voice.voiceSettingsStability.what, kind: 'percent' },
				{ key: 'voice_settings.similarity_boost', label: eleven.voice.voiceSettingsSimilarityBoost.label, what: eleven.voice.voiceSettingsSimilarityBoost.what, kind: 'percent' },
				{ key: 'voice_settings.style', label: eleven.voice.voiceSettingsStyle.label, what: eleven.voice.voiceSettingsStyle.what, kind: 'percent' },
				{ key: 'voice_settings.use_speaker_boost', label: eleven.voice.voiceSettingsUseSpeakerBoost.label, what: eleven.voice.voiceSettingsUseSpeakerBoost.what, kind: 'switch' },
				{ key: 'use_request_context', label: eleven.voice.useRequestContext.label, what: eleven.voice.useRequestContext.what, kind: 'switch' },
				{ key: 'output_format', label: eleven.voice.outputFormat.label, what: eleven.voice.outputFormat.what, kind: 'text', placeholder: 'pcm_16000' },
				{ key: 'optimize_streaming_latency', label: eleven.voice.optimizeStreamingLatency.label, what: eleven.voice.optimizeStreamingLatency.what, kind: 'count', min: 0, max: 4 }
			]
		}
	],
	xtts: [{ title: eng.xtts.voice.title, fields: [{ key: 'enable_live_cloning', label: eng.xtts.voice.enableLiveCloning.label, what: eng.xtts.voice.enableLiveCloning.what, kind: 'switch' }] }],
	zonos: [
		{
			title: zonos.voice.title,
			fields: [
				{
					key: 'model',
					label: zonos.voice.model.label,
					what: zonos.voice.model.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'Zyphra/Zonos-v0.1-hybrid', label: zonos.voice.model.choices.zyphraZonosV01Hybrid.label },
						{ value: 'Zyphra/Zonos-v0.1-transformer', label: zonos.voice.model.choices.zyphraZonosV01Transformer.label }
					]
				},
				tune('speaking_rate', zonos.voice.speakingRate, 0.1, 50),
				tune('pitch_std', zonos.voice.pitchStd),
				tune('cfg_scale', zonos.voice.cfgScale, 1, 20),
				tune('vq_score', zonos.voice.vqScore),
				tune('dnsmos_overall', zonos.voice.dnsmosOverall, 1, 5),
				{ key: 'fmax', label: zonos.voice.fmax.label, what: zonos.voice.fmax.what, kind: 'count', unit: zonos.voice.fmax.unit },
				{ key: 'denoise_speaker', label: zonos.voice.denoiseSpeaker.label, what: zonos.voice.denoiseSpeaker.what, kind: 'switch' },
				{ key: 'randomize_seed', label: zonos.voice.randomizeSeed.label, what: zonos.voice.randomizeSeed.what, kind: 'switch' },
				{ key: 'unconditional_keys', label: zonos.voice.unconditionalKeys.label, what: zonos.voice.unconditionalKeys.what, kind: 'lines', noun: zonos.voice.unconditionalKeys.noun }
			]
		},
		{
			title: zonos.sampling.title,
			fields: [
				tune('top_p', zonos.sampling.topP, 0, 1),
				tune('min_k', zonos.sampling.minK),
				tune('min_p', zonos.sampling.minP, 0, 1),
				tune('linear', zonos.sampling.linear),
				tune('confidence', zonos.sampling.confidence),
				tune('quadratic', zonos.sampling.quadratic)
			]
		}
	],
	chatterbox: [
		{
			title: chatterbox.voice.title,
			fields: [
				tune('exaggeration', chatterbox.voice.exaggeration),
				tune('cfg_weight', chatterbox.voice.cfgWeight, 0, 1),
				tune('temperature', chatterbox.voice.temperature, 0.05, 5),
				{ key: 'seed', label: chatterbox.voice.seed.label, what: chatterbox.voice.seed.what, kind: 'count' },
				{ key: 'voice_sample_rate', label: chatterbox.voice.voiceSampleRate.label, what: chatterbox.voice.voiceSampleRate.what, kind: 'count', min: 8000, max: 48000, unit: chatterbox.voice.voiceSampleRate.unit }
			]
		},
		{ title: chatterbox.sampling.title, fields: [tune('min_p', chatterbox.sampling.minP, 0, 1), tune('top_p', chatterbox.sampling.topP, 0, 1), tune('repetition_penalty', chatterbox.sampling.repetitionPenalty)] }
	],
	higgs: [
		{
			title: higgs.title,
			fields: [
				{ key: 'model', label: higgs.model.label, what: higgs.model.what, kind: 'text' },
				tune('temperature', higgs.temperature, 0.05, 5),
				{ key: 'enable_tag_propagation', label: higgs.enableTagPropagation.label, what: higgs.enableTagPropagation.what, kind: 'switch' }
			]
		}
	],
	'openai-compatible': [
		{
			title: openai.account.title,
			fields: [
				{ key: 'service_profile', label: openai.account.serviceProfile.label, what: openai.account.serviceProfile.what, kind: 'choice', custom: true, choices: [{ value: 'self_hosted', label: openai.account.serviceProfile.choices.selfHosted.label }] },
				{ key: 'api_key', label: openai.account.apiKey.label, what: openai.account.apiKey.what, kind: 'secret' }
			]
		},
		{
			title: openai.voice.title,
			fields: [
				{ key: 'model', label: openai.voice.model.label, what: openai.voice.model.what, kind: 'text' },
				tune('temperature', openai.voice.temperature),
				tune('speed', openai.voice.speed, 0.25, 4),
				tune('top_p', openai.voice.topP),
				{ key: 'top_k', label: openai.voice.topK.label, what: openai.voice.topK.what, kind: 'count' },
				{ key: 'transcribe_missing_reference_samples', label: openai.voice.transcribeMissingReferenceSamples.label, what: openai.voice.transcribeMissingReferenceSamples.what, kind: 'switch' },
				{ key: 'enable_higgs_tags', label: openai.voice.enableHiggsTags.label, what: openai.voice.enableHiggsTags.what, kind: 'switch' },
				{ key: 'enable_tag_propagation', label: openai.voice.enableTagPropagation.label, what: openai.voice.enableTagPropagation.what, kind: 'switch' }
			]
		}
	]
};

// The settings for what's said, and voices to fall back on, every engine has.
const tags = eng.tags;
const say = eng.pronunciation.pronunciationDictionary;
const narrative = eng.narration;
const fallback = eng.fallbackVoices;
const shared: Section[] = [
	{
		title: tags.title,
		what: tags.what,
		fields: [
			{ key: 'enable_audio_tags', label: tags.enableAudioTags.label, what: tags.enableAudioTags.what, kind: 'switch' },
			{ key: 'allowed_audio_tags', label: tags.allowedAudioTags.label, what: tags.allowedAudioTags.what, kind: 'lines', noun: tags.allowedAudioTags.noun }
		]
	},
	{
		title: eng.pronunciation.title,
		fields: [
			{
				key: 'pronunciation_dictionary',
				label: say.label,
				what: say.what,
				kind: 'pairs',
				columns: [say.columns.first, say.columns.second],
				pairKeys: ['original', 'replacement'],
				noun: say.noun,
				add: say.add
			}
		]
	},
	{
		title: narrative.title,
		fields: [
			{ key: 'narrative.enabled', label: narrative.narrativeEnabled.label, what: narrative.narrativeEnabled.what, kind: 'switch' },
			{ key: 'narrative.use_narrator_voice', label: narrative.narrativeUseNarratorVoice.label, what: narrative.narrativeUseNarratorVoice.what, kind: 'switch' },
			{ key: 'narrative.show_narrative_in_subtitles', label: narrative.narrativeShowNarrativeInSubtitles.label, what: narrative.narrativeShowNarrativeInSubtitles.what, kind: 'switch' }
		]
	},
	{
		title: fallback.title,
		what: fallback.what,
		fields: [
			{ key: 'voice_fallback.enabled', label: fallback.voiceFallbackEnabled.label, what: fallback.voiceFallbackEnabled.what, kind: 'switch' },
			{ key: 'voice_fallback.default_fallback_voice_male', label: fallback.voiceFallbackDefaultFallbackVoiceMale.label, what: fallback.voiceFallbackDefaultFallbackVoiceMale.what, kind: 'text', placeholder: 'malecommoner' },
			{ key: 'voice_fallback.default_fallback_voice_female', label: fallback.voiceFallbackDefaultFallbackVoiceFemale.label, what: fallback.voiceFallbackDefaultFallbackVoiceFemale.what, kind: 'text', placeholder: 'femaleeventoned' },
			{
				key: 'voice_fallback.race_gender_fallbacks',
				label: fallback.voiceFallbackRaceGenderFallbacks.label,
				what: fallback.voiceFallbackRaceGenderFallbacks.what,
				kind: 'map',
				columns: [fallback.voiceFallbackRaceGenderFallbacks.columns.first, fallback.voiceFallbackRaceGenderFallbacks.columns.second],
				noun: fallback.voiceFallbackRaceGenderFallbacks.noun
			}
		]
	}
];

// engineFeature is an engine's page: its file's settings, in its own terms.
export function engineFeature(e: Engine): Feature {
	return { id: `engines/${e.id}`, file: e.file, title: e.label, what: e.what, sections: [connection(e), ...(own[e.id] ?? []), ...shared] };
}

const backends = m.features.speechToText.recognizer.backend.choices;
const sttBackends = [
	{ value: 'whisper', label: backends.whisper.label, what: backends.whisper.what },
	{ value: 'parakeet', label: backends.parakeet.label, what: backends.parakeet.what },
	{ value: 'external_whisper', label: backends.externalWhisper.label, what: backends.externalWhisper.what },
	{ value: 'fateless', label: backends.fateless.label, what: backends.fateless.what },
	{ value: 'none', label: backends.none.label, what: backends.none.what }
];

const stt = m.features.speechToText;

export const speechToText: Feature = {
	id: 'speech-to-text',
	file: 'STT',
	title: stt.title,
	what: stt.what,
	sections: [
		{
			title: stt.recognizer.title,
			fields: [
				{ key: 'backend', label: stt.recognizer.backend.label, what: stt.recognizer.backend.what, kind: 'choice', custom: true, choices: sttBackends },
				{ key: 'language', label: stt.recognizer.language.label, what: stt.recognizer.language.what, kind: 'text', placeholder: 'en' },
				{ key: 'name_hints', label: stt.recognizer.nameHints.label, what: stt.recognizer.nameHints.what, kind: 'switch' },
				{ key: 'enable_partial_results', label: stt.recognizer.enablePartialResults.label, what: stt.recognizer.enablePartialResults.what, kind: 'switch' }
			]
		},
		{
			title: stt.whisper.title,
			what: stt.whisper.what,
			fields: [
				{ key: 'whisper.name_hints', label: stt.whisper.whisperNameHints.label, what: stt.whisper.whisperNameHints.what, kind: 'switch' },
				{
					key: 'whisper.model',
					label: stt.whisper.whisperModel.label,
					what: stt.whisper.whisperModel.what,
					kind: 'choice',
					custom: true,
					choices: ['tiny', 'base', 'small', 'medium', 'large-v1', 'large-v2', 'large-v3', 'large-v3-turbo'].map((v) => ({ value: v, label: v }))
				},
				{ key: 'whisper.use_whisper_finetune', label: stt.whisper.whisperUseWhisperFinetune.label, what: stt.whisper.whisperUseWhisperFinetune.what, kind: 'switch' },
				{ key: 'whisper.use_gpu', label: stt.whisper.whisperUseGpu.label, what: stt.whisper.whisperUseGpu.what, kind: 'switch' },
				{ key: 'whisper.gpu_device_index', label: stt.whisper.whisperGpuDeviceIndex.label, what: stt.whisper.whisperGpuDeviceIndex.what, kind: 'count' },
				{ key: 'whisper.threads', label: stt.whisper.whisperThreads.label, what: stt.whisper.whisperThreads.what, kind: 'count' },
				{ key: 'whisper.input_sample_rate', label: stt.whisper.whisperInputSampleRate.label, what: stt.whisper.whisperInputSampleRate.what, kind: 'count', unit: stt.whisper.whisperInputSampleRate.unit }
			]
		},
		{
			title: stt.parakeet.title,
			what: stt.parakeet.what,
			fields: [
				{
					key: 'parakeet.model',
					label: stt.parakeet.parakeetModel.label,
					what: stt.parakeet.parakeetModel.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'q8_0', label: 'Q8_0', what: stt.parakeet.parakeetModel.choices.q80.what },
						{ value: 'q4_k', label: 'Q4_K' },
						{ value: 'q4_0', label: 'Q4_0' },
						{ value: 'f16', label: 'F16', what: stt.parakeet.parakeetModel.choices.f16.what }
					]
				},
				{ key: 'parakeet.use_gpu', label: stt.parakeet.parakeetUseGpu.label, what: stt.parakeet.parakeetUseGpu.what, kind: 'switch' },
				{ key: 'parakeet.gpu_device_index', label: stt.parakeet.parakeetGpuDeviceIndex.label, what: stt.parakeet.parakeetGpuDeviceIndex.what, kind: 'count' },
				{ key: 'parakeet.threads', label: stt.parakeet.parakeetThreads.label, what: stt.parakeet.parakeetThreads.what, kind: 'count' }
			]
		},
		{
			title: stt.fateless.title,
			what: stt.fateless.what,
			fields: [
				{ key: 'fateless_model', label: stt.fateless.fatelessModel.label, what: stt.fateless.fatelessModel.what, kind: 'text', placeholder: 'fateless/transcribe' },
				{ key: 'fateless_streaming', label: stt.fateless.fatelessStreaming.label, what: stt.fateless.fatelessStreaming.what, kind: 'switch' }
			]
		},
		{
			title: stt.whisperService.title,
			what: stt.whisperService.what,
			fields: [
				{ key: 'external_whisper.api_endpoint', label: stt.whisperService.externalWhisperApiEndpoint.label, what: stt.whisperService.externalWhisperApiEndpoint.what, kind: 'text' },
				{ key: 'external_whisper.api_key', label: stt.whisperService.externalWhisperApiKey.label, what: stt.whisperService.externalWhisperApiKey.what, kind: 'secret' },
				{ key: 'external_whisper.model', label: stt.whisperService.externalWhisperModel.label, what: stt.whisperService.externalWhisperModel.what, kind: 'text' },
				{ key: 'external_whisper.timeout_ms', label: stt.whisperService.externalWhisperTimeoutMs.label, what: stt.whisperService.externalWhisperTimeoutMs.what, kind: 'count', unit: stt.whisperService.externalWhisperTimeoutMs.unit }
			]
		},
		{
			title: stt.openMic.title,
			what: stt.openMic.what,
			fields: [
				{ key: 'streaming.silence_threshold', label: stt.openMic.streamingSilenceThreshold.label, what: stt.openMic.streamingSilenceThreshold.what, kind: 'percent' },
				{ key: 'streaming.silence_duration', label: stt.openMic.streamingSilenceDuration.label, what: stt.openMic.streamingSilenceDuration.what, kind: 'count', unit: stt.openMic.streamingSilenceDuration.unit },
				{ key: 'streaming.chunks', label: stt.openMic.streamingChunks.label, what: stt.openMic.streamingChunks.what, kind: 'count' },
				{ key: 'streaming.openMicVerbose', label: stt.openMic.streamingOpenMicVerbose.label, what: stt.openMic.streamingOpenMicVerbose.what, kind: 'switch' }
			]
		}
	]
};

const vs = m.features.voiceSamples;

export const voiceSamples: Feature = {
	id: 'samples',
	file: 'VoiceSamples',
	title: vs.title,
	what: vs.what,
	sections: [
		{
			title: vs.whichLines.title,
			fields: [
				{ key: 'filtering.min_duration_seconds', label: vs.whichLines.filteringMinDurationSeconds.label, what: vs.whichLines.filteringMinDurationSeconds.what, kind: 'number' },
				{ key: 'filtering.max_duration_seconds', label: vs.whichLines.filteringMaxDurationSeconds.label, what: vs.whichLines.filteringMaxDurationSeconds.what, kind: 'number' },
				{ key: 'filtering.banned_keywords', label: vs.whichLines.filteringBannedKeywords.label, what: vs.whichLines.filteringBannedKeywords.what, kind: 'lines', noun: vs.whichLines.filteringBannedKeywords.noun },
				{ key: 'filtering.banned_voice_types', label: vs.whichLines.filteringBannedVoiceTypes.label, what: vs.whichLines.filteringBannedVoiceTypes.what, kind: 'lines', noun: vs.whichLines.filteringBannedVoiceTypes.noun },
				{ key: 'filtering.excluded_plugins', label: vs.whichLines.filteringExcludedPlugins.label, what: vs.whichLines.filteringExcludedPlugins.what, kind: 'lines', noun: vs.whichLines.filteringExcludedPlugins.noun }
			]
		},
		{
			title: vs.samples.title,
			fields: [{ key: 'voice_types', leaf: 'selected_sample', label: vs.samples.voiceTypes.label, what: vs.samples.voiceTypes.what, kind: 'map', columns: [vs.samples.voiceTypes.columns.first, vs.samples.voiceTypes.columns.second], noun: vs.samples.voiceTypes.noun }]
		}
	]
};
