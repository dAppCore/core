package agentic

import (
	"slices"

	"forge.lthn.ai/core/go/pkg/log"
)

// AllowanceService enforces agent quota limits. It provides pre-dispatch checks,
// runtime usage recording, and quota recovery for failed/cancelled jobs.
type AllowanceService struct {
	store AllowanceStore
}

// NewAllowanceService creates a new AllowanceService with the given store.
func NewAllowanceService(store AllowanceStore) *AllowanceService {
	return &AllowanceService{store: store}
}

// Check performs a pre-dispatch allowance check for the given agent and model.
// It verifies daily token limits, daily job limits, concurrent job limits, and
// model allowlists. Returns a QuotaCheckResult indicating whether the agent may proceed.
func (s *AllowanceService) Check(agentID, model string) (*QuotaCheckResult, error) {
	const op = "AllowanceService.Check"

	allowance, err := s.store.GetAllowance(agentID)
	if err != nil {
		return nil, log.E(op, "failed to get allowance", err)
	}

	usage, err := s.store.GetUsage(agentID)
	if err != nil {
		return nil, log.E(op, "failed to get usage", err)
	}

	result := &QuotaCheckResult{
		Allowed:         true,
		Status:          AllowanceOK,
		RemainingTokens: -1, // unlimited
		RemainingJobs:   -1, // unlimited
	}

	// Check model allowlist
	if len(allowance.ModelAllowlist) > 0 && model != "" {
		if !slices.Contains(allowance.ModelAllowlist, model) {
			result.Allowed = false
			result.Status = AllowanceExceeded
			result.Reason = "model not in allowlist: " + model
			return result, nil
		}
	}

	// Check daily token limit
	if allowance.DailyTokenLimit > 0 {
		remaining := allowance.DailyTokenLimit - usage.TokensUsed
		result.RemainingTokens = remaining
		if remaining <= 0 {
			result.Allowed = false
			result.Status = AllowanceExceeded
			result.Reason = "daily token limit exceeded"
			return result, nil
		}
		ratio := float64(usage.TokensUsed) / float64(allowance.DailyTokenLimit)
		if ratio >= 0.8 {
			result.Status = AllowanceWarning
		}
	}

	// Check daily job limit
	if allowance.DailyJobLimit > 0 {
		remaining := allowance.DailyJobLimit - usage.JobsStarted
		result.RemainingJobs = remaining
		if remaining <= 0 {
			result.Allowed = false
			result.Status = AllowanceExceeded
			result.Reason = "daily job limit exceeded"
			return result, nil
		}
	}

	// Check concurrent jobs
	if allowance.ConcurrentJobs > 0 && usage.ActiveJobs >= allowance.ConcurrentJobs {
		result.Allowed = false
		result.Status = AllowanceExceeded
		result.Reason = "concurrent job limit reached"
		return result, nil
	}

	// Check global model quota
	if model != "" {
		modelQuota, err := s.store.GetModelQuota(model)
		if err == nil && modelQuota.DailyTokenBudget > 0 {
			modelUsage, err := s.store.GetModelUsage(model)
			if err == nil && modelUsage >= modelQuota.DailyTokenBudget {
				result.Allowed = false
				result.Status = AllowanceExceeded
				result.Reason = "global model token budget exceeded for: " + model
				return result, nil
			}
		}
	}

	return result, nil
}

// RecordUsage processes a usage report, updating counters and handling quota recovery.
func (s *AllowanceService) RecordUsage(report UsageReport) error {
	const op = "AllowanceService.RecordUsage"

	totalTokens := report.TokensIn + report.TokensOut

	switch report.Event {
	case QuotaEventJobStarted:
		if err := s.store.IncrementUsage(report.AgentID, 0, 1); err != nil {
			return log.E(op, "failed to increment job count", err)
		}

	case QuotaEventJobCompleted:
		if err := s.store.IncrementUsage(report.AgentID, totalTokens, 0); err != nil {
			return log.E(op, "failed to record token usage", err)
		}
		if err := s.store.DecrementActiveJobs(report.AgentID); err != nil {
			return log.E(op, "failed to decrement active jobs", err)
		}
		// Record model-level usage
		if report.Model != "" {
			if err := s.store.IncrementModelUsage(report.Model, totalTokens); err != nil {
				return log.E(op, "failed to record model usage", err)
			}
		}

	case QuotaEventJobFailed:
		// Record partial usage, return 50% of tokens
		if err := s.store.IncrementUsage(report.AgentID, totalTokens, 0); err != nil {
			return log.E(op, "failed to record token usage", err)
		}
		if err := s.store.DecrementActiveJobs(report.AgentID); err != nil {
			return log.E(op, "failed to decrement active jobs", err)
		}
		returnAmount := totalTokens / 2
		if returnAmount > 0 {
			if err := s.store.ReturnTokens(report.AgentID, returnAmount); err != nil {
				return log.E(op, "failed to return tokens", err)
			}
		}
		// Still record model-level usage (net of return)
		if report.Model != "" {
			if err := s.store.IncrementModelUsage(report.Model, totalTokens-returnAmount); err != nil {
				return log.E(op, "failed to record model usage", err)
			}
		}

	case QuotaEventJobCancelled:
		// Return 100% of tokens
		if err := s.store.DecrementActiveJobs(report.AgentID); err != nil {
			return log.E(op, "failed to decrement active jobs", err)
		}
		if totalTokens > 0 {
			if err := s.store.ReturnTokens(report.AgentID, totalTokens); err != nil {
				return log.E(op, "failed to return tokens", err)
			}
		}
		// No model-level usage for cancelled jobs
	}

	return nil
}

// ResetAgent clears daily usage counters for the given agent (midnight reset).
func (s *AllowanceService) ResetAgent(agentID string) error {
	const op = "AllowanceService.ResetAgent"
	if err := s.store.ResetUsage(agentID); err != nil {
		return log.E(op, "failed to reset usage", err)
	}
	return nil
}
