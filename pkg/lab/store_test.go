package lab

import (
	"errors"
	"testing"
	"time"
)

// ── NewStore ────────────────────────────────────────────────────────

func TestNewStore_Good(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
	if s.subs == nil {
		t.Fatal("subs map not initialised")
	}
	if s.errors == nil {
		t.Fatal("errors map not initialised")
	}
}

// ── Subscribe / Unsubscribe ────────────────────────────────────────

func TestSubscribe_Good(t *testing.T) {
	s := NewStore()
	ch := s.Subscribe()
	if ch == nil {
		t.Fatal("Subscribe returned nil channel")
	}

	s.subMu.Lock()
	_, ok := s.subs[ch]
	s.subMu.Unlock()
	if !ok {
		t.Fatal("subscriber not registered")
	}
}

func TestUnsubscribe_Good(t *testing.T) {
	s := NewStore()
	ch := s.Subscribe()
	s.Unsubscribe(ch)

	s.subMu.Lock()
	_, ok := s.subs[ch]
	s.subMu.Unlock()
	if ok {
		t.Fatal("subscriber not removed after Unsubscribe")
	}
}

func TestUnsubscribe_Bad_NeverSubscribed(t *testing.T) {
	s := NewStore()
	ch := make(chan struct{}, 1)
	// Should not panic.
	s.Unsubscribe(ch)
}

// ── Notify ─────────────────────────────────────────────────────────

func TestNotify_Good_SubscriberReceivesSignal(t *testing.T) {
	s := NewStore()
	ch := s.Subscribe()
	defer s.Unsubscribe(ch)

	s.SetMachines([]Machine{{Name: "test"}})

	select {
	case <-ch:
		// good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("subscriber did not receive notification")
	}
}

func TestNotify_Good_NonBlockingWhenFull(t *testing.T) {
	s := NewStore()
	ch := s.Subscribe()
	defer s.Unsubscribe(ch)

	// Fill the buffer.
	ch <- struct{}{}

	// Should not block.
	s.SetMachines([]Machine{{Name: "a"}})
	s.SetMachines([]Machine{{Name: "b"}})
}

func TestNotify_Good_MultipleSubscribers(t *testing.T) {
	s := NewStore()
	ch1 := s.Subscribe()
	ch2 := s.Subscribe()
	defer s.Unsubscribe(ch1)
	defer s.Unsubscribe(ch2)

	s.SetAgents(AgentSummary{Available: true})

	for _, ch := range []chan struct{}{ch1, ch2} {
		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("subscriber missed notification")
		}
	}
}

// ── SetMachines / Overview ─────────────────────────────────────────

func TestSetMachines_Good(t *testing.T) {
	s := NewStore()
	machines := []Machine{{Name: "noc", Host: "77.42.42.205"}, {Name: "de1", Host: "116.202.82.115"}}
	s.SetMachines(machines)

	ov := s.Overview()
	if len(ov.Machines) != 2 {
		t.Fatalf("expected 2 machines, got %d", len(ov.Machines))
	}
	if ov.Machines[0].Name != "noc" {
		t.Fatalf("expected noc, got %s", ov.Machines[0].Name)
	}
}

func TestOverview_Good_ContainersMergedIntoFirstMachine(t *testing.T) {
	s := NewStore()
	s.SetMachines([]Machine{{Name: "primary"}, {Name: "secondary"}})
	s.SetContainers([]Container{{Name: "forgejo", Status: "running"}})

	ov := s.Overview()
	if len(ov.Machines[0].Containers) != 1 {
		t.Fatal("containers not merged into first machine")
	}
	if ov.Machines[0].Containers[0].Name != "forgejo" {
		t.Fatalf("unexpected container name: %s", ov.Machines[0].Containers[0].Name)
	}
	if len(ov.Machines[1].Containers) != 0 {
		t.Fatal("containers leaked into second machine")
	}
}

func TestOverview_Good_EmptyMachinesNoContainerPanic(t *testing.T) {
	s := NewStore()
	s.SetContainers([]Container{{Name: "c1"}})

	// No machines set — should not panic.
	ov := s.Overview()
	if len(ov.Machines) != 0 {
		t.Fatal("expected zero machines")
	}
}

func TestOverview_Good_ErrorsCopied(t *testing.T) {
	s := NewStore()
	s.SetError("prometheus", errors.New("connection refused"))

	ov := s.Overview()
	if ov.Errors["prometheus"] != "connection refused" {
		t.Fatal("error not in overview")
	}

	// Mutating the copy should not affect the store.
	ov.Errors["prometheus"] = "hacked"
	ov2 := s.Overview()
	if ov2.Errors["prometheus"] != "connection refused" {
		t.Fatal("overview errors map is not a copy")
	}
}

// ── SetAgents / GetAgents ──────────────────────────────────────────

func TestAgents_Good(t *testing.T) {
	s := NewStore()
	s.SetAgents(AgentSummary{Available: true, RegisteredTotal: 3, QueuePending: 1})

	got := s.GetAgents()
	if !got.Available {
		t.Fatal("expected Available=true")
	}
	if got.RegisteredTotal != 3 {
		t.Fatalf("expected 3, got %d", got.RegisteredTotal)
	}
}

// ── SetTraining / GetTraining ──────────────────────────────────────

func TestTraining_Good(t *testing.T) {
	s := NewStore()
	s.SetTraining(TrainingSummary{GoldGenerated: 404, GoldTarget: 15000, GoldPercent: 2.69})

	got := s.GetTraining()
	if got.GoldGenerated != 404 {
		t.Fatalf("expected 404, got %d", got.GoldGenerated)
	}
}

// ── SetModels / GetModels ──────────────────────────────────────────

func TestModels_Good(t *testing.T) {
	s := NewStore()
	s.SetModels([]HFModel{{ModelID: "lthn/lem-gemma3-1b", Downloads: 42}})

	got := s.GetModels()
	if len(got) != 1 {
		t.Fatal("expected 1 model")
	}
	if got[0].Downloads != 42 {
		t.Fatalf("expected 42 downloads, got %d", got[0].Downloads)
	}
}

// ── SetCommits ─────────────────────────────────────────────────────

func TestCommits_Good(t *testing.T) {
	s := NewStore()
	s.SetCommits([]Commit{{SHA: "abc123", Message: "feat: test coverage", Author: "virgil"}})

	ov := s.Overview()
	if len(ov.Commits) != 1 {
		t.Fatal("expected 1 commit")
	}
	if ov.Commits[0].Author != "virgil" {
		t.Fatalf("expected virgil, got %s", ov.Commits[0].Author)
	}
}

// ── SetContainers / GetContainers ──────────────────────────────────

func TestContainers_Good(t *testing.T) {
	s := NewStore()
	s.SetContainers([]Container{{Name: "traefik", Status: "running"}, {Name: "forgejo", Status: "running"}})

	got := s.GetContainers()
	if len(got) != 2 {
		t.Fatal("expected 2 containers")
	}
}

// ── SetError / GetErrors ───────────────────────────────────────────

func TestSetError_Good_SetAndClear(t *testing.T) {
	s := NewStore()
	s.SetError("hf", errors.New("rate limited"))

	errs := s.GetErrors()
	if errs["hf"] != "rate limited" {
		t.Fatal("error not stored")
	}

	// Clear by passing nil.
	s.SetError("hf", nil)
	errs = s.GetErrors()
	if _, ok := errs["hf"]; ok {
		t.Fatal("error not cleared")
	}
}

func TestGetErrors_Good_ReturnsCopy(t *testing.T) {
	s := NewStore()
	s.SetError("forge", errors.New("timeout"))

	errs := s.GetErrors()
	errs["forge"] = "tampered"

	fresh := s.GetErrors()
	if fresh["forge"] != "timeout" {
		t.Fatal("GetErrors did not return a copy")
	}
}

// ── SetServices / GetServices ──────────────────────────────────────

func TestServices_Good(t *testing.T) {
	s := NewStore()
	s.SetServices([]Service{{Name: "Forgejo", URL: "https://forge.lthn.ai", Status: "ok"}})

	got := s.GetServices()
	if len(got) != 1 {
		t.Fatal("expected 1 service")
	}
	if got[0].Name != "Forgejo" {
		t.Fatalf("expected Forgejo, got %s", got[0].Name)
	}
}

// ── SetBenchmarks / GetBenchmarks ──────────────────────────────────

func TestBenchmarks_Good(t *testing.T) {
	s := NewStore()
	s.SetBenchmarks(BenchmarkData{
		Runs: []BenchmarkRun{{RunID: "run-1", Model: "gemma3-4b", Type: "training"}},
	})

	got := s.GetBenchmarks()
	if len(got.Runs) != 1 {
		t.Fatal("expected 1 benchmark run")
	}
}

// ── SetGoldenSet / GetGoldenSet ────────────────────────────────────

func TestGoldenSet_Good(t *testing.T) {
	s := NewStore()
	s.SetGoldenSet(GoldenSetSummary{Available: true, TotalExamples: 15000, TargetTotal: 15000, CompletionPct: 100})

	got := s.GetGoldenSet()
	if !got.Available {
		t.Fatal("expected Available=true")
	}
	if got.TotalExamples != 15000 {
		t.Fatalf("expected 15000, got %d", got.TotalExamples)
	}
}

// ── SetTrainingRuns / GetTrainingRuns ───────────────────────────────

func TestTrainingRuns_Good(t *testing.T) {
	s := NewStore()
	s.SetTrainingRuns([]TrainingRunStatus{
		{Model: "gemma3-4b", RunID: "r1", Status: "training", Iteration: 100, TotalIters: 300},
	})

	got := s.GetTrainingRuns()
	if len(got) != 1 {
		t.Fatal("expected 1 training run")
	}
	if got[0].Iteration != 100 {
		t.Fatalf("expected iter 100, got %d", got[0].Iteration)
	}
}

// ── SetDataset / GetDataset ────────────────────────────────────────

func TestDataset_Good(t *testing.T) {
	s := NewStore()
	s.SetDataset(DatasetSummary{
		Available: true,
		Tables:    []DatasetTable{{Name: "golden_set", Rows: 15000}},
	})

	got := s.GetDataset()
	if !got.Available {
		t.Fatal("expected Available=true")
	}
	if len(got.Tables) != 1 {
		t.Fatal("expected 1 table")
	}
}

// ── Concurrent access (race detector) ──────────────────────────────

func TestConcurrentAccess_Good(t *testing.T) {
	s := NewStore()
	done := make(chan struct{})

	// Writer goroutine.
	go func() {
		for i := range 100 {
			s.SetMachines([]Machine{{Name: "noc"}})
			s.SetAgents(AgentSummary{Available: true})
			s.SetTraining(TrainingSummary{GoldGenerated: i})
			s.SetModels([]HFModel{{ModelID: "m1"}})
			s.SetCommits([]Commit{{SHA: "abc"}})
			s.SetContainers([]Container{{Name: "c1"}})
			s.SetError("test", errors.New("e"))
			s.SetServices([]Service{{Name: "s1"}})
			s.SetBenchmarks(BenchmarkData{})
			s.SetGoldenSet(GoldenSetSummary{})
			s.SetTrainingRuns([]TrainingRunStatus{})
			s.SetDataset(DatasetSummary{})
		}
		close(done)
	}()

	// Reader goroutine.
	for range 100 {
		_ = s.Overview()
		_ = s.GetModels()
		_ = s.GetTraining()
		_ = s.GetAgents()
		_ = s.GetContainers()
		_ = s.GetServices()
		_ = s.GetBenchmarks()
		_ = s.GetGoldenSet()
		_ = s.GetTrainingRuns()
		_ = s.GetDataset()
		_ = s.GetErrors()
	}

	<-done
}
