package rag

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IngestConfig holds ingestion configuration.
type IngestConfig struct {
	Directory  string
	Collection string
	Recreate   bool
	Verbose    bool
	BatchSize  int
	Chunk      ChunkConfig
}

// DefaultIngestConfig returns default ingestion configuration.
func DefaultIngestConfig() IngestConfig {
	return IngestConfig{
		Collection: "hostuk-docs",
		BatchSize:  100,
		Chunk:      DefaultChunkConfig(),
	}
}

// IngestStats holds statistics from ingestion.
type IngestStats struct {
	Files  int
	Chunks int
	Errors int
}

// IngestProgress is called during ingestion to report progress.
type IngestProgress func(file string, chunks int, total int)

// Ingest processes a directory of documents and stores them in Qdrant.
func Ingest(ctx context.Context, qdrant *QdrantClient, ollama *OllamaClient, cfg IngestConfig, progress IngestProgress) (*IngestStats, error) {
	stats := &IngestStats{}

	// Validate batch size to prevent infinite loop
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100 // Safe default
	}

	// Resolve directory
	absDir, err := filepath.Abs(cfg.Directory)
	if err != nil {
		return nil, fmt.Errorf("error resolving directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("error accessing directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absDir)
	}

	// Check/create collection
	exists, err := qdrant.CollectionExists(ctx, cfg.Collection)
	if err != nil {
		return nil, fmt.Errorf("error checking collection: %w", err)
	}

	if cfg.Recreate && exists {
		if err := qdrant.DeleteCollection(ctx, cfg.Collection); err != nil {
			return nil, fmt.Errorf("error deleting collection: %w", err)
		}
		exists = false
	}

	if !exists {
		vectorDim := ollama.EmbedDimension()
		if err := qdrant.CreateCollection(ctx, cfg.Collection, vectorDim); err != nil {
			return nil, fmt.Errorf("error creating collection: %w", err)
		}
	}

	// Find markdown files
	var files []string
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && ShouldProcess(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no markdown files found in %s", absDir)
	}

	// Process files
	var points []Point
	for _, filePath := range files {
		relPath, err := filepath.Rel(absDir, filePath)
		if err != nil {
			stats.Errors++
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			stats.Errors++
			continue
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			continue
		}

		// Chunk the content
		category := Category(relPath)
		chunks := ChunkMarkdown(string(content), cfg.Chunk)

		for _, chunk := range chunks {
			// Generate embedding
			embedding, err := ollama.Embed(ctx, chunk.Text)
			if err != nil {
				stats.Errors++
				if cfg.Verbose {
					fmt.Printf("  Error embedding %s chunk %d: %v\n", relPath, chunk.Index, err)
				}
				continue
			}

			// Create point
			points = append(points, Point{
				ID:     ChunkID(relPath, chunk.Index, chunk.Text),
				Vector: embedding,
				Payload: map[string]any{
					"text":        chunk.Text,
					"source":      relPath,
					"section":     chunk.Section,
					"category":    category,
					"chunk_index": chunk.Index,
				},
			})
			stats.Chunks++
		}

		stats.Files++
		if progress != nil {
			progress(relPath, stats.Chunks, len(files))
		}
	}

	// Batch upsert to Qdrant
	if len(points) > 0 {
		for i := 0; i < len(points); i += cfg.BatchSize {
			end := i + cfg.BatchSize
			if end > len(points) {
				end = len(points)
			}
			batch := points[i:end]
			if err := qdrant.UpsertPoints(ctx, cfg.Collection, batch); err != nil {
				return stats, fmt.Errorf("error upserting batch %d: %w", i/cfg.BatchSize+1, err)
			}
		}
	}

	return stats, nil
}

// IngestFile processes a single file and stores it in Qdrant.
func IngestFile(ctx context.Context, qdrant *QdrantClient, ollama *OllamaClient, collection string, filePath string, chunkCfg ChunkConfig) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	if len(strings.TrimSpace(string(content))) == 0 {
		return 0, nil
	}

	category := Category(filePath)
	chunks := ChunkMarkdown(string(content), chunkCfg)

	var points []Point
	for _, chunk := range chunks {
		embedding, err := ollama.Embed(ctx, chunk.Text)
		if err != nil {
			return 0, fmt.Errorf("error embedding chunk %d: %w", chunk.Index, err)
		}

		points = append(points, Point{
			ID:     ChunkID(filePath, chunk.Index, chunk.Text),
			Vector: embedding,
			Payload: map[string]any{
				"text":        chunk.Text,
				"source":      filePath,
				"section":     chunk.Section,
				"category":    category,
				"chunk_index": chunk.Index,
			},
		})
	}

	if err := qdrant.UpsertPoints(ctx, collection, points); err != nil {
		return 0, fmt.Errorf("error upserting points: %w", err)
	}

	return len(points), nil
}
