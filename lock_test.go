package core_test

import (
	"os"
	"os/exec"
	"testing"

	. "dappco.re/go/core"
)

func TestLock_Good(t *testing.T) {
	c := New()
	lock := c.Lock("test")
	AssertNotNil(t, lock)
	AssertNotNil(t, lock.Mutex)
}

func TestLock_SameName_Good(t *testing.T) {
	c := New()
	l1 := c.Lock("shared")
	l2 := c.Lock("shared")
	AssertEqual(t, l1, l2)
}

func TestLock_DifferentName_Good(t *testing.T) {
	c := New()
	l1 := c.Lock("a")
	l2 := c.Lock("b")
	AssertNotEqual(t, l1, l2)
}

func TestLock_LockEnable_Good(t *testing.T) {
	c := New()
	c.Service("early", Service{})
	c.LockEnable()
	c.LockApply()

	r := c.Service("late", Service{})
	AssertFalse(t, r.OK)
}

func TestLock_Startables_Good(t *testing.T) {
	c := New()
	c.Service("s", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Startables()
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_Stoppables_Good(t *testing.T) {
	c := New()
	c.Service("s", Service{OnStop: func() Result { return Result{OK: true} }})
	r := c.Stoppables()
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_LockUnlock_Good(t *testing.T) {
	c := New()
	l := c.Lock("a")
	l.Lock()
	l.Unlock()
}

func TestLock_LockUnlock_Bad(t *testing.T) {
	c := New()
	l := c.Lock("held")
	l.Lock()
	defer l.Unlock()
	r := l.TryLock()
	AssertFalse(t, r.OK, "TryLock on already-held lock must report not-acquired")
}

func TestLock_LockUnlock_Ugly(t *testing.T) {
	c := New()
	l := c.Lock("reentry")
	l.Lock()
	l.Unlock()
	l.Lock()
	l.Unlock()
}

func TestLock_RLockRUnlock_Good(t *testing.T) {
	c := New()
	l := c.Lock("a")
	l.RLock()
	l.RUnlock()
}

func TestLock_RLockRUnlock_Bad(t *testing.T) {
	if os.Getenv("CORE_LOCK_RUNLOCK_BAD") == "1" {
		c := New()
		l := c.Lock("not-rlocked")
		l.RUnlock()
		return
	}

	t.Run("without-prior-rlock", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=^TestLock_RLockRUnlock_Bad$")
		cmd.Env = append(os.Environ(), "CORE_LOCK_RUNLOCK_BAD=1")
		out, err := cmd.CombinedOutput()

		AssertError(t, err)
		AssertContains(t, string(out), "sync: RUnlock of unlocked RWMutex")
	})
}

func TestLock_RLockRUnlock_Ugly(t *testing.T) {
	c := New()
	l := c.Lock("a")
	l.RLock()
	l.RLock()
	l.RUnlock()
	l.RUnlock()
}

func TestLock_TryLock_Good(t *testing.T) {
	c := New()
	l := c.Lock("a")
	r := l.TryLock()
	AssertTrue(t, r.OK)
	l.Unlock()
}

func TestLock_TryLock_Bad(t *testing.T) {
	c := New()
	l := c.Lock("held")
	l.Lock()
	defer l.Unlock()
	r := l.TryLock()
	AssertFalse(t, r.OK)
}

func TestLock_TryLock_Ugly(t *testing.T) {
	c := New()
	l := c.Lock("a")
	r1 := l.TryLock()
	AssertTrue(t, r1.OK)
	r2 := l.TryLock()
	AssertFalse(t, r2.OK)
	l.Unlock()
	r3 := l.TryLock()
	AssertTrue(t, r3.OK)
	l.Unlock()
}
