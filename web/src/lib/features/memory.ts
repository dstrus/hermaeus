import type { Feature } from './schema.ts';

// Memories: how SkyrimNet turns what happens into memories, and which it
// recalls. Memory.yaml, in the dashboard's sections.
export const memory: Feature = {
	id: 'memories',
	file: 'Memory',
	title: 'Memories',
	what: 'How SkyrimNet turns what happens into memories, and which of them an NPC recalls when they speak.',
	sections: [
		{
			title: 'General',
			fields: [
				{ key: 'performance.enable_background_processing', label: 'Make memories in the background', what: 'Turn events into memories while you play, rather than not at all.', kind: 'switch' },
				{ key: 'memory_retrieval.max_memories_per_retrieval', label: 'Memories recalled at once', what: 'How many memories a single recall brings into a prompt.', kind: 'count', min: 1, max: 100 },
				{ key: 'memory_lifecycle.max_memories_per_actor', label: 'Memory index starts at', what: 'The starting capacity of each character’s search index. It grows past this, so it isn’t a limit.', kind: 'count', min: 100, max: 50000, step: 100 },
				{ key: 'generation_rules.min_importance_for_storage', label: 'Keep memories above', what: 'The importance a new memory needs to be kept at all.', kind: 'percent' }
			]
		},
		{
			title: 'Embedding model',
			what: 'The model that turns memories into vectors for recall. Changing it re-embeds every memory in the background, and recall is partial until that finishes.',
			fields: [
				{
					key: 'embedding.model',
					label: 'Model',
					what: 'Bekko has the best recall, in English and other languages, and embeds about half as fast. all-MiniLM-L6-v2 is lighter and English only. A custom model goes in a folder of its own, holding its model.onnx and tokenizer.json, in SkyrimNet’s folder for custom models, and is chosen by its folder’s name.',
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'bekko-embedding-v1-a25m', label: 'Bekko Embedding v1', what: 'Best recall, multilingual' },
						{ value: 'all-MiniLM-L6-v2', label: 'all-MiniLM-L6-v2', what: 'Lighter, English only' }
					]
				}
			]
		},
		{
			title: 'Performance',
			fields: [
				{ key: 'performance.max_concurrent_generations', label: 'Memories written at once', what: 'How many memory generations run at the same time.', kind: 'count', min: 1, max: 50 },
				{ key: 'performance.reembed_threads', label: 'Re-embedding threads', what: 'CPU threads that re-index memories in the background after an update, at low priority. Lower it if the game stutters while that runs; it applies the next time re-embedding starts.', kind: 'count', min: 1, max: 4 },
				{ key: 'performance.memory_batch_size', label: 'Memories per cycle', what: 'How many memories one processing cycle generates.', kind: 'count', min: 10, max: 1000, step: 10 }
			]
		},
		{
			title: 'Segment boundaries',
			what: 'Events are cut into segments, and each segment becomes memories. These are in real minutes.',
			fields: [
				{ key: 'boundary_config.max_time_gap_minutes_real', label: 'Gap that ends a segment', what: 'A pause between events longer than this starts a new segment.', kind: 'minutes', min: 1, max: 1440 },
				{ key: 'boundary_config.min_segment_duration_minutes_real', label: 'Shortest segment', what: 'A segment shorter than this isn’t turned into memories.', kind: 'minutes', min: 1, max: 120 },
				{ key: 'boundary_config.max_segment_duration_minutes_real', label: 'Longest segment', what: 'A segment this long is cut, whatever is happening.', kind: 'minutes', min: 1, max: 2880 },
				{ key: 'boundary_config.max_events_per_segment', label: 'Most events in a segment', what: 'A segment with this many events is cut.', kind: 'count', min: 10, max: 1000, step: 10, unit: 'events' },
				{ key: 'boundary_config.min_events_per_segment', label: 'Fewest events for a memory', what: 'A segment with fewer events than this makes no memory.', kind: 'count', min: 1, max: 50, unit: 'events' }
			]
		},
		{
			title: 'Recall thresholds',
			fields: [
				{ key: 'vector_storage.similarity_threshold', label: 'Similarity needed', what: 'How alike a memory and the moment must be to count: 50% is related, 100% a strong match.', kind: 'percent' },
				{ key: 'vector_storage.min_relevance_score', label: 'Relevance needed', what: 'The overall score a memory needs to be recalled, once the weights below are applied.', kind: 'percent' }
			]
		},
		{
			title: 'Scoring weights',
			what: 'How much each part of a memory’s fit counts toward its relevance.',
			fields: [
				{ key: 'scoring_weights.semantic', label: 'Meaning', what: 'How alike the memory and the moment are in meaning.', kind: 'percent' },
				{ key: 'scoring_weights.temporal', label: 'Recency', what: 'How recent the memory is.', kind: 'percent' },
				{ key: 'scoring_weights.actor', label: 'Who was there', what: 'Whether the people in the memory are the people here.', kind: 'percent' },
				{ key: 'scoring_weights.emotional', label: 'Feeling', what: 'How the memory’s mood fits the moment.', kind: 'percent' },
				{ key: 'scoring_weights.location', label: 'Place', what: 'Whether the memory happened near here.', kind: 'percent' },
				{ key: 'scoring_weights.keyword', label: 'Words', what: 'Whether the memory shares words with what was said.', kind: 'percent' }
			]
		},
		{
			title: 'World knowledge',
			what: 'Lore and facts any NPC can draw on, from the History section’s Knowledge page.',
			fields: [
				{ key: 'world_knowledge.max_semantic_results', label: 'Entries recalled', what: 'How many knowledge entries, found by relevance, go into an NPC’s prompt.', kind: 'count', min: 1, max: 20 },
				{ key: 'world_knowledge.max_always_inject', label: 'Always-included entries', what: 'How many of the entries marked “always include” go in. 0 means all of them.', kind: 'count', min: 0, max: 50 },
				{ key: 'world_knowledge.similarity_threshold', label: 'Similarity needed', what: 'How alike an entry and the moment must be to be considered, on the model’s scale.', kind: 'percent' },
				{ key: 'world_knowledge.min_relevance_score', label: 'Relevance needed', what: 'The combined score an entry needs to be included: the filter that keeps unrelated lore out, however high it ranks.', kind: 'percent', step: 0.005 },
				{ key: 'world_knowledge.weight_keyword', label: 'Words', what: 'How much shared words count in an entry’s score.', kind: 'percent' },
				{ key: 'world_knowledge.weight_embedding', label: 'Meaning', what: 'How much likeness in meaning counts.', kind: 'percent' },
				{ key: 'world_knowledge.weight_importance', label: 'Importance', what: 'How much an entry’s own importance counts.', kind: 'percent' }
			]
		},
		{
			title: 'Vector index',
			what: 'The search index’s build settings. The defaults suit almost everyone.',
			fields: [
				{ key: 'vector_storage.hnsw_ef_construction', label: 'Build effort', what: 'HNSW’s ef_construction: higher builds a better index, more slowly.', kind: 'count', min: 50, max: 1000, step: 10 },
				{ key: 'vector_storage.hnsw_m', label: 'Connectivity', what: 'HNSW’s M: how many links each entry keeps. Higher uses more memory.', kind: 'count', min: 4, max: 64 }
			]
		}
	]
};
