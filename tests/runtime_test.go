package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- ServiceRuntime ---

type testOpts struct {
	URL     string
	Timeout int
}

func TestServiceRuntime_Good(t *testing.T) {
	c := New()
	opts := testOpts{URL: "https://api.lthn.ai", Timeout: 30}
	rt := NewServiceRuntime(c, opts)

	assert.Equal(t, c, rt.Core())
	assert.Equal(t, opts, rt.Options())
	assert.Equal(t, "https://api.lthn.ai", rt.Options().URL)
	assert.NotNil(t, rt.Config())
}

// --- NewWithFactories ---

func TestNewWithFactories_Good(t *testing.T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"svc1": func() Result { return Result{Value: Service{}, OK: true} },
		"svc2": func() Result { return Result{Value: Service{}, OK: true} },
	})
	assert.True(t, r.OK)
	rt := r.Value.(*Runtime)
	assert.NotNil(t, rt.Core)
}

func TestNewWithFactories_NilFactory_Good(t *testing.T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"bad": nil,
	})
	assert.True(t, r.OK) // nil factories skipped
}

func TestNewRuntime_Good(t *testing.T) {
	r := NewRuntime(nil)
	assert.True(t, r.OK)
}

func TestRuntime_ServiceName_Good(t *testing.T) {
	r := NewRuntime(nil)
	rt := r.Value.(*Runtime)
	assert.Equal(t, "Core", rt.ServiceName())
}

// --- Lifecycle via Runtime ---

func TestRuntime_Lifecycle_Good(t *testing.T) {
	started := false
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"test": func() Result {
			return Result{Value: Service{
				OnStart: func() Result { started = true; return Result{OK: true} },
			}, OK: true}
		},
	})
	assert.True(t, r.OK)
	rt := r.Value.(*Runtime)

	result := rt.ServiceStartup(context.Background(), nil)
	assert.True(t, result.OK)
	assert.True(t, started)
}
