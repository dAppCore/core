package jobrunner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock source ---

type mockSource struct {
	name    string
	signals []*PipelineSignal
	reports []*ActionResult
	mu      sync.Mutex
}

func (m *mockSource) Name() string { return m.name }

func (m *mockSource) Poll(_ context.Context) ([]*PipelineSignal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals, nil
}

func (m *mockSource) Report(_ context.Context, result *ActionResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, result)
	return nil
}

// --- Mock handler ---

type mockHandler struct {
	name     string
	matchFn  func(*PipelineSignal) bool
	executed []*PipelineSignal
	mu       sync.Mutex
}

func (m *mockHandler) Name() string { return m.name }

func (m *mockHandler) Match(sig *PipelineSignal) bool {
	if m.matchFn != nil {
		return m.matchFn(sig)
	}
	return true
}

func (m *mockHandler) Execute(_ context.Context, sig *PipelineSignal) (*ActionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executed = append(m.executed, sig)
	return &ActionResult{
		Action:    m.name,
		RepoOwner: sig.RepoOwner,
		RepoName:  sig.RepoName,
		PRNumber:  sig.PRNumber,
		Success:   true,
		Timestamp: time.Now(),
	}, nil
}

func TestPoller_RunOnce_Good(t *testing.T) {
	sig := &PipelineSignal{
		EpicNumber:  1,
		ChildNumber: 2,
		PRNumber:    10,
		RepoOwner:   "host-uk",
		RepoName:    "core-php",
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
		Mergeable:   "MERGEABLE",
	}

	src := &mockSource{
		name:    "test-source",
		signals: []*PipelineSignal{sig},
	}

	handler := &mockHandler{
		name: "test-handler",
		matchFn: func(s *PipelineSignal) bool {
			return s.PRNumber == 10
		},
	}

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{src},
		Handlers: []JobHandler{handler},
	})

	err := p.RunOnce(context.Background())
	require.NoError(t, err)

	// Handler should have been called with our signal.
	handler.mu.Lock()
	defer handler.mu.Unlock()
	require.Len(t, handler.executed, 1)
	assert.Equal(t, 10, handler.executed[0].PRNumber)

	// Source should have received a report.
	src.mu.Lock()
	defer src.mu.Unlock()
	require.Len(t, src.reports, 1)
	assert.Equal(t, "test-handler", src.reports[0].Action)
	assert.True(t, src.reports[0].Success)
	assert.Equal(t, 1, src.reports[0].Cycle)
	assert.Equal(t, 1, src.reports[0].EpicNumber)
	assert.Equal(t, 2, src.reports[0].ChildNumber)

	// Cycle counter should have incremented.
	assert.Equal(t, 1, p.Cycle())
}

func TestPoller_RunOnce_Good_NoSignals(t *testing.T) {
	src := &mockSource{
		name:    "empty-source",
		signals: nil,
	}

	handler := &mockHandler{
		name: "unused-handler",
	}

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{src},
		Handlers: []JobHandler{handler},
	})

	err := p.RunOnce(context.Background())
	require.NoError(t, err)

	// Handler should not have been called.
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Empty(t, handler.executed)

	// Source should not have received reports.
	src.mu.Lock()
	defer src.mu.Unlock()
	assert.Empty(t, src.reports)

	assert.Equal(t, 1, p.Cycle())
}

func TestPoller_RunOnce_Good_NoMatchingHandler(t *testing.T) {
	sig := &PipelineSignal{
		EpicNumber:  5,
		ChildNumber: 8,
		PRNumber:    42,
		RepoOwner:   "host-uk",
		RepoName:    "core-tenant",
		PRState:     "OPEN",
	}

	src := &mockSource{
		name:    "test-source",
		signals: []*PipelineSignal{sig},
	}

	handler := &mockHandler{
		name: "picky-handler",
		matchFn: func(s *PipelineSignal) bool {
			return false // never matches
		},
	}

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{src},
		Handlers: []JobHandler{handler},
	})

	err := p.RunOnce(context.Background())
	require.NoError(t, err)

	// Handler should not have been called.
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Empty(t, handler.executed)

	// Source should not have received reports (no action taken).
	src.mu.Lock()
	defer src.mu.Unlock()
	assert.Empty(t, src.reports)
}

func TestPoller_RunOnce_Good_DryRun(t *testing.T) {
	sig := &PipelineSignal{
		EpicNumber:  1,
		ChildNumber: 3,
		PRNumber:    20,
		RepoOwner:   "host-uk",
		RepoName:    "core-admin",
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
		Mergeable:   "MERGEABLE",
	}

	src := &mockSource{
		name:    "test-source",
		signals: []*PipelineSignal{sig},
	}

	handler := &mockHandler{
		name: "merge-handler",
		matchFn: func(s *PipelineSignal) bool {
			return true
		},
	}

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{src},
		Handlers: []JobHandler{handler},
		DryRun:   true,
	})

	assert.True(t, p.DryRun())

	err := p.RunOnce(context.Background())
	require.NoError(t, err)

	// Handler should NOT have been called in dry-run mode.
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Empty(t, handler.executed)

	// Source should not have received reports.
	src.mu.Lock()
	defer src.mu.Unlock()
	assert.Empty(t, src.reports)
}

func TestPoller_SetDryRun_Good(t *testing.T) {
	p := NewPoller(PollerConfig{})

	assert.False(t, p.DryRun())
	p.SetDryRun(true)
	assert.True(t, p.DryRun())
	p.SetDryRun(false)
	assert.False(t, p.DryRun())
}

func TestPoller_AddSourceAndHandler_Good(t *testing.T) {
	p := NewPoller(PollerConfig{})

	sig := &PipelineSignal{
		EpicNumber:  1,
		ChildNumber: 1,
		PRNumber:    5,
		RepoOwner:   "host-uk",
		RepoName:    "core-php",
		PRState:     "OPEN",
	}

	src := &mockSource{
		name:    "added-source",
		signals: []*PipelineSignal{sig},
	}

	handler := &mockHandler{
		name:    "added-handler",
		matchFn: func(s *PipelineSignal) bool { return true },
	}

	p.AddSource(src)
	p.AddHandler(handler)

	err := p.RunOnce(context.Background())
	require.NoError(t, err)

	handler.mu.Lock()
	defer handler.mu.Unlock()
	require.Len(t, handler.executed, 1)
	assert.Equal(t, 5, handler.executed[0].PRNumber)
}

func TestPoller_Run_Good(t *testing.T) {
	src := &mockSource{
		name:    "tick-source",
		signals: nil,
	}

	p := NewPoller(PollerConfig{
		Sources:      []JobSource{src},
		PollInterval: 50 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	err := p.Run(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Should have completed at least 2 cycles (one immediate + at least one tick).
	assert.GreaterOrEqual(t, p.Cycle(), 2)
}

func TestPoller_DefaultInterval_Good(t *testing.T) {
	p := NewPoller(PollerConfig{})
	assert.Equal(t, 60*time.Second, p.interval)
}
