package core

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCore_PerformAsync_Good(t *testing.T) {
	c, _ := New()
	
	var completed atomic.Bool
	var resultReceived any
	
	c.RegisterAction(func(c *Core, msg Message) error {
		if tc, ok := msg.(ActionTaskCompleted); ok {
			resultReceived = tc.Result
			completed.Store(true)
		}
		return nil
	})
	
	c.RegisterTask(func(c *Core, task Task) (any, bool, error) {
		return "async-result", true, nil
	})
	
	taskID := c.PerformAsync(TestTask{})
	assert.NotEmpty(t, taskID)
	
	// Wait for completion
	assert.Eventually(t, func() bool {
		return completed.Load()
	}, 1*time.Second, 10*time.Millisecond)
	
	assert.Equal(t, "async-result", resultReceived)
}

func TestCore_PerformAsync_Shutdown(t *testing.T) {
	c, _ := New()
	_ = c.ServiceShutdown(context.Background())
	
	taskID := c.PerformAsync(TestTask{})
	assert.Empty(t, taskID, "PerformAsync should return empty string if already shut down")
}

func TestCore_Progress_Good(t *testing.T) {
	c, _ := New()
	
	var progressReceived float64
	var messageReceived string
	
	c.RegisterAction(func(c *Core, msg Message) error {
		if tp, ok := msg.(ActionTaskProgress); ok {
			progressReceived = tp.Progress
			messageReceived = tp.Message
		}
		return nil
	})
	
	c.Progress("task-1", 0.5, "halfway", TestTask{})
	
	assert.Equal(t, 0.5, progressReceived)
	assert.Equal(t, "halfway", messageReceived)
}

func TestCore_WithService_UnnamedType(t *testing.T) {
	// Primitive types have no package path
	factory := func(c *Core) (any, error) {
		s := "primitive"
		return &s, nil
	}
	
	_, err := New(WithService(factory))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service name could not be discovered")
}

func TestRuntime_ServiceStartup_ErrorPropagation(t *testing.T) {
	rt, _ := NewRuntime(nil)
	
	// Register a service that fails startup
	errSvc := &MockStartable{err: errors.New("startup failed")}
	_ = rt.Core.RegisterService("error-svc", errSvc)
	
	err := rt.ServiceStartup(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startup failed")
}

func TestCore_ServiceStartup_ContextCancellation(t *testing.T) {
	c, _ := New()
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	s1 := &MockStartable{}
	_ = c.RegisterService("s1", s1)
	
	err := c.ServiceStartup(ctx, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, s1.started, "Service should not have started if context was cancelled before loop")
}

func TestCore_ServiceShutdown_ContextCancellation(t *testing.T) {
	c, _ := New()
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	s1 := &MockStoppable{}
	_ = c.RegisterService("s1", s1)
	
	err := c.ServiceShutdown(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, s1.stopped, "Service should not have stopped if context was cancelled before loop")
}

type TaskWithIDImpl struct {
	id string
}
func (t *TaskWithIDImpl) SetTaskID(id string) { t.id = id }
func (t *TaskWithIDImpl) GetTaskID() string { return t.id }

func TestCore_PerformAsync_InjectsID(t *testing.T) {
	c, _ := New()
	c.RegisterTask(func(c *Core, t Task) (any, bool, error) { return nil, true, nil })
	
	task := &TaskWithIDImpl{}
	taskID := c.PerformAsync(task)
	
	assert.Equal(t, taskID, task.GetTaskID())
}
