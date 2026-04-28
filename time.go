// SPDX-License-Identifier: EUPL-1.2

// Time helpers for the Core framework.

package core

import "time"

// Now returns the current local time.
//
//	started := core.Now()
func Now() time.Time {
	return time.Now()
}

// UnixNow returns the current Unix timestamp in seconds.
//
//	ts := core.UnixNow()
func UnixNow() int64 {
	return Now().Unix()
}

// Sleep pauses the current goroutine for at least d.
//
//	core.Sleep(50 * time.Millisecond)
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// Since returns the time elapsed since t.
//
//	elapsed := core.Since(started)
func Since(t time.Time) time.Duration {
	return time.Since(t)
}

// Until returns the duration until t.
//
//	wait := core.Until(deadline)
func Until(t time.Time) time.Duration {
	return time.Until(t)
}

// ParseDuration parses a duration string and returns a Result containing
// time.Duration.
//
//	r := core.ParseDuration("250ms")
//	if r.OK { timeout := r.Value.(time.Duration) }
func ParseDuration(s string) Result {
	d, err := time.ParseDuration(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{d, true}
}

// Duration is a time span — alias of time.Duration so consumers can
// pass timeouts and intervals without importing the time package.
//
//	timeout := 5 * core.Second
//	ctx, cancel := core.WithTimeout(core.Background(), timeout)
type Duration = time.Duration

// Time is a moment — alias of time.Time so consumers can take
// timestamps without importing the time package.
//
//	deadline := core.Now().Add(2 * core.Minute)
type Time = time.Time

// Common Duration units. Multiply by an integer to build a Duration.
//
//	timeout := 5 * core.Second
//	pause   := 250 * core.Millisecond
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// Common time format constants. Layouts compatible with time.Format.
//
//	stamp := core.TimeFormat(core.Now(), core.TimeRFC3339)
const (
	TimeRFC3339     = time.RFC3339
	TimeRFC3339Nano = time.RFC3339Nano
	TimeRFC1123     = time.RFC1123
	TimeRFC822      = time.RFC822
	TimeKitchen     = time.Kitchen  // "3:04PM"
	TimeStamp       = time.Stamp    // "Jan _2 15:04:05"
	TimeDateTime    = time.DateTime // "2006-01-02 15:04:05"
	TimeDateOnly    = time.DateOnly // "2006-01-02"
	TimeTimeOnly    = time.TimeOnly // "15:04:05"
)

// TimeFormat formats t as a string using the given layout. Layout
// constants are exported as TimeRFC3339, TimeDateTime, etc.
//
//	s := core.TimeFormat(core.Now(), core.TimeRFC3339)
func TimeFormat(t time.Time, layout string) string {
	return t.Format(layout)
}

// TimeParse parses value into a time.Time using the given layout.
// Returns Result wrapping time.Time on success or the parse error.
//
//	r := core.TimeParse(core.TimeRFC3339, "2026-04-28T07:00:00Z")
//	if r.OK { ts := r.Value.(time.Time) }
func TimeParse(layout, value string) Result {
	t, err := time.Parse(layout, value)
	if err != nil {
		return Result{err, false}
	}
	return Result{t, true}
}

// UnixTime returns the time corresponding to the given Unix timestamp
// (seconds since 1970-01-01 UTC).
//
//	ts := core.UnixTime(1714291200)
func UnixTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}
