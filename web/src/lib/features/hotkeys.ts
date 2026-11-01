import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.hotkeys;

// Hotkeys: the keys that drive SkyrimNet in game. Hotkey.yaml, as key codes.
export const hotkeys: Feature = {
	id: 'hotkeys',
	file: 'Hotkey',
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.input.title,
			fields: [
				{ key: 'textInput', label: f.input.textInput.label, what: f.input.textInput.what, kind: 'hotkey' },
				{ key: 'recordSpeech', label: f.input.recordSpeech.label, what: f.input.recordSpeech.what, kind: 'hotkey' },
				{ key: 'toggleOpenMic', label: f.input.toggleOpenMic.label, what: f.input.toggleOpenMic.what, kind: 'hotkey' },
				{ key: 'voiceThought', label: f.input.voiceThought.label, what: f.input.voiceThought.what, kind: 'hotkey' },
				{ key: 'voiceDialogueTransform', label: f.input.voiceDialogueTransform.label, what: f.input.voiceDialogueTransform.what, kind: 'hotkey' },
				{ key: 'voiceDirectInput', label: f.input.voiceDirectInput.label, what: f.input.voiceDirectInput.what, kind: 'hotkey' },
				{ key: 'dialogueChoice', label: f.input.dialogueChoice.label, what: f.input.dialogueChoice.what, kind: 'hotkey' }
			]
		},
		{
			title: f.dashboard.title,
			fields: [
				{ key: 'wheelMenu', label: f.dashboard.wheelMenu.label, what: f.dashboard.wheelMenu.what, kind: 'hotkey' },
				{ key: 'openLabelPanel', label: f.dashboard.openLabelPanel.label, what: f.dashboard.openLabelPanel.what, kind: 'hotkey' },
				{ key: 'toggleDashboard', label: f.dashboard.toggleDashboard.label, what: f.dashboard.toggleDashboard.what, kind: 'hotkey' }
			]
		},
		{
			title: f.switches.title,
			fields: [
				{ key: 'toggleGlobalAI', label: f.switches.toggleGlobalAI.label, what: f.switches.toggleGlobalAI.what, kind: 'hotkey' },
				{ key: 'toggleGameMaster', label: f.switches.toggleGameMaster.label, what: f.switches.toggleGameMaster.what, kind: 'hotkey' },
				{ key: 'toggleContinuousMode', label: f.switches.toggleContinuousMode.label, what: f.switches.toggleContinuousMode.what, kind: 'hotkey' },
				{ key: 'toggleWhisperMode', label: f.switches.toggleWhisperMode.label, what: f.switches.toggleWhisperMode.what, kind: 'hotkey' },
				{ key: 'toggleWorldEventReactions', label: f.switches.toggleWorldEventReactions.label, what: f.switches.toggleWorldEventReactions.what, kind: 'hotkey' }
			]
		},
		{
			title: f.actions.title,
			fields: [
				{ key: 'continueNarration', label: f.actions.continueNarration.label, what: f.actions.continueNarration.what, kind: 'hotkey' },
				{ key: 'captureCrosshair', label: f.actions.captureCrosshair.label, what: f.actions.captureCrosshair.what, kind: 'hotkey' },
				{ key: 'generateDiaryBio', label: f.actions.generateDiaryBio.label, what: f.actions.generateDiaryBio.what, kind: 'hotkey' },
				{ key: 'interruptDialogue', label: f.actions.interruptDialogue.label, what: f.actions.interruptDialogue.what, kind: 'hotkey' }
			]
		}
	]
};
