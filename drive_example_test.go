package core_test

import (
	. "dappco.re/go"
)

// ExampleDriveHandle defines drive handle metadata through `DriveHandle` for remote drive
// metadata. Drive handles carry names and transports before remote API calls use them.
func ExampleDriveHandle() {
	handle := DriveHandle{Name: "forge", Transport: "https://forge.example"}
	Println(handle.Name)
	Println(handle.Transport)
	// Output:
	// forge
	// https://forge.example
}

// ExampleDrive reaches drive registration through `Drive` for remote drive metadata. Drive
// handles carry names and transports before remote API calls use them.
func ExampleDrive() {
	d := &Drive{Registry: NewRegistry[*DriveHandle]()}
	d.New(NewOptions(Option{Key: "name", Value: "forge"}))
	Println(d.Names())
	// Output: [forge]
}

// ExampleDrive_New registers a remote drive handle from name and transport options. Drive
// handles carry names and transports before remote API calls use them.
func ExampleDrive_New() {
	c := New()
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "forge"},
		Option{Key: "transport", Value: "https://forge.lthn.ai"},
	))

	Println(c.Drive().Has("forge"))
	Println(c.Drive().Names())
	// Output:
	// true
	// [forge]
}

// ExampleDrive_Get retrieves a value through `Drive.Get` for remote drive metadata. Drive
// handles carry names and transports before remote API calls use them.
func ExampleDrive_Get() {
	c := New()
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101"},
	))

	r := c.Drive().Get("charon")
	if r.OK {
		h := r.Value.(*DriveHandle)
		Println(h.Transport)
	}
	// Output: http://10.69.69.165:9101
}
