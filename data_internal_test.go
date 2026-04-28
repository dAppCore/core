// SPDX-License-Identifier: EUPL-1.2

package core

func TestData_Data_resolve_Good(t *T) {
	c := New()
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, "prompts"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, "prompts", "agent.md"), []byte("ready"), 0o644).OK)
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "."},
	))
	RequireTrue(t, r.OK)

	embed, rel := c.Data().resolve("agent/prompts/agent.md")

	AssertNotNil(t, embed)
	AssertEqual(t, "prompts/agent.md", rel)
}
func TestData_Data_resolve_Bad(t *T) {
	embed, rel := New().Data().resolve("agent")

	AssertNil(t, embed)
	AssertEqual(t, "", rel)
}
func TestData_Data_resolve_Ugly(t *T) {
	embed, rel := New().Data().resolve("missing/agent.md")

	AssertNil(t, embed)
	AssertEqual(t, "", rel)
}
