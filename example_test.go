package core_test

import (
	"context"

	. "dappco.re/go/core"
)

// --- Core Creation ---

func ExampleNew() {
	c := New(
		WithOption("name", "my-app"),
		WithServiceLock(),
	)
	Println(c.App().Name)
	// Output: my-app
}

func ExampleNew_withService() {
	c := New(
		WithOption("name", "example"),
		WithService(func(c *Core) Result {
			return c.Service("greeter", Service{
				OnStart: func() Result {
					Info("greeter started", "app", c.App().Name)
					return Result{OK: true}
				},
			})
		}),
	)
	c.ServiceStartup(context.Background(), nil)
	Println(c.Services())
	c.ServiceShutdown(context.Background())
	// Output is non-deterministic (map order), so no Output comment
}

// --- Options ---

func ExampleNewOptions() {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
		Option{Key: "debug", Value: true},
	)
	Println(opts.String("name"))
	Println(opts.Int("port"))
	Println(opts.Bool("debug"))
	// Output:
	// brain
	// 8080
	// true
}

// --- Result ---

func ExampleResult() {
	r := Result{Value: "hello", OK: true}
	if r.OK {
		Println(r.Value)
	}
	// Output: hello
}

// --- Action ---

func ExampleCore_Action_register() {
	c := New()
	c.Action("greet", func(_ context.Context, opts Options) Result {
		name := opts.String("name")
		return Result{Value: Concat("hello ", name), OK: true}
	})
	Println(c.Action("greet").Exists())
	// Output: true
}

func ExampleCore_Action_invoke() {
	c := New()
	c.Action("add", func(_ context.Context, opts Options) Result {
		a := opts.Int("a")
		b := opts.Int("b")
		return Result{Value: a + b, OK: true}
	})

	r := c.Action("add").Run(context.Background(), NewOptions(
		Option{Key: "a", Value: 3},
		Option{Key: "b", Value: 4},
	))
	Println(r.Value)
	// Output: 7
}

func ExampleCore_Actions() {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("brain.recall", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	Println(c.Actions())
	// Output: [process.run brain.recall]
}

// --- Task ---

func ExampleCore_Task() {
	c := New()
	order := ""

	c.Action("step.a", func(_ context.Context, _ Options) Result {
		order += "a"
		return Result{Value: "from-a", OK: true}
	})
	c.Action("step.b", func(_ context.Context, opts Options) Result {
		order += "b"
		return Result{OK: true}
	})

	c.Task("pipeline", Task{
		Steps: []Step{
			{Action: "step.a"},
			{Action: "step.b", Input: "previous"},
		},
	})

	c.Task("pipeline").Run(context.Background(), c, NewOptions())
	Println(order)
	// Output: ab
}

// --- Registry ---

func ExampleNewRegistry() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("bravo", "second")

	Println(r.Has("alpha"))
	Println(r.Names())
	Println(r.Len())
	// Output:
	// true
	// [alpha bravo]
	// 2
}

func ExampleRegistry_Lock() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()

	result := r.Set("beta", "second")
	Println(result.OK)
	// Output: false
}

func ExampleRegistry_Seal() {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()

	// Can update existing
	Println(r.Set("alpha", "updated").OK)
	// Can't add new
	Println(r.Set("beta", "new").OK)
	// Output:
	// true
	// false
}

// --- Entitlement ---

func ExampleCore_Entitled_default() {
	c := New()
	e := c.Entitled("anything")
	Println(e.Allowed)
	Println(e.Unlimited)
	// Output:
	// true
	// true
}

func ExampleCore_Entitled_custom() {
	c := New()
	c.SetEntitlementChecker(func(action string, qty int, _ context.Context) Entitlement {
		if action == "premium" {
			return Entitlement{Allowed: false, Reason: "upgrade required"}
		}
		return Entitlement{Allowed: true, Unlimited: true}
	})

	Println(c.Entitled("basic").Allowed)
	Println(c.Entitled("premium").Allowed)
	Println(c.Entitled("premium").Reason)
	// Output:
	// true
	// false
	// upgrade required
}

func ExampleEntitlement_NearLimit() {
	e := Entitlement{Allowed: true, Limit: 100, Used: 85, Remaining: 15}
	Println(e.NearLimit(0.8))
	Println(e.UsagePercent())
	// Output:
	// true
	// 85
}

// --- Process ---

func ExampleCore_Process() {
	c := New()
	// No go-process registered — permission by registration
	Println(c.Process().Exists())

	// Register a mock process handler
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: Concat("output of ", opts.String("command")), OK: true}
	})
	Println(c.Process().Exists())

	r := c.Process().Run(context.Background(), "echo", "hello")
	Println(r.Value)
	// Output:
	// false
	// true
	// output of echo
}

// --- JSON ---

func ExampleJSONMarshal() {
	type config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	r := JSONMarshal(config{Host: "localhost", Port: 8080})
	Println(string(r.Value.([]byte)))
	// Output: {"host":"localhost","port":8080}
}

func ExampleJSONUnmarshalString() {
	type config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	var cfg config
	JSONUnmarshalString(`{"host":"localhost","port":8080}`, &cfg)
	Println(cfg.Host, cfg.Port)
	// Output: localhost 8080
}

// --- Utilities ---

func ExampleID() {
	id := ID()
	Println(HasPrefix(id, "id-"))
	// Output: true
}

func ExampleValidateName() {
	Println(ValidateName("brain").OK)
	Println(ValidateName("").OK)
	Println(ValidateName("..").OK)
	Println(ValidateName("path/traversal").OK)
	// Output:
	// true
	// false
	// false
	// false
}

func ExampleSanitisePath() {
	Println(SanitisePath("../../etc/passwd"))
	Println(SanitisePath(""))
	Println(SanitisePath("/some/path/file.txt"))
	// Output:
	// passwd
	// invalid
	// file.txt
}

// --- Command ---

func ExampleCore_Command() {
	c := New()
	c.Command("deploy/to/homelab", Command{
		Action: func(opts Options) Result {
			return Result{Value: Concat("deployed to ", opts.String("_arg")), OK: true}
		},
	})

	r := c.Cli().Run("deploy", "to", "homelab")
	Println(r.OK)
	// Output: true
}

// --- Config ---

func ExampleConfig() {
	c := New()
	c.Config().Set("database.host", "localhost")
	c.Config().Set("database.port", 5432)
	c.Config().Enable("dark-mode")

	Println(c.Config().String("database.host"))
	Println(c.Config().Int("database.port"))
	Println(c.Config().Enabled("dark-mode"))
	// Output:
	// localhost
	// 5432
	// true
}

// Error examples in error_example_test.go
