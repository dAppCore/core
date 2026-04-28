package core_test

import (
	. "dappco.re/go"
)

// --- Drive (Transport Handles) ---

func TestDrive_New_Good(t *T) {
	c := New()
	r := c.Drive().New(NewOptions(
		Option{Key: "name", Value: "api"},
		Option{Key: "transport", Value: "https://api.lthn.ai"},
	))
	AssertTrue(t, r.OK)
	AssertEqual(t, "api", r.Value.(*DriveHandle).Name)
	AssertEqual(t, "https://api.lthn.ai", r.Value.(*DriveHandle).Transport)
}

func TestDrive_New_Bad(t *T) {
	c := New()
	// Missing name
	r := c.Drive().New(NewOptions(
		Option{Key: "transport", Value: "https://api.lthn.ai"},
	))
	AssertFalse(t, r.OK)
}

func TestDrive_Get_Good(t *T) {
	c := New()
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "ssh"},
		Option{Key: "transport", Value: "ssh://claude@10.69.69.165"},
	))
	r := c.Drive().Get("ssh")
	AssertTrue(t, r.OK)
	handle := r.Value.(*DriveHandle)
	AssertEqual(t, "ssh://claude@10.69.69.165", handle.Transport)
}

func TestDrive_Get_Bad(t *T) {
	c := New()
	r := c.Drive().Get("nonexistent")
	AssertFalse(t, r.OK)
}

func TestDrive_Has_Good(t *T) {
	c := New()
	c.Drive().New(NewOptions(Option{Key: "name", Value: "mcp"}, Option{Key: "transport", Value: "mcp://mcp.lthn.sh"}))
	AssertTrue(t, c.Drive().Has("mcp"))
	AssertFalse(t, c.Drive().Has("missing"))
}

func TestDrive_Names_Good(t *T) {
	c := New()
	c.Drive().New(NewOptions(Option{Key: "name", Value: "api"}, Option{Key: "transport", Value: "https://api.lthn.ai"}))
	c.Drive().New(NewOptions(Option{Key: "name", Value: "ssh"}, Option{Key: "transport", Value: "ssh://claude@10.69.69.165"}))
	c.Drive().New(NewOptions(Option{Key: "name", Value: "mcp"}, Option{Key: "transport", Value: "mcp://mcp.lthn.sh"}))
	names := c.Drive().Names()
	AssertLen(t, names, 3)
	AssertContains(t, names, "api")
	AssertContains(t, names, "ssh")
	AssertContains(t, names, "mcp")
}

func TestDrive_OptionsPreserved_Good(t *T) {
	c := New()
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "api"},
		Option{Key: "transport", Value: "https://api.lthn.ai"},
		Option{Key: "timeout", Value: 30},
	))
	r := c.Drive().Get("api")
	AssertTrue(t, r.OK)
	handle := r.Value.(*DriveHandle)
	AssertEqual(t, 30, handle.Options.Int("timeout"))
}
