package core_test

import . "dappco.re/go"

// ExampleNow reads the current time through `Now` for health-check timing. Durations,
// parsing, and timestamps use core time wrappers for service code.
func ExampleNow() {
	Println(!Now().IsZero())
	// Output: true
}

// ExampleUnixNow reads the current Unix timestamp through `UnixNow` for health-check
// timing. Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleUnixNow() {
	Println(UnixNow() > 0)
	// Output: true
}

// ExampleSleep pauses execution through `Sleep` for health-check timing. Durations,
// parsing, and timestamps use core time wrappers for service code.
func ExampleSleep() {
	Sleep(0)
	Println("awake")
	// Output: awake
}

// ExampleSince measures elapsed time through `Since` for health-check timing. Durations,
// parsing, and timestamps use core time wrappers for service code.
func ExampleSince() {
	Println(Since(UnixTime(0)) > 0)
	// Output: true
}

// ExampleUntil measures time until a deadline through `Until` for health-check timing.
// Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleUntil() {
	Println(Until(UnixTime(32503680000)) > 0)
	// Output: true
}

// ExampleParseDuration parses duration text through `ParseDuration` for health-check
// timing. Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleParseDuration() {
	r := ParseDuration("250ms")
	Println(r.Value)
	// Output: 250ms
}

// ExampleTimeFormat formats a timestamp through `TimeFormat` for health-check timing.
// Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleTimeFormat() {
	t := UnixTime(1714262400)
	Println(TimeFormat(t, TimeDateOnly))
	// Output: 2024-04-28
}

// ExampleTimeParse parses a timestamp through `TimeParse` for health-check timing.
// Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleTimeParse() {
	r := TimeParse(TimeRFC3339, "2026-04-28T07:00:00Z")
	Println(Contains(Sprint(r.Value), "2026-04-28 07:00:00"))
	// Output: true
}

// ExampleUnixTime builds a timestamp from seconds through `UnixTime` for health-check
// timing. Durations, parsing, and timestamps use core time wrappers for service code.
func ExampleUnixTime() {
	Println(Contains(Sprint(UnixTime(0)), "1970-01-01"))
	// Output: true
}
