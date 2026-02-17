package coredeno

import "context"

// Service wraps the CoreDeno sidecar for framework lifecycle integration.
// Implements Startable (OnStartup) and Stoppable (OnShutdown) interfaces.
type Service struct {
	sidecar *Sidecar
	opts    Options
}

// NewService creates a CoreDeno service ready for framework registration.
func NewService(opts Options) *Service {
	return &Service{
		sidecar: NewSidecar(opts),
		opts:    opts,
	}
}

// OnStartup starts the Deno sidecar. Called by the framework.
func (s *Service) OnStartup(ctx context.Context) error {
	return nil
}

// OnShutdown stops the Deno sidecar. Called by the framework.
func (s *Service) OnShutdown() error {
	return s.sidecar.Stop()
}

// Sidecar returns the underlying sidecar for direct access.
func (s *Service) Sidecar() *Sidecar {
	return s.sidecar
}
