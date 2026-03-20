package core_test

import (
	"context"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

type testService struct {
	name    string
	started bool
	stopped bool
}

func (s *testService) OnStartup(_ context.Context) error  { s.started = true; return nil }
func (s *testService) OnShutdown(_ context.Context) error { s.stopped = true; return nil }

// --- Service Registration ---

func TestService_Register_Good(t *testing.T) {
	c := New()
	svc := &testService{name: "auth"}
	result := c.Service("auth", svc)
	assert.Nil(t, result) // nil = success

	got := c.Service("auth")
	assert.Equal(t, svc, got)
}

func TestService_Register_Bad(t *testing.T) {
	c := New()
	svc := &testService{name: "auth"}

	// Register once — ok
	c.Service("auth", svc)

	// Register duplicate — returns error
	result := c.Service("auth", svc)
	assert.NotNil(t, result)

	// Empty name — returns error
	result = c.Service("", svc)
	assert.NotNil(t, result)
}

func TestService_Get_Good(t *testing.T) {
	c := New()
	c.Service("brain", &testService{name: "brain"})

	svc := c.Service("brain")
	assert.NotNil(t, svc)

	ts, ok := svc.(*testService)
	assert.True(t, ok)
	assert.Equal(t, "brain", ts.name)
}

func TestService_Get_Bad(t *testing.T) {
	c := New()
	svc := c.Service("nonexistent")
	assert.Nil(t, svc)
}

func TestService_Registry_Good(t *testing.T) {
	c := New()
	// Zero args returns *Service
	registry := c.Service()
	assert.NotNil(t, registry)
}

// --- Service Lifecycle ---

func TestService_Lifecycle_Good(t *testing.T) {
	c := New()
	svc := &testService{name: "lifecycle"}
	c.Service("lifecycle", svc)

	// Startup
	err := c.ServiceStartup(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, svc.started)

	// Shutdown
	err = c.ServiceShutdown(context.Background())
	assert.NoError(t, err)
	assert.True(t, svc.stopped)
}

func TestService_Lock_Good(t *testing.T) {
	c := New()
	c.Service("early", &testService{name: "early"})

	// Lock service registration
	c.LockEnable()
	c.LockApply()

	// Attempt to register after lock
	result := c.Service("late", &testService{name: "late"})
	assert.NotNil(t, result) // error — locked
}
