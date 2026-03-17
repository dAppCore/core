package log

import (
	"context"
	"testing"

	"forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService_Good(t *testing.T) {
	opts := Options{Level: LevelInfo}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log")
	require.NotNil(t, svc)

	logSvc, ok := svc.(*Service)
	require.True(t, ok)
	assert.NotNil(t, logSvc.Logger)
	assert.NotNil(t, logSvc.ServiceRuntime)
}

func TestService_OnStartup_Good(t *testing.T) {
	opts := Options{Level: LevelInfo}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log").(*Service)

	err = svc.OnStartup(context.Background())
	assert.NoError(t, err)
}

func TestService_QueryLevel_Good(t *testing.T) {
	opts := Options{Level: LevelDebug}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log").(*Service)
	err = svc.OnStartup(context.Background())
	require.NoError(t, err)

	result, handled, err := c.QUERY(QueryLevel{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, LevelDebug, result)
}

func TestService_QueryLevel_Bad(t *testing.T) {
	opts := Options{Level: LevelInfo}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log").(*Service)
	err = svc.OnStartup(context.Background())
	require.NoError(t, err)

	// Unknown query type should not be handled
	result, handled, err := c.QUERY("unknown")
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}

func TestService_TaskSetLevel_Good(t *testing.T) {
	opts := Options{Level: LevelInfo}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log").(*Service)
	err = svc.OnStartup(context.Background())
	require.NoError(t, err)

	// Change level via task
	_, handled, err := c.PERFORM(TaskSetLevel{Level: LevelError})
	assert.NoError(t, err)
	assert.True(t, handled)

	// Verify level changed via query
	result, handled, err := c.QUERY(QueryLevel{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, LevelError, result)
}

func TestService_TaskSetLevel_Bad(t *testing.T) {
	opts := Options{Level: LevelInfo}
	factory := NewService(opts)

	c, err := core.New(core.WithName("log", func(cc *core.Core) (any, error) {
		return factory(cc)
	}))
	require.NoError(t, err)

	svc := c.Service("log").(*Service)
	err = svc.OnStartup(context.Background())
	require.NoError(t, err)

	// Unknown task type should not be handled
	_, handled, err := c.PERFORM("unknown")
	assert.NoError(t, err)
	assert.False(t, handled)
}
