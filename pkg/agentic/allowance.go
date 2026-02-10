package agentic

import (
	"sync"
	"time"
)

// AllowanceStatus indicates the current state of an agent's quota.
type AllowanceStatus string

const (
	// AllowanceOK indicates the agent has remaining quota.
	AllowanceOK AllowanceStatus = "ok"
	// AllowanceWarning indicates the agent is at 80%+ usage.
	AllowanceWarning AllowanceStatus = "warning"
	// AllowanceExceeded indicates the agent has exceeded its quota.
	AllowanceExceeded AllowanceStatus = "exceeded"
)

// AgentAllowance defines the quota limits for a single agent.
type AgentAllowance struct {
	// AgentID is the unique identifier for the agent.
	AgentID string `json:"agent_id" yaml:"agent_id"`
	// DailyTokenLimit is the maximum tokens (in+out) per 24h. 0 means unlimited.
	DailyTokenLimit int64 `json:"daily_token_limit" yaml:"daily_token_limit"`
	// DailyJobLimit is the maximum jobs per 24h. 0 means unlimited.
	DailyJobLimit int `json:"daily_job_limit" yaml:"daily_job_limit"`
	// ConcurrentJobs is the maximum simultaneous jobs. 0 means unlimited.
	ConcurrentJobs int `json:"concurrent_jobs" yaml:"concurrent_jobs"`
	// MaxJobDuration is the maximum job duration before kill. 0 means unlimited.
	MaxJobDuration time.Duration `json:"max_job_duration" yaml:"max_job_duration"`
	// ModelAllowlist restricts which models this agent can use. Empty means all.
	ModelAllowlist []string `json:"model_allowlist,omitempty" yaml:"model_allowlist"`
}

// ModelQuota defines global per-model limits across all agents.
type ModelQuota struct {
	// Model is the model identifier (e.g. "claude-sonnet-4-5-20250929").
	Model string `json:"model" yaml:"model"`
	// DailyTokenBudget is the total tokens across all agents per 24h.
	DailyTokenBudget int64 `json:"daily_token_budget" yaml:"daily_token_budget"`
	// HourlyRateLimit is the max requests per hour.
	HourlyRateLimit int `json:"hourly_rate_limit" yaml:"hourly_rate_limit"`
	// CostCeiling stops all usage if cumulative cost exceeds this (in cents).
	CostCeiling int64 `json:"cost_ceiling" yaml:"cost_ceiling"`
}

// RepoLimit defines per-repository rate limits.
type RepoLimit struct {
	// Repo is the repository identifier (e.g. "owner/repo").
	Repo string `json:"repo" yaml:"repo"`
	// MaxDailyPRs is the maximum PRs per day. 0 means unlimited.
	MaxDailyPRs int `json:"max_daily_prs" yaml:"max_daily_prs"`
	// MaxDailyIssues is the maximum issues per day. 0 means unlimited.
	MaxDailyIssues int `json:"max_daily_issues" yaml:"max_daily_issues"`
	// CooldownAfterFailure is the wait time after a failure before retrying.
	CooldownAfterFailure time.Duration `json:"cooldown_after_failure" yaml:"cooldown_after_failure"`
}

// UsageRecord tracks an agent's current usage within a quota period.
type UsageRecord struct {
	// AgentID is the agent this record belongs to.
	AgentID string `json:"agent_id"`
	// TokensUsed is the total tokens consumed in the current period.
	TokensUsed int64 `json:"tokens_used"`
	// JobsStarted is the total jobs started in the current period.
	JobsStarted int `json:"jobs_started"`
	// ActiveJobs is the number of currently running jobs.
	ActiveJobs int `json:"active_jobs"`
	// PeriodStart is when the current quota period began.
	PeriodStart time.Time `json:"period_start"`
}

// QuotaCheckResult is the outcome of a pre-dispatch allowance check.
type QuotaCheckResult struct {
	// Allowed indicates whether the agent may proceed.
	Allowed bool `json:"allowed"`
	// Status is the current allowance state.
	Status AllowanceStatus `json:"status"`
	// Remaining is the number of tokens remaining in the period.
	RemainingTokens int64 `json:"remaining_tokens"`
	// RemainingJobs is the number of jobs remaining in the period.
	RemainingJobs int `json:"remaining_jobs"`
	// Reason explains why the check failed (if !Allowed).
	Reason string `json:"reason,omitempty"`
}

// QuotaEvent represents a change in quota usage, used for recovery.
type QuotaEvent string

const (
	// QuotaEventJobStarted deducts quota when a job begins.
	QuotaEventJobStarted QuotaEvent = "job_started"
	// QuotaEventJobCompleted deducts nothing (already counted).
	QuotaEventJobCompleted QuotaEvent = "job_completed"
	// QuotaEventJobFailed returns 50% of token quota.
	QuotaEventJobFailed QuotaEvent = "job_failed"
	// QuotaEventJobCancelled returns 100% of token quota.
	QuotaEventJobCancelled QuotaEvent = "job_cancelled"
)

// UsageReport is emitted by the agent runner to report token consumption.
type UsageReport struct {
	// AgentID is the agent that consumed tokens.
	AgentID string `json:"agent_id"`
	// JobID identifies the specific job.
	JobID string `json:"job_id"`
	// Model is the model used.
	Model string `json:"model"`
	// TokensIn is the number of input tokens consumed.
	TokensIn int64 `json:"tokens_in"`
	// TokensOut is the number of output tokens consumed.
	TokensOut int64 `json:"tokens_out"`
	// Event is the type of quota event.
	Event QuotaEvent `json:"event"`
	// Timestamp is when the usage occurred.
	Timestamp time.Time `json:"timestamp"`
}

// AllowanceStore is the interface for persisting and querying allowance data.
// Implementations may use Redis, SQLite, or any backing store.
type AllowanceStore interface {
	// GetAllowance returns the quota limits for an agent.
	GetAllowance(agentID string) (*AgentAllowance, error)
	// SetAllowance persists quota limits for an agent.
	SetAllowance(a *AgentAllowance) error
	// GetUsage returns the current usage record for an agent.
	GetUsage(agentID string) (*UsageRecord, error)
	// IncrementUsage atomically adds to an agent's usage counters.
	IncrementUsage(agentID string, tokens int64, jobs int) error
	// DecrementActiveJobs reduces the active job count by 1.
	DecrementActiveJobs(agentID string) error
	// ReturnTokens adds tokens back to the agent's remaining quota.
	ReturnTokens(agentID string, tokens int64) error
	// ResetUsage clears usage counters for an agent (daily reset).
	ResetUsage(agentID string) error
	// GetModelQuota returns global limits for a model.
	GetModelQuota(model string) (*ModelQuota, error)
	// GetModelUsage returns current token usage for a model.
	GetModelUsage(model string) (int64, error)
	// IncrementModelUsage atomically adds to a model's usage counter.
	IncrementModelUsage(model string, tokens int64) error
}

// MemoryStore is an in-memory AllowanceStore for testing and single-node use.
type MemoryStore struct {
	mu          sync.RWMutex
	allowances  map[string]*AgentAllowance
	usage       map[string]*UsageRecord
	modelQuotas map[string]*ModelQuota
	modelUsage  map[string]int64
}

// NewMemoryStore creates a new in-memory allowance store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		allowances:  make(map[string]*AgentAllowance),
		usage:       make(map[string]*UsageRecord),
		modelQuotas: make(map[string]*ModelQuota),
		modelUsage:  make(map[string]int64),
	}
}

// GetAllowance returns the quota limits for an agent.
func (m *MemoryStore) GetAllowance(agentID string) (*AgentAllowance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.allowances[agentID]
	if !ok {
		return nil, &APIError{Code: 404, Message: "allowance not found for agent: " + agentID}
	}
	cp := *a
	return &cp, nil
}

// SetAllowance persists quota limits for an agent.
func (m *MemoryStore) SetAllowance(a *AgentAllowance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.allowances[a.AgentID] = &cp
	return nil
}

// GetUsage returns the current usage record for an agent.
func (m *MemoryStore) GetUsage(agentID string) (*UsageRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usage[agentID]
	if !ok {
		return &UsageRecord{
			AgentID:     agentID,
			PeriodStart: startOfDay(time.Now().UTC()),
		}, nil
	}
	cp := *u
	return &cp, nil
}

// IncrementUsage atomically adds to an agent's usage counters.
func (m *MemoryStore) IncrementUsage(agentID string, tokens int64, jobs int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usage[agentID]
	if !ok {
		u = &UsageRecord{
			AgentID:     agentID,
			PeriodStart: startOfDay(time.Now().UTC()),
		}
		m.usage[agentID] = u
	}
	u.TokensUsed += tokens
	u.JobsStarted += jobs
	if jobs > 0 {
		u.ActiveJobs += jobs
	}
	return nil
}

// DecrementActiveJobs reduces the active job count by 1.
func (m *MemoryStore) DecrementActiveJobs(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usage[agentID]
	if !ok {
		return nil
	}
	if u.ActiveJobs > 0 {
		u.ActiveJobs--
	}
	return nil
}

// ReturnTokens adds tokens back to the agent's remaining quota.
func (m *MemoryStore) ReturnTokens(agentID string, tokens int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usage[agentID]
	if !ok {
		return nil
	}
	u.TokensUsed -= tokens
	if u.TokensUsed < 0 {
		u.TokensUsed = 0
	}
	return nil
}

// ResetUsage clears usage counters for an agent.
func (m *MemoryStore) ResetUsage(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usage[agentID] = &UsageRecord{
		AgentID:     agentID,
		PeriodStart: startOfDay(time.Now().UTC()),
	}
	return nil
}

// GetModelQuota returns global limits for a model.
func (m *MemoryStore) GetModelQuota(model string) (*ModelQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	q, ok := m.modelQuotas[model]
	if !ok {
		return nil, &APIError{Code: 404, Message: "model quota not found: " + model}
	}
	cp := *q
	return &cp, nil
}

// GetModelUsage returns current token usage for a model.
func (m *MemoryStore) GetModelUsage(model string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.modelUsage[model], nil
}

// IncrementModelUsage atomically adds to a model's usage counter.
func (m *MemoryStore) IncrementModelUsage(model string, tokens int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelUsage[model] += tokens
	return nil
}

// SetModelQuota sets global limits for a model (used in testing).
func (m *MemoryStore) SetModelQuota(q *ModelQuota) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *q
	m.modelQuotas[q.Model] = &cp
}

// startOfDay returns midnight UTC for the given time.
func startOfDay(t time.Time) time.Time {
	y, mo, d := t.Date()
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}
