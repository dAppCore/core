package rag

import (
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Shared flags
var (
	qdrantHost string
	qdrantPort int
	ollamaHost string
	ollamaPort int
	model      string
	verbose    bool
)

var ragCmd = &cobra.Command{
	Use:   "rag",
	Short: i18n.T("cmd.rag.short"),
	Long:  i18n.T("cmd.rag.long"),
}

func initFlags() {
	// Qdrant connection flags (persistent) - defaults to localhost for local development
	ragCmd.PersistentFlags().StringVar(&qdrantHost, "qdrant-host", "localhost", i18n.T("cmd.rag.flag.qdrant_host"))
	ragCmd.PersistentFlags().IntVar(&qdrantPort, "qdrant-port", 6334, i18n.T("cmd.rag.flag.qdrant_port"))

	// Ollama connection flags (persistent) - defaults to localhost for local development
	ragCmd.PersistentFlags().StringVar(&ollamaHost, "ollama-host", "localhost", i18n.T("cmd.rag.flag.ollama_host"))
	ragCmd.PersistentFlags().IntVar(&ollamaPort, "ollama-port", 11434, i18n.T("cmd.rag.flag.ollama_port"))
	ragCmd.PersistentFlags().StringVar(&model, "model", "nomic-embed-text", i18n.T("cmd.rag.flag.model"))

	// Verbose flag (persistent)
	ragCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, i18n.T("common.flag.verbose"))

	// Ingest command flags
	ingestCmd.Flags().StringVar(&collection, "collection", "hostuk-docs", i18n.T("cmd.rag.ingest.flag.collection"))
	ingestCmd.Flags().BoolVar(&recreate, "recreate", false, i18n.T("cmd.rag.ingest.flag.recreate"))
	ingestCmd.Flags().IntVar(&chunkSize, "chunk-size", 500, i18n.T("cmd.rag.ingest.flag.chunk_size"))
	ingestCmd.Flags().IntVar(&chunkOverlap, "chunk-overlap", 50, i18n.T("cmd.rag.ingest.flag.chunk_overlap"))

	// Query command flags
	queryCmd.Flags().StringVar(&queryCollection, "collection", "hostuk-docs", i18n.T("cmd.rag.query.flag.collection"))
	queryCmd.Flags().IntVar(&limit, "top", 5, i18n.T("cmd.rag.query.flag.top"))
	queryCmd.Flags().Float32Var(&threshold, "threshold", 0.5, i18n.T("cmd.rag.query.flag.threshold"))
	queryCmd.Flags().StringVar(&category, "category", "", i18n.T("cmd.rag.query.flag.category"))
	queryCmd.Flags().StringVar(&format, "format", "text", i18n.T("cmd.rag.query.flag.format"))

	// Collections command flags
	collectionsCmd.Flags().BoolVar(&listCollections, "list", false, i18n.T("cmd.rag.collections.flag.list"))
	collectionsCmd.Flags().BoolVar(&showStats, "stats", false, i18n.T("cmd.rag.collections.flag.stats"))
	collectionsCmd.Flags().StringVar(&deleteCollection, "delete", "", i18n.T("cmd.rag.collections.flag.delete"))
}
