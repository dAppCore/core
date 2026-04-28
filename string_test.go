package core_test

import (
	. "dappco.re/go"
)

// --- String Operations ---

func TestString_HasPrefix_Good(t *T) {
	AssertTrue(t, HasPrefix("--verbose", "--"))
	AssertTrue(t, HasPrefix("-v", "-"))
	AssertFalse(t, HasPrefix("hello", "-"))
}

func TestString_HasSuffix_Good(t *T) {
	AssertTrue(t, HasSuffix("test.go", ".go"))
	AssertFalse(t, HasSuffix("test.go", ".py"))
}

func TestString_TrimPrefix_Good(t *T) {
	AssertEqual(t, "verbose", TrimPrefix("--verbose", "--"))
	AssertEqual(t, "hello", TrimPrefix("hello", "--"))
}

func TestString_TrimSuffix_Good(t *T) {
	AssertEqual(t, "test", TrimSuffix("test.go", ".go"))
	AssertEqual(t, "test.go", TrimSuffix("test.go", ".py"))
}

func TestString_Contains_Good(t *T) {
	AssertTrue(t, Contains("hello world", "world"))
	AssertFalse(t, Contains("hello world", "mars"))
}

func TestString_Split_Good(t *T) {
	AssertEqual(t, []string{"a", "b", "c"}, Split("a/b/c", "/"))
}

func TestString_SplitN_Good(t *T) {
	AssertEqual(t, []string{"key", "value=extra"}, SplitN("key=value=extra", "=", 2))
}

func TestString_Join_Good(t *T) {
	AssertEqual(t, "a/b/c", Join("/", "a", "b", "c"))
}

func TestString_Replace_Good(t *T) {
	AssertEqual(t, "deploy.to.homelab", Replace("deploy/to/homelab", "/", "."))
}

func TestString_Lower_Good(t *T) {
	AssertEqual(t, "hello", Lower("HELLO"))
}

func TestString_Upper_Good(t *T) {
	AssertEqual(t, "HELLO", Upper("hello"))
}

func TestString_Trim_Good(t *T) {
	AssertEqual(t, "hello", Trim("  hello  "))
}

func TestString_RuneCount_Good(t *T) {
	AssertEqual(t, 5, RuneCount("hello"))
	AssertEqual(t, 1, RuneCount("🔥"))
	AssertEqual(t, 0, RuneCount(""))
}

func TestString_Concat_Good(t *T) {
	AssertEqual(t, "agent.dispatch.ready", Concat("agent.", "dispatch", ".ready"))
}

func TestString_Concat_Bad(t *T) {
	AssertEqual(t, "", Concat())
}

func TestString_Concat_Ugly(t *T) {
	AssertEqual(t, "token=", Concat("token", "=", ""))
}

func TestString_Contains_Bad(t *T) {
	AssertFalse(t, Contains("agent dispatch", "homelab"))
}

func TestString_Contains_Ugly(t *T) {
	AssertTrue(t, Contains("", ""))
}

func TestString_HasPrefix_Bad(t *T) {
	AssertFalse(t, HasPrefix("agent.dispatch", "task."))
}

func TestString_HasPrefix_Ugly(t *T) {
	AssertTrue(t, HasPrefix("agent.dispatch", ""))
}

func TestString_HasSuffix_Bad(t *T) {
	AssertFalse(t, HasSuffix("agent.yaml", ".json"))
}

func TestString_HasSuffix_Ugly(t *T) {
	AssertTrue(t, HasSuffix("agent.yaml", ""))
}

func TestString_Join_Bad(t *T) {
	AssertEqual(t, "", Join("/"))
}

func TestString_Join_Ugly(t *T) {
	AssertEqual(t, "agentdispatchready", Join("", "agent", "dispatch", "ready"))
}

func TestString_Lower_Bad(t *T) {
	AssertEqual(t, "agent-01", Lower("agent-01"))
}

func TestString_Lower_Ugly(t *T) {
	AssertEqual(t, "", Lower(""))
}

func TestString_NewBuilder_Good(t *T) {
	b := NewBuilder()
	n, err := b.WriteString("agent")

	AssertNoError(t, err)
	AssertEqual(t, 5, n)
	AssertEqual(t, "agent", b.String())
}

func TestString_NewBuilder_Bad(t *T) {
	b := NewBuilder()

	AssertEqual(t, "", b.String())
	AssertEqual(t, 0, b.Len())
}

func TestString_NewBuilder_Ugly(t *T) {
	b := NewBuilder()
	_, err := b.WriteString("session")
	RequireNoError(t, err)
	b.Reset()

	AssertEqual(t, "", b.String())
	AssertEqual(t, 0, b.Len())
}

func TestString_NewReader_Good(t *T) {
	reader := NewReader("agent")
	buf := make([]byte, 5)
	n, err := reader.Read(buf)

	AssertNoError(t, err)
	AssertEqual(t, 5, n)
	AssertEqual(t, "agent", string(buf))
}

func TestString_NewReader_Bad(t *T) {
	reader := NewReader("")
	buf := make([]byte, 1)
	n, err := reader.Read(buf)

	AssertEqual(t, 0, n)
	AssertErrorIs(t, err, EOF)
}

func TestString_NewReader_Ugly(t *T) {
	reader := NewReader("agent dispatch")
	offset, err := reader.Seek(6, 0)
	RequireNoError(t, err)
	buf := make([]byte, 8)
	n, err := reader.Read(buf)

	AssertEqual(t, int64(6), offset)
	AssertNoError(t, err)
	AssertEqual(t, 8, n)
	AssertEqual(t, "dispatch", string(buf))
}

func TestString_Replace_Bad(t *T) {
	AssertEqual(t, "agent/dispatch", Replace("agent/dispatch", ".", "/"))
}

func TestString_Replace_Ugly(t *T) {
	AssertEqual(t, ".a.g.e.n.t.", Replace("agent", "", "."))
}

func TestString_RuneCount_Bad(t *T) {
	AssertNotEqual(t, len("agent"), RuneCount("agent dispatch"))
}

func TestString_RuneCount_Ugly(t *T) {
	AssertEqual(t, 2, RuneCount(string([]byte{0xff, 'a'})))
}

func TestString_Split_Bad(t *T) {
	AssertEqual(t, []string{"agent.dispatch"}, Split("agent.dispatch", "/"))
}

func TestString_Split_Ugly(t *T) {
	AssertEqual(t, []string{"a", "b", "c"}, Split("abc", ""))
}

func TestString_SplitN_Bad(t *T) {
	AssertNil(t, SplitN("agent=dispatch", "=", 0))
}

func TestString_SplitN_Ugly(t *T) {
	AssertEqual(t, []string{"agent", "dispatch", "ready"}, SplitN("agent=dispatch=ready", "=", -1))
}

func TestString_Trim_Bad(t *T) {
	AssertEqual(t, "agent", Trim("agent"))
}

func TestString_Trim_Ugly(t *T) {
	AssertEqual(t, "agent", Trim("\n\tagent\r\n"))
}

func TestString_TrimPrefix_Bad(t *T) {
	AssertEqual(t, "agent.dispatch", TrimPrefix("agent.dispatch", "task."))
}

func TestString_TrimPrefix_Ugly(t *T) {
	AssertEqual(t, "agent.dispatch", TrimPrefix("agent.dispatch", ""))
}

func TestString_TrimSuffix_Bad(t *T) {
	AssertEqual(t, "agent.yaml", TrimSuffix("agent.yaml", ".json"))
}

func TestString_TrimSuffix_Ugly(t *T) {
	AssertEqual(t, "agent.yaml", TrimSuffix("agent.yaml", ""))
}

func TestString_Upper_Bad(t *T) {
	AssertEqual(t, "AGENT-01", Upper("AGENT-01"))
}

func TestString_Upper_Ugly(t *T) {
	AssertEqual(t, "", Upper(""))
}
