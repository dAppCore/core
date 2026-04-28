package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

// --- NewOptions ---

func TestOptions_NewOptions_Good(t *testing.T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	AssertEqual(t, 2, opts.Len())
}

func TestOptions_NewOptions_Empty_Good(t *testing.T) {
	opts := NewOptions()
	AssertEqual(t, 0, opts.Len())
	AssertFalse(t, opts.Has("anything"))
}

// --- Options.Set ---

func TestOptions_Set_Good(t *testing.T) {
	opts := NewOptions()
	opts.Set("name", "brain")
	AssertEqual(t, "brain", opts.String("name"))
}

func TestOptions_Set_Update_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "old"})
	opts.Set("name", "new")
	AssertEqual(t, "new", opts.String("name"))
	AssertEqual(t, 1, opts.Len())
}

// --- Options.Get ---

func TestOptions_Get_Good(t *testing.T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	r := opts.Get("name")
	AssertTrue(t, r.OK)
	AssertEqual(t, "brain", r.Value)
}

func TestOptions_Get_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	r := opts.Get("missing")
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

// --- Options.Has ---

func TestOptions_Has_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	AssertTrue(t, opts.Has("debug"))
	AssertFalse(t, opts.Has("missing"))
}

// --- Options.String ---

func TestOptions_String_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertEqual(t, "brain", opts.String("name"))
}

func TestOptions_String_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	AssertEqual(t, "", opts.String("port"))
	AssertEqual(t, "", opts.String("missing"))
}

// --- Options.Int ---

func TestOptions_Int_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	AssertEqual(t, 8080, opts.Int("port"))
}

func TestOptions_Int_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertEqual(t, 0, opts.Int("name"))
	AssertEqual(t, 0, opts.Int("missing"))
}

// --- Options.Bool ---

func TestOptions_Bool_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	AssertTrue(t, opts.Bool("debug"))
}

func TestOptions_Bool_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	AssertFalse(t, opts.Bool("name"))
	AssertFalse(t, opts.Bool("missing"))
}

// --- Options.Items ---

func TestOptions_Items_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "a", Value: 1}, Option{Key: "b", Value: 2})
	items := opts.Items()
	AssertLen(t, items, 2)
}

// --- Options with typed struct ---

func TestOptions_TypedStruct_Good(t *testing.T) {
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

func TestOptions_Result_New_Good(t *testing.T) {
	r := Result{}.New("value")
	AssertEqual(t, "value", r.Value)
}

func TestOptions_Result_New_Error_Bad(t *testing.T) {
	err := E("test", "failed", nil)
	r := Result{}.New(err)
	AssertFalse(t, r.OK)
	AssertEqual(t, err, r.Value)
}

func TestOptions_Result_Result_Good(t *testing.T) {
	r := Result{Value: "hello", OK: true}
	AssertEqual(t, r, r.Result())
}

func TestOptions_Result_Result_WithArgs_Good(t *testing.T) {
	r := Result{}.Result("value")
	AssertEqual(t, "value", r.Value)
}

func TestOptions_Result_Get_Good(t *testing.T) {
	r := Result{Value: "hello", OK: true}
	AssertTrue(t, r.Get().OK)
}

func TestOptions_Result_Get_Bad(t *testing.T) {
	r := Result{Value: "err", OK: false}
	AssertFalse(t, r.Get().OK)
}

// --- WithOption ---

func TestOptions_WithOption_Good(t *testing.T) {
	c := New(
		WithOption("name", "myapp"),
		WithOption("port", 8080),
	)
	AssertEqual(t, "myapp", c.App().Name)
	AssertEqual(t, 8080, c.Options().Int("port"))
}
