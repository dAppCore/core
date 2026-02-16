package jobrunner

import (
	"context"
	"sync"
	"time"

	"forge.lthn.ai/core/go/pkg/log"
)

// PollerConfig configures a Poller.
type PollerConfig struct {
	Sources      []JobSource
	Handlers     []JobHandler
	Journal      *Journal
	PollInterval time.Duration
	DryRun       bool
}

// Poller discovers signals from sources and dispatches them to handlers.
type Poller struct {
	mu       sync.RWMutex
	sources  []JobSource
	handlers []JobHandler
	journal  *Journal
	interval time.Duration
	dryRun   bool
	cycle    int
}

// NewPoller creates a Poller from the given config.
func NewPoller(cfg PollerConfig) *Poller {
	interval := cfg.PollInterval
	if interval <= 0 {
		interval = 60 * time.Second
	}

	return &Poller{
		sources:  cfg.Sources,
		handlers: cfg.Handlers,
		journal:  cfg.Journal,
		interval: interval,
		dryRun:   cfg.DryRun,
	}
}

// Cycle returns the number of completed poll-dispatch cycles.
func (p *Poller) Cycle() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cycle
}

// DryRun returns whether dry-run mode is enabled.
func (p *Poller) DryRun() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dryRun
}

// SetDryRun enables or disables dry-run mode.
func (p *Poller) SetDryRun(v bool) {
	p.mu.Lock()
	p.dryRun = v
	p.mu.Unlock()
}

// AddSource appends a source to the poller.
func (p *Poller) AddSource(s JobSource) {
	p.mu.Lock()
	p.sources = append(p.sources, s)
	p.mu.Unlock()
}

// AddHandler appends a handler to the poller.
func (p *Poller) AddHandler(h JobHandler) {
	p.mu.Lock()
	p.handlers = append(p.handlers, h)
	p.mu.Unlock()
}

// Run starts a blocking poll-dispatch loop. It runs one cycle immediately,
// then repeats on each tick of the configured interval until the context
// is cancelled.
func (p *Poller) Run(ctx context.Context) error {
	if err := p.RunOnce(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.RunOnce(ctx); err != nil {
				return err
			}
		}
	}
}

// RunOnce performs a single poll-dispatch cycle: iterate sources, poll each,
// find the first matching handler for each signal, and execute it.
func (p *Poller) RunOnce(ctx context.Context) error {
	p.mu.Lock()
	p.cycle++
	cycle := p.cycle
	dryRun := p.dryRun
	sources := make([]JobSource, len(p.sources))
	copy(sources, p.sources)
	handlers := make([]JobHandler, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.Unlock()

	log.Info("poller cycle starting", "cycle", cycle, "sources", len(sources), "handlers", len(handlers))

	for _, src := range sources {
		signals, err := src.Poll(ctx)
		if err != nil {
			log.Error("poll failed", "source", src.Name(), "err", err)
			continue
		}

		log.Info("polled source", "source", src.Name(), "signals", len(signals))

		for _, sig := range signals {
			handler := p.findHandler(handlers, sig)
			if handler == nil {
				log.Debug("no matching handler", "epic", sig.EpicNumber, "child", sig.ChildNumber)
				continue
			}

			if dryRun {
				log.Info("dry-run: would execute",
					"handler", handler.Name(),
					"epic", sig.EpicNumber,
					"child", sig.ChildNumber,
					"pr", sig.PRNumber,
				)
				continue
			}

			start := time.Now()
			result, err := handler.Execute(ctx, sig)
			elapsed := time.Since(start)

			if err != nil {
				log.Error("handler execution failed",
					"handler", handler.Name(),
					"epic", sig.EpicNumber,
					"child", sig.ChildNumber,
					"err", err,
				)
				continue
			}

			result.Cycle = cycle
			result.EpicNumber = sig.EpicNumber
			result.ChildNumber = sig.ChildNumber
			result.Duration = elapsed

			if p.journal != nil {
				if jErr := p.journal.Append(sig, result); jErr != nil {
					log.Error("journal append failed", "err", jErr)
				}
			}

			if rErr := src.Report(ctx, result); rErr != nil {
				log.Error("source report failed", "source", src.Name(), "err", rErr)
			}

			log.Info("handler executed",
				"handler", handler.Name(),
				"action", result.Action,
				"success", result.Success,
				"duration", elapsed,
			)
		}
	}

	return nil
}

// findHandler returns the first handler that matches the signal, or nil.
func (p *Poller) findHandler(handlers []JobHandler, sig *PipelineSignal) JobHandler {
	for _, h := range handlers {
		if h.Match(sig) {
			return h
		}
	}
	return nil
}
