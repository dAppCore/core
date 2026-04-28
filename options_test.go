package core_test

import (
	. "dappco.re/go"
)

// --- NewOptions ---

func TestOptions_NewOptions_Good(t *T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	AssertEqual(t, 2, opts.Len())
}

func TestOptions_NewOptions_Empty_Good(t *T) {
	opts := NewOptions()
	AssertEqual(t, 0, opts.Len())
	AssertFalse(t, opts.Has("anything"))
}

// --- Options.Set ---

func TestOptions_Set_Good(t *T) {
	opts := NewOptions()
	opts.Set("name", "brain")
	AssertEqual(t, "brain", opts.String("name"))
}

func TestOptions_Set_Update_Good(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "old"})
	opts.Set("name", "new")
	AssertEqual(t, "new", opts.String("name"))
	AssertEqual(t, 1, opts.Len())
}

// --- Options.Get ---

func TestOptions_Get_Good(t *T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	r := opts.Get("name")
	AssertTrue(t, r.OK)
	AssertEqual(t, "brain", r.Value)
}

func TestOptions_Get_Bad(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	r := opts.Get("missing")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

// --- Options.Has ---

func TestOptions_Has_Good(t *T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	AssertTrue(t, opts.Has("debug"))
	AssertFalse(t, opts.Has("missing"))
}

// --- Options.String ---

func TestOptions_String_Good(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertEqual(t, "brain", opts.String("name"))
}

func TestOptions_String_Bad(t *T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	AssertEqual(t, "", opts.String("port"))
	AssertEqual(t, "", opts.String("missing"))
}

// --- Options.Int ---

func TestOptions_Int_Good(t *T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	AssertEqual(t, 8080, opts.Int("port"))
}

func TestOptions_Int_Bad(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertEqual(t, 0, opts.Int("name"))
	AssertEqual(t, 0, opts.Int("missing"))
}

// --- Options.Bool ---

func TestOptions_Bool_Good(t *T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	AssertTrue(t, opts.Bool("debug"))
}

func TestOptions_Bool_Bad(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertFalse(t, opts.Bool("name"))
	AssertFalse(t, opts.Bool("missing"))
}

// --- Options.Items ---

func TestOptions_Items_Good(t *T) {
	opts := NewOptions(Option{Key: "a", Value: 1}, Option{Key: "b", Value: 2})
	items := opts.Items()
	AssertLen(t, items, 2)
}

// --- Options with typed struct ---

func TestOptions_TypedStruct_Good(t *T) {
	type BrainConfig struct {
		Name       string
		OllamaURL  string
		Collection string
	}
	cfg := BrainConfig{Name: "brain", OllamaURL: "http://localhost:11434", Collection: "openbrain"}
	opts := NewOptions(Option{Key: "config", Value: cfg})

	r := opts.Get("config")
	AssertTrue(t, r.OK)
	bc, ok := r.Value.(BrainConfig)
	AssertTrue(t, ok)
	AssertEqual(t, "brain", bc.Name)
	AssertEqual(t, "http://localhost:11434", bc.OllamaURL)
}

// --- Result ---

func TestOptions_Result_New_Good(t *T) {
	r := Result{}.New("value")
	AssertEqual(t, "value", r.Value)
}

func TestOptions_Result_New_Error_Bad(t *T) {
	err := E("test", "failed", nil)
	r := Result{}.New(err)
	AssertFalse(t, r.OK)
	AssertEqual(t, err, r.Value)
}

// --- WithOption ---

func TestOptions_WithOption_Good(t *T) {
	c := New(
		WithOption("name", "myapp"),
		WithOption("port", 8080),
	)
	AssertEqual(t, "myapp", c.App().Name)
	AssertEqual(t, 8080, c.Options().Int("port"))
}

func TestOptions_NewOptions_Bad(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"}, Option{Key: "agent", Value: "hades"})

	AssertEqual(t, 2, opts.Len())
	AssertEqual(t, "codex", opts.String("agent"))
}

func TestOptions_NewOptions_Ugly(t *T) {
	items := []Option{{Key: "agent", Value: "codex"}}
	opts := NewOptions(items...)
	items[0].Value = "hades"

	AssertEqual(t, "codex", opts.String("agent"))
}

func TestOptions_Options_Bool_Good(t *T) {
	opts := NewOptions(Option{Key: "enabled", Value: true})

	AssertTrue(t, opts.Bool("enabled"))
}

func TestOptions_Options_Bool_Bad(t *T) {
	opts := NewOptions(Option{Key: "enabled", Value: "true"})

	AssertFalse(t, opts.Bool("enabled"))
}

func TestOptions_Options_Bool_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: true})

	AssertTrue(t, opts.Bool(""))
}

func TestOptions_Options_Get_Good(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})
	r := opts.Get("agent")

	AssertTrue(t, r.OK)
	AssertEqual(t, "codex", r.Value)
}

func TestOptions_Options_Get_Bad(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})
	r := opts.Get("missing")

	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestOptions_Options_Get_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: "empty-key"})
	r := opts.Get("")

	AssertTrue(t, r.OK)
	AssertEqual(t, "empty-key", r.Value)
}

func TestOptions_Options_Has_Good(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})

	AssertTrue(t, opts.Has("agent"))
}

func TestOptions_Options_Has_Bad(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})

	AssertFalse(t, opts.Has("missing"))
}

func TestOptions_Options_Has_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: "empty-key"})

	AssertTrue(t, opts.Has(""))
}

func TestOptions_Options_Int_Good(t *T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})

	AssertEqual(t, 8080, opts.Int("port"))
}

func TestOptions_Options_Int_Bad(t *T) {
	opts := NewOptions(Option{Key: "port", Value: "8080"})

	AssertEqual(t, 0, opts.Int("port"))
}

func TestOptions_Options_Int_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: -1})

	AssertEqual(t, -1, opts.Int(""))
}

func TestOptions_Options_Items_Good(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"}, Option{Key: "region", Value: "homelab"})

	AssertEqual(t, []Option{{Key: "agent", Value: "codex"}, {Key: "region", Value: "homelab"}}, opts.Items())
}

func TestOptions_Options_Items_Bad(t *T) {
	opts := NewOptions()

	AssertEmpty(t, opts.Items())
}

func TestOptions_Options_Items_Ugly(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})
	items := opts.Items()
	items[0].Value = "hades"

	AssertEqual(t, "codex", opts.String("agent"))
}

func TestOptions_Options_Len_Good(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"}, Option{Key: "debug", Value: true})

	AssertEqual(t, 2, opts.Len())
}

func TestOptions_Options_Len_Bad(t *T) {
	opts := NewOptions()

	AssertEqual(t, 0, opts.Len())
}

func TestOptions_Options_Len_Ugly(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})
	opts.Set("agent", "hades")

	AssertEqual(t, 1, opts.Len())
}

func TestOptions_Options_Set_Good(t *T) {
	opts := NewOptions()
	opts.Set("agent", "codex")

	AssertEqual(t, "codex", opts.String("agent"))
}

func TestOptions_Options_Set_Bad(t *T) {
	var opts *Options

	AssertPanics(t, func() { opts.Set("agent", "codex") })
}

func TestOptions_Options_Set_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: "before"})
	opts.Set("", "after")

	AssertEqual(t, 1, opts.Len())
	AssertEqual(t, "after", opts.String(""))
}

func TestOptions_Options_String_Good(t *T) {
	opts := NewOptions(Option{Key: "agent", Value: "codex"})

	AssertEqual(t, "codex", opts.String("agent"))
}

func TestOptions_Options_String_Bad(t *T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})

	AssertEqual(t, "", opts.String("port"))
}

func TestOptions_Options_String_Ugly(t *T) {
	opts := NewOptions(Option{Key: "", Value: "empty-key"})

	AssertEqual(t, "empty-key", opts.String(""))
}

// --- Options.Float64 ---

func TestOptions_Options_Float64_Good(t *T) {
	opts := NewOptions(Option{Key: "weight", Value: 0.75})
	AssertEqual(t, 0.75, opts.Float64("weight"))
}

func TestOptions_Options_Float64_Bad(t *T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertEqual(t, 0.0, opts.Float64("name"))
	AssertEqual(t, 0.0, opts.Float64("missing"))
}

func TestOptions_Options_Float64_Ugly(t *T) {
	// int and float32 promote to float64 — JSON-decoded numbers work.
	opts := NewOptions(
		Option{Key: "i", Value: 42},
		Option{Key: "i64", Value: int64(99)},
		Option{Key: "f32", Value: float32(1.5)},
	)
	AssertEqual(t, 42.0, opts.Float64("i"))
	AssertEqual(t, 99.0, opts.Float64("i64"))
	AssertEqual(t, 1.5, opts.Float64("f32"))
}

// --- Options.Duration ---

func TestOptions_Options_Duration_Good(t *T) {
	opts := NewOptions(Option{Key: "timeout", Value: 5 * Second})
	AssertEqual(t, 5*Second, opts.Duration("timeout"))
}

func TestOptions_Options_Duration_Bad(t *T) {
	opts := NewOptions(Option{Key: "name", Value: 12345})
	AssertEqual(t, Duration(0), opts.Duration("name"))
	AssertEqual(t, Duration(0), opts.Duration("missing"))
}

func TestOptions_Options_Duration_Ugly(t *T) {
	// String values parse via ParseDuration; bad strings return 0.
	opts := NewOptions(
		Option{Key: "good", Value: "250ms"},
		Option{Key: "bad", Value: "not-a-duration"},
	)
	AssertEqual(t, 250*Millisecond, opts.Duration("good"))
	AssertEqual(t, Duration(0), opts.Duration("bad"))
}

func TestOptions_Result_New_Bad(t *T) {
	r := Result{}.New(AnError)

	AssertFalse(t, r.OK)
	AssertEqual(t, AnError, r.Value)
}

func TestOptions_Result_New_Ugly(t *T) {
	r := Result{Value: "existing", OK: true}.New()

	AssertTrue(t, r.OK)
	AssertEqual(t, "existing", r.Value)
}
