// How SkyrimNet gives characters voices: Speech, from SkyrimNet.yaml, picks
// the engine and covers what every engine shares (the narrator, silent
// NPCs, subtitles); each engine has a page of its own file's settings.
import type { Feature, Field, Section } from './schema.ts';

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
	{ id: 'pockettts', value: 'pocket_tts', file: 'PocketTTS', label: 'PocketTTS', what: 'Voices made on your own machine, cloned from each voice type’s sample. It needs nothing else installed.', local: true },
	{ id: 'piper', value: 'piper', file: 'Piper', label: 'Piper', what: 'Fast voices made on your own machine from voice model files, which are installed as a mod of their own.', local: true },
	{ id: 'inworld', value: 'inworld', file: 'Inworld', label: 'Inworld', what: 'Inworld’s service, with voices cloned from each voice type’s sample. It needs an Inworld account and key.', local: false },
	{ id: 'elevenlabs', value: 'elevenlabs', file: 'ElevenLabs', label: 'ElevenLabs', what: 'ElevenLabs’s service, with voices cloned from each voice type’s sample. It needs an ElevenLabs key.', local: false },
	{ id: 'xtts', value: 'xtts', file: 'XTTS', label: 'XTTS', what: 'Cloned voices from an XTTS server you run, on your own machine or rented through VastAI.', local: true },
	{ id: 'zonos', value: 'zonos', file: 'Zonos', label: 'Zonos', what: 'Voices that follow a character’s mood, from a Zonos server you run, on your own machine or rented through VastAI.', local: true },
	{ id: 'chatterbox', value: 'chatterbox', file: 'Chatterbox', label: 'Chatterbox', what: 'Expressive voices that take emotion tags, from a Chatterbox server you run.', local: true },
	{ id: 'higgs', value: 'higgs', file: 'Higgs', label: 'Higgs', what: 'Voices from a Higgs Audio server you run, with tags for emotion, pace, and style.', local: true },
	{ id: 'openai-compatible', value: 'openai_compatible', file: 'OpenAICompatibleTTS', label: 'OpenAI-compatible', what: 'Any speech service with an OpenAI-style API that clones voices, such as Boson.ai’s Higgs or a server you run yourself.', local: true }
];

export const noVoice = 'noop'; // tts.engine when SkyrimNet speaks with no engine

export function engineByValue(value: unknown): Engine | undefined {
	return engines.find((e) => e.value === value);
}

export function engineById(id: string): Engine | undefined {
	return engines.find((e) => e.id === id);
}

export const speech: Feature = {
	id: 'speech',
	file: 'game',
	title: 'Speech',
	what: 'Which engine gives characters their voices, and what every engine shares: the narrator, characters without voice lines of their own, and subtitles. Each engine’s own settings are under Engines.',
	sections: [
		{
			title: 'Engine',
			fields: [
				{
					key: 'tts.engine',
					label: 'Engine',
					what: 'What turns characters’ lines into speech. Without one, they aren’t spoken at all.',
					kind: 'choice',
					choices: [...engines.map((e) => ({ value: e.value, label: e.label, what: e.local ? 'runs on a machine of yours' : 'a service you sign up for' })), { value: noVoice, label: 'None', what: 'lines are only shown, never spoken' }]
				},
				{ key: 'tts.maxSegmentWords', label: 'Words sent at once', what: 'Lines go to the engine a sentence at a time. Above 0, short sentences in a row are joined, up to this many words, before they’re sent. At 0, every sentence goes on its own.', kind: 'count', min: 0, max: 500, unit: 'words', zero: 'Each sentence' }
			]
		},
		{
			title: 'Narrator',
			fields: [
				{ key: 'tts.enableNarratorAudioEffect', label: 'Narrator effect', what: 'Play narration through an audio effect, so the narrator sounds apart from the characters.', kind: 'switch' },
				{ key: 'tts.enablePlayerThoughtAudioEffect', label: 'Thoughts effect', what: 'Play your character’s thoughts through an audio effect, so they sound apart from speech.', kind: 'switch' },
				{ key: 'tts.useKokoroForNarrator', label: 'Kokoro narrator', what: 'Narrate with Kokoro, a small voice model that runs on your own machine, rather than the engine.', kind: 'switch' },
				{ key: 'tts.kokoroNarratorVoice', label: 'Kokoro voice', what: 'Which of Kokoro’s voices narrates.', kind: 'text', placeholder: 'af_bella' },
				{ key: 'tts.kokoroVolumeLevel', label: 'Kokoro volume', what: 'How loud Kokoro’s narration plays.', kind: 'percent' }
			]
		},
		{
			title: 'Silent characters',
			what: 'Characters with no voice lines of their own, voiced through DBVO.',
			fields: [
				{ key: 'dbvo.enabled', label: 'Use DBVO', what: 'Work with DBVO, the mod that voices lines the game has no recording for.', kind: 'switch' },
				{ key: 'dbvo.voiceSilentNPCs', label: 'Voice silent characters', what: 'Speak the game’s own lines for characters who have no recording of them. Universal Translator depends on it.', kind: 'switch' },
				{ key: 'dbvo.fixSubtitleDuringTTS', label: 'Subtitle fix', what: 'SkyrimNet’s fix for the game’s subtitle while such a line is spoken.', kind: 'switch' }
			]
		},
		{
			title: 'Subtitles',
			fields: [
				{ key: 'subtitles.enableNPCSubtitles', label: 'Characters’ lines', what: 'Show what characters say as subtitles.', kind: 'switch' },
				{ key: 'subtitles.enablePlayerSubtitles', label: 'Your lines', what: 'Show what you say as subtitles.', kind: 'switch' },
				{ key: 'subtitles.enableConsoleOutput', label: 'In the console', what: 'Also write lines to the game’s console.', kind: 'switch' }
			]
		},
		{
			title: 'Volume',
			fields: [{ key: 'audio.universalMaxVolume', label: 'Loudest', what: 'The most any of SkyrimNet’s audio plays at. 1 is full volume.', kind: 'number', min: 0, max: 4 }]
		}
	]
};

const seconds = (key: string, label: string, what: string): Field => ({ key, label, what, kind: 'count', min: 0, max: 3600, unit: 'seconds' });
const tune = (key: string, label: string, what: string, min?: number, max?: number): Field => ({ key, label, what, kind: 'number', min, max });

// The connection settings every engine file has.
function connection(e: Engine): Section {
	const endpoint: Field =
		e.id === 'inworld'
			? { key: 'endpoint', label: 'Address', what: 'Where Inworld’s service is reached. It can name your account, so it’s kept like a key: Hermaeus never shows it.', kind: 'secret' }
			: { key: 'endpoint', label: 'Address', what: e.id === 'piper' || e.id === 'pockettts' ? 'Left empty, the engine runs inside SkyrimNet itself.' : 'Where the server is reached.', kind: 'text', placeholder: 'http://localhost:7860' };
	return {
		title: 'Connection',
		fields: [
			endpoint,
			{ key: 'language', label: 'Language', what: 'The language lines are spoken in, as the engine names it.', kind: 'text', placeholder: 'en' },
			seconds('timeout', 'Wait for a line', 'How long to wait for the engine to speak a line before giving up.'),
			seconds('connect_timeout', 'Wait to connect', 'How long to wait to reach the engine.'),
			{ key: 'max_parallel_requests', label: 'Lines at once', what: 'How many lines the engine may work on at the same time. More starts speech sooner but takes more of the machine or account.', kind: 'count', min: 1, max: 16, unit: 'lines' },
			{ key: 'max_retries', label: 'Retries', what: 'How many times to ask again when the engine fails with a server error. At 0, a failed line isn’t retried.', kind: 'count', min: 0, max: 10, zero: 'None' },
			{ key: 'stream_audio', label: 'Stream', what: 'Start playing a line while the rest of it is still being made, so speech begins sooner.', kind: 'switch' }
		]
	};
}

// What each engine has of its own.
const own: Record<string, Section[]> = {
	pockettts: [
		{
			title: 'Voice',
			fields: [
				{ key: 'max_lsd_steps', label: 'Steps', what: 'How many passes go into each line. More sound better but take longer.', kind: 'count', min: 1, max: 64 },
				tune('temperature', 'Variety', 'How much each line varies. Lower is steadier; higher is livelier.', 0, 2),
				{
					key: 'model_precision',
					label: 'Precision',
					what: 'How the model runs.',
					kind: 'choice',
					choices: [
						{ value: 'FP32', label: 'FP32', what: 'better quality, best on a graphics card' },
						{ value: 'INT8', label: 'INT8', what: 'faster on the processor, a little rougher' }
					]
				},
				{ key: 'device_index', label: 'Device', what: 'Which graphics card to use, counting from 0.', kind: 'count', min: 0, max: 16 }
			]
		}
	],
	piper: [
		{
			title: 'Voice',
			fields: [
				{ key: 'enable_preloading', label: 'Load at start', what: 'Load the voice models when the game starts, rather than when each is first needed.', kind: 'switch' },
				tune('piper_settings.noise_scale', 'Expressiveness', 'How much feeling voices carry. 1 is normal; higher is more dramatic, lower is flatter.', 0, 5),
				tune('piper_settings.noise_w', 'Rhythm', 'How natural the pacing sounds. 1 is normal; higher varies more, lower is more mechanical.', 0, 5),
				tune('piper_settings.length_scale', 'Slowness', 'How drawn out speech is. 1 is normal; higher is slower, lower is faster.', 0.1, 5),
				{ key: 'piper_settings.num_threads', label: 'Threads', what: 'How many of the processor’s threads Piper uses.', kind: 'count', min: 1, max: 64 },
				{ key: 'speaker_id', label: 'Speaker', what: 'Which speaker to use in a model with several, counting from 0. The usual English models have one.', kind: 'count', min: 0, max: 1000 }
			]
		}
	],
	inworld: [
		{
			title: 'Account',
			fields: [
				{ key: 'api_key_base64', label: 'Key', what: 'Your Inworld key, as Inworld gives it, encoded in Base64.', kind: 'secret' },
				{ key: 'workspace_id', label: 'Workspace', what: 'The Inworld workspace voices are cloned into.', kind: 'secret' }
			]
		},
		{
			title: 'Voice',
			fields: [
				{
					key: 'model_id',
					label: 'Model',
					what: 'Which of Inworld’s models speaks.',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'inworld-tts-2', label: 'inworld-tts-2', what: 'the best quality' },
						{ value: 'inworld-tts-2-flash', label: 'inworld-tts-2-flash', what: 'the fastest and cheapest' }
					]
				},
				tune('temperature', 'Variety', 'How much each line varies, from 0 to 2. Higher is more expressive.', 0, 2),
				tune('speaking_rate', 'Speed', 'How fast lines are spoken. 1 is normal.', 0.5, 2),
				{ key: 'sample_rate_hz', label: 'Sample rate', what: 'The quality of the audio Inworld sends back.', kind: 'count', min: 8000, max: 48000, unit: 'Hz' },
				{ key: 'clone_cache.max_voices', label: 'Cloned voices kept', what: 'How many voices may be cloned into the account. Past it, the voice used longest ago is deleted first; voices cloned outside SkyrimNet count, but are never deleted. At 0, there’s no limit.', kind: 'count', min: 0, max: 10000, unit: 'voices', zero: 'No limit' }
			]
		}
	],
	elevenlabs: [
		{
			title: 'Account',
			fields: [
				{ key: 'api_key', label: 'Key', what: 'Your ElevenLabs key.', kind: 'secret' },
				{ key: 'clone_cache.shared_account', label: 'Shared account', what: 'Switch on when more than one SkyrimNet install uses the account, so they reuse each other’s cloned voices rather than cloning over them.', kind: 'switch' },
				{ key: 'clone_cache.max_voices', label: 'Cloned voices kept', what: 'How many voices may be cloned into the account. Past it, the voice used longest ago is deleted first. At 0, there’s no limit.', kind: 'count', min: 0, max: 10000, unit: 'voices', zero: 'No limit' }
			]
		},
		{
			title: 'Voice',
			fields: [
				{
					key: 'model_id',
					label: 'Model',
					what: 'Which of ElevenLabs’s models speaks.',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'eleven_v4_turbo', label: 'eleven_v4_turbo' },
						{ value: 'eleven_multilingual_v2', label: 'eleven_multilingual_v2' }
					]
				},
				{ key: 'voice_settings.stability', label: 'Stability', what: 'How consistent a voice is from line to line.', kind: 'percent' },
				{ key: 'voice_settings.similarity_boost', label: 'Likeness', what: 'How closely a voice keeps to the sample it was cloned from.', kind: 'percent' },
				{ key: 'voice_settings.style', label: 'Style', what: 'How much a voice exaggerates its delivery.', kind: 'percent' },
				{ key: 'voice_settings.use_speaker_boost', label: 'Speaker boost', what: 'Make voices clearer.', kind: 'switch' },
				{ key: 'use_request_context', label: 'Earlier lines as context', what: 'Send a speaker’s earlier lines with each request, for a steadier delivery. Only eleven_v4_turbo uses it, and lines are then made one at a time.', kind: 'switch' },
				{ key: 'output_format', label: 'Audio format', what: 'The audio ElevenLabs sends back. Streaming needs a PCM format.', kind: 'text', placeholder: 'pcm_16000' },
				{ key: 'optimize_streaming_latency', label: 'Speed over quality', what: 'From 0 to 4: higher starts speech sooner, at some cost to quality.', kind: 'count', min: 0, max: 4 }
			]
		}
	],
	xtts: [{ title: 'Voice', fields: [{ key: 'enable_live_cloning', label: 'Clone as needed', what: 'Clone a voice from its sample the first time it’s used, and keep it for next time.', kind: 'switch' }] }],
	zonos: [
		{
			title: 'Voice',
			fields: [
				{
					key: 'model',
					label: 'Model',
					what: 'Which Zonos model speaks.',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'Zyphra/Zonos-v0.1-hybrid', label: 'Hybrid' },
						{ value: 'Zyphra/Zonos-v0.1-transformer', label: 'Transformer' }
					]
				},
				tune('speaking_rate', 'Speed', 'How fast lines are spoken. Lower is slower.', 0, 40),
				tune('pitch_std', 'Pitch variation', 'How much the pitch moves. Higher is more varied.', 0, 400),
				tune('cfg_scale', 'Guidance', 'How closely speech follows the voice and mood it’s given.', 1, 10),
				tune('vq_score', 'VQ score', 'Trades audio quality against speed.', 0, 1),
				tune('dnsmos_overall', 'Clarity target', 'The audio quality aimed for, from 1 to 5.', 1, 5),
				{ key: 'fmax', label: 'Highest frequency', what: 'The highest frequency in the audio.', kind: 'count', min: 0, max: 48000, unit: 'Hz' },
				{ key: 'denoise_speaker', label: 'Clean the sample', what: 'Remove noise from a voice’s sample before cloning it.', kind: 'switch' },
				{ key: 'randomize_seed', label: 'New seed each line', what: 'Vary every line, rather than repeating the same delivery.', kind: 'switch' },
				{ key: 'unconditional_keys', label: 'Unconditioned', what: 'Conditions Zonos leaves for itself to decide, by name.', kind: 'lines', noun: ['condition', 'conditions'] }
			]
		},
		{
			title: 'Sampling',
			fields: [
				tune('top_p', 'Top P', 'A sampling setting. 0 leaves it off.', 0, 1),
				tune('min_k', 'Min K', 'A sampling setting. 0 leaves it off.', 0, 1024),
				tune('min_p', 'Min P', 'A sampling setting. 0 leaves it off.', 0, 1),
				tune('linear', 'Linear', 'A sampling setting.', -2, 2),
				tune('confidence', 'Confidence', 'A sampling setting.', -2, 2),
				tune('quadratic', 'Quadratic', 'A sampling setting.', -2, 2)
			]
		}
	],
	chatterbox: [
		{
			title: 'Voice',
			fields: [
				tune('exaggeration', 'Expressiveness', 'How dramatic emotions sound. Higher is more expressive.', 0, 2),
				tune('cfg_weight', 'Pace', 'How the pace of speech is held.', 0, 1),
				tune('temperature', 'Variety', 'How much each line varies.', 0, 2),
				{ key: 'seed', label: 'Seed', what: 'Repeat the same delivery for the same line. At -1, every line varies.', kind: 'count', min: -1, max: 2147483647 },
				{ key: 'voice_sample_rate', label: 'Sample rate', what: 'The quality voice samples are sent at for cloning.', kind: 'count', min: 8000, max: 48000, unit: 'Hz' }
			]
		},
		{ title: 'Sampling', fields: [tune('min_p', 'Min P', 'A sampling setting.', 0, 1), tune('top_p', 'Top P', 'A sampling setting.', 0, 1), tune('repetition_penalty', 'Repetition penalty', 'Discourages the same sounds repeating.', 1, 3)] }
	],
	higgs: [
		{
			title: 'Voice',
			fields: [{ key: 'model', label: 'Model', what: 'Which model the server uses.', kind: 'text' }, tune('temperature', 'Variety', 'How much each line varies.', 0, 2), { key: 'enable_tag_propagation', label: 'Carry tags on', what: 'A tag for emotion, style, or pace in one sentence of a reply carries on to the sentences after it.', kind: 'switch' }]
		}
	],
	'openai-compatible': [
		{
			title: 'Account',
			fields: [
				{ key: 'service_profile', label: 'Service', what: 'Your own server, or Boson.ai’s Higgs, whose address and model are fixed.', kind: 'choice', custom: true, choices: [{ value: 'self_hosted', label: 'Your own server' }] },
				{ key: 'api_key', label: 'Key', what: 'Boson.ai’s key, or a token your own server asks for. Leave it empty if it asks for none.', kind: 'secret' }
			]
		},
		{
			title: 'Voice',
			fields: [
				{ key: 'model', label: 'Model', what: 'Which model the service uses.', kind: 'text' },
				tune('temperature', 'Variety', 'How much each line varies.', 0, 2),
				tune('speed', 'Speed', 'How fast lines are spoken. 1 is normal.', 0.25, 4),
				tune('top_p', 'Top P', 'A sampling setting.', 0, 1),
				{ key: 'top_k', label: 'Top K', what: 'A sampling setting.', kind: 'count', min: 0, max: 1000 },
				{ key: 'transcribe_missing_reference_samples', label: 'Transcribe samples', what: 'When a voice’s sample has no transcript, make one with your speech recognition. Off, your own server gets an empty transcript.', kind: 'switch' },
				{ key: 'enable_higgs_tags', label: 'Higgs tags', what: 'Turn SkyrimNet’s tags into Higgs’s own. Switch off for a model whose tags are written differently.', kind: 'switch' },
				{ key: 'enable_tag_propagation', label: 'Carry tags on', what: 'A tag for emotion, style, or pace in one sentence of a reply carries on to the sentences after it.', kind: 'switch' }
			]
		}
	]
};

// The settings for what's said, and voices to fall back on, every engine has.
const shared: Section[] = [
	{
		title: 'Tags',
		what: 'Tags the model writes into a line, such as [sighs], for how it’s spoken.',
		fields: [
			{ key: 'enable_audio_tags', label: 'Use tags', what: 'Pass tags on to the engine. Only engines that understand them do anything with them.', kind: 'switch' },
			{ key: 'allowed_audio_tags', label: 'Allowed tags', what: 'The tags passed on; others are taken out of the line. With none listed, every tag is passed on.', kind: 'lines', noun: ['tag', 'tags'] }
		]
	},
	{
		title: 'Pronunciation',
		fields: [
			{
				key: 'pronunciation_dictionary',
				label: 'Pronunciations',
				what: 'Words the engine says wrong, each spelled the way it should sound.',
				kind: 'pairs',
				columns: ['Word', 'Say it as'],
				pairKeys: ['original', 'replacement'],
				noun: ['word', 'words']
			}
		]
	},
	{
		title: 'Narration',
		fields: [
			{ key: 'narrative.enabled', label: 'Speak narration', what: 'Speak the narrator’s lines through this engine.', kind: 'switch' },
			{ key: 'narrative.use_narrator_voice', label: 'Narrator’s voice', what: 'Narrate in a narrator’s voice of its own.', kind: 'switch' },
			{ key: 'narrative.show_narrative_in_subtitles', label: 'Narration subtitles', what: 'Show narration as subtitles too.', kind: 'switch' }
		]
	},
	{
		title: 'Fallback voices',
		what: 'Voices used when a character’s own voice can’t be made.',
		fields: [
			{ key: 'voice_fallback.enabled', label: 'Fall back', what: 'Speak in a fallback voice when a character’s own fails, rather than not at all.', kind: 'switch' },
			{ key: 'voice_fallback.default_fallback_voice_male', label: 'Men', what: 'The voice type a man falls back to, when his race has none below.', kind: 'text', placeholder: 'malecommoner' },
			{ key: 'voice_fallback.default_fallback_voice_female', label: 'Women', what: 'The voice type a woman falls back to, when her race has none below.', kind: 'text', placeholder: 'femaleeventoned' },
			{ key: 'voice_fallback.race_gender_fallbacks', label: 'By race', what: 'The voice type each race and gender falls back to, before the two above.', kind: 'map', columns: ['Race and gender', 'Voice type'], noun: ['voice', 'voices'] }
		]
	}
];

// engineFeature is an engine's page: its file's settings, in its own terms.
export function engineFeature(e: Engine): Feature {
	return { id: `engines/${e.id}`, file: e.file, title: e.label, what: e.what, sections: [connection(e), ...(own[e.id] ?? []), ...shared] };
}
