package core_test

import (
	. "dappco.re/go"
)

func TestUser_UserCurrent_Good(t *T) {
	r := UserCurrent()
	if !r.OK {
		t.Skip("current user lookup unavailable on this host")
	}
	u := r.Value.(*User)

	AssertNotEmpty(t, u.Username)
	AssertNotEmpty(t, u.Uid)
}

func TestUser_UserCurrent_Bad(t *T) {
	var r Result

	AssertNotPanics(t, func() { r = UserCurrent() })
	if !r.OK {
		AssertError(t, r.Value.(error))
	}
}

func TestUser_UserCurrent_Ugly(t *T) {
	u := currentUserForUserTest(t)
	r := UserLookupID(u.Uid)

	AssertTrue(t, r.OK)
	AssertEqual(t, u.Uid, r.Value.(*User).Uid)
}

func TestUser_UserGroupLookup_Good(t *T) {
	name := knownGroupNameForUserTest(t)
	r := UserGroupLookup(name)

	AssertTrue(t, r.OK)
	AssertEqual(t, name, r.Value.(*Group).Name)
}

func TestUser_UserGroupLookup_Bad(t *T) {
	r := UserGroupLookup("dappcore-no-such-group-000000")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "group not found")
}

func TestUser_UserGroupLookup_Ugly(t *T) {
	r := UserGroupLookup("")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestUser_UserLookup_Good(t *T) {
	u := currentUserForUserTest(t)
	r := UserLookup(u.Username)

	AssertTrue(t, r.OK)
	AssertEqual(t, u.Uid, r.Value.(*User).Uid)
}

func TestUser_UserLookup_Bad(t *T) {
	r := UserLookup("dappcore-no-such-user-000000")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "user not found")
}

func TestUser_UserLookup_Ugly(t *T) {
	r := UserLookup("")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestUser_UserLookupID_Good(t *T) {
	u := currentUserForUserTest(t)
	r := UserLookupID(u.Uid)

	AssertTrue(t, r.OK)
	AssertEqual(t, u.Username, r.Value.(*User).Username)
}

func TestUser_UserLookupID_Bad(t *T) {
	r := UserLookupID("-1")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "uid not found")
}

func TestUser_UserLookupID_Ugly(t *T) {
	r := UserLookupID("")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func currentUserForUserTest(t *T) *User {
	t.Helper()
	r := UserCurrent()
	if !r.OK {
		t.Skip("current user lookup unavailable on this host")
	}
	return r.Value.(*User)
}

func knownGroupNameForUserTest(t *T) string {
	t.Helper()
	for _, name := range []string{"staff", "users", "wheel", "root"} {
		if r := UserGroupLookup(name); r.OK {
			return name
		}
	}
	t.Skip("no common group name available on this host")
	return ""
}
