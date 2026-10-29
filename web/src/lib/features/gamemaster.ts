import type { Feature } from './schema.ts';

// The GameMaster: the agent that reads a scene and has characters speak or
// act. Part of SkyrimNet.yaml; its model is the GameMaster task on Models.
export const gameMaster: Feature = {
	id: 'gamemaster',
	file: 'game',
	title: 'GameMaster',
	what: 'The GameMaster reads the scene around you and decides when characters speak up or act, without being asked.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'gamemaster.enabled', label: 'GameMaster', what: 'Let the GameMaster shape scenes at all.', kind: 'switch' },
				{ key: 'gamemaster.agentEnabled', label: 'Act on its own', what: 'Let the GameMaster start scenes by itself, prompting characters to speak now and then.', kind: 'switch' },
				{ key: 'gamemaster.excludeActorsInScenes', label: 'Leave scenes alone', what: 'Don’t choose characters who are busy in one of the game’s own scenes, such as scripted events or vanilla dialogue.', kind: 'switch' }
			]
		},
		{
			title: 'Timing',
			fields: [
				{ key: 'gamemaster.actionCooldownSeconds', label: 'Between actions', what: 'How long the GameMaster waits after one action before the next.', kind: 'count', min: 30, max: 600, step: 5, unit: 'seconds' },
				{ key: 'gamemaster.requestTimeoutSeconds', label: 'Give the model', what: 'How long a request to the GameMaster’s model can take before it’s given up.', kind: 'count', min: 10, max: 120, unit: 'seconds' },
				{
					key: 'gamemaster.continuousSceneCooldownSeconds',
					label: 'Restart a stalled scene after',
					what: 'In a continuous scene, the next character speaks as the last one finishes. If the conversation dies instead, the GameMaster tries to restart it after this much silence.',
					kind: 'count',
					min: 10,
					max: 120,
					unit: 'seconds'
				}
			]
		},
		{
			title: 'Context',
			fields: [
				{ key: 'gamemaster.nearbyActorRadius', label: 'Who counts as nearby', what: 'How far away, in game units, a character can be and still be part of the scene.', kind: 'count', min: 400, max: 2000, step: 10, unit: 'units' },
				{ key: 'gamemaster.recentEventsCount', label: 'Recent events considered', what: 'How many of the latest events the GameMaster reads before deciding.', kind: 'count', min: 10, max: 100, unit: 'events' }
			]
		}
	]
};
