// SPDX-License-Identifier: EUPL-1.2

package core

func TestEmbed_Embed_path_Good(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path("agent/readme.md")

	AssertTrue(t, r.OK)
	AssertEqual(t, "assets/agent/readme.md", r.Value)
}
func TestEmbed_Embed_path_Bad(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path("../../secrets/token")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path traversal rejected")
}
func TestEmbed_Embed_path_Ugly(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path(".")

	AssertTrue(t, r.OK)
	AssertEqual(t, "assets", r.Value)
}
func TestEmbed_compress_Good(t *T) {
	packed, err := compress("agent dispatch ready")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "agent dispatch ready", plain)
}
func TestEmbed_compress_Bad(t *T) {
	packed, err := compress("")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "", plain)
}
func TestEmbed_compress_Ugly(t *T) {
	input := Join("\n", "agent", "dispatch", "retry")
	packed, err := compress(input)
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, input, plain)
}
func TestEmbed_compressFile_Good(t *T) {
	path := Path(t.TempDir(), "agent.txt")
	RequireTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	packed, err := compressFile(path)
	RequireNoError(t, err)
	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "ready", plain)
}
func TestEmbed_compressFile_Bad(t *T) {
	_, err := compressFile(Path(t.TempDir(), "missing.txt"))

	AssertError(t, err)
}
func TestEmbed_compressFile_Ugly(t *T) {
	path := Path(t.TempDir(), "empty.txt")
	RequireTrue(t, WriteFile(path, nil, 0o644).OK)

	packed, err := compressFile(path)
	RequireNoError(t, err)
	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "", plain)
}
func TestEmbed_decompress_Good(t *T) {
	packed, err := compress("homelab")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "homelab", plain)
}
func TestEmbed_decompress_Bad(t *T) {
	_, err := decompress("not base64")

	AssertError(t, err)
}
func TestEmbed_decompress_Ugly(t *T) {
	_, err := decompress(Base64Encode([]byte("plain text")))

	AssertError(t, err)
}
func TestEmbed_getAllFiles_Good(t *T) {
	dir := t.TempDir()
	agent := Path(dir, "agent.txt")
	task := Path(dir, "nested", "task.txt")
	RequireTrue(t, WriteFile(agent, []byte("agent"), 0o644).OK)
	RequireTrue(t, MkdirAll(Path(dir, "nested"), 0o755).OK)
	RequireTrue(t, WriteFile(task, []byte("task"), 0o644).OK)

	files, err := getAllFiles(dir)

	RequireNoError(t, err)
	AssertContains(t, files, agent)
	AssertContains(t, files, task)
}
func TestEmbed_getAllFiles_Bad(t *T) {
	_, err := getAllFiles(Path(t.TempDir(), "missing"))

	AssertError(t, err)
}
func TestEmbed_getAllFiles_Ugly(t *T) {
	files, err := getAllFiles(t.TempDir())

	RequireNoError(t, err)
	AssertEmpty(t, files)
}
func TestEmbed_isTemplate_Good(t *T) {
	AssertTrue(t, isTemplate("README.md.tmpl", []string{".tmpl"}))
}
func TestEmbed_isTemplate_Bad(t *T) {
	AssertFalse(t, isTemplate("README.md", []string{".tmpl"}))
}
func TestEmbed_isTemplate_Ugly(t *T) {
	AssertTrue(t, isTemplate("agent.go.tpl", []string{".tmpl", ".tpl"}))
}
func TestEmbed_renderPath_Good(t *T) {
	path := renderPath("workspace/{{.Name}}/README.md", map[string]string{"Name": "agent"})

	AssertEqual(t, "workspace/agent/README.md", path)
}
func TestEmbed_renderPath_Bad(t *T) {
	path := "workspace/{{.Name/README.md"

	AssertEqual(t, path, renderPath(path, map[string]string{"Name": "agent"}))
}
func TestEmbed_renderPath_Ugly(t *T) {
	path := "workspace/{{.Name}}/README.md"

	AssertEqual(t, path, renderPath(path, nil))
}
func TestEmbed_copyFile_Good(t *T) {
	src := t.TempDir()
	target := Path(t.TempDir(), "agent.txt")
	RequireTrue(t, WriteFile(Path(src, "agent.txt"), []byte("ready"), 0o644).OK)

	err := copyFile(DirFS(src), "agent.txt", target)

	RequireNoError(t, err)
	AssertEqual(t, "ready", string(ReadFile(target).Value.([]byte)))
}
func TestEmbed_copyFile_Bad(t *T) {
	err := copyFile(DirFS(t.TempDir()), "missing.txt", Path(t.TempDir(), "out.txt"))

	AssertError(t, err)
}
func TestEmbed_copyFile_Ugly(t *T) {
	src := t.TempDir()
	target := Path(t.TempDir(), "nested", "agent.txt")
	RequireTrue(t, WriteFile(Path(src, "agent.txt"), []byte("nested"), 0o644).OK)

	err := copyFile(DirFS(src), "agent.txt", target)

	RequireNoError(t, err)
	AssertEqual(t, "nested", string(ReadFile(target).Value.([]byte)))
}
