package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Collector interface {
	Name() string
	Collect(ctx context.Context) error
}

type Registry struct {
	mu      sync.Mutex
	entries []entry
	logger  *slog.Logger
}

type entry struct {
	c        Collector
	interval time.Duration
	cancel   context.CancelFunc
}

func NewRegistry(logger *slog.Logger) *Registry {
	return &Registry{logger: logger}
}

func (r *Registry) Register(c Collector, interval time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry{c: c, interval: interval})
}

func (r *Registry) Start(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.entries {
		e := &r.entries[i]
		cctx, cancel := context.WithCancel(ctx)
		e.cancel = cancel
		go r.run(cctx, e.c, e.interval)
	}
}

func (r *Registry) run(ctx context.Context, c Collector, interval time.Duration) {
	r.logger.Info("collector started", "name", c.Name(), "interval", interval)

	// Run immediately on start.
	if err := c.Collect(ctx); err != nil {
		r.logger.Warn("collector error", "name", c.Name(), "err", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("collector stopped", "name", c.Name())
			return
		case <-ticker.C:
			if err := c.Collect(ctx); err != nil {
				r.logger.Warn("collector error", "name", c.Name(), "err", err)
			}
		}
	}
}

func (r *Registry) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.entries {
		if e.cancel != nil {
			e.cancel()
		}
	}
}
