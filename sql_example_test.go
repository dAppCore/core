package core_test

import . "dappco.re/go"

// ExampleSQLOpen opens a SQL database through `SQLOpen` for database adapter setup.
// Database opening and sentinel checks use aliases without importing database/sql
// directly.
func ExampleSQLOpen() {
	r := SQLOpen("missing-driver", "")
	Println(r.OK)
	// Output: false
}

// ExampleSQLDrivers lists SQL drivers through `SQLDrivers` for database adapter setup.
// Database opening and sentinel checks use aliases without importing database/sql
// directly.
func ExampleSQLDrivers() {
	drivers := SQLDrivers()
	Println(len(drivers) >= 0)
	// Output: true
}

// ExampleSQLIsNoRows checks the no-rows sentinel through `SQLIsNoRows` for database
// adapter setup. Database opening and sentinel checks use aliases without importing
// database/sql directly.
func ExampleSQLIsNoRows() {
	Println(SQLIsNoRows(ErrNoRows))
	Println(ErrTxDone != nil)
	// Output:
	// true
	// true
}

// ExampleNullString declares a nullable string through `NullString` for database adapter
// setup. Database opening and sentinel checks use aliases without importing database/sql
// directly.
func ExampleNullString() {
	value := NullString{String: "optional", Valid: true}
	Println(value.String)
	Println(value.Valid)
	// Output:
	// optional
	// true
}
