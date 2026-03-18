package core_test

import (
	. "forge.lthn.ai/core/go/pkg/core"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceManager_RegisterService_Good(t *testing.T) {
	c, _ := New()

	err := c.RegisterService("svc1", &MockService{Name: "one"})
	assert.NoError(t, err)

	got := c.Service("svc1")
	assert.NotNil(t, got)
	assert.Equal(t, "one", got.(*MockService).GetName())
}

func TestServiceManager_RegisterService_Bad(t *testing.T) {
	c, _ := New()

	// Empty name
	err := c.RegisterService("", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Duplicate
	err = c.RegisterService("dup", &MockService{})
	assert.NoError(t, err)
	err = c.RegisterService("dup", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Locked
	c2, _ := New(WithServiceLock())
	err = c2.RegisterService("late", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serviceLock")
}

func TestServiceManager_ServiceNotFound_Good(t *testing.T) {
	c, _ := New()
	assert.Nil(t, c.Service("nonexistent"))
}

func TestServiceManager_Startables_Good(t *testing.T) {
	s1 := &MockStartable{}
	s2 := &MockStartable{}

	c, _ := New(
		WithName("s1", func(_ *Core) (any, error) { return s1, nil }),
		WithName("s2", func(_ *Core) (any, error) { return s2, nil }),
	)

	// Startup should call both
	err := c.ServiceStartup(context.Background(), nil)
	assert.NoError(t, err)
}

func TestServiceManager_Stoppables_Good(t *testing.T) {
	s1 := &MockStoppable{}
	s2 := &MockStoppable{}

	c, _ := New(
		WithName("s1", func(_ *Core) (any, error) { return s1, nil }),
		WithName("s2", func(_ *Core) (any, error) { return s2, nil }),
	)

	// Shutdown should call both
	err := c.ServiceShutdown(context.Background())
	assert.NoError(t, err)
}

func TestServiceManager_Lock_Good(t *testing.T) {
	c, _ := New(
		WithName("early", func(_ *Core) (any, error) { return &MockService{}, nil }),
		WithServiceLock(),
	)

	// Register after lock — should fail
	err := c.RegisterService("late", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serviceLock")

	// Early service is still accessible
	assert.NotNil(t, c.Service("early"))
}

func TestServiceManager_LockNotAppliedWithoutEnable_Good(t *testing.T) {
	// No WithServiceLock — should allow registration after New()
	c, _ := New()
	err := c.RegisterService("svc", &MockService{})
	assert.NoError(t, err)
}

type mockFullLifecycle struct{}

func (m *mockFullLifecycle) OnStartup(_ context.Context) error  { return nil }
func (m *mockFullLifecycle) OnShutdown(_ context.Context) error { return nil }

func TestServiceManager_LifecycleBoth_Good(t *testing.T) {
	svc := &mockFullLifecycle{}

	c, _ := New(
		WithName("both", func(_ *Core) (any, error) { return svc, nil }),
	)

	// Should participate in both startup and shutdown
	err := c.ServiceStartup(context.Background(), nil)
	assert.NoError(t, err)
	err = c.ServiceShutdown(context.Background())
	assert.NoError(t, err)
}
