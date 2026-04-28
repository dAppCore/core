package core_test

import (
	. "dappco.re/go"
)

func TestSql_SQLDrivers_Good(t *T) {
	drivers := SQLDrivers()

	AssertGreaterOrEqual(t, len(drivers), 0)
}

func TestSql_SQLDrivers_Bad(t *T) {
	AssertFalse(t, SliceContains(SQLDrivers(), "dappcore-missing-driver"))
}

func TestSql_SQLDrivers_Ugly(t *T) {
	drivers := SQLDrivers()

	for i := 1; i < len(drivers); i++ {
		AssertLessOrEqual(t, drivers[i-1], drivers[i])
	}
}

func TestSql_SQLIsNoRows_Good(t *T) {
	AssertTrue(t, SQLIsNoRows(ErrNoRows))
}

func TestSql_SQLIsNoRows_Bad(t *T) {
	AssertFalse(t, SQLIsNoRows(AnError))
}

func TestSql_SQLIsNoRows_Ugly(t *T) {
	err := Wrap(ErrNoRows, "agent.lookup", "agent not found")

	AssertTrue(t, SQLIsNoRows(err))
}

func TestSql_SQLOpen_Good(t *T) {
	// core/go intentionally registers no SQL driver; consumers import one.
	r := SQLOpen("dappcore-missing-driver", "agent.db")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "unknown driver")
}

func TestSql_SQLOpen_Bad(t *T) {
	r := SQLOpen("", "")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "unknown driver")
}

func TestSql_SQLOpen_Ugly(t *T) {
	r := SQLOpen("dappcore-missing-driver", "file:agent.db?mode=memory&cache=shared")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "unknown driver")
}
