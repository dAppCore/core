package core_test

import (
	"context"

	. "dappco.re/go"
)

type contractLifecycleService struct {
	started bool
	stopped bool
}

func (s *contractLifecycleService) OnStartup(_ context.Context) Result {
	s.started = true
	return Result{OK: true}
}

func (s *contractLifecycleService) OnShutdown(_ context.Context) Result {
	s.stopped = true
	return Result{OK: true}
}

func ExampleMessage() {
	var msg Message = ActionServiceStartup{}
	Println(Sprint(msg))
	// Output: {}
}

func ExampleQuery() {
	var q Query = "status"
	Println(q)
	// Output: status
}

func ExampleQueryHandler() {
	var handler QueryHandler = func(_ *Core, q Query) Result {
		return Result{Value: Concat("query:", q.(string)), OK: true}
	}
	Println(handler(New(), "status").Value)
	// Output: query:status
}

func ExampleStartable() {
	var _ Startable = (*contractLifecycleService)(nil)
}

func ExampleStoppable() {
	var _ Stoppable = (*contractLifecycleService)(nil)
}

func ExampleActionServiceStartup() {
	Println(Sprint(ActionServiceStartup{}))
	// Output: {}
}

func ExampleActionServiceShutdown() {
	Println(Sprint(ActionServiceShutdown{}))
	// Output: {}
}

func ExampleActionTaskStarted() {
	ev := ActionTaskStarted{TaskIdentifier: "task-1", Action: "deploy"}
	Println(ev.TaskIdentifier)
	Println(ev.Action)
	// Output:
	// task-1
	// deploy
}

func ExampleActionTaskProgress() {
	ev := ActionTaskProgress{TaskIdentifier: "task-1", Action: "deploy", Progress: 0.5, Message: "halfway"}
	Println(ev.TaskIdentifier)
	Println(ev.Progress)
	Println(ev.Message)
	// Output:
	// task-1
	// 0.5
	// halfway
}

func ExampleActionTaskCompleted() {
	ev := ActionTaskCompleted{TaskIdentifier: "task-1", Action: "deploy", Result: Result{Value: "done", OK: true}}
	Println(ev.Action)
	Println(ev.Result.Value)
	// Output:
	// deploy
	// done
}

func ExampleCoreOption() {
	var opt CoreOption = WithOption("name", "ops")
	c := New(opt)
	Println(c.App().Name)
	// Output: ops
}

func ExampleNew_withOptions() {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "ops"})))
	Println(c.App().Name)
	// Output: ops
}

func ExampleWithOptions() {
	c := New(WithOptions(NewOptions(Option{Key: "debug", Value: true})))
	Println(c.Options().Bool("debug"))
	// Output: true
}

func ExampleWithService_factory() {
	c := New(WithService(func(c *Core) Result {
		return c.Service("worker", Service{OnStart: func() Result { return Result{OK: true} }})
	}))
	Println(c.Service("worker").OK)
	// Output: true
}

func ExampleWithName() {
	c := New(WithName("lifecycle", func(_ *Core) Result {
		return Result{Value: &contractLifecycleService{}, OK: true}
	}))
	Println(c.Service("lifecycle").OK)
	// Output: true
}

func ExampleWithOption() {
	c := New(WithOption("name", "ops"))
	Println(c.Options().String("name"))
	Println(c.App().Name)
	// Output:
	// ops
	// ops
}

func ExampleWithServiceLock_contract() {
	c := New(WithServiceLock())
	r := c.Service("late", Service{})
	Println(r.OK)
	// Output: false
}
