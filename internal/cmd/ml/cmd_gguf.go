package ml

import (
	"fmt"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	ggufInput  string
	ggufConfig string
	ggufOutput string
	ggufArch   string
)

var ggufCmd = &cli.Command{
	Use:   "gguf",
	Short: "Convert MLX LoRA adapter to GGUF format",
	Long:  "Converts an MLX safetensors LoRA adapter to GGUF v3 format for use with llama.cpp.",
	RunE:  runGGUF,
}

func init() {
	ggufCmd.Flags().StringVar(&ggufInput, "input", "", "Input safetensors file (required)")
	ggufCmd.Flags().StringVar(&ggufConfig, "config", "", "Adapter config JSON (required)")
	ggufCmd.Flags().StringVar(&ggufOutput, "output", "", "Output GGUF file (required)")
	ggufCmd.Flags().StringVar(&ggufArch, "arch", "gemma3", "GGUF architecture name")
	ggufCmd.MarkFlagRequired("input")
	ggufCmd.MarkFlagRequired("config")
	ggufCmd.MarkFlagRequired("output")
}

func runGGUF(cmd *cli.Command, args []string) error {
	if err := ml.ConvertMLXtoGGUFLoRA(ggufInput, ggufConfig, ggufOutput, ggufArch); err != nil {
		return fmt.Errorf("convert to GGUF: %w", err)
	}
	fmt.Printf("GGUF LoRA adapter written to %s\n", ggufOutput)
	return nil
}
