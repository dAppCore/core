// SPDX-License-Identifier: EUPL-1.2

// Formatting and printing primitives — re-exports of Go's fmt package
// as Core helpers, so consumers never need to write `import "fmt"`.
// SPOR file for fmt: every Sprintf/Sprint/Println/Print in core/go and
// downstream packages routes through here.
//
//	core.Println("agent", agent.Name, "ready")
//	msg := core.Sprintf("%s connected on %d", host, port)
package core

import "fmt"

// Sprint converts variadic values to their default string
// representation. Identical to fmt.Sprint.
//
//	core.Sprint(42)       // "42"
//	core.Sprint(err)      // "connection refused"
func Sprint(args ...any) string {
	return fmt.Sprint(args...)
}

// Sprintf formats according to format and returns the resulting string.
// Identical to fmt.Sprintf.
//
//	core.Sprintf("%v=%q", "key", "value")  // `key="value"`
//	core.Sprintf("[%s] %d agents online", region, count)
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Sprintln formats values with default formatting plus a trailing
// newline.
//
//	line := core.Sprintln("agent", "ready")
func Sprintln(args ...any) string {
	return fmt.Sprintln(args...)
}

// Println prints values to stdout separated by spaces with a trailing
// newline. The canonical replacement for fmt.Println.
//
//	core.Println("hello", 42, true)
func Println(args ...any) {
	fmt.Println(args...)
}

// Print writes a formatted line to a writer (defaulting to stdout when
// w is nil), appending a newline. The canonical formatted-output helper.
//
//	core.Print(nil, "hello %s", "world")    // → stdout
//	core.Print(buf, "port: %d", 8080)       // → buf
func Print(w Writer, format string, args ...any) {
	if w == nil {
		w = Stdout()
	}
	fmt.Fprintf(w, format+"\n", args...)
}

// Errorf returns an error formatted per fmt.Errorf semantics. Use
// core.E or core.NewCode for structured errors; reach for Errorf only
// when interoperating with code that expects an %w-style chain.
//
//	err := core.Errorf("connect %s: %w", host, cause)
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
