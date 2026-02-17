package coredeno

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService_Good(t *testing.T) {
	opts := Options{
		DenoPath:   "echo",
		SocketPath: "/tmp/test-service.sock",
	}
	svc := NewService(opts)
	require.NotNil(t, svc)
	assert.NotNil(t, svc.sidecar)
	assert.Equal(t, "echo", svc.sidecar.opts.DenoPath)
}

func TestService_OnShutdown_Good_NotStarted(t *testing.T) {
	svc := NewService(Options{DenoPath: "echo"})
	err := svc.OnShutdown()
	assert.NoError(t, err)
}

func TestService_Sidecar_Good(t *testing.T) {
	svc := NewService(Options{DenoPath: "echo"})
	assert.NotNil(t, svc.Sidecar())
}
