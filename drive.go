// SPDX-License-Identifier: EUPL-1.2

// Drive is the resource handle registry for transport connections.
// Packages register their transport handles (API, MCP, SSH, VPN)
// and other packages access them by name.
//
// Register a transport:
//
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "api"},
//	    core.Option{Key: "transport", Value: "https://api.lthn.ai"},
//	))
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "ssh"},
//	    core.Option{Key: "transport", Value: "ssh://claude@10.69.69.165"},
//	))
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "mcp"},
//	    core.Option{Key: "transport", Value: "mcp://mcp.lthn.sh"},
//	))
//
// Retrieve a handle:
//
//	api := c.Drive().Get("api")
package core

// DriveHandle holds a named transport resource.
//
//	handle := &core.DriveHandle{Name: "homelab", Transport: "ssh://agent@10.69.69.165"}
//	core.Println(handle.Transport)
type DriveHandle struct {
	Name      string
	Transport string
	Options   Options
}

// Drive manages named transport handles. Embeds Registry[*DriveHandle].
//
//	c := core.New()
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "homelab"},
//	    core.Option{Key: "transport", Value: "ssh://agent@10.69.69.165"},
//	))
type Drive struct {
	*Registry[*DriveHandle]
}

// New registers a transport handle.
//
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "api"},
//	    core.Option{Key: "transport", Value: "https://api.lthn.ai"},
//	))
func (d *Drive) New(opts Options) Result {
	name := opts.String("name")
	if name == "" {
		return Result{}
	}

	handle := &DriveHandle{
		Name:      name,
		Transport: opts.String("transport"),
		Options:   opts,
	}

	d.Set(name, handle)
	return Result{handle, true}
}
