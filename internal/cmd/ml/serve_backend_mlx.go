//go:build darwin && arm64 && mlx

package ml

import (
	"fmt"
	"log/slog"

	"forge.lthn.ai/core/cli/pkg/ml"
)

func createServeBackend() (ml.Backend, error) {
	if serveModelPath != "" {
		slog.Info("ml serve: loading native MLX backend", "path", serveModelPath)
		b, err := ml.NewMLXBackend(serveModelPath)
		if err != nil {
			return nil, fmt.Errorf("mlx backend: %w", err)
		}
		return b, nil
	}
	return ml.NewHTTPBackend(apiURL, modelName), nil
}
