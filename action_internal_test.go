// SPDX-License-Identifier: EUPL-1.2

package core

func TestAction_Action_safeName_Good(t *T) {
	action := &Action{Name: "agent.dispatch"}

	AssertEqual(t, "agent.dispatch", action.safeName())
}
func TestAction_Action_safeName_Bad(t *T) {
	var action *Action

	AssertEqual(t, "<nil>", action.safeName())
}
func TestAction_Action_safeName_Ugly(t *T) {
	action := &Action{}

	AssertEqual(t, "", action.safeName())
}
func TestAction_Task_safeName_Good(t *T) {
	task := &Task{Name: "deploy.to.homelab"}

	AssertEqual(t, "deploy.to.homelab", task.safeName())
}
func TestAction_Task_safeName_Bad(t *T) {
	var task *Task

	AssertEqual(t, "<nil>", task.safeName())
}
func TestAction_Task_safeName_Ugly(t *T) {
	task := &Task{}

	AssertEqual(t, "", task.safeName())
}
func TestAction_stepOptions_Good(t *T) {
	opts := stepOptions(Step{With: NewOptions(Option{Key: "site", Value: "homelab"})})

	AssertEqual(t, "homelab", opts.String("site"))
}
func TestAction_stepOptions_Bad(t *T) {
	opts := stepOptions(Step{})

	AssertEqual(t, 0, opts.Len())
}
func TestAction_stepOptions_Ugly(t *T) {
	opts := stepOptions(Step{With: NewOptions(
		Option{Key: "agent", Value: "codex"},
		Option{Key: "retry", Value: 3},
	)})

	AssertEqual(t, "codex", opts.String("agent"))
	AssertEqual(t, 3, opts.Int("retry"))
}
