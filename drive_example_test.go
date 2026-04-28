package core_test

import (
	. "dappco.re/go"
)

func ExampleDriveHandle() {
	handle := DriveHandle{Name: "forge", Transport: "https://forge.example"}
	Println(handle.Name)
	Println(handle.Transport)
	// Output:
	// forge
	// https://forge.example
}

func ExampleDrive() {
	d := &Drive{Registry: NewRegistry[*DriveHandle]()}
	d.New(NewOptions(Option{Key: "name", Value: "forge"}))
	Println(d.Names())
	// Output: [forge]
}

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
