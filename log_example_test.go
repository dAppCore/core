package core_test

import . "dappco.re/go/core"

func ExampleInfo() {
	Info("server started", "port", 8080)
}

func ExampleWarn() {
	Warn("deprecated", "feature", "old-api")
}

func ExampleSecurity() {
	Security("access denied", "user", "unknown", "action", "admin.nuke")
}
