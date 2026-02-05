package handlers

import "os/exec"

// execCommand is a package-level variable for creating exec.Cmd instances.
// It defaults to exec.CommandContext and can be replaced in tests for
// mocking shell commands.
var execCommand = exec.CommandContext
