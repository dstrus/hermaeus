import { m } from '#lib/copy/messages.ts';
import type { Feature } from './schema.ts';

const f = m.features.memories;

// Memories: how SkyrimNet turns what happens into memories, and which it
// recalls. Memory.yaml, in the dashboard's sections.
export const memory: Feature = {
	id: 'memories',
	file: 'Memory',
	title: f.title,
	what: f.what,
	sections: [
		{
			title: f.general.title,
			fields: [
				{ key: 'performance.enable_background_processing', label: f.general.performanceEnableBackgroundProcessing.label, what: f.general.performanceEnableBackgroundProcessing.what, kind: 'switch' },
				{ key: 'memory_retrieval.max_memories_per_retrieval', label: f.general.memoryRetrievalMaxMemoriesPerRetrieval.label, what: f.general.memoryRetrievalMaxMemoriesPerRetrieval.what, kind: 'count', min: 1, max: 100 },
				{ key: 'memory_lifecycle.max_memories_per_actor', label: f.general.memoryLifecycleMaxMemoriesPerActor.label, what: f.general.memoryLifecycleMaxMemoriesPerActor.what, kind: 'count', min: 100, max: 50000, step: 100 },
				{ key: 'generation_rules.min_importance_for_storage', label: f.general.generationRulesMinImportanceForStorage.label, what: f.general.generationRulesMinImportanceForStorage.what, kind: 'percent' }
			]
		},
		{
			title: f.embeddingModel.title,
			what: f.embeddingModel.what,
			fields: [
				{
					key: 'embedding.model',
					label: f.embeddingModel.embeddingModel.label,
					what: f.embeddingModel.embeddingModel.what,
					kind: 'choice',
					custom: true,
					choices: [
						{ value: 'bekko-embedding-v1-a25m', label: f.embeddingModel.embeddingModel.choices.bekkoEmbeddingV1A25m.label, what: f.embeddingModel.embeddingModel.choices.bekkoEmbeddingV1A25m.what },
						{ value: 'all-MiniLM-L6-v2', label: 'all-MiniLM-L6-v2', what: f.embeddingModel.embeddingModel.choices.allMiniLML6V2.what }
					]
				}
			]
		},
		{
			title: f.performance.title,
			fields: [
				{ key: 'performance.max_concurrent_generations', label: f.performance.performanceMaxConcurrentGenerations.label, what: f.performance.performanceMaxConcurrentGenerations.what, kind: 'count', min: 1, max: 50 },
				{ key: 'performance.reembed_threads', label: f.performance.performanceReembedThreads.label, what: f.performance.performanceReembedThreads.what, kind: 'count', min: 1, max: 4 },
				{ key: 'performance.memory_batch_size', label: f.performance.performanceMemoryBatchSize.label, what: f.performance.performanceMemoryBatchSize.what, kind: 'count', min: 10, max: 1000, step: 10 }
			]
		},
		{
			title: f.segments.title,
			what: f.segments.what,
			fields: [
				{ key: 'boundary_config.max_time_gap_minutes_real', label: f.segments.boundaryConfigMaxTimeGapMinutesReal.label, what: f.segments.boundaryConfigMaxTimeGapMinutesReal.what, kind: 'minutes', min: 1, max: 1440 },
				{ key: 'boundary_config.min_segment_duration_minutes_real', label: f.segments.boundaryConfigMinSegmentDurationMinutesReal.label, what: f.segments.boundaryConfigMinSegmentDurationMinutesReal.what, kind: 'minutes', min: 1, max: 120 },
				{ key: 'boundary_config.max_segment_duration_minutes_real', label: f.segments.boundaryConfigMaxSegmentDurationMinutesReal.label, what: f.segments.boundaryConfigMaxSegmentDurationMinutesReal.what, kind: 'minutes', min: 1, max: 2880 },
				{ key: 'boundary_config.max_events_per_segment', label: f.segments.boundaryConfigMaxEventsPerSegment.label, what: f.segments.boundaryConfigMaxEventsPerSegment.what, kind: 'count', min: 10, max: 1000, step: 10, unit: f.segments.boundaryConfigMaxEventsPerSegment.unit },
				{ key: 'boundary_config.min_events_per_segment', label: f.segments.boundaryConfigMinEventsPerSegment.label, what: f.segments.boundaryConfigMinEventsPerSegment.what, kind: 'count', min: 1, max: 50, unit: f.segments.boundaryConfigMinEventsPerSegment.unit }
			]
		},
		{
			title: f.recall.title,
			fields: [
				{ key: 'vector_storage.similarity_threshold', label: f.recall.vectorStorageSimilarityThreshold.label, what: f.recall.vectorStorageSimilarityThreshold.what, kind: 'percent' },
				{ key: 'vector_storage.min_relevance_score', label: f.recall.vectorStorageMinRelevanceScore.label, what: f.recall.vectorStorageMinRelevanceScore.what, kind: 'percent' }
			]
		},
		{
			title: f.weights.title,
			what: f.weights.what,
			fields: [
				{ key: 'scoring_weights.semantic', label: f.weights.scoringWeightsSemantic.label, what: f.weights.scoringWeightsSemantic.what, kind: 'percent' },
				{ key: 'scoring_weights.temporal', label: f.weights.scoringWeightsTemporal.label, what: f.weights.scoringWeightsTemporal.what, kind: 'percent' },
				{ key: 'scoring_weights.actor', label: f.weights.scoringWeightsActor.label, what: f.weights.scoringWeightsActor.what, kind: 'percent' },
				{ key: 'scoring_weights.emotional', label: f.weights.scoringWeightsEmotional.label, what: f.weights.scoringWeightsEmotional.what, kind: 'percent' },
				{ key: 'scoring_weights.location', label: f.weights.scoringWeightsLocation.label, what: f.weights.scoringWeightsLocation.what, kind: 'percent' },
				{ key: 'scoring_weights.keyword', label: f.weights.scoringWeightsKeyword.label, what: f.weights.scoringWeightsKeyword.what, kind: 'percent' }
			]
		},
		{
			title: f.worldKnowledge.title,
			what: f.worldKnowledge.what,
			fields: [
				{ key: 'world_knowledge.max_semantic_results', label: f.worldKnowledge.worldKnowledgeMaxSemanticResults.label, what: f.worldKnowledge.worldKnowledgeMaxSemanticResults.what, kind: 'count', min: 1, max: 20 },
				{ key: 'world_knowledge.max_always_inject', label: f.worldKnowledge.worldKnowledgeMaxAlwaysInject.label, what: f.worldKnowledge.worldKnowledgeMaxAlwaysInject.what, kind: 'count', min: 0, max: 50 },
				{ key: 'world_knowledge.similarity_threshold', label: f.worldKnowledge.worldKnowledgeSimilarityThreshold.label, what: f.worldKnowledge.worldKnowledgeSimilarityThreshold.what, kind: 'percent' },
				{ key: 'world_knowledge.min_relevance_score', label: f.worldKnowledge.worldKnowledgeMinRelevanceScore.label, what: f.worldKnowledge.worldKnowledgeMinRelevanceScore.what, kind: 'percent', step: 0.005 },
				{ key: 'world_knowledge.weight_keyword', label: f.worldKnowledge.worldKnowledgeWeightKeyword.label, what: f.worldKnowledge.worldKnowledgeWeightKeyword.what, kind: 'percent' },
				{ key: 'world_knowledge.weight_embedding', label: f.worldKnowledge.worldKnowledgeWeightEmbedding.label, what: f.worldKnowledge.worldKnowledgeWeightEmbedding.what, kind: 'percent' },
				{ key: 'world_knowledge.weight_importance', label: f.worldKnowledge.worldKnowledgeWeightImportance.label, what: f.worldKnowledge.worldKnowledgeWeightImportance.what, kind: 'percent' }
			]
		},
		{
			title: f.vectorIndex.title,
			what: f.vectorIndex.what,
			fields: [
				{ key: 'vector_storage.hnsw_ef_construction', label: f.vectorIndex.vectorStorageHnswEfConstruction.label, what: f.vectorIndex.vectorStorageHnswEfConstruction.what, kind: 'count', min: 50, max: 1000, step: 10 },
				{ key: 'vector_storage.hnsw_m', label: f.vectorIndex.vectorStorageHnswM.label, what: f.vectorIndex.vectorStorageHnswM.what, kind: 'count', min: 4, max: 64 }
			]
		}
	]
};
