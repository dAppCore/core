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
//	if core.Is(err, core.ErrNoRows) { /* not found */ }
package core

import "database/sql"

// DB is a connection-pooled handle to a SQL database.
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
type DB = sql.DB

// Tx is an in-progress SQL transaction.
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
//	tx, err := db.Begin()
//	if err == nil { defer tx.Rollback() }
type Tx = sql.Tx

// Stmt is a prepared SQL statement.
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
//	stmt, err := db.Prepare("select name from agents where id = ?")
//	if err == nil { defer stmt.Close() }
type Stmt = sql.Stmt

// Rows is the result of a query — iterate with rs.Next().
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
//	rows, err := db.Query("select name from agents")
//	if err == nil { defer rows.Close() }
type Rows = sql.Rows

// Row is a single-row query result.
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
//	row := db.QueryRow("select name from agents where id = ?", 42)
//	_ = row
type Row = sql.Row

// Result is shadowed by core.Result; the SQL exec result type is
// re-exported as SQLResult to disambiguate.
//
//	r := core.SQLOpen("sqlite3", "file:./data/homelab.db")
//	if !r.OK { return r }
//	db := r.Value.(*core.DB)
//	defer db.Close()
//	res, err := db.Exec("update agents set status = ? where name = ?", "ready", "codex")
//	if err == nil { affected, _ := res.RowsAffected(); core.Println(affected) }
type SQLResult = sql.Result

// NullString is sql.NullString — string column that may be NULL.
//
//	name := core.NullString{String: "codex", Valid: true}
//	if name.Valid { core.Println(name.String) }
type NullString = sql.NullString

// NullInt64 is sql.NullInt64 — int64 column that may be NULL.
//
//	count := core.NullInt64{Int64: 42, Valid: true}
//	if count.Valid { core.Println(count.Int64) }
type NullInt64 = sql.NullInt64

// NullBool is sql.NullBool — bool column that may be NULL.
//
//	enabled := core.NullBool{Bool: true, Valid: true}
//	if enabled.Valid && enabled.Bool { core.Println("enabled") }
type NullBool = sql.NullBool

// NullFloat64 is sql.NullFloat64 — float64 column that may be NULL.
//
//	progress := core.NullFloat64{Float64: 0.75, Valid: true}
//	if progress.Valid { core.Println(progress.Float64) }
type NullFloat64 = sql.NullFloat64

// NullTime is sql.NullTime — time.Time column that may be NULL.
//
//	seen := core.NullTime{Time: core.Now(), Valid: true}
//	if seen.Valid { core.Println(core.TimeFormat(seen.Time, "2006-01-02")) }
type NullTime = sql.NullTime

// ErrNoRows is the canonical "no row matched" sentinel returned by
// QueryRow when the query returns zero rows.
//
//	if Is(err, core.ErrNoRows) { /* handle "not found" */ }
var ErrNoRows = sql.ErrNoRows

// ErrTxDone is returned when a transaction operation runs after
// Commit or Rollback has already been called.
//
//	err := core.ErrTxDone
//	if core.Is(err, core.ErrTxDone) { core.Println("transaction closed") }
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
//
//	drivers := core.SQLDrivers()
//	core.Println(core.Join(", ", drivers...))
func SQLDrivers() []string {
	return sql.Drivers()
}

// SQLIsNoRows is a convenience for the most common error check.
//
//	if core.SQLIsNoRows(err) { /* not found */ }
func SQLIsNoRows(err error) bool {
	return Is(err, sql.ErrNoRows)
}
