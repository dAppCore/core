package core_test

import . "dappco.re/go"

var testFS EmbedFS = EmbeddedTestFS

// --- Data (Embedded Content Mounts) ---

func mountTestData(t *T, c *Core, name string) {
	t.Helper()

	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: name},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "tests/data"},
	))
	AssertTrue(t, r.OK)
}

func TestData_New_Good(t *T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "test"},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "tests/data"},
	))
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestData_New_Bad(t *T) {
	c := New()

	r := c.Data().New(NewOptions(Option{Key: "source", Value: testFS}))
	AssertFalse(t, r.OK)

	r = c.Data().New(NewOptions(Option{Key: "name", Value: "test"}))
	AssertFalse(t, r.OK)

	r = c.Data().New(NewOptions(Option{Key: "name", Value: "test"}, Option{Key: "source", Value: "not-an-fs"}))
	AssertFalse(t, r.OK)
}

func TestData_ReadString_Good(t *T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ReadString("app/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestData_ReadString_Bad(t *T) {
	c := New()
	r := c.Data().ReadString("nonexistent/file.txt")
	AssertFalse(t, r.OK)
}

func TestData_ReadFile_Good(t *T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ReadFile("app/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestData_Get_Good(t *T) {
	c := New()
	mountTestData(t, c, "brain")
	gr := c.Data().Get("brain")
	AssertTrue(t, gr.OK)
	emb := gr.Value.(*Embed)

	r := emb.Open("test.txt")
	AssertTrue(t, r.OK)
	cr := ReadAll(r.Value)
	AssertTrue(t, cr.OK)
	AssertEqual(t, "hello from testdata\n", cr.Value)
}

func TestData_Get_Bad(t *T) {
	c := New()
	r := c.Data().Get("nonexistent")
	AssertFalse(t, r.OK)
}

func TestData_Mounts_Good(t *T) {
	c := New()
	mountTestData(t, c, "a")
	mountTestData(t, c, "b")
	mounts := c.Data().Mounts()
	AssertLen(t, mounts, 2)
}

func TestData_List_Good(t *T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().List("app/.")
	AssertTrue(t, r.OK)
}

func TestData_List_Bad(t *T) {
	c := New()
	r := c.Data().List("nonexistent/path")
	AssertFalse(t, r.OK)
}

func TestData_ListNames_Good(t *T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ListNames("app/.")
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.([]string), "test")
}

func TestData_Extract_Good(t *T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().Extract("app/.", t.TempDir(), nil)
	AssertTrue(t, r.OK)
}

func TestData_Extract_Bad(t *T) {
	c := New()
	r := c.Data().Extract("nonexistent/path", t.TempDir(), nil)
	AssertFalse(t, r.OK)
}

// --- AX-7 canonical triplets ---

func TestData_Data_New_Good(t *T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "tests/data"},
	))
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value.(*Embed))
}

func TestData_Data_New_Bad(t *T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: "not-a-filesystem"},
	))
	AssertFalse(t, r.OK)
}

func TestData_Data_New_Ugly(t *T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "root"},
		Option{Key: "source", Value: testFS},
	))
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(*Embed).BaseDirectory())
}

func TestData_Data_ReadFile_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ReadFile("agent/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestData_Data_ReadFile_Bad(t *T) {
	c := New()
	r := c.Data().ReadFile("agent/test.txt")
	AssertFalse(t, r.OK)
}

func TestData_Data_ReadFile_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ReadFile("agent")
	AssertFalse(t, r.OK)
}

func TestData_Data_ReadString_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ReadString("agent/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestData_Data_ReadString_Bad(t *T) {
	c := New()
	r := c.Data().ReadString("missing/test.txt")
	AssertFalse(t, r.OK)
}

func TestData_Data_ReadString_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ReadString("")
	AssertFalse(t, r.OK)
}

func TestData_Data_List_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().List("agent/.")
	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.([]FsDirEntry))
}

func TestData_Data_List_Bad(t *T) {
	c := New()
	r := c.Data().List("missing/.")
	AssertFalse(t, r.OK)
}

func TestData_Data_List_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().List("agent/test.txt")
	AssertFalse(t, r.OK)
}

func TestData_Data_ListNames_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ListNames("agent/.")
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.([]string), "test")
}

func TestData_Data_ListNames_Bad(t *T) {
	c := New()
	r := c.Data().ListNames("missing/.")
	AssertFalse(t, r.OK)
}

func TestData_Data_ListNames_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	r := c.Data().ListNames("agent/.")
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.([]string), "_scantest")
}

func TestData_Data_Extract_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	target := t.TempDir()
	r := c.Data().Extract("agent/.", target, nil)
	AssertTrue(t, r.OK)
	read := (&Fs{}).New("/").Read(Path(target, "test.txt"))
	AssertTrue(t, read.OK)
}

func TestData_Data_Extract_Bad(t *T) {
	c := New()
	r := c.Data().Extract("missing/.", t.TempDir(), nil)
	AssertFalse(t, r.OK)
}

func TestData_Data_Extract_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	target := Path(t.TempDir(), "nested", "workspace")
	r := c.Data().Extract("agent/.", target, map[string]string{"Agent": "codex"})
	AssertTrue(t, r.OK)
	AssertTrue(t, (&Fs{}).New("/").Exists(Path(target, "test.txt")))
}

func TestData_Data_Mounts_Good(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	mountTestData(t, c, "docs")
	AssertEqual(t, []string{"agent", "docs"}, c.Data().Mounts())
}

func TestData_Data_Mounts_Bad(t *T) {
	c := New()
	AssertEmpty(t, c.Data().Mounts())
}

func TestData_Data_Mounts_Ugly(t *T) {
	c := New()
	mountTestData(t, c, "agent")
	mounts := c.Data().Mounts()
	mounts[0] = "mutated"
	AssertEqual(t, []string{"agent"}, c.Data().Mounts())
}
