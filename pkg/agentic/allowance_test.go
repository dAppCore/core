package agentic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- MemoryStore tests ---

func TestMemoryStore_SetGetAllowance_Good(t *testing.T) {
	store := NewMemoryStore()
	a := &AgentAllowance{
		AgentID:         "agent-1",
		DailyTokenLimit: 100000,
		DailyJobLimit:   10,
		ConcurrentJobs:  2,
		MaxJobDuration:  30 * time.Minute,
		ModelAllowlist:  []string{"claude-sonnet-4-5-20250929"},
	}

	err := store.SetAllowance(a)
	require.NoError(t, err)

	got, err := store.GetAllowance("agent-1")
	require.NoError(t, err)
	assert.Equal(t, a.AgentID, got.AgentID)
	assert.Equal(t, a.DailyTokenLimit, got.DailyTokenLimit)
	assert.Equal(t, a.DailyJobLimit, got.DailyJobLimit)
	assert.Equal(t, a.ConcurrentJobs, got.ConcurrentJobs)
	assert.Equal(t, a.ModelAllowlist, got.ModelAllowlist)
}

func TestMemoryStore_GetAllowance_Bad_NotFound(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.GetAllowance("nonexistent")
	require.Error(t, err)
}

func TestMemoryStore_IncrementUsage_Good(t *testing.T) {
	store := NewMemoryStore()

	err := store.IncrementUsage("agent-1", 5000, 1)
	require.NoError(t, err)

	usage, err := store.GetUsage("agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(5000), usage.TokensUsed)
	assert.Equal(t, 1, usage.JobsStarted)
	assert.Equal(t, 1, usage.ActiveJobs)
}

func TestMemoryStore_DecrementActiveJobs_Good(t *testing.T) {
	store := NewMemoryStore()

	_ = store.IncrementUsage("agent-1", 0, 2)
	_ = store.DecrementActiveJobs("agent-1")

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, 1, usage.ActiveJobs)
}

func TestMemoryStore_DecrementActiveJobs_Good_FloorAtZero(t *testing.T) {
	store := NewMemoryStore()

	_ = store.DecrementActiveJobs("agent-1") // no-op, no usage record
	_ = store.IncrementUsage("agent-1", 0, 0)
	_ = store.DecrementActiveJobs("agent-1") // should stay at 0

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, 0, usage.ActiveJobs)
}

func TestMemoryStore_ReturnTokens_Good(t *testing.T) {
	store := NewMemoryStore()

	_ = store.IncrementUsage("agent-1", 10000, 0)
	err := store.ReturnTokens("agent-1", 5000)
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, int64(5000), usage.TokensUsed)
}

func TestMemoryStore_ReturnTokens_Good_FloorAtZero(t *testing.T) {
	store := NewMemoryStore()

	_ = store.IncrementUsage("agent-1", 1000, 0)
	_ = store.ReturnTokens("agent-1", 5000) // more than used

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, int64(0), usage.TokensUsed)
}

func TestMemoryStore_ResetUsage_Good(t *testing.T) {
	store := NewMemoryStore()

	_ = store.IncrementUsage("agent-1", 50000, 5)
	err := store.ResetUsage("agent-1")
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, int64(0), usage.TokensUsed)
	assert.Equal(t, 0, usage.JobsStarted)
	assert.Equal(t, 0, usage.ActiveJobs)
}

func TestMemoryStore_ModelUsage_Good(t *testing.T) {
	store := NewMemoryStore()

	_ = store.IncrementModelUsage("claude-sonnet", 10000)
	_ = store.IncrementModelUsage("claude-sonnet", 5000)

	usage, err := store.GetModelUsage("claude-sonnet")
	require.NoError(t, err)
	assert.Equal(t, int64(15000), usage)
}

// --- AllowanceService.Check tests ---

func TestAllowanceServiceCheck_Good(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:         "agent-1",
		DailyTokenLimit: 100000,
		DailyJobLimit:   10,
		ConcurrentJobs:  2,
	})

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, AllowanceOK, result.Status)
	assert.Equal(t, int64(100000), result.RemainingTokens)
	assert.Equal(t, 10, result.RemainingJobs)
}

func TestAllowanceServiceCheck_Good_Warning(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:         "agent-1",
		DailyTokenLimit: 100000,
	})
	_ = store.IncrementUsage("agent-1", 85000, 0)

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, AllowanceWarning, result.Status)
	assert.Equal(t, int64(15000), result.RemainingTokens)
}

func TestAllowanceServiceCheck_Bad_TokenLimitExceeded(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:         "agent-1",
		DailyTokenLimit: 100000,
	})
	_ = store.IncrementUsage("agent-1", 100001, 0)

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, AllowanceExceeded, result.Status)
	assert.Contains(t, result.Reason, "daily token limit")
}

func TestAllowanceServiceCheck_Bad_JobLimitExceeded(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:       "agent-1",
		DailyJobLimit: 5,
	})
	_ = store.IncrementUsage("agent-1", 0, 5)

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "daily job limit")
}

func TestAllowanceServiceCheck_Bad_ConcurrentLimitReached(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:        "agent-1",
		ConcurrentJobs: 1,
	})
	_ = store.IncrementUsage("agent-1", 0, 1) // 1 active job

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "concurrent job limit")
}

func TestAllowanceServiceCheck_Bad_ModelNotInAllowlist(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:        "agent-1",
		ModelAllowlist: []string{"claude-sonnet-4-5-20250929"},
	})

	result, err := svc.Check("agent-1", "claude-opus-4-6")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "model not in allowlist")
}

func TestAllowanceServiceCheck_Good_ModelInAllowlist(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:        "agent-1",
		ModelAllowlist: []string{"claude-sonnet-4-5-20250929", "claude-haiku-4-5-20251001"},
	})

	result, err := svc.Check("agent-1", "claude-sonnet-4-5-20250929")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestAllowanceServiceCheck_Good_EmptyModelSkipsCheck(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID:        "agent-1",
		ModelAllowlist: []string{"claude-sonnet-4-5-20250929"},
	})

	result, err := svc.Check("agent-1", "")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestAllowanceServiceCheck_Bad_GlobalModelBudgetExceeded(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.SetAllowance(&AgentAllowance{
		AgentID: "agent-1",
	})
	store.SetModelQuota(&ModelQuota{
		Model:            "claude-opus-4-6",
		DailyTokenBudget: 500000,
	})
	_ = store.IncrementModelUsage("claude-opus-4-6", 500001)

	result, err := svc.Check("agent-1", "claude-opus-4-6")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "global model token budget")
}

func TestAllowanceServiceCheck_Bad_NoAllowance(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_, err := svc.Check("unknown-agent", "")
	require.Error(t, err)
}

// --- AllowanceService.RecordUsage tests ---

func TestAllowanceServiceRecordUsage_Good_JobStarted(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	err := svc.RecordUsage(UsageReport{
		AgentID: "agent-1",
		JobID:   "job-1",
		Event:   QuotaEventJobStarted,
	})
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, 1, usage.JobsStarted)
	assert.Equal(t, 1, usage.ActiveJobs)
	assert.Equal(t, int64(0), usage.TokensUsed)
}

func TestAllowanceServiceRecordUsage_Good_JobCompleted(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	// Start a job first
	_ = svc.RecordUsage(UsageReport{
		AgentID: "agent-1",
		JobID:   "job-1",
		Event:   QuotaEventJobStarted,
	})

	err := svc.RecordUsage(UsageReport{
		AgentID:   "agent-1",
		JobID:     "job-1",
		Model:     "claude-sonnet",
		TokensIn:  1000,
		TokensOut: 500,
		Event:     QuotaEventJobCompleted,
	})
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, int64(1500), usage.TokensUsed)
	assert.Equal(t, 0, usage.ActiveJobs)

	modelUsage, _ := store.GetModelUsage("claude-sonnet")
	assert.Equal(t, int64(1500), modelUsage)
}

func TestAllowanceServiceRecordUsage_Good_JobFailed_ReturnsHalf(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = svc.RecordUsage(UsageReport{
		AgentID: "agent-1",
		JobID:   "job-1",
		Event:   QuotaEventJobStarted,
	})

	err := svc.RecordUsage(UsageReport{
		AgentID:   "agent-1",
		JobID:     "job-1",
		Model:     "claude-sonnet",
		TokensIn:  1000,
		TokensOut: 1000,
		Event:     QuotaEventJobFailed,
	})
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	// 2000 tokens used, 1000 returned (50%) = 1000 net
	assert.Equal(t, int64(1000), usage.TokensUsed)
	assert.Equal(t, 0, usage.ActiveJobs)

	// Model sees net usage (2000 - 1000 = 1000)
	modelUsage, _ := store.GetModelUsage("claude-sonnet")
	assert.Equal(t, int64(1000), modelUsage)
}

func TestAllowanceServiceRecordUsage_Good_JobCancelled_ReturnsAll(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.IncrementUsage("agent-1", 5000, 1) // simulate pre-existing usage

	err := svc.RecordUsage(UsageReport{
		AgentID:   "agent-1",
		JobID:     "job-1",
		TokensIn:  500,
		TokensOut: 500,
		Event:     QuotaEventJobCancelled,
	})
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	// 5000 pre-existing - 1000 returned = 4000
	assert.Equal(t, int64(4000), usage.TokensUsed)
	assert.Equal(t, 0, usage.ActiveJobs)
}

// --- AllowanceService.ResetAgent tests ---

func TestAllowanceServiceResetAgent_Good(t *testing.T) {
	store := NewMemoryStore()
	svc := NewAllowanceService(store)

	_ = store.IncrementUsage("agent-1", 50000, 5)

	err := svc.ResetAgent("agent-1")
	require.NoError(t, err)

	usage, _ := store.GetUsage("agent-1")
	assert.Equal(t, int64(0), usage.TokensUsed)
	assert.Equal(t, 0, usage.JobsStarted)
}

// --- startOfDay helper test ---

func TestStartOfDay_Good(t *testing.T) {
	input := time.Date(2026, 2, 10, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, startOfDay(input))
}

// --- AllowanceStatus tests ---

func TestAllowanceStatus_Good_Values(t *testing.T) {
	assert.Equal(t, AllowanceStatus("ok"), AllowanceOK)
	assert.Equal(t, AllowanceStatus("warning"), AllowanceWarning)
	assert.Equal(t, AllowanceStatus("exceeded"), AllowanceExceeded)
}
