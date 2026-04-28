package core_test

import (
	. "dappco.re/go"
)

// ExampleBackground creates a background context through `Background` for request lifetime
// control. Cancellation and timeout lifetimes are created through the core context
// surface.
func ExampleBackground() {
	ctx := Background()
	Println(ctx.Err() == nil)
	// Output: true
}

// ExampleWithTimeout creates a timeout context through `WithTimeout` for request lifetime
// control. Cancellation and timeout lifetimes are created through the core context
// surface.
func ExampleWithTimeout() {
	ctx, cancel := WithTimeout(Background(), 50*Millisecond)
	defer cancel()
	Println(ctx.Err() == nil)
	// Output: true
}

// ExampleWithCancel creates a cancellable context through `WithCancel` for request
// lifetime control. Cancellation and timeout lifetimes are created through the core
// context surface.
func ExampleWithCancel() {
	ctx, cancel := WithCancel(Background())
	cancel()
	Println(ctx.Err() != nil)
	// Output: true
}
