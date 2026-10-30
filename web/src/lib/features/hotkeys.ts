import type { Feature } from './schema.ts';

// Hotkeys: the keys that drive SkyrimNet in game. Hotkey.yaml, as key codes.
export const hotkeys: Feature = {
	id: 'hotkeys',
	file: 'Hotkey',
	title: 'Hotkeys',
	what: 'The keys that drive SkyrimNet while you play. Press a key to set one, or click with a mouse button; any key can be left unset.',
	sections: [
		{
			title: 'Input',
			fields: [
				{ key: 'textInput', label: 'Type to speak', what: 'Say something to the characters around you, typed.', kind: 'hotkey' },
				{ key: 'recordSpeech', label: 'Push to talk', what: 'Hold to record what you say, release to send it.', kind: 'hotkey' },
				{ key: 'toggleOpenMic', label: 'Open mic', what: 'Switch the microphone to always listening, instead of push to talk.', kind: 'hotkey' },
				{ key: 'voiceThought', label: 'Think aloud', what: 'Hold to voice a thought to yourself. A quick tap prompts a thought with nothing said.', kind: 'hotkey' },
				{ key: 'voiceDialogueTransform', label: 'Speak in character', what: 'Hold to say something that’s rewritten as your character would say it. A quick tap prompts a line with nothing said.', kind: 'hotkey' },
				{ key: 'voiceDirectInput', label: 'Tell the GameMaster', what: 'Hold to state something as fact, such as “lightning suddenly strikes”, and characters react as though it happened.', kind: 'hotkey' },
				{ key: 'dialogueChoice', label: 'Choose a dialogue option', what: 'With the game’s dialogue menu open: a quick tap has the model pick an option for you to confirm; hold and speak to say what you mean, and it picks the closest option.', kind: 'hotkey' }
			]
		},
		{
			title: 'Dashboard and menus',
			fields: [
				{ key: 'wheelMenu', label: 'Wheel menu', what: 'Open the wheel of every SkyrimNet action.', kind: 'hotkey' },
				{ key: 'toggleDashboard', label: 'Dashboard', what: 'Open or close SkyrimNet’s dashboard in game.', kind: 'hotkey' }
			]
		},
		{
			title: 'Switches',
			fields: [
				{ key: 'toggleGlobalAI', label: 'SkyrimNet on or off', what: 'Switch all of SkyrimNet’s AI on or off.', kind: 'hotkey' },
				{ key: 'toggleGameMaster', label: 'GameMaster on or off', what: 'Switch the GameMaster, which prompts characters to speak now and then.', kind: 'hotkey' },
				{ key: 'toggleContinuousMode', label: 'Continuous scene', what: 'Keep a conversation going: as one character finishes, the GameMaster has another speak.', kind: 'hotkey' },
				{ key: 'toggleWhisperMode', label: 'Whisper', what: 'Switch between whisper and normal speaking distance.', kind: 'hotkey' },
				{ key: 'toggleWorldEventReactions', label: 'Reactions on or off', what: 'Switch characters reacting to what happens around them, including each other’s lines.', kind: 'hotkey' }
			]
		},
		{
			title: 'Actions',
			fields: [
				{ key: 'continueNarration', label: 'Continue', what: 'Have a character speak up, to react to something or pick a stalled conversation back up.', kind: 'hotkey' },
				{ key: 'captureCrosshair', label: 'Capture', what: 'Opens the capture window, for the character or furniture in your crosshairs, yourself, or where you are. With Ctrl, it captures straight away: a quick press for what’s in your crosshairs, or hold for a second for yourself. With a book open, it captures the book.', kind: 'hotkey' },
				{ key: 'generateDiaryBio', label: 'Diary and bio now', what: 'Write diary entries and update bios now, for you, those nearby, those pinned, or whoever is in your crosshairs.', kind: 'hotkey' },
				{ key: 'interruptDialogue', label: 'Interrupt', what: 'Stop every character speaking, and clear what’s waiting to be said.', kind: 'hotkey' }
			]
		}
	]
};
