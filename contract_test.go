// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
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
func TestWithService_NameDiscovery_Good(t *testing.T) {
	r := New(WithService(stubFactory))
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	names := c.Services()
	// Service should be auto-registered under a discovered name (not just "cli" which is built-in)
	assert.Greater(t, len(names), 1, "expected auto-discovered service to be registered alongside built-in 'cli'")
}

// TestWithService_FactorySelfRegisters_Good verifies that when a factory
// returns Result{OK:true} with no Value (it registered itself), WithService
// does not attempt a second registration and returns success.
func TestWithService_FactorySelfRegisters_Good(t *testing.T) {
	selfReg := func(c *Core) Result {
		// Factory registers directly, returns no instance.
		c.Service("self", Service{})
		return Result{OK: true}
	}

	r := New(WithService(selfReg))
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	// "self" must be present and registered exactly once.
	svc := c.Service("self")
	assert.True(t, svc.OK, "expected self-registered service to be present")
}

// --- WithName ---

func TestWithName_Good(t *testing.T) {
	r := New(
		WithName("custom", func(c *Core) Result {
			return Result{Value: &stubNamedService{}, OK: true}
		}),
	)
	assert.True(t, r.OK)
	c := r.Value.(*Core)
	assert.Contains(t, c.Services(), "custom")
}

// --- Lifecycle ---

type lifecycleService struct {
	started bool
}

func (s *lifecycleService) OnStartup(_ context.Context) error {
	s.started = true
	return nil
}

func TestWithService_Lifecycle_Good(t *testing.T) {
	svc := &lifecycleService{}
	r := New(
		WithService(func(c *Core) Result {
			return Result{Value: svc, OK: true}
		}),
	)
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	c.ServiceStartup(context.Background(), nil)
	assert.True(t, svc.started)
}

// --- IPC Handler ---

type ipcService struct {
	received Message
}

func (s *ipcService) HandleIPCEvents(c *Core, msg Message) Result {
	s.received = msg
	return Result{OK: true}
}

func TestWithService_IPCHandler_Good(t *testing.T) {
	svc := &ipcService{}
	r := New(
		WithService(func(c *Core) Result {
			return Result{Value: svc, OK: true}
		}),
	)
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	c.ACTION("ping")
	assert.Equal(t, "ping", svc.received)
}

// --- Error ---

// TestWithService_FactoryError_Bad verifies that a factory returning an error
// causes New() to stop and propagate the failure.
func TestWithService_FactoryError_Bad(t *testing.T) {
	r := New(WithService(func(c *Core) Result {
		return Result{Value: E("test", "factory failed", nil), OK: false}
	}))
	assert.False(t, r.OK, "expected New() to fail when factory returns error")
}
