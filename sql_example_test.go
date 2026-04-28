package core_test

import . "dappco.re/go"

func ExampleSQLOpen() {
	r := SQLOpen("missing-driver", "")
	Println(r.OK)
	// Output: false
}

func ExampleSQLDrivers() {
	drivers := SQLDrivers()
	Println(len(drivers) >= 0)
	// Output: true
}

func ExampleSQLIsNoRows() {
	Println(SQLIsNoRows(ErrNoRows))
	Println(ErrTxDone != nil)
	// Output:
	// true
	// true
}

func ExampleNullString() {
	value := NullString{String: "optional", Valid: true}
	Println(value.String)
	Println(value.Valid)
	// Output:
	// optional
	// true
}
