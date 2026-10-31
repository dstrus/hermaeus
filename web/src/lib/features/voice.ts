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
	{ id: 'fateless', value: 'fateless', file: 'FatelessTTS', label: 'Fateless', what: 'Fateless’s service, from SkyrimNet Beta 29: characters speak in their own in-game voices, with no server or key to set up. It needs a Fateless subscription, signed in from SkyrimNet’s dashboard.', local: false },
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
				{ key: 'tts.maxSegmentWords', label: 'Words sent at once', what: 'Lines go to the engine a sentence at a time. Above 0, short sentences in a row are joined, up to this many words, before they’re sent. At 0, every sentence goes on its own.', kind: 'count', unit: 'words', zero: 'Each sentence' }
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
			title: 'Game dialogue',
			what: 'The game’s own dialogue: the lines you pick, and characters with no recording of a line.',
			fields: [
				{ key: 'vanillaDialogue.voicePlayerChoices', was: 'dbvo.enabled', label: 'Voice your choices', what: 'Speak the line you pick in the dialogue menu in your character’s voice before the NPC answers.', kind: 'switch' },
				{ key: 'vanillaDialogue.voiceSilentNPCs', was: 'dbvo.voiceSilentNPCs', label: 'Voice silent characters', what: 'Speak the game’s own lines for characters who have no recording of them. Universal Translator depends on it.', kind: 'switch' },
				{ key: 'vanillaDialogue.fixSubtitleDuringTTS', was: 'dbvo.fixSubtitleDuringTTS', label: 'Subtitle fix', what: 'SkyrimNet’s fix for the game’s subtitle while such a line is spoken.', kind: 'switch' }
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
			fields: [{ key: 'audio.universalMaxVolume', label: 'Loudest', what: 'The most any of SkyrimNet’s audio plays at. 1 is full volume.', kind: 'number', min: 0, max: 1 }]
		}
	]
};

const seconds = (key: string, label: string, what: string): Field => ({ key, label, what, kind: 'count', unit: 'seconds' });
const tune = (key: string, label: string, what: string, min?: number, max?: number): Field => ({ key, label, what, kind: 'number', min, max });

// The connection settings every engine file has.
function connection(e: Engine): Section {
	const endpoint: Field =
		e.id === 'inworld'
			? { key: 'endpoint', label: 'Address', what: 'Where Inworld’s service is reached. It can name your account, so it’s kept like a key: Hermaeus never shows it.', kind: 'secret' }
			: { key: 'endpoint', label: 'Address', what: e.id === 'piper' || e.id === 'pockettts' ? 'Left empty for the engine that comes with SkyrimNet.' : 'Where the server is reached.', kind: 'text', placeholder: 'http://localhost:7860' };
	return {
		title: 'Connection',
		fields: [
			// Fateless speaks through the gateway SkyrimNet's sign-in pairs with.
			...(e.id === 'fateless' ? [] : [endpoint]),
			{ key: 'language', label: 'Language', what: 'The language lines are spoken in, as the engine names it.', kind: 'text', placeholder: 'en' },
			seconds('timeout', 'Wait for a line', 'How long to wait for the engine to speak a line before giving up.'),
			seconds('connect_timeout', 'Wait to connect', 'How long to wait to reach the engine.'),
			{ key: 'max_parallel_requests', label: 'Lines at once', what: 'How many lines the engine may work on at the same time. More starts speech sooner but takes more of the machine or account.', kind: 'count', unit: 'lines' },
			{ key: 'max_retries', label: 'Retries', what: 'How many times to ask again when the engine fails with a server error. At 0, a failed line isn’t retried.', kind: 'count', zero: 'None' },
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
				{ key: 'max_lsd_steps', label: 'Steps', what: 'How many passes go into each line. More sound better but take longer.', kind: 'count' },
				tune('temperature', 'Variety', 'How much each line varies. Lower is steadier; higher is livelier.'),
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
				{ key: 'device_index', label: 'Device', what: 'Which graphics card to use, counting from 0.', kind: 'count' }
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
				{ key: 'piper_settings.num_threads', label: 'Threads', what: 'How many of the processor’s threads Piper uses.', kind: 'count' },
				{ key: 'speaker_id', label: 'Speaker', what: 'Which speaker to use in a model with several, counting from 0. The usual English models have one.', kind: 'count' }
			]
		}
	],
	fateless: [
		{
			title: 'Voice',
			fields: [
				{
					key: 'model',
					label: 'Model',
					what: 'Which of Fateless’s speech models speaks. Directions such as [say bitterly in a low voice] need fateless/speech-plus; fateless/speech is faster and cheaper, and plays only sounds such as [sigh].',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'fateless/speech-plus', label: 'Speech Plus', what: 'the default, with directions' },
						{ value: 'fateless/speech', label: 'Speech', what: 'faster and cheaper, sounds only' }
					]
				},
				tune('speed', 'Speed', 'How fast lines are spoken. 1 is normal.', 0.5, 1.5),
				tune('temperature', 'Variety', 'How much each line varies.', 0, 2)
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
				tune('speaking_rate', 'Speed', 'How fast lines are spoken. 1 is normal.'),
				{ key: 'sample_rate_hz', label: 'Sample rate', what: 'The quality of the audio Inworld sends back.', kind: 'count', min: 8000, max: 48000, unit: 'Hz' },
				{ key: 'clone_cache.max_voices', label: 'Cloned voices kept', what: 'How many voices may be cloned into the account. Past it, the voice used longest ago is deleted first; voices cloned outside SkyrimNet count, but are never deleted. At 0, there’s no limit.', kind: 'count', unit: 'voices', zero: 'No limit' }
			]
		}
	],
	elevenlabs: [
		{
			title: 'Account',
			fields: [
				{ key: 'api_key', label: 'Key', what: 'Your ElevenLabs key.', kind: 'secret' },
				{ key: 'clone_cache.shared_account', label: 'Shared account', what: 'Switch on when more than one SkyrimNet install uses the account, so they reuse each other’s cloned voices rather than cloning over them.', kind: 'switch' },
				{ key: 'clone_cache.max_voices', label: 'Cloned voices kept', what: 'How many voices may be cloned into the account. Past it, the voice used longest ago is deleted first. At 0, there’s no limit.', kind: 'count', unit: 'voices', zero: 'No limit' }
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
				tune('speaking_rate', 'Speed', 'How fast lines are spoken. Lower is slower.', 0.1, 50),
				tune('pitch_std', 'Pitch variation', 'How much the pitch moves. Higher is more varied.'),
				tune('cfg_scale', 'Guidance', 'How closely speech follows the voice and mood it’s given.', 1, 20),
				tune('vq_score', 'VQ score', 'Trades audio quality against speed.'),
				tune('dnsmos_overall', 'Clarity target', 'The audio quality aimed for, from 1 to 5.', 1, 5),
				{ key: 'fmax', label: 'Highest frequency', what: 'The highest frequency in the audio.', kind: 'count', unit: 'Hz' },
				{ key: 'denoise_speaker', label: 'Clean the sample', what: 'Remove noise from a voice’s sample before cloning it.', kind: 'switch' },
				{ key: 'randomize_seed', label: 'New seed each line', what: 'Vary every line, rather than repeating the same delivery.', kind: 'switch' },
				{ key: 'unconditional_keys', label: 'Unconditioned', what: 'Conditions Zonos leaves for itself to decide, by name.', kind: 'lines', noun: ['condition', 'conditions'] }
			]
		},
		{
			title: 'Sampling',
			fields: [
				tune('top_p', 'Top P', 'A sampling setting. 0 leaves it off.', 0, 1),
				tune('min_k', 'Min K', 'A sampling setting. 0 leaves it off.'),
				tune('min_p', 'Min P', 'A sampling setting. 0 leaves it off.', 0, 1),
				tune('linear', 'Linear', 'A sampling setting.'),
				tune('confidence', 'Confidence', 'A sampling setting.'),
				tune('quadratic', 'Quadratic', 'A sampling setting.')
			]
		}
	],
	chatterbox: [
		{
			title: 'Voice',
			fields: [
				tune('exaggeration', 'Expressiveness', 'How dramatic emotions sound. Higher is more expressive.'),
				tune('cfg_weight', 'Pace', 'How the pace of speech is held.', 0, 1),
				tune('temperature', 'Variety', 'How much each line varies.', 0.05, 5),
				{ key: 'seed', label: 'Seed', what: 'Repeat the same delivery for the same line. At -1, every line varies.', kind: 'count' },
				{ key: 'voice_sample_rate', label: 'Sample rate', what: 'The quality voice samples are sent at for cloning.', kind: 'count', min: 8000, max: 48000, unit: 'Hz' }
			]
		},
		{ title: 'Sampling', fields: [tune('min_p', 'Min P', 'A sampling setting.', 0, 1), tune('top_p', 'Top P', 'A sampling setting.', 0, 1), tune('repetition_penalty', 'Repetition penalty', 'Discourages the same sounds repeating.')] }
	],
	higgs: [
		{
			title: 'Voice',
			fields: [{ key: 'model', label: 'Model', what: 'Which model the server uses.', kind: 'text' }, tune('temperature', 'Variety', 'How much each line varies.', 0.05, 5), { key: 'enable_tag_propagation', label: 'Carry tags on', what: 'A tag for emotion, style, or pace in one sentence of a reply carries on to the sentences after it.', kind: 'switch' }]
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
				tune('temperature', 'Variety', 'How much each line varies.'),
				tune('speed', 'Speed', 'How fast lines are spoken. 1 is normal.', 0.25, 4),
				tune('top_p', 'Top P', 'A sampling setting.'),
				{ key: 'top_k', label: 'Top K', what: 'A sampling setting.', kind: 'count' },
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
			{ key: 'allowed_audio_tags', label: 'Allowed tags', what: 'The tags passed on; others are taken out of the line.', kind: 'lines', noun: ['tag', 'tags'] }
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

const sttBackends = [
	{ value: 'whisper', label: 'Whisper', what: 'runs on your machine' },
	{ value: 'parakeet', label: 'Parakeet', what: 'runs on your machine, and tells languages apart itself' },
	{ value: 'external_whisper', label: 'Whisper service', what: 'any OpenAI-style transcription API; a hosted one needs a key' },
	{ value: 'fateless', label: 'Fateless', what: 'fateless.ai, with an account there' },
	{ value: 'none', label: 'None', what: 'no microphone; you can still type' }
];

export const speechToText: Feature = {
	id: 'speech-to-text',
	file: 'STT',
	title: 'Speech to text',
	what: 'How what you say into your microphone becomes text for characters to answer: which recognizer listens, and when it decides you’ve finished speaking.',
	sections: [
		{
			title: 'Recognizer',
			fields: [
				{ key: 'backend', label: 'Recognizer', what: 'What turns your speech into text.', kind: 'choice', custom: true, choices: sttBackends },
				{ key: 'language', label: 'Language', what: 'The language you speak, as a two-letter code. Whisper also takes auto-detect.', kind: 'text', placeholder: 'en' },
				{ key: 'name_hints', label: 'Name hints', what: 'Tell the recognizer the names of you, nearby and pinned characters, and nearby places, so they come back spelled right. The stt_hints prompt sets what’s sent. Whisper on your machine has its own switch, under Whisper.', kind: 'switch' },
				{ key: 'enable_partial_results', label: 'Text as you speak', what: 'Show what’s been recognized so far while you’re still speaking.', kind: 'switch' }
			]
		},
		{
			title: 'Whisper',
			what: 'When the recognizer is Whisper.',
			fields: [
				{ key: 'whisper.name_hints', label: 'Name hints', what: 'Send Whisper the names, as above. Off unless set: with the names sent, Whisper can sometimes turn a whole message into a stray word.', kind: 'switch' },
				{
					key: 'whisper.model',
					label: 'Model',
					what: 'Larger models are more accurate, and slower. A model is downloaded the first time it’s used.',
					kind: 'choice',
					custom: true,
					choices: ['tiny', 'base', 'small', 'medium', 'large-v1', 'large-v2', 'large-v3', 'large-v3-turbo'].map((v) => ({ value: v, label: v }))
				},
				{ key: 'whisper.use_whisper_finetune', label: 'Skyrim’s version', what: 'Use a version of the model trained for the game’s words, where there is one.', kind: 'switch' },
				{ key: 'whisper.use_gpu', label: 'Graphics card', what: 'Run on the graphics card rather than the processor.', kind: 'switch' },
				{ key: 'whisper.gpu_device_index', label: 'Card', what: 'Which graphics card, counting from 0.', kind: 'count' },
				{ key: 'whisper.threads', label: 'Threads', what: 'How many of the processor’s threads it uses.', kind: 'count' },
				{ key: 'whisper.input_sample_rate', label: 'Sample rate', what: 'The sample rate of the microphone’s audio as it reaches Whisper.', kind: 'count', unit: 'Hz' }
			]
		},
		{
			title: 'Parakeet',
			what: 'When the recognizer is Parakeet.',
			fields: [
				{
					key: 'parakeet.model',
					label: 'Model',
					what: 'Smaller versions are faster; larger ones are more accurate.',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'q8_0', label: 'Q8_0', what: 'recommended' },
						{ value: 'q4_k', label: 'Q4_K' },
						{ value: 'q4_0', label: 'Q4_0' },
						{ value: 'f16', label: 'F16', what: 'the largest' }
					]
				},
				{ key: 'parakeet.use_gpu', label: 'Graphics card', what: 'Run on the graphics card rather than the processor.', kind: 'switch' },
				{ key: 'parakeet.gpu_device_index', label: 'Card', what: 'Which graphics card, counting from 0.', kind: 'count' },
				{ key: 'parakeet.threads', label: 'Threads', what: 'How many of the processor’s threads it uses.', kind: 'count' }
			]
		},
		{
			title: 'Fateless',
			what: 'When the recognizer is Fateless, through SkyrimNet’s Fateless sign-in.',
			fields: [
				{ key: 'fateless_model', label: 'Model', what: 'Which of Fateless’s transcription models listens.', kind: 'text', placeholder: 'fateless/transcribe' },
				{ key: 'fateless_streaming', label: 'Stream while talking', what: 'With push-to-talk, send your voice while you hold the key, once it hears you speak, so the text is ready moments after you let go. It costs a little more per minute.', kind: 'switch' }
			]
		},
		{
			title: 'Whisper service',
			what: 'When the recognizer is a Whisper service.',
			fields: [
				{ key: 'external_whisper.api_endpoint', label: 'Address', what: 'The service’s transcription address.', kind: 'text' },
				{ key: 'external_whisper.api_key', label: 'Key', what: 'The service’s key. A server of your own may not need one.', kind: 'secret' },
				{ key: 'external_whisper.model', label: 'Model', what: 'The model as the service names it, such as openai/whisper-1 on OpenRouter or whisper-1 on OpenAI.', kind: 'text' },
				{ key: 'external_whisper.timeout_ms', label: 'Wait for text', what: 'How long to wait for the service before giving up.', kind: 'count', unit: 'ms' }
			]
		},
		{
			title: 'Open mic',
			what: 'Listening without a key held down.',
			fields: [
				{ key: 'streaming.silence_threshold', label: 'Speech threshold', what: 'How sure the voice detector must be that a sound is speech, in your microphone and in characters’ voice files.', kind: 'percent' },
				{ key: 'streaming.silence_duration', label: 'Pause that ends it', what: 'How long you must be quiet before what you said is sent.', kind: 'count', unit: 'ms' },
				{ key: 'streaming.chunks', label: 'Chunks held', what: 'How many pieces of audio are held at once for recognizing.', kind: 'count' },
				{ key: 'streaming.openMicVerbose', label: 'Notices', what: 'Show a notice when open mic hears you and when it sends what you said.', kind: 'switch' }
			]
		}
	]
};

export const voiceSamples: Feature = {
	id: 'samples',
	file: 'VoiceSamples',
	title: 'Voice samples',
	what: 'The recordings voices are cloned from: which of the game’s own lines may serve as a voice type’s sample, and the sample each voice type uses.',
	sections: [
		{
			title: 'Which lines may be samples',
			fields: [
				{ key: 'filtering.min_duration_seconds', label: 'Shortest', what: 'Lines shorter than this aren’t used as samples.', kind: 'number' },
				{ key: 'filtering.max_duration_seconds', label: 'Longest', what: 'Lines longer than this aren’t used as samples.', kind: 'number' },
				{ key: 'filtering.banned_keywords', label: 'Left out by word', what: 'Lines whose file name has one of these words aren’t used, such as screams, music, and coughs.', kind: 'lines', noun: ['word', 'words'] },
				{ key: 'filtering.banned_voice_types', label: 'Voice types left out', what: 'Voice types never cloned.', kind: 'lines', noun: ['voice type', 'voice types'] },
				{ key: 'filtering.excluded_plugins', label: 'Plugins left out', what: 'Game plugins whose lines are left out of the new audio scan and found the older way instead.', kind: 'lines', noun: ['plugin', 'plugins'] }
			]
		},
		{
			title: 'Samples',
			fields: [{ key: 'voice_types', leaf: 'selected_sample', label: 'Chosen samples', what: 'The recording each voice type is cloned from: a voice file in the game’s data, or an audio file of your own.', kind: 'map', columns: ['Voice type', 'Sample'], noun: ['voice type', 'voice types'] }]
		}
	]
};
