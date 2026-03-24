package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Service Registration ---

func TestService_Register_Good(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Service("auth", Service{})
	assert.True(t, r.OK)
}

func TestService_Register_Duplicate_Bad(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("auth", Service{})
	r := c.Service("auth", Service{})
	assert.False(t, r.OK)
}

func TestService_Register_Empty_Bad(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Service("", Service{})
	assert.False(t, r.OK)
}

func TestService_Get_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("brain", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Service("brain")
	assert.True(t, r.OK)
	assert.NotNil(t, r.Value)
}

func TestService_Get_Bad(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Service("nonexistent")
	assert.False(t, r.OK)
}

func TestService_Names_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("a", Service{})
	c.Service("b", Service{})
	names := c.Services()
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
	assert.Contains(t, names, "cli") // auto-registered by CliRegister in New()
}

// --- Service Lifecycle ---

func TestService_Lifecycle_Good(t *testing.T) {
	c := New().Value.(*Core)
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
