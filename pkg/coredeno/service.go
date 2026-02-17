package coredeno

import (
	"context"
	"fmt"

	core "forge.lthn.ai/core/go/pkg/framework/core"
	"forge.lthn.ai/core/go/pkg/io"
	"forge.lthn.ai/core/go/pkg/manifest"
	"forge.lthn.ai/core/go/pkg/store"
)

// Service wraps the CoreDeno sidecar as a framework service.
// Implements Startable and Stoppable for lifecycle management.
//
// Registration:
//
//	core.New(core.WithService(coredeno.NewServiceFactory(opts)))
type Service struct {
	*core.ServiceRuntime[Options]
	sidecar    *Sidecar
	grpcServer *Server
	store      *store.Store
	grpcCancel context.CancelFunc
	grpcDone   chan error
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

// OnStartup boots the CoreDeno subsystem. Called by the framework on app startup.
//
// Sequence: medium → store → server → manifest → gRPC listener → sidecar.
func (s *Service) OnStartup(ctx context.Context) error {
	opts := s.Opts()

	// 1. Create sandboxed Medium (or mock if no AppRoot)
	var medium io.Medium
	if opts.AppRoot != "" {
		var err error
		medium, err = io.NewSandboxed(opts.AppRoot)
		if err != nil {
			return fmt.Errorf("coredeno: medium: %w", err)
		}
	} else {
		medium = io.NewMockMedium()
	}

	// 2. Create Store
	dbPath := opts.StoreDBPath
	if dbPath == "" {
		dbPath = ":memory:"
	}
	var err error
	s.store, err = store.New(dbPath)
	if err != nil {
		return fmt.Errorf("coredeno: store: %w", err)
	}

	// 3. Create gRPC Server
	s.grpcServer = NewServer(medium, s.store)

	// 4. Load manifest if AppRoot set (non-fatal if missing)
	if opts.AppRoot != "" {
		m, loadErr := manifest.Load(medium, ".")
		if loadErr == nil && m != nil {
			if opts.PublicKey != nil {
				if ok, verr := manifest.Verify(m, opts.PublicKey); verr == nil && ok {
					s.grpcServer.RegisterModule(m)
				}
			} else {
				s.grpcServer.RegisterModule(m)
			}
		}
	}

	// 5. Start gRPC listener in background
	grpcCtx, grpcCancel := context.WithCancel(ctx)
	s.grpcCancel = grpcCancel
	s.grpcDone = make(chan error, 1)
	go func() {
		s.grpcDone <- ListenGRPC(grpcCtx, opts.SocketPath, s.grpcServer)
	}()

	// 6. Start sidecar (if args provided)
	if len(opts.SidecarArgs) > 0 {
		if err := s.sidecar.Start(ctx, opts.SidecarArgs...); err != nil {
			return fmt.Errorf("coredeno: sidecar: %w", err)
		}
	}

	return nil
}

// OnShutdown stops the CoreDeno subsystem. Called by the framework on app shutdown.
func (s *Service) OnShutdown(_ context.Context) error {
	// Stop sidecar first
	_ = s.sidecar.Stop()

	// Stop gRPC listener
	if s.grpcCancel != nil {
		s.grpcCancel()
		<-s.grpcDone
	}

	// Close store
	if s.store != nil {
		s.store.Close()
	}

	return nil
}

// Sidecar returns the underlying sidecar for direct access.
func (s *Service) Sidecar() *Sidecar {
	return s.sidecar
}

// GRPCServer returns the gRPC server for direct access.
func (s *Service) GRPCServer() *Server {
	return s.grpcServer
}
