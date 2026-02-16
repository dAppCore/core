package ml

import (
	"fmt"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	convertInput     string
	convertConfig    string
	convertOutputDir string
	convertBaseModel string
)

var convertCmd = &cli.Command{
	Use:   "convert",
	Short: "Convert MLX LoRA adapter to PEFT format",
	Long:  "Converts an MLX safetensors LoRA adapter to HuggingFace PEFT format for Ollama.",
	RunE:  runConvert,
}

func init() {
	convertCmd.Flags().StringVar(&convertInput, "input", "", "Input safetensors file (required)")
	convertCmd.Flags().StringVar(&convertConfig, "config", "", "Adapter config JSON (required)")
	convertCmd.Flags().StringVar(&convertOutputDir, "output-dir", "", "Output directory (required)")
	convertCmd.Flags().StringVar(&convertBaseModel, "base-model", "", "Base model name for adapter_config.json")
	convertCmd.MarkFlagRequired("input")
	convertCmd.MarkFlagRequired("config")
	convertCmd.MarkFlagRequired("output-dir")
}

func runConvert(cmd *cli.Command, args []string) error {
	if err := ml.ConvertMLXtoPEFT(convertInput, convertConfig, convertOutputDir, convertBaseModel); err != nil {
		return fmt.Errorf("convert to PEFT: %w", err)
	}
	fmt.Printf("PEFT adapter written to %s\n", convertOutputDir)
	return nil
}
