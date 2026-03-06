package log

import (
	"context"

	"forge.lthn.ai/core/go/pkg/core"
)

// Service wraps Logger for Core framework integration.
type Service struct {
	*core.ServiceRuntime[Options]
	*Logger
}

// NewService creates a log service factory for Core.
func NewService(opts Options) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		logger := New(opts)

		return &Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
			Logger:         logger,
		}, nil
	}
}

// OnStartup registers query and task handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	s.Core().RegisterTask(s.handleTask)
	return nil
}

// QueryLevel returns the current log level.
type QueryLevel struct{}

// TaskSetLevel changes the log level.
type TaskSetLevel struct {
	Level Level
}

func (s *Service) handleQuery(c *core.Core, q core.Query) (any, bool, error) {
	switch q.(type) {
	case QueryLevel:
		return s.Level(), true, nil
	}
	return nil, false, nil
}

func (s *Service) handleTask(c *core.Core, t core.Task) (any, bool, error) {
	switch m := t.(type) {
	case TaskSetLevel:
		s.SetLevel(m.Level)
		return nil, true, nil
	}
	return nil, false, nil
}
