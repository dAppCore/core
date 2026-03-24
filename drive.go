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

import (
	"sync"
)

// DriveHandle holds a named transport resource.
type DriveHandle struct {
	Name      string
	Transport string
	Options   Options
}

// Drive manages named transport handles.
type Drive struct {
	handles map[string]*DriveHandle
	mu      sync.RWMutex
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

	transport := opts.String("transport")

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.handles == nil {
		d.handles = make(map[string]*DriveHandle)
	}

	handle := &DriveHandle{
		Name:      name,
		Transport: transport,
		Options:   opts,
	}

	d.handles[name] = handle
	return Result{handle, true}
}

// Get returns a handle by name.
//
//	r := c.Drive().Get("api")
//	if r.OK { handle := r.Value.(*DriveHandle) }
func (d *Drive) Get(name string) Result {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.handles == nil {
		return Result{}
	}
	h, ok := d.handles[name]
	if !ok {
		return Result{}
	}
	return Result{h, true}
}

// Has returns true if a handle is registered.
//
//	if c.Drive().Has("ssh") { ... }
func (d *Drive) Has(name string) bool {
	return d.Get(name).OK
}

// Names returns all registered handle names.
//
//	names := c.Drive().Names()
func (d *Drive) Names() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var names []string
	for k := range d.handles {
		names = append(names, k)
	}
	return names
}
