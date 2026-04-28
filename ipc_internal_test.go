// SPDX-License-Identifier: EUPL-1.2

package core

func TestIpc_Core_broadcast_Good(t *T) {
	c := New()
	called := 0
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})

	r := c.broadcast(ActionTaskStarted{TaskIdentifier: "task-1117"})

	AssertTrue(t, r.OK)
	AssertEqual(t, 2, called)
}
func TestIpc_Core_broadcast_Bad(t *T) {
	r := New().broadcast(ActionTaskStarted{})

	AssertTrue(t, r.OK)
}
func TestIpc_Core_broadcast_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: NewBuffer()}))

	c := New()
	called := 0
	c.RegisterAction(func(_ *Core, _ Message) Result {
		panic("handler refused")
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})

	r := c.broadcast(ActionTaskStarted{TaskIdentifier: "task-1117"})

	AssertTrue(t, r.OK)
	AssertEqual(t, 1, called)
}
