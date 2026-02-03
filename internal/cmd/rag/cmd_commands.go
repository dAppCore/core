// Package rag provides RAG (Retrieval Augmented Generation) commands.
//
// Commands:
//   - core rag ingest: Ingest markdown files into Qdrant
//   - core rag query: Query the vector database
//   - core rag collections: List and manage collections
package rag

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddRAGCommands)
}

// AddRAGCommands registers the 'rag' command and all subcommands.
func AddRAGCommands(root *cobra.Command) {
	initFlags()
	ragCmd.AddCommand(ingestCmd)
	ragCmd.AddCommand(queryCmd)
	ragCmd.AddCommand(collectionsCmd)
	root.AddCommand(ragCmd)
}
