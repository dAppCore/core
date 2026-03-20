package core_test

import (
	"context"
	"errors"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- ServiceRuntime ---

type testOpts struct {
	URL     string
	Timeout int
}

type runtimeService struct {
	*ServiceRuntime[testOpts]
}

func TestServiceRuntime_Good(t *testing.T) {
	c := New()
	opts := testOpts{URL: "https://api.lthn.ai", Timeout: 30}
	rt := NewServiceRuntime(c, opts)

	assert.Equal(t, c, rt.Core())
	assert.Equal(t, opts, rt.Opts())
	assert.Equal(t, "https://api.lthn.ai", rt.Opts().URL)
	assert.Equal(t, 30, rt.Opts().Timeout)
	assert.NotNil(t, rt.Config())
}

func TestServiceRuntime_Embedded_Good(t *testing.T) {
	c := New()
	svc := &runtimeService{
		ServiceRuntime: NewServiceRuntime(c, testOpts{URL: "https://lthn.sh"}),
	}
	assert.Equal(t, "https://lthn.sh", svc.Opts().URL)
}

// --- NewWithFactories ---

func TestNewWithFactories_Good(t *testing.T) {
	rt, err := NewWithFactories(nil, map[string]ServiceFactory{
		"svc1": func() (any, error) { return &testService{name: "one"}, nil },
		"svc2": func() (any, error) { return &testService{name: "two"}, nil },
	})
	assert.NoError(t, err)
	assert.NotNil(t, rt)
	assert.NotNil(t, rt.Core)

	svc := rt.Core.Service("svc1")
	assert.NotNil(t, svc)
	ts, ok := svc.(*testService)
	assert.True(t, ok)
	assert.Equal(t, "one", ts.name)
}

func TestNewWithFactories_Bad(t *testing.T) {
	// Nil factory
	_, err := NewWithFactories(nil, map[string]ServiceFactory{
		"bad": nil,
	})
	assert.Error(t, err)

	// Factory returns error
	_, err = NewWithFactories(nil, map[string]ServiceFactory{
		"fail": func() (any, error) { return nil, errors.New("factory failed") },
	})
	assert.Error(t, err)
}

func TestNewRuntime_Good(t *testing.T) {
	rt, err := NewRuntime(nil)
	assert.NoError(t, err)
	assert.NotNil(t, rt)
}

// --- Lifecycle via Runtime ---

func TestRuntime_Lifecycle_Good(t *testing.T) {
	svc := &testService{name: "lifecycle"}
	rt, err := NewWithFactories(nil, map[string]ServiceFactory{
		"test": func() (any, error) { return svc, nil },
	})
	assert.NoError(t, err)

	err = rt.ServiceStartup(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, svc.started)

	err = rt.ServiceShutdown(context.Background())
	assert.NoError(t, err)
	assert.True(t, svc.stopped)
}
