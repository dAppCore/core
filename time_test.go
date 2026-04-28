package core_test

import (
	. "dappco.re/go"
)

func TestTime_Now_Good(t *T) {
	before := Now()
	value := Now()
	after := Now()

	AssertFalse(t, value.Before(before))
	AssertFalse(t, value.After(after))
}

func TestTime_Now_Bad(t *T) {
	AssertFalse(t, Now().IsZero())
}

func TestTime_Now_Ugly(t *T) {
	first := Now()
	second := Now()

	AssertFalse(t, second.Before(first))
}

func TestTime_ParseDuration_Good(t *T) {
	r := ParseDuration("250ms")

	AssertTrue(t, r.OK)
	AssertEqual(t, 250*Millisecond, r.Value.(Duration))
}

func TestTime_ParseDuration_Bad(t *T) {
	r := ParseDuration("not-a-duration")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTime_ParseDuration_Ugly(t *T) {
	r := ParseDuration("-1h30m")

	AssertTrue(t, r.OK)
	AssertEqual(t, -90*Minute, r.Value.(Duration))
}

func TestTime_Since_Good(t *T) {
	start := Now().Add(-Second)

	AssertGreaterOrEqual(t, Since(start), Second)
}

func TestTime_Since_Bad(t *T) {
	future := Now().Add(Second)

	AssertLess(t, Since(future), Duration(0))
}

func TestTime_Since_Ugly(t *T) {
	start := Now()
	Sleep(Millisecond)

	AssertGreater(t, Since(start), Duration(0))
}

func TestTime_Sleep_Good(t *T) {
	start := Now()
	Sleep(Millisecond)

	AssertGreaterOrEqual(t, Since(start), Millisecond)
}

func TestTime_Sleep_Bad(t *T) {
	start := Now()
	Sleep(-Millisecond)

	AssertLess(t, Since(start), 50*Millisecond)
}

func TestTime_Sleep_Ugly(t *T) {
	start := Now()
	Sleep(0)

	AssertLess(t, Since(start), 50*Millisecond)
}

func TestTime_TimeFormat_Good(t *T) {
	r := TimeParse(TimeRFC3339, "2026-04-28T07:00:00Z")
	RequireTrue(t, r.OK)

	AssertEqual(t, "2026-04-28T07:00:00Z", TimeFormat(r.Value.(Time), TimeRFC3339))
}

func TestTime_TimeFormat_Bad(t *T) {
	AssertEqual(t, "agent", TimeFormat(UnixTime(0), "agent"))
}

func TestTime_TimeFormat_Ugly(t *T) {
	AssertEqual(t, "1970-01-01", TimeFormat(UnixTime(0), TimeDateOnly))
}

func TestTime_TimeParse_Good(t *T) {
	r := TimeParse(TimeRFC3339, "2026-04-28T07:00:00Z")

	AssertTrue(t, r.OK)
	AssertEqual(t, int64(1777359600), r.Value.(Time).Unix())
}

func TestTime_TimeParse_Bad(t *T) {
	r := TimeParse(TimeRFC3339, "not-a-time")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTime_TimeParse_Ugly(t *T) {
	r := TimeParse(TimeDateOnly, "2026-04-28")

	AssertTrue(t, r.OK)
	AssertEqual(t, "2026-04-28", TimeFormat(r.Value.(Time), TimeDateOnly))
}

func TestTime_Until_Good(t *T) {
	future := Now().Add(Second)

	AssertGreater(t, Until(future), Duration(0))
}

func TestTime_Until_Bad(t *T) {
	past := Now().Add(-Second)

	AssertLess(t, Until(past), Duration(0))
}

func TestTime_Until_Ugly(t *T) {
	deadline := Now().Add(Millisecond)
	Sleep(2 * Millisecond)

	AssertLessOrEqual(t, Until(deadline), Duration(0))
}

func TestTime_UnixNow_Good(t *T) {
	before := Now().Unix()
	value := UnixNow()
	after := Now().Unix()

	AssertGreaterOrEqual(t, value, before)
	AssertLessOrEqual(t, value, after)
}

func TestTime_UnixNow_Bad(t *T) {
	AssertGreater(t, UnixNow(), int64(0))
}

func TestTime_UnixNow_Ugly(t *T) {
	AssertLessOrEqual(t, UnixNow()-Now().Unix(), int64(1))
}

func TestTime_UnixTime_Good(t *T) {
	AssertEqual(t, int64(1714291200), UnixTime(1714291200).Unix())
}

func TestTime_UnixTime_Bad(t *T) {
	AssertEqual(t, int64(-1), UnixTime(-1).Unix())
}

func TestTime_UnixTime_Ugly(t *T) {
	AssertEqual(t, "1970-01-01", TimeFormat(UnixTime(0), TimeDateOnly))
}
