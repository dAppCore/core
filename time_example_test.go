package core_test

import . "dappco.re/go"

func ExampleNow() {
	Println(!Now().IsZero())
	// Output: true
}

func ExampleUnixNow() {
	Println(UnixNow() > 0)
	// Output: true
}

func ExampleSleep() {
	Sleep(0)
	Println("awake")
	// Output: awake
}

func ExampleSince() {
	Println(Since(UnixTime(0)) > 0)
	// Output: true
}

func ExampleUntil() {
	Println(Until(UnixTime(32503680000)) > 0)
	// Output: true
}

func ExampleParseDuration() {
	r := ParseDuration("250ms")
	Println(r.Value)
	// Output: 250ms
}

func ExampleTimeFormat() {
	t := UnixTime(1714262400)
	Println(TimeFormat(t, TimeDateOnly))
	// Output: 2024-04-28
}

func ExampleTimeParse() {
	r := TimeParse(TimeRFC3339, "2026-04-28T07:00:00Z")
	Println(Contains(Sprint(r.Value), "2026-04-28 07:00:00"))
	// Output: true
}

func ExampleUnixTime() {
	Println(Contains(Sprint(UnixTime(0)), "1970-01-01"))
	// Output: true
}
