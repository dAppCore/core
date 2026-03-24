package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestLock_Good(t *testing.T) {
	c := New().Value.(*Core)
	lock := c.Lock("test")
	assert.NotNil(t, lock)
	assert.NotNil(t, lock.Mutex)
}

func TestLock_SameName_Good(t *testing.T) {
	c := New().Value.(*Core)
	l1 := c.Lock("shared")
	l2 := c.Lock("shared")
	assert.Equal(t, l1, l2)
}

func TestLock_DifferentName_Good(t *testing.T) {
	c := New().Value.(*Core)
	l1 := c.Lock("a")
	l2 := c.Lock("b")
	assert.NotEqual(t, l1, l2)
}

func TestLockEnable_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("early", Service{})
	c.LockEnable()
	c.LockApply()

	r := c.Service("late", Service{})
	assert.False(t, r.OK)
}

func TestStartables_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("s", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Startables()
	assert.True(t, r.OK)
	assert.Len(t, r.Value.([]*Service), 1)
}

func TestStoppables_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Service("s", Service{OnStop: func() Result { return Result{OK: true} }})
	r := c.Stoppables()
	assert.True(t, r.OK)
	assert.Len(t, r.Value.([]*Service), 1)
}
