package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Service Registration ---

func TestService_Register_Good(t *testing.T) {
	c := New()
	r := c.Service("auth", Service{})
	assert.True(t, r.OK)
}

func TestService_Register_Duplicate_Bad(t *testing.T) {
	c := New()
	c.Service("auth", Service{})
	r := c.Service("auth", Service{})
	assert.False(t, r.OK)
}

func TestService_Register_Empty_Bad(t *testing.T) {
	c := New()
	r := c.Service("", Service{})
	assert.False(t, r.OK)
}

func TestService_Get_Good(t *testing.T) {
	c := New()
	c.Service("brain", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Service("brain")
	assert.True(t, r.OK)
	assert.NotNil(t, r.Value)
}

func TestService_Get_Bad(t *testing.T) {
	c := New()
	r := c.Service("nonexistent")
	assert.False(t, r.OK)
}

func TestService_Names_Good(t *testing.T) {
	c := New()
	c.Service("a", Service{})
	c.Service("b", Service{})
	names := c.Services()
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
	assert.Contains(t, names, "cli") // auto-registered by CliRegister in New()
}

// --- Service Lifecycle ---

func TestService_Lifecycle_Good(t *testing.T) {
	c := New()
	started := false
	stopped := false
	c.Service("lifecycle", Service{
		OnStart: func() Result { started = true; return Result{OK: true} },
		OnStop:  func() Result { stopped = true; return Result{OK: true} },
	})

	sr := c.Startables()
	assert.True(t, sr.OK)
	startables := sr.Value.([]*Service)
	assert.Len(t, startables, 1)
	startables[0].OnStart()
	assert.True(t, started)

	tr := c.Stoppables()
	assert.True(t, tr.OK)
	stoppables := tr.Value.([]*Service)
	assert.Len(t, stoppables, 1)
	stoppables[0].OnStop()
	assert.True(t, stopped)
}

type autoLifecycleService struct {
	started  bool
	stopped  bool
	messages []Message
}

func (s *autoLifecycleService) OnStartup(_ context.Context) Result {
	s.started = true
	return Result{OK: true}
}

func (s *autoLifecycleService) OnShutdown(_ context.Context) Result {
	s.stopped = true
	return Result{OK: true}
}

func (s *autoLifecycleService) HandleIPCEvents(_ *Core, msg Message) Result {
	s.messages = append(s.messages, msg)
	return Result{OK: true}
}

func TestService_RegisterService_Bad(t *testing.T) {
	t.Run("EmptyName", func(t *testing.T) {
		c := New()
		r := c.RegisterService("", "value")
		assert.False(t, r.OK)

		err, ok := r.Value.(error)
		if assert.True(t, ok) {
			assert.Equal(t, "core.RegisterService", Operation(err))
		}
	})

	t.Run("DuplicateName", func(t *testing.T) {
		c := New()
		assert.True(t, c.RegisterService("svc", "first").OK)

		r := c.RegisterService("svc", "second")
		assert.False(t, r.OK)
	})

	t.Run("LockedRegistry", func(t *testing.T) {
		c := New()
		c.LockEnable()
		c.LockApply()

		r := c.RegisterService("blocked", "value")
		assert.False(t, r.OK)
	})
}

func TestService_RegisterService_Ugly(t *testing.T) {
	t.Run("AutoDiscoversLifecycleAndIPCHandlers", func(t *testing.T) {
		c := New()
		svc := &autoLifecycleService{}

		r := c.RegisterService("auto", svc)
		assert.True(t, r.OK)
		assert.True(t, c.ServiceStartup(context.Background(), nil).OK)
		assert.True(t, c.ACTION("ping").OK)
		assert.True(t, c.ServiceShutdown(context.Background()).OK)
		assert.True(t, svc.started)
		assert.True(t, svc.stopped)
		assert.Contains(t, svc.messages, Message("ping"))
	})

	t.Run("NilInstanceReturnsServiceDTO", func(t *testing.T) {
		c := New()
		assert.True(t, c.RegisterService("nil", nil).OK)

		r := c.Service("nil")
		if assert.True(t, r.OK) {
			svc, ok := r.Value.(*Service)
			if assert.True(t, ok) {
				assert.Equal(t, "nil", svc.Name)
				assert.Nil(t, svc.Instance)
			}
		}
	})
}

func TestService_ServiceFor_Bad(t *testing.T) {
	typed, ok := ServiceFor[string](New(), "missing")
	assert.False(t, ok)
	assert.Equal(t, "", typed)
}

func TestService_ServiceFor_Ugly(t *testing.T) {
	c := New()
	assert.True(t, c.RegisterService("value", "hello").OK)

	typed, ok := ServiceFor[int](c, "value")
	assert.False(t, ok)
	assert.Equal(t, 0, typed)
}

func TestService_MustServiceFor_Bad(t *testing.T) {
	c := New()
	assert.PanicsWithError(t, `core.MustServiceFor: service "missing" not found or wrong type`, func() {
		_ = MustServiceFor[string](c, "missing")
	})
}

func TestService_MustServiceFor_Ugly(t *testing.T) {
	var c *Core
	assert.Panics(t, func() {
		_ = MustServiceFor[string](c, "missing")
	})
}
