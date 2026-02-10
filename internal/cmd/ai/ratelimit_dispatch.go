package ai

import (
	"context"

	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/ratelimit"
)

// executeWithRateLimit wraps an agent execution with rate limiting logic.
// It estimates token usage, waits for capacity, executes the runner, and records usage.
func executeWithRateLimit(ctx context.Context, model, prompt string, runner func() (bool, int, error)) (bool, int, error) {
	rl, err := ratelimit.New()
	if err != nil {
		log.Warn("Failed to initialize rate limiter, proceeding without limits", "error", err)
		return runner()
	}

	if err := rl.Load(); err != nil {
		log.Warn("Failed to load rate limit state", "error", err)
	}

	// Estimate tokens from prompt length (1 token ≈ 4 chars)
	estTokens := len(prompt) / 4
	if estTokens == 0 {
		estTokens = 1
	}

	log.Info("Checking rate limits", "model", model, "est_tokens", estTokens)

	if err := rl.WaitForCapacity(ctx, model, estTokens); err != nil {
		return false, -1, err
	}

	success, exitCode, runErr := runner()

	// Record usage with conservative output estimate (actual tokens unknown from shell runner).
	outputEst := estTokens / 10
	if outputEst < 50 {
		outputEst = 50
	}
	rl.RecordUsage(model, estTokens, outputEst)

	if err := rl.Persist(); err != nil {
		log.Warn("Failed to persist rate limit state", "error", err)
	}

	return success, exitCode, runErr
}
