import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.gamemaster;

// The GameMaster: the agent that reads a scene and has characters speak or
// act. Part of SkyrimNet.yaml; its model is the GameMaster task on Models.
export const gameMaster: Feature = {
	id: 'gamemaster',
	file: 'game',
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.general.title,
			fields: [
				{ key: 'gamemaster.enabled', label: f.general.gamemasterEnabled.label, what: f.general.gamemasterEnabled.what, kind: 'switch' },
				{ key: 'gamemaster.agentEnabled', label: f.general.gamemasterAgentEnabled.label, what: f.general.gamemasterAgentEnabled.what, kind: 'switch' },
				{ key: 'gamemaster.excludeActorsInScenes', label: f.general.gamemasterExcludeActorsInScenes.label, what: f.general.gamemasterExcludeActorsInScenes.what, kind: 'switch' }
			]
		},
		{
			title: f.timing.title,
			fields: [
				{ key: 'gamemaster.actionCooldownSeconds', label: f.timing.gamemasterActionCooldownSeconds.label, what: f.timing.gamemasterActionCooldownSeconds.what, kind: 'count', min: 30, max: 600, step: 5, unit: f.timing.gamemasterActionCooldownSeconds.unit },
				{ key: 'gamemaster.requestTimeoutSeconds', label: f.timing.gamemasterRequestTimeoutSeconds.label, what: f.timing.gamemasterRequestTimeoutSeconds.what, kind: 'count', min: 10, max: 120, unit: f.timing.gamemasterRequestTimeoutSeconds.unit },
				{
					key: 'gamemaster.continuousSceneCooldownSeconds',
					label: f.timing.gamemasterContinuousSceneCooldownSeconds.label,
					what: f.timing.gamemasterContinuousSceneCooldownSeconds.what,
					kind: 'count',
					min: 10,
					max: 120,
					unit: f.timing.gamemasterContinuousSceneCooldownSeconds.unit
				}
			]
		},
		{
			title: f.context.title,
			fields: [
				{ key: 'gamemaster.nearbyActorRadius', label: f.context.gamemasterNearbyActorRadius.label, what: f.context.gamemasterNearbyActorRadius.what, kind: 'count', min: 400, max: 2000, step: 10, unit: f.context.gamemasterNearbyActorRadius.unit },
				{ key: 'gamemaster.recentEventsCount', label: f.context.gamemasterRecentEventsCount.label, what: f.context.gamemasterRecentEventsCount.what, kind: 'count', min: 10, max: 100, unit: f.context.gamemasterRecentEventsCount.unit }
			]
		}
	]
};
