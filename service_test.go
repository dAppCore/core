package core_test

import (
	. "dappco.re/go"
)

// --- Service Registration ---

func TestService_Register_Good(t *T) {
	c := New()
	r := c.Service("auth", Service{})
	AssertTrue(t, r.OK)
}

func TestService_Register_Duplicate_Bad(t *T) {
	c := New()
	c.Service("auth", Service{})
	r := c.Service("auth", Service{})
	AssertFalse(t, r.OK)
}

func TestService_Register_Empty_Bad(t *T) {
	c := New()
	r := c.Service("", Service{})
	AssertFalse(t, r.OK)
}

func TestService_Get_Good(t *T) {
	c := New()
	c.Service("brain", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Service("brain")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestService_Get_Bad(t *T) {
	c := New()
	r := c.Service("nonexistent")
	AssertFalse(t, r.OK)
}

func TestService_Names_Good(t *T) {
	c := New()
	c.Service("a", Service{})
	c.Service("b", Service{})
	names := c.Services()
	AssertContains(t, names, "a")
	AssertContains(t, names, "b")
	AssertContains(t, names, "cli") // auto-registered by CliRegister in New()
}

// --- Service Lifecycle ---

func TestService_Lifecycle_Good(t *T) {
	c := New()
	started := false
	stopped := false
	c.Service("lifecycle", Service{
		OnStart: func() Result { started = true; return Result{OK: true} },
		OnStop:  func() Result { stopped = true; return Result{OK: true} },
	})

	sr := c.Startables()
	AssertTrue(t, sr.OK)
	startables := sr.Value.([]*Service)
	AssertLen(t, startables, 1)
	startables[0].OnStart()
	AssertTrue(t, started)

	tr := c.Stoppables()
	AssertTrue(t, tr.OK)
	stoppables := tr.Value.([]*Service)
	AssertLen(t, stoppables, 1)
	stoppables[0].OnStop()
	AssertTrue(t, stopped)
}

type autoLifecycleService struct {
	started  bool
	stopped  bool
	messages []Message
}

func (s *autoLifecycleService) OnStartup(_ Context) Result {
	s.started = true
	return Result{OK: true}
}

func (s *autoLifecycleService) OnShutdown(_ Context) Result {
	s.stopped = true
	return Result{OK: true}
}

func (s *autoLifecycleService) HandleIPCEvents(_ *Core, msg Message) Result {
	s.messages = append(s.messages, msg)
	return Result{OK: true}
}

func TestService_RegisterService_Bad(t *T) {
	t.Run("EmptyName", func(t *T) {
		c := New()
		r := c.RegisterService("", "value")
		AssertFalse(t, r.OK)

		err, ok := r.Value.(error)
		AssertTrue(t, ok)
		if ok {
			AssertEqual(t, "core.RegisterService", Operation(err))
		}
	})

	t.Run("DuplicateName", func(t *T) {
		c := New()
		AssertTrue(t, c.RegisterService("svc", "first").OK)

		r := c.RegisterService("svc", "second")
		AssertFalse(t, r.OK)
	})

	t.Run("LockedRegistry", func(t *T) {
		c := New()
		c.LockEnable()
		c.LockApply()

		r := c.RegisterService("blocked", "value")
		AssertFalse(t, r.OK)
	})
}

func TestService_RegisterService_Ugly(t *T) {
	t.Run("AutoDiscoversLifecycleAndIPCHandlers", func(t *T) {
		c := New()
		svc := &autoLifecycleService{}

		r := c.RegisterService("auto", svc)
		AssertTrue(t, r.OK)
		AssertTrue(t, c.ServiceStartup(Background(), nil).OK)
		AssertTrue(t, c.ACTION("ping").OK)
		AssertTrue(t, c.ServiceShutdown(Background()).OK)
		AssertTrue(t, svc.started)
		AssertTrue(t, svc.stopped)
		AssertContains(t, svc.messages, Message("ping"))
	})

	t.Run("NilInstanceReturnsServiceDTO", func(t *T) {
		c := New()
		AssertTrue(t, c.RegisterService("nil", nil).OK)

		r := c.Service("nil")
		AssertTrue(t, r.OK)
		if r.OK {
			svc, ok := r.Value.(*Service)
			AssertTrue(t, ok)
			if ok {
				AssertEqual(t, "nil", svc.Name)
				AssertNil(t, svc.Instance)
			}
		}
	})
}

func TestService_ServiceFor_Bad(t *T) {
	typed, ok := ServiceFor[string](New(), "missing")
	AssertFalse(t, ok)
	AssertEqual(t, "", typed)
}

func TestService_ServiceFor_Ugly(t *T) {
	c := New()
	AssertTrue(t, c.RegisterService("value", "hello").OK)

	typed, ok := ServiceFor[int](c, "value")
	AssertFalse(t, ok)
	AssertEqual(t, 0, typed)
}

func TestService_MustServiceFor_Bad(t *T) {
	c := New()
	AssertPanicsWithError(t, `core.MustServiceFor: service "missing" not found or wrong type`, func() {
		_ = MustServiceFor[string](c, "missing")
	})
}

func TestService_MustServiceFor_Ugly(t *T) {
	var c *Core
	AssertPanics(t, func() {
		_ = MustServiceFor[string](c, "missing")
	})
}

func TestService_Core_RegisterService_Good(t *T) {
	c := New()
	r := c.RegisterService("agent", "dispatch")
	AssertTrue(t, r.OK)

	got, ok := ServiceFor[string](c, "agent")
	AssertTrue(t, ok)
	AssertEqual(t, "dispatch", got)
}

func TestService_Core_RegisterService_Bad(t *T) {
	c := New()
	r := c.RegisterService("", "dispatch")
	AssertFalse(t, r.OK)
	AssertEqual(t, "core.RegisterService", Operation(r.Value.(error)))
}

func TestService_Core_RegisterService_Ugly(t *T) {
	c := New()
	r := c.RegisterService("empty", nil)
	AssertTrue(t, r.OK)

	svc := c.Service("empty").Value.(*Service)
	AssertEqual(t, "empty", svc.Name)
	AssertNil(t, svc.Instance)
}

func TestService_Core_Service_Good(t *T) {
	c := New()
	r := c.Service("agent", Service{OnStart: func() Result { return Result{OK: true} }})
	AssertTrue(t, r.OK)

	got := c.Service("agent")
	AssertTrue(t, got.OK)
	AssertEqual(t, "agent", got.Value.(*Service).Name)
}

func TestService_Core_Service_Bad(t *T) {
	r := New().Service("missing")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestService_Core_Service_Ugly(t *T) {
	c := New()
	r := c.Service("agent", Service{Instance: "dispatch-runtime"})
	AssertTrue(t, r.OK)

	got := c.Service("agent")
	AssertTrue(t, got.OK)
	AssertEqual(t, "dispatch-runtime", got.Value)
}

func TestService_Core_Services_Good(t *T) {
	c := New()
	c.Service("agent", Service{})
	c.Service("health", Service{})

	names := c.Services()

	AssertContains(t, names, "agent")
	AssertContains(t, names, "health")
}

func TestService_Core_Services_Bad(t *T) {
	c := &Core{}
	AssertNil(t, c.Services())
}

func TestService_Core_Services_Ugly(t *T) {
	names := New().Services()
	AssertContains(t, names, "cli")
}

func TestService_ServiceFor_Good(t *T) {
	c := New()
	AssertTrue(t, c.RegisterService("agent", "dispatch").OK)

	got, ok := ServiceFor[string](c, "agent")

	AssertTrue(t, ok)
	AssertEqual(t, "dispatch", got)
}

func TestService_MustServiceFor_Good(t *T) {
	c := New()
	AssertTrue(t, c.RegisterService("agent", "dispatch").OK)

	got := MustServiceFor[string](c, "agent")

	AssertEqual(t, "dispatch", got)
}
