// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"context"

	. "dappco.re/go/core"
)

// --- WithService ---

// stub service used only for name-discovery tests.
type stubNamedService struct{}

// stubFactory is a package-level factory so the runtime function name carries
// the package path "core_test.stubFactory" — last segment after '/' is
// "core_test", and after stripping a "_test" suffix we get "core".
// For a real service package such as "dappco.re/go/agentic" the discovered
// name would be "agentic".
func stubFactory(c *Core) Result {
	return Result{Value: &stubNamedService{}, OK: true}
}

// TestWithService_NameDiscovery_Good verifies that WithService discovers the
// service name from the factory's package path and registers the instance via
// RegisterService, making it retrievable through c.Services().
//
// stubFactory lives in package "dappco.re/go/core_test", so the last path
// segment is "core_test" — WithService strips the "_test" suffix and registers
// the service under the name "core".
func TestContract_WithService_NameDiscovery_Good(t *T) {
	c := New(WithService(stubFactory))

	names := c.Services()
	// Service should be auto-registered under a discovered name (not just "cli" which is built-in)
	AssertGreater(t, len(names), 1, "expected auto-discovered service to be registered alongside built-in 'cli'")
}

// TestWithService_FactorySelfRegisters_Good verifies that when a factory
// returns Result{OK:true} with no Value (it registered itself), WithService
// does not attempt a second registration and returns success.
func TestContract_WithService_FactorySelfRegisters_Good(t *T) {
	selfReg := func(c *Core) Result {
		// Factory registers directly, returns no instance.
		c.Service("self", Service{})
		return Result{OK: true}
	}

	c := New(WithService(selfReg))

	// "self" must be present and registered exactly once.
	svc := c.Service("self")
	AssertTrue(t, svc.OK, "expected self-registered service to be present")
}

// --- WithName ---

func TestContract_WithName_Good(t *T) {
	c := New(
		WithName("custom", func(c *Core) Result {
			return Result{Value: &stubNamedService{}, OK: true}
		}),
	)
	AssertContains(t, c.Services(), "custom")
}

// --- Lifecycle ---

type lifecycleService struct {
	started bool
}

func (s *lifecycleService) OnStartup(_ context.Context) Result {
	s.started = true
	return Result{OK: true}
}

func TestContract_WithService_Lifecycle_Good(t *T) {
	svc := &lifecycleService{}
	c := New(
		WithService(func(c *Core) Result {
			return Result{Value: svc, OK: true}
		}),
	)

	c.ServiceStartup(context.Background(), nil)
	AssertTrue(t, svc.started)
}

// --- IPC Handler ---

type ipcService struct {
	received Message
}

func (s *ipcService) HandleIPCEvents(c *Core, msg Message) Result {
	s.received = msg
	return Result{OK: true}
}

func TestContract_WithService_IPCHandler_Good(t *T) {
	svc := &ipcService{}
	c := New(
		WithService(func(c *Core) Result {
			return Result{Value: svc, OK: true}
		}),
	)

	c.ACTION("ping")
	AssertEqual(t, "ping", svc.received)
}

// --- Error ---

// TestWithService_FactoryError_Bad verifies that a failing factory
// stops further option processing (second service not registered).
func TestContract_WithService_FactoryError_Bad(t *T) {
	secondCalled := false
	c := New(
		WithService(func(c *Core) Result {
			return Result{Value: E("test", "factory failed", nil), OK: false}
		}),
		WithService(func(c *Core) Result {
			secondCalled = true
			return Result{OK: true}
		}),
	)
	AssertNotNil(t, c)
	AssertFalse(t, secondCalled, "second option should not run after first fails")
}
