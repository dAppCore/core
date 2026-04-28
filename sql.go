// SPDX-License-Identifier: EUPL-1.2

// SQL database primitive for the Core framework.
//
// Re-exports stdlib database/sql types and provides a Result-shape
// Open. Consumer packages declare DB / Tx / Rows / Stmt parameters
// via core without importing database/sql directly.
//
// Driver registration stays the canonical Go pattern — import the
// driver package for its init-time side effect:
//
//	import _ "github.com/mattn/go-sqlite3"
//	r := core.SQLOpen("sqlite3", "file:./data.db?_pragma=journal_mode(WAL)")
//	if !r.OK { return r }
//	db := r.Value.(*DB)
//	defer db.Close()
//
// ErrNoRows is the canonical sentinel for "no row matched" queries.
//
//	if errors.Is(err, core.ErrNoRows) { /* not found */ }
package core

import (
	"database/sql"
	"errors"
)

// DB is a connection-pooled handle to a SQL database.
type DB = sql.DB

// Tx is an in-progress SQL transaction.
type Tx = sql.Tx

// Stmt is a prepared SQL statement.
type Stmt = sql.Stmt

// Rows is the result of a query — iterate with rs.Next().
type Rows = sql.Rows

// Row is a single-row query result.
type Row = sql.Row

// Result is shadowed by core.Result; the SQL exec result type is
// re-exported as SQLResult to disambiguate.
type SQLResult = sql.Result

// NullString is sql.NullString — string column that may be NULL.
type NullString = sql.NullString

// NullInt64 is sql.NullInt64 — int64 column that may be NULL.
type NullInt64 = sql.NullInt64

// NullBool is sql.NullBool — bool column that may be NULL.
type NullBool = sql.NullBool

// NullFloat64 is sql.NullFloat64 — float64 column that may be NULL.
type NullFloat64 = sql.NullFloat64

// NullTime is sql.NullTime — time.Time column that may be NULL.
type NullTime = sql.NullTime

// ErrNoRows is the canonical "no row matched" sentinel returned by
// QueryRow when the query returns zero rows.
//
//	if errors.Is(err, core.ErrNoRows) { /* handle "not found" */ }
var ErrNoRows = sql.ErrNoRows

// ErrTxDone is returned when a transaction operation runs after
// Commit or Rollback has already been called.
var ErrTxDone = sql.ErrTxDone

// SQLOpen opens a database handle for the given driver and data source
// name. The driver must be imported for its side-effect registration.
//
//	r := core.SQLOpen("sqlite3", "file:./data.db")
//	if !r.OK { return r }
//	db := r.Value.(*DB)
func SQLOpen(driverName, dataSourceName string) Result {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return Result{err, false}
	}
	return Result{db, true}
}

// SQLDrivers returns a sorted list of registered driver names. Useful
// for diagnostics ("which drivers were imported?").
func SQLDrivers() []string {
	return sql.Drivers()
}

// SQLIsNoRows is a convenience for the most common error check.
//
//	if core.SQLIsNoRows(err) { /* not found */ }
func SQLIsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
