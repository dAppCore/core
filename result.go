// SPDX-License-Identifier: EUPL-1.2

// Result ergonomics — methods and free functions that collapse the
// common Result-handling patterns into one-liners. Result itself is
// defined in options.go alongside Options as a Core primitive; this
// file extends it with the call-site helpers.
//
//	user := core.MustCast[*User](c.Drive().Get(opts))
//	timeout := core.HTTPGet(url).Or(defaultResp)
package core

// Ok wraps v in a successful Result. The canonical "happy path"
// constructor — replaces the awkward `Result{v, true}` literal.
//
//	return core.Ok(parsed)
func Ok(v any) Result {
	return Result{Value: v, OK: true}
}

// Fail wraps err in a failed Result. The canonical "sad path"
// constructor — pair Result.Code() / Result.Error() to introspect.
// (Named Fail rather than Err to avoid colliding with the *Err type
// in error.go.)
//
//	if err := decode(b); err != nil { return core.Fail(err) }
func Fail(err error) Result {
	return Result{Value: err, OK: false}
}

// ResultOf adapts a stdlib (value, error) pair into a Result —
// OK=false carrying err when err != nil, OK=true carrying v otherwise.
// The replacement for Result{}.New(v, err).
//
//	r := core.ResultOf(os.ReadFile(path))
//	if !r.OK { return r }
func ResultOf(v any, err error) Result {
	if err != nil {
		return Result{Value: err, OK: false}
	}
	return Result{Value: v, OK: true}
}

// Error returns the error message when the Result represents a failure,
// or "" when OK. Convenience for logging without unwrapping Value.
//
//	if !r.OK { core.Error("dispatch failed", "err", r.Error()) }
func (r Result) Error() string {
	if r.OK {
		return ""
	}
	if err, ok := r.Value.(error); ok {
		return err.Error()
	}
	if s, ok := r.Value.(string); ok {
		return s
	}
	return "unknown error"
}

// Code returns the stable error code from the Result's failure, or ""
// when OK or when the failure isn't a *core.Err with a Code populated.
// Codes form a flat keyspace agents grep on (e.g. "fs.notfound",
// "json.invalid", "http.timeout", "crypto.algo.unsupported"). See the
// Stable codespace section in AGENTS.md for the canonical list.
//
//	r := c.Fs().Read("/missing")
//	if r.Code() == "fs.notfound" { core.Println("first run") }
//	switch r.Code() {
//	case "http.timeout": retry()
//	case "http.refused": fallback()
//	}
func (r Result) Code() string {
	if r.OK {
		return ""
	}
	if e, ok := r.Value.(*Err); ok {
		return e.Code
	}
	return ""
}

// Must returns Value when OK; panics with the underlying error when
// not. Use for fast-fail paths — init, test setup, must-have config.
// Production request paths should check r.OK and return r.
//
//	cfg := core.JSONUnmarshal(data, &Config{}).Must().(*Config)
//	dir := core.PathAbs(".").Must().(string)
func (r Result) Must() any {
	if !r.OK {
		if err, ok := r.Value.(error); ok {
			panic(err)
		}
		panic(r.Value)
	}
	return r.Value
}

// Or returns Value when OK, fallback otherwise. Convenience for
// optional reads where a default is acceptable.
//
//	port := core.EnvGet("PORT").Or("8080").(string)
//	timeout := core.ParseDuration(s).Or(5 * core.Second).(Duration)
func (r Result) Or(fallback any) any {
	if r.OK {
		return r.Value
	}
	return fallback
}

// Cast extracts a typed value from a Result. Returns (zero, false) when
// the Result is not OK or Value isn't assignable to T. Single
// expression replacing the (Result.OK check + type assertion) pair.
//
//	cfg, ok := core.Cast[*Config](core.JSONUnmarshal(data, &Config{}))
//	if !ok { return r }
//	if user, ok := core.Cast[*User](svc.Get(id)); ok { use(user) }
func Cast[T any](r Result) (T, bool) {
	var zero T
	if !r.OK {
		return zero, false
	}
	v, ok := r.Value.(T)
	if !ok {
		return zero, false
	}
	return v, true
}

// MustCast extracts a typed value from a Result. Panics with the
// underlying error when the Result is not OK or when Value isn't
// assignable to T. Use for fast-fail paths — init, test setup,
// must-have config — where the type-assertion is part of the contract.
//
//	cfg := core.MustCast[*Config](core.JSONUnmarshal(data, &Config{}))
//	dir := core.MustCast[string](core.PathAbs("."))
func MustCast[T any](r Result) T {
	var zero T
	if !r.OK {
		if err, ok := r.Value.(error); ok {
			panic(err)
		}
		panic(r.Value)
	}
	v, ok := r.Value.(T)
	if !ok {
		panic(E("MustCast", Sprintf("Value is %T, not %T", r.Value, zero), nil))
	}
	return v
}

// Try runs fn and converts its outcome into a Result. A nil error or
// a returned value sets OK=true; a returned error or a panic sets
// OK=false. Bridges legacy code that panics or returns (T, error).
//
//	r := core.Try(func() any {
//	    return riskyParse(input)  // may panic
//	})
//	if !r.OK { core.Error("parse failed", "err", r.Error()) }
func Try(fn func() any) (r Result) {
	defer func() {
		if rec := recover(); rec != nil {
			if err, ok := rec.(error); ok {
				r = Result{Value: err, OK: false}
				return
			}
			r = Result{Value: E("Try", "panic recovered", nil), OK: false}
		}
	}()
	v := fn()
	if err, ok := v.(error); ok && err != nil {
		return Result{Value: err, OK: false}
	}
	return Result{Value: v, OK: true}
}
