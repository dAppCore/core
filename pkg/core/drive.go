// SPDX-License-Identifier: EUPL-1.2

// Drive is the resource handle registry for transport connections.
// Packages register their transport handles (API, MCP, SSH, VPN)
// and other packages access them by name.
//
// Register a transport:
//
//	c.Drive().New(core.Options{
//	    {K: "name", V: "api"},
//	    {K: "transport", V: "https://api.lthn.ai"},
//	})
//	c.Drive().New(core.Options{
//	    {K: "name", V: "ssh"},
//	    {K: "transport", V: "ssh://claude@10.69.69.165"},
//	})
//	c.Drive().New(core.Options{
//	    {K: "name", V: "mcp"},
//	    {K: "transport", V: "mcp://mcp.lthn.sh"},
//	})
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
//	c.Drive().New(core.Options{
//	    {K: "name", V: "api"},
//	    {K: "transport", V: "https://api.lthn.ai"},
//	})
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
	return Result{Value: handle, OK: true}
}

// Get returns a handle by name.
//
//	api := c.Drive().Get("api")
func (d *Drive) Get(name string) *DriveHandle {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.handles == nil {
		return nil
	}
	return d.handles[name]
}

// Has returns true if a handle is registered.
//
//	if c.Drive().Has("ssh") { ... }
func (d *Drive) Has(name string) bool {
	return d.Get(name) != nil
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
