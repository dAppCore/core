package core_test

import (
	"testing"
	"time"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestTime_Now_Good(t *testing.T) {
	before := time.Now()
	value := Now()
	after := time.Now()

	assert.False(t, value.Before(before))
	assert.False(t, value.After(after))
}

func TestTime_Now_Bad(t *testing.T) {
	assert.False(t, Now().IsZero())
}

func TestTime_Now_Ugly(t *testing.T) {
	first := Now()
	second := Now()

	assert.False(t, second.Before(first))
}

func TestTime_UnixNow_Good(t *testing.T) {
	before := time.Now().Unix()
	value := UnixNow()
	after := time.Now().Unix()

	assert.GreaterOrEqual(t, value, before)
	assert.LessOrEqual(t, value, after)
}

func TestTime_UnixNow_Bad(t *testing.T) {
	assert.Greater(t, UnixNow(), int64(0))
}

func TestTime_UnixNow_Ugly(t *testing.T) {
	assert.LessOrEqual(t, UnixNow()-time.Now().Unix(), int64(1))
}

func TestTime_Sleep_Good(t *testing.T) {
	start := time.Now()
	Sleep(time.Millisecond)

	assert.GreaterOrEqual(t, time.Since(start), time.Millisecond)
}

func TestTime_Sleep_Bad(t *testing.T) {
	start := time.Now()
	Sleep(-time.Millisecond)

	assert.Less(t, time.Since(start), 50*time.Millisecond)
}

func TestTime_Sleep_Ugly(t *testing.T) {
	start := time.Now()
	Sleep(0)

	assert.Less(t, time.Since(start), 50*time.Millisecond)
}

func TestTime_Since_Good(t *testing.T) {
	start := time.Now().Add(-time.Second)

	assert.GreaterOrEqual(t, Since(start), time.Second)
}

func TestTime_Since_Bad(t *testing.T) {
	future := time.Now().Add(time.Second)

	assert.Less(t, Since(future), time.Duration(0))
}

func TestTime_Since_Ugly(t *testing.T) {
	start := time.Now()
	Sleep(time.Millisecond)

	assert.Greater(t, Since(start), time.Duration(0))
}

func TestTime_Until_Good(t *testing.T) {
	future := time.Now().Add(time.Second)

	assert.Greater(t, Until(future), time.Duration(0))
}

func TestTime_Until_Bad(t *testing.T) {
	past := time.Now().Add(-time.Second)

	assert.Less(t, Until(past), time.Duration(0))
}

func TestTime_Until_Ugly(t *testing.T) {
	deadline := time.Now().Add(time.Millisecond)
	Sleep(2 * time.Millisecond)

	assert.LessOrEqual(t, Until(deadline), time.Duration(0))
}

func TestTime_ParseDuration_Good(t *testing.T) {
	r := ParseDuration("250ms")

	assert.True(t, r.OK)
	assert.Equal(t, 250*time.Millisecond, r.Value.(time.Duration))
}

func TestTime_ParseDuration_Bad(t *testing.T) {
	r := ParseDuration("not-a-duration")

	assert.False(t, r.OK)
	assert.Error(t, r.Value.(error))
}

func TestTime_ParseDuration_Ugly(t *testing.T) {
	r := ParseDuration("-1h30m")

	assert.True(t, r.OK)
	assert.Equal(t, -90*time.Minute, r.Value.(time.Duration))
}
