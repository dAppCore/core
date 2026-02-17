package coredeno

import (
	"context"
	"testing"

	core "forge.lthn.ai/core/go/pkg/framework/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceFactory_Good(t *testing.T) {
	opts := Options{
		DenoPath:   "echo",
		SocketPath: "/tmp/test-service.sock",
	}
	c, err := core.New()
	require.NoError(t, err)

	factory := NewServiceFactory(opts)
	result, err := factory(c)
	require.NoError(t, err)

	svc, ok := result.(*Service)
	require.True(t, ok)
	assert.NotNil(t, svc.sidecar)
	assert.Equal(t, "echo", svc.sidecar.opts.DenoPath)
	assert.NotNil(t, svc.Core(), "ServiceRuntime should provide Core access")
	assert.Equal(t, opts, svc.Opts(), "ServiceRuntime should provide Options access")
}

func TestService_WithService_Good(t *testing.T) {
	opts := Options{DenoPath: "echo"}
	c, err := core.New(core.WithService(NewServiceFactory(opts)))
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestService_Lifecycle_Good(t *testing.T) {
	c, err := core.New()
	require.NoError(t, err)

	factory := NewServiceFactory(Options{DenoPath: "echo"})
	result, _ := factory(c)
	svc := result.(*Service)

	// Verify Startable
	err = svc.OnStartup(context.Background())
	assert.NoError(t, err)

	// Verify Stoppable (not started, should be no-op)
	err = svc.OnShutdown(context.Background())
	assert.NoError(t, err)
}

func TestService_Sidecar_Good(t *testing.T) {
	c, err := core.New()
	require.NoError(t, err)

	factory := NewServiceFactory(Options{DenoPath: "echo"})
	result, _ := factory(c)
	svc := result.(*Service)

	assert.NotNil(t, svc.Sidecar())
}
