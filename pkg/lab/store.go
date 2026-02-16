package lab

import (
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	// SSE subscriber channels -- notified on any data change.
	subMu sync.Mutex
	subs  map[chan struct{}]struct{}

	machines   []Machine
	machinesAt time.Time

	agents   AgentSummary
	agentsAt time.Time

	training   TrainingSummary
	trainingAt time.Time

	models   []HFModel
	modelsAt time.Time

	commits   []Commit
	commitsAt time.Time

	containers   []Container
	containersAt time.Time

	services   []Service
	servicesAt time.Time

	benchmarks   BenchmarkData
	benchmarksAt time.Time

	goldenSet   GoldenSetSummary
	goldenSetAt time.Time

	trainingRuns   []TrainingRunStatus
	trainingRunsAt time.Time

	dataset   DatasetSummary
	datasetAt time.Time

	errors map[string]string
}

func NewStore() *Store {
	return &Store{
		subs:   make(map[chan struct{}]struct{}),
		errors: make(map[string]string),
	}
}

// Subscribe returns a channel that receives a signal on every data update.
// Call Unsubscribe when done to avoid leaks.
func (s *Store) Subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	s.subMu.Lock()
	s.subs[ch] = struct{}{}
	s.subMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (s *Store) Unsubscribe(ch chan struct{}) {
	s.subMu.Lock()
	delete(s.subs, ch)
	s.subMu.Unlock()
}

// notify sends a non-blocking signal to all subscribers.
func (s *Store) notify() {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *Store) SetMachines(m []Machine) {
	s.mu.Lock()
	s.machines = m
	s.machinesAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetAgents(a AgentSummary) {
	s.mu.Lock()
	s.agents = a
	s.agentsAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetTraining(t TrainingSummary) {
	s.mu.Lock()
	s.training = t
	s.trainingAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetModels(m []HFModel) {
	s.mu.Lock()
	s.models = m
	s.modelsAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetCommits(c []Commit) {
	s.mu.Lock()
	s.commits = c
	s.commitsAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetContainers(c []Container) {
	s.mu.Lock()
	s.containers = c
	s.containersAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) SetError(collector string, err error) {
	s.mu.Lock()
	if err != nil {
		s.errors[collector] = err.Error()
	} else {
		delete(s.errors, collector)
	}
	s.mu.Unlock()
	s.notify()
}

func (s *Store) Overview() Overview {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errCopy := make(map[string]string, len(s.errors))
	for k, v := range s.errors {
		errCopy[k] = v
	}

	// Merge containers into the first machine (snider-linux / local Docker host).
	machines := make([]Machine, len(s.machines))
	copy(machines, s.machines)
	if len(machines) > 0 {
		machines[0].Containers = s.containers
	}

	return Overview{
		UpdatedAt: time.Now(),
		Machines:  machines,
		Agents:    s.agents,
		Training:  s.training,
		Models:    s.models,
		Commits:   s.commits,
		Errors:    errCopy,
	}
}

func (s *Store) GetModels() []HFModel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.models
}

func (s *Store) GetTraining() TrainingSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.training
}

func (s *Store) GetAgents() AgentSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agents
}

func (s *Store) GetContainers() []Container {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.containers
}

func (s *Store) SetServices(svc []Service) {
	s.mu.Lock()
	s.services = svc
	s.servicesAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) GetServices() []Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.services
}

func (s *Store) SetBenchmarks(b BenchmarkData) {
	s.mu.Lock()
	s.benchmarks = b
	s.benchmarksAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) GetBenchmarks() BenchmarkData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.benchmarks
}

func (s *Store) SetGoldenSet(g GoldenSetSummary) {
	s.mu.Lock()
	s.goldenSet = g
	s.goldenSetAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) GetGoldenSet() GoldenSetSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.goldenSet
}

func (s *Store) SetTrainingRuns(runs []TrainingRunStatus) {
	s.mu.Lock()
	s.trainingRuns = runs
	s.trainingRunsAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) GetTrainingRuns() []TrainingRunStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trainingRuns
}

func (s *Store) SetDataset(d DatasetSummary) {
	s.mu.Lock()
	s.dataset = d
	s.datasetAt = time.Now()
	s.mu.Unlock()
	s.notify()
}

func (s *Store) GetDataset() DatasetSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dataset
}

func (s *Store) GetErrors() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	errCopy := make(map[string]string, len(s.errors))
	for k, v := range s.errors {
		errCopy[k] = v
	}
	return errCopy
}
