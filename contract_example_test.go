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

// ExampleMessage assigns a lifecycle event to the message contract through `Message` for
// service contract wiring. Service lifecycle contracts remain small interfaces and option
// hooks.
func ExampleMessage() {
	var msg Message = ActionServiceStartup{}
	Println(Sprint(msg))
	// Output: {}
}

// ExampleQuery runs or declares a query through `Query` for service contract wiring.
// Service lifecycle contracts remain small interfaces and option hooks.
func ExampleQuery() {
	var q Query = "status"
	Println(q)
	// Output: status
}

// ExampleQueryHandler declares a query handler contract through `QueryHandler` for service
// contract wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleQueryHandler() {
	var handler QueryHandler = func(_ *Core, q Query) Result {
		return Result{Value: Concat("query:", q.(string)), OK: true}
	}
	Println(handler(New(), "status").Value)
	// Output: query:status
}

// ExampleStartable declares a startable service contract through `Startable` for service
// contract wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleStartable() {
	var _ Startable = (*contractLifecycleService)(nil)
}

// ExampleStoppable declares a stoppable service contract through `Stoppable` for service
// contract wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleStoppable() {
	var _ Stoppable = (*contractLifecycleService)(nil)
}

// ExampleActionServiceStartup emits the service-startup action contract through
// `ActionServiceStartup` for service contract wiring. Service lifecycle contracts remain
// small interfaces and option hooks.
func ExampleActionServiceStartup() {
	Println(Sprint(ActionServiceStartup{}))
	// Output: {}
}

// ExampleActionServiceShutdown emits the service-shutdown action contract through
// `ActionServiceShutdown` for service contract wiring. Service lifecycle contracts remain
// small interfaces and option hooks.
func ExampleActionServiceShutdown() {
	Println(Sprint(ActionServiceShutdown{}))
	// Output: {}
}

// ExampleActionTaskStarted creates a task-started action event through `ActionTaskStarted`
// for service contract wiring. Service lifecycle contracts remain small interfaces and
// option hooks.
func ExampleActionTaskStarted() {
	ev := ActionTaskStarted{TaskIdentifier: "task-1", Action: "deploy"}
	Println(ev.TaskIdentifier)
	Println(ev.Action)
	// Output:
	// task-1
	// deploy
}

// ExampleActionTaskProgress creates a task-progress action event through
// `ActionTaskProgress` for service contract wiring. Service lifecycle contracts remain
// small interfaces and option hooks.
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

// ExampleActionTaskCompleted creates a task-completed action event through
// `ActionTaskCompleted` for service contract wiring. Service lifecycle contracts remain
// small interfaces and option hooks.
func ExampleActionTaskCompleted() {
	ev := ActionTaskCompleted{TaskIdentifier: "task-1", Action: "deploy", Result: Result{Value: "done", OK: true}}
	Println(ev.Action)
	Println(ev.Result.Value)
	// Output:
	// deploy
	// done
}

// ExampleCoreOption declares a Core option function through `CoreOption` for service
// contract wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleCoreOption() {
	var opt CoreOption = WithOption("name", "ops")
	c := New(opt)
	Println(c.App().Name)
	// Output: ops
}

// ExampleNew_withOptions passes construction options through `New` for service contract
// wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleNew_withOptions() {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "ops"})))
	Println(c.App().Name)
	// Output: ops
}

// ExampleWithOptions applies options through `WithOptions` for service contract wiring.
// Service lifecycle contracts remain small interfaces and option hooks.
func ExampleWithOptions() {
	c := New(WithOptions(NewOptions(Option{Key: "debug", Value: true})))
	Println(c.Options().Bool("debug"))
	// Output: true
}

// ExampleWithService_factory wires a service factory through `WithService` for service
// contract wiring. Service lifecycle contracts remain small interfaces and option hooks.
func ExampleWithService_factory() {
	c := New(WithService(func(c *Core) Result {
		return c.Service("worker", Service{OnStart: func() Result { return Result{OK: true} }})
	}))
	Println(c.Service("worker").OK)
	// Output: true
}

// ExampleWithName applies a service name through `WithName` for service contract wiring.
// Service lifecycle contracts remain small interfaces and option hooks.
func ExampleWithName() {
	c := New(WithName("lifecycle", func(_ *Core) Result {
		return Result{Value: &contractLifecycleService{}, OK: true}
	}))
	Println(c.Service("lifecycle").OK)
	// Output: true
}

// ExampleWithOption applies one option through `WithOption` for service contract wiring.
// Service lifecycle contracts remain small interfaces and option hooks.
func ExampleWithOption() {
	c := New(WithOption("name", "ops"))
	Println(c.Options().String("name"))
	Println(c.App().Name)
	// Output:
	// ops
	// ops
}

// ExampleWithServiceLock_contract documents the locking contract through `WithServiceLock`
// for service contract wiring. Service lifecycle contracts remain small interfaces and
// option hooks.
func ExampleWithServiceLock_contract() {
	c := New(WithServiceLock())
	r := c.Service("late", Service{})
	Println(r.OK)
	// Output: false
}
