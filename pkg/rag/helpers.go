package rag

import (
	"context"
	"fmt"
)

// QueryDocs queries the RAG database with default clients.
func QueryDocs(ctx context.Context, question, collectionName string, topK int) ([]QueryResult, error) {
	qdrantClient, err := NewQdrantClient(DefaultQdrantConfig())
	if err != nil {
		return nil, err
	}
	defer func() { _ = qdrantClient.Close() }()

	ollamaClient, err := NewOllamaClient(DefaultOllamaConfig())
	if err != nil {
		return nil, err
	}

	cfg := DefaultQueryConfig()
	cfg.Collection = collectionName
	cfg.Limit = uint64(topK)

	return Query(ctx, qdrantClient, ollamaClient, question, cfg)
}

// QueryDocsContext queries the RAG database and returns context-formatted results.
func QueryDocsContext(ctx context.Context, question, collectionName string, topK int) (string, error) {
	results, err := QueryDocs(ctx, question, collectionName, topK)
	if err != nil {
		return "", err
	}
	return FormatResultsContext(results), nil
}

// IngestDirectory ingests all documents in a directory with default clients.
func IngestDirectory(ctx context.Context, directory, collectionName string, recreateCollection bool) error {
	qdrantClient, err := NewQdrantClient(DefaultQdrantConfig())
	if err != nil {
		return err
	}
	defer func() { _ = qdrantClient.Close() }()

	if err := qdrantClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("qdrant health check failed: %w", err)
	}

	ollamaClient, err := NewOllamaClient(DefaultOllamaConfig())
	if err != nil {
		return err
	}

	if err := ollamaClient.VerifyModel(ctx); err != nil {
		return err
	}

	cfg := DefaultIngestConfig()
	cfg.Directory = directory
	cfg.Collection = collectionName
	cfg.Recreate = recreateCollection

	_, err = Ingest(ctx, qdrantClient, ollamaClient, cfg, nil)
	return err
}

// IngestSingleFile ingests a single file with default clients.
func IngestSingleFile(ctx context.Context, filePath, collectionName string) (int, error) {
	qdrantClient, err := NewQdrantClient(DefaultQdrantConfig())
	if err != nil {
		return 0, err
	}
	defer func() { _ = qdrantClient.Close() }()

	if err := qdrantClient.HealthCheck(ctx); err != nil {
		return 0, fmt.Errorf("qdrant health check failed: %w", err)
	}

	ollamaClient, err := NewOllamaClient(DefaultOllamaConfig())
	if err != nil {
		return 0, err
	}

	if err := ollamaClient.VerifyModel(ctx); err != nil {
		return 0, err
	}

	return IngestFile(ctx, qdrantClient, ollamaClient, collectionName, filePath, DefaultChunkConfig())
}
