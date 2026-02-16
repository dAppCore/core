package ml

import (
	"fmt"
	"sort"
)

// RunCompare reads two score files and prints a comparison table for each
// model showing Old, New, and Delta values for every metric.
func RunCompare(oldPath, newPath string) error {
	oldOutput, err := ReadScorerOutput(oldPath)
	if err != nil {
		return fmt.Errorf("read old file: %w", err)
	}

	newOutput, err := ReadScorerOutput(newPath)
	if err != nil {
		return fmt.Errorf("read new file: %w", err)
	}

	// Collect all models present in both files.
	models := make(map[string]bool)
	for m := range oldOutput.ModelAverages {
		models[m] = true
	}
	for m := range newOutput.ModelAverages {
		models[m] = true
	}

	// Sort model names for deterministic output.
	sortedModels := make([]string, 0, len(models))
	for m := range models {
		sortedModels = append(sortedModels, m)
	}
	sort.Strings(sortedModels)

	for _, model := range sortedModels {
		oldAvgs := oldOutput.ModelAverages[model]
		newAvgs := newOutput.ModelAverages[model]

		if oldAvgs == nil && newAvgs == nil {
			continue
		}

		fmt.Printf("\nModel: %s\n", model)
		fmt.Printf("%-25s %11s  %11s  %6s\n", "", "Old", "New", "Delta")

		// Collect all metrics from both old and new.
		metrics := make(map[string]bool)
		for k := range oldAvgs {
			metrics[k] = true
		}
		for k := range newAvgs {
			metrics[k] = true
		}

		sortedMetrics := make([]string, 0, len(metrics))
		for k := range metrics {
			sortedMetrics = append(sortedMetrics, k)
		}
		sort.Strings(sortedMetrics)

		for _, metric := range sortedMetrics {
			oldVal := oldAvgs[metric]
			newVal := newAvgs[metric]
			delta := newVal - oldVal

			deltaStr := fmt.Sprintf("%+.2f", delta)

			fmt.Printf("%-25s %11.2f  %11.2f  %6s\n", metric, oldVal, newVal, deltaStr)
		}
	}

	return nil
}
