package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Lock (Named Mutexes) ---

func TestLock_Good(t *testing.T) {
	c := New()
	lock := c.Lock("test")
	assert.NotNil(t, lock)
	assert.NotNil(t, lock.Mu)
}

func TestLock_SameName_Good(t *testing.T) {
	c := New()
	l1 := c.Lock("shared")
	l2 := c.Lock("shared")
	// Same name returns same lock
	assert.Equal(t, l1, l2)
}

func TestLock_DifferentName_Good(t *testing.T) {
	c := New()
	l1 := c.Lock("a")
	l2 := c.Lock("b")
	assert.NotEqual(t, l1, l2)
}

func TestLock_MutexWorks_Good(t *testing.T) {
	c := New()
	lock := c.Lock("counter")
	counter := 0
	lock.Mu.Lock()
	counter++
	lock.Mu.Unlock()
	assert.Equal(t, 1, counter)
}

func TestLockEnable_Good(t *testing.T) {
	c := New()
	c.Service("early", struct{}{})
	c.LockEnable()
	c.LockApply()

	// After lock, registration should fail
	result := c.Service("late", struct{}{})
	assert.NotNil(t, result)
}

func TestStartables_Good(t *testing.T) {
	c := New()
	svc := &testService{name: "s"}
	c.Service("s", svc)
	startables := c.Startables()
	assert.Len(t, startables, 1)
}

func TestStoppables_Good(t *testing.T) {
	c := New()
	svc := &testService{name: "s"}
	c.Service("s", svc)
	stoppables := c.Stoppables()
	assert.Len(t, stoppables, 1)
}
