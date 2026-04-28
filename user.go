// SPDX-License-Identifier: EUPL-1.2

// User identity primitives — re-exports of Go's os/user package as
// Core types, so consumers never need to write `import "os/user"`.
// Used by filesystem path resolution (~ expansion) and log identity.
//
//	r := core.UserCurrent()
//	if r.OK { homeDir := r.Value.(*core.User).HomeDir }
package core

import "os/user"

// User represents a system user with username, uid, gid, and home dir.
// Alias of os/user.User.
//
//	u := core.UserCurrent().Value.(*core.User)
//	core.Println(u.Username, u.HomeDir)
type User = user.User

// Group represents a system user group. Alias of os/user.Group.
//
//	g := core.UserGroupLookup("staff").Value.(*core.Group)
//	core.Println(g.Name, g.Gid)
type Group = user.Group

// UserCurrent returns the current process's user wrapped in a Result.
// OK=false with Code "user.lookup.failed" when the lookup fails (e.g.
// in containers without a /etc/passwd entry).
//
//	r := core.UserCurrent()
//	if r.OK { home := r.Value.(*core.User).HomeDir }
func UserCurrent() Result {
	u, err := user.Current()
	if err != nil {
		return Result{Value: WrapCode(err, "user.lookup.failed", "UserCurrent", "current user lookup failed"), OK: false}
	}
	return Result{Value: u, OK: true}
}

// UserLookup finds a user by username. OK=false with Code
// "user.notfound" when the username doesn't exist on the system.
//
//	r := core.UserLookup("snider")
//	if r.OK { uid := r.Value.(*core.User).Uid }
func UserLookup(username string) Result {
	u, err := user.Lookup(username)
	if err != nil {
		return Result{Value: WrapCode(err, "user.notfound", "UserLookup", Concat("user not found: ", username)), OK: false}
	}
	return Result{Value: u, OK: true}
}

// UserLookupID finds a user by uid string. OK=false with Code
// "user.notfound" when no user has that uid.
//
//	r := core.UserLookupID("1000")
//	if r.OK { username := r.Value.(*core.User).Username }
func UserLookupID(uid string) Result {
	u, err := user.LookupId(uid)
	if err != nil {
		return Result{Value: WrapCode(err, "user.notfound", "UserLookupID", Concat("uid not found: ", uid)), OK: false}
	}
	return Result{Value: u, OK: true}
}

// UserGroupLookup finds a group by name. OK=false with Code
// "user.group.notfound" when the group doesn't exist.
//
//	r := core.UserGroupLookup("staff")
//	if r.OK { gid := r.Value.(*core.Group).Gid }
func UserGroupLookup(name string) Result {
	g, err := user.LookupGroup(name)
	if err != nil {
		return Result{Value: WrapCode(err, "user.group.notfound", "UserGroupLookup", Concat("group not found: ", name)), OK: false}
	}
	return Result{Value: g, OK: true}
}
