package core_test

import . "dappco.re/go"

func TestLock_Good(t *T) {
	c := New()
	lock := c.Lock("test")
	AssertNotNil(t, lock)
	AssertNotNil(t, lock.Mutex)
}

func TestLock_SameName_Good(t *T) {
	c := New()
	l1 := c.Lock("shared")
	l2 := c.Lock("shared")
	AssertEqual(t, l1, l2)
}

func TestLock_DifferentName_Good(t *T) {
	c := New()
	l1 := c.Lock("a")
	l2 := c.Lock("b")
	AssertNotEqual(t, l1, l2)
}

func TestLock_LockEnable_Good(t *T) {
	c := New()
	c.Service("early", Service{})
	c.LockEnable()
	c.LockApply()

	r := c.Service("late", Service{})
	AssertFalse(t, r.OK)
}

func TestLock_Startables_Good(t *T) {
	c := New()
	c.Service("s", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Startables()
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_Stoppables_Good(t *T) {
	c := New()
	c.Service("s", Service{OnStop: func() Result { return Result{OK: true} }})
	r := c.Stoppables()
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_LockUnlock_Good(t *T) {
	c := New()
	l := c.Lock("a")
	l.Lock()
	l.Unlock()
}

func TestLock_LockUnlock_Bad(t *T) {
	c := New()
	l := c.Lock("held")
	l.Lock()
	defer l.Unlock()
	r := l.TryLock()
	AssertFalse(t, r.OK, "TryLock on already-held lock must report not-acquired")
}

func TestLock_LockUnlock_Ugly(t *T) {
	c := New()
	l := c.Lock("reentry")
	l.Lock()
	l.Unlock()
	l.Lock()
	l.Unlock()
}

func TestLock_RLockRUnlock_Good(t *T) {
	c := New()
	l := c.Lock("a")
	l.RLock()
	l.RUnlock()
}

func TestLock_RLockRUnlock_Bad(t *T) {
	if Getenv("CORE_LOCK_RUNLOCK_BAD") == "1" {
		c := New()
		l := c.Lock("not-rlocked")
		l.RUnlock()
		return
	}

	t.Run("without-prior-rlock", func(t *T) {
		cmd := ExecCmdForTest(Args()[0], "-test.run=^TestLock_RLockRUnlock_Bad$")
		cmd.Env = append(Environ(), "CORE_LOCK_RUNLOCK_BAD=1")
		out, err := cmd.CombinedOutput()

		AssertError(t, err)
		AssertContains(t, string(out), "sync: RUnlock of unlocked RWMutex")
	})
}

func TestLock_RLockRUnlock_Ugly(t *T) {
	c := New()
	l := c.Lock("a")
	l.RLock()
	l.RLock()
	l.RUnlock()
	l.RUnlock()
}

func TestLock_TryLock_Good(t *T) {
	c := New()
	l := c.Lock("a")
	r := l.TryLock()
	AssertTrue(t, r.OK)
	l.Unlock()
}

func TestLock_TryLock_Bad(t *T) {
	c := New()
	l := c.Lock("held")
	l.Lock()
	defer l.Unlock()
	r := l.TryLock()
	AssertFalse(t, r.OK)
}

func TestLock_TryLock_Ugly(t *T) {
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

func TestLock_Core_Lock_Good(t *T) {
	c := New()

	lock := c.Lock("agent.dispatch")

	AssertEqual(t, "agent.dispatch", lock.Name)
	AssertNotNil(t, lock.Mutex)
}

func TestLock_Core_Lock_Bad(t *T) {
	c := New()

	lock := c.Lock("")

	AssertEqual(t, "", lock.Name)
	AssertNotNil(t, lock.Mutex)
}

func TestLock_Core_Lock_Ugly(t *T) {
	c := New()

	first := c.Lock("agent.dispatch")
	second := c.Lock("agent.dispatch")

	AssertSame(t, first.Mutex, second.Mutex)
}

func TestLock_Core_LockEnable_Good(t *T) {
	c := New()

	c.LockEnable()
	c.LockApply()
	r := c.Service("late", Service{})

	AssertFalse(t, r.OK)
}

func TestLock_Core_LockEnable_Bad(t *T) {
	c := New()

	c.LockEnable("ignored")
	c.LockApply("ignored")
	r := c.Service("late", Service{})

	AssertFalse(t, r.OK)
}

func TestLock_Core_LockEnable_Ugly(t *T) {
	c := New()

	c.LockEnable()
	c.LockEnable()
	c.LockApply()
	r := c.Service("late", Service{})

	AssertFalse(t, r.OK)
}

func TestLock_Core_LockApply_Good(t *T) {
	c := New()
	c.LockEnable()

	c.LockApply()
	r := c.Service("late", Service{})

	AssertFalse(t, r.OK)
}

func TestLock_Core_LockApply_Bad(t *T) {
	c := New()

	c.LockApply()
	r := c.Service("late", Service{})

	AssertTrue(t, r.OK)
}

func TestLock_Core_LockApply_Ugly(t *T) {
	c := New()
	c.LockEnable()

	c.LockApply()
	c.LockApply()
	r := c.Service("late", Service{})

	AssertFalse(t, r.OK)
}

func TestLock_Core_Startables_Good(t *T) {
	c := New()
	c.Service("agent.dispatch", Service{OnStart: func() Result { return Result{OK: true} }})

	r := c.Startables()

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_Core_Startables_Bad(t *T) {
	c := New()
	c.Service("agent.dispatch", Service{})

	r := c.Startables()

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 0)
}

func TestLock_Core_Startables_Ugly(t *T) {
	c := New()
	c.Service("agent.prepare", Service{OnStart: func() Result { return Result{OK: true} }})
	c.Service("agent.cache", Service{})
	c.Service("agent.dispatch", Service{OnStart: func() Result { return Result{OK: true} }})

	r := c.Startables()

	AssertTrue(t, r.OK)
	services := r.Value.([]*Service)
	AssertLen(t, services, 2)
	AssertEqual(t, "agent.prepare", services[0].Name)
	AssertEqual(t, "agent.dispatch", services[1].Name)
}

func TestLock_Core_Stoppables_Good(t *T) {
	c := New()
	c.Service("agent.dispatch", Service{OnStop: func() Result { return Result{OK: true} }})

	r := c.Stoppables()

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 1)
}

func TestLock_Core_Stoppables_Bad(t *T) {
	c := New()
	c.Service("agent.dispatch", Service{})

	r := c.Stoppables()

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]*Service), 0)
}

func TestLock_Core_Stoppables_Ugly(t *T) {
	c := New()
	c.Service("agent.prepare", Service{OnStop: func() Result { return Result{OK: true} }})
	c.Service("agent.cache", Service{})
	c.Service("agent.dispatch", Service{OnStop: func() Result { return Result{OK: true} }})

	r := c.Stoppables()

	AssertTrue(t, r.OK)
	services := r.Value.([]*Service)
	AssertLen(t, services, 2)
	AssertEqual(t, "agent.prepare", services[0].Name)
	AssertEqual(t, "agent.dispatch", services[1].Name)
}

func TestLock_Lock_Lock_Good(t *T) {
	lock := New().Lock("agent.dispatch")

	lock.Lock()
	lock.Unlock()
}

func TestLock_Lock_Lock_Bad(t *T) {
	lock := New().Lock("agent.dispatch")
	lock.Lock()
	defer lock.Unlock()

	r := lock.TryLock()

	AssertFalse(t, r.OK)
}

func TestLock_Lock_Lock_Ugly(t *T) {
	lock := New().Lock("agent.dispatch")
	var count AtomicInt32
	var wg WaitGroup

	for i := 0; i < 16; i++ {
		wg.Go(func() {
			lock.Lock()
			count.Add(1)
			lock.Unlock()
		})
	}
	wg.Wait()

	AssertEqual(t, int32(16), count.Load())
}

func TestLock_Lock_Unlock_Good(t *T) {
	lock := New().Lock("agent.dispatch")

	lock.Lock()
	lock.Unlock()

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_Unlock_Bad(t *T) {
	lock := New().Lock("agent.dispatch")
	lock.Lock()

	lock.Unlock()

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_Unlock_Ugly(t *T) {
	lock := New().Lock("agent.dispatch")

	for i := 0; i < 3; i++ {
		lock.Lock()
		lock.Unlock()
	}

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_RLock_Good(t *T) {
	lock := New().Lock("agent.dispatch")

	lock.RLock()
	lock.RUnlock()
}

func TestLock_Lock_RLock_Bad(t *T) {
	lock := New().Lock("agent.dispatch")
	lock.Lock()
	defer lock.Unlock()

	r := lock.Mutex.TryRLock()

	AssertFalse(t, r.OK)
}

func TestLock_Lock_RLock_Ugly(t *T) {
	lock := New().Lock("agent.dispatch")

	lock.RLock()
	lock.RLock()
	lock.RUnlock()
	lock.RUnlock()

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_RUnlock_Good(t *T) {
	lock := New().Lock("agent.dispatch")

	lock.RLock()
	lock.RUnlock()

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_RUnlock_Bad(t *T) {
	lock := New().Lock("agent.dispatch")
	lock.RLock()

	lock.RUnlock()

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_RUnlock_Ugly(t *T) {
	lock := New().Lock("agent.dispatch")

	for i := 0; i < 3; i++ {
		lock.RLock()
		lock.RUnlock()
	}

	r := lock.TryLock()
	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_TryLock_Good(t *T) {
	lock := New().Lock("agent.dispatch")

	r := lock.TryLock()

	AssertTrue(t, r.OK)
	lock.Unlock()
}

func TestLock_Lock_TryLock_Bad(t *T) {
	lock := New().Lock("agent.dispatch")
	lock.Lock()
	defer lock.Unlock()

	r := lock.TryLock()

	AssertFalse(t, r.OK)
}

func TestLock_Lock_TryLock_Ugly(t *T) {
	lock := New().Lock("agent.dispatch")

	first := lock.TryLock()
	second := lock.TryLock()

	AssertTrue(t, first.OK)
	AssertFalse(t, second.OK)
	lock.Unlock()
}
