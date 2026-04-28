package core_test

import (
	. "dappco.re/go"
)

func ExampleBackground() {
	ctx := Background()
	Println(ctx.Err() == nil)
	// Output: true
}

func ExampleWithTimeout() {
	ctx, cancel := WithTimeout(Background(), 50*Millisecond)
	defer cancel()
	Println(ctx.Err() == nil)
	// Output: true
}

func ExampleWithCancel() {
	ctx, cancel := WithCancel(Background())
	cancel()
	Println(ctx.Err() != nil)
	// Output: true
}
