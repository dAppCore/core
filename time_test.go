package core_test

import (
	"testing"
	"time"

	. "dappco.re/go/core"
)

func TestTime_Now_Good(t *testing.T) {
	before := time.Now()
	value := Now()
	after := time.Now()

	AssertFalse(t, value.Before(before))
	AssertFalse(t, value.After(after))
}

func TestTime_Now_Bad(t *testing.T) {
	AssertFalse(t, Now().IsZero())
}

func TestTime_Now_Ugly(t *testing.T) {
	first := Now()
	second := Now()

	AssertFalse(t, second.Before(first))
}

func TestTime_UnixNow_Good(t *testing.T) {
	before := time.Now().Unix()
	value := UnixNow()
	after := time.Now().Unix()

	AssertGreaterOrEqual(t, value, before)
	AssertLessOrEqual(t, value, after)
}

func TestTime_UnixNow_Bad(t *testing.T) {
	AssertGreater(t, UnixNow(), int64(0))
}

func TestTime_UnixNow_Ugly(t *testing.T) {
	AssertLessOrEqual(t, UnixNow()-time.Now().Unix(), int64(1))
}

func TestTime_Sleep_Good(t *testing.T) {
	start := time.Now()
	Sleep(time.Millisecond)

	AssertGreaterOrEqual(t, time.Since(start), time.Millisecond)
}

func TestTime_Sleep_Bad(t *testing.T) {
	start := time.Now()
	Sleep(-time.Millisecond)

	AssertLess(t, time.Since(start), 50*time.Millisecond)
}

func TestTime_Sleep_Ugly(t *testing.T) {
	start := time.Now()
	Sleep(0)

	AssertLess(t, time.Since(start), 50*time.Millisecond)
}

func TestTime_Since_Good(t *testing.T) {
	start := time.Now().Add(-time.Second)

	AssertGreaterOrEqual(t, Since(start), time.Second)
}

func TestTime_Since_Bad(t *testing.T) {
	future := time.Now().Add(time.Second)

	AssertLess(t, Since(future), time.Duration(0))
}

func TestTime_Since_Ugly(t *testing.T) {
	start := time.Now()
	Sleep(time.Millisecond)

	AssertGreater(t, Since(start), time.Duration(0))
}

func TestTime_Until_Good(t *testing.T) {
	future := time.Now().Add(time.Second)

	AssertGreater(t, Until(future), time.Duration(0))
}

func TestTime_Until_Bad(t *testing.T) {
	past := time.Now().Add(-time.Second)

	AssertLess(t, Until(past), time.Duration(0))
}

func TestTime_Until_Ugly(t *testing.T) {
	deadline := time.Now().Add(time.Millisecond)
	Sleep(2 * time.Millisecond)

	AssertLessOrEqual(t, Until(deadline), time.Duration(0))
}

func TestTime_ParseDuration_Good(t *testing.T) {
	r := ParseDuration("250ms")

	AssertTrue(t, r.OK)
	AssertEqual(t, 250*time.Millisecond, r.Value.(time.Duration))
}

func TestTime_ParseDuration_Bad(t *testing.T) {
	r := ParseDuration("not-a-duration")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTime_ParseDuration_Ugly(t *testing.T) {
	r := ParseDuration("-1h30m")

	AssertTrue(t, r.OK)
	AssertEqual(t, -90*time.Minute, r.Value.(time.Duration))
}
