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
