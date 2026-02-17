package coredeno

import (
	"context"

	core "forge.lthn.ai/core/go/pkg/framework/core"
)

// Service wraps the CoreDeno sidecar as a framework service.
// Implements Startable and Stoppable for lifecycle management.
//
// Registration:
//
//	core.New(core.WithService(coredeno.NewServiceFactory(opts)))
type Service struct {
	*core.ServiceRuntime[Options]
	sidecar *Sidecar
}

// NewServiceFactory returns a factory function for framework registration via WithService.
func NewServiceFactory(opts Options) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		return &Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
			sidecar:        NewSidecar(opts),
		}, nil
	}
}

// OnStartup starts the Deno sidecar. Called by the framework on app startup.
func (s *Service) OnStartup(ctx context.Context) error {
	return nil
}

// OnShutdown stops the Deno sidecar. Called by the framework on app shutdown.
func (s *Service) OnShutdown(_ context.Context) error {
	return s.sidecar.Stop()
}

// Sidecar returns the underlying sidecar for direct access.
func (s *Service) Sidecar() *Sidecar {
	return s.sidecar
}
