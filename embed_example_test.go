package core_test

import . "dappco.re/go"

func ExampleAddAsset() {
	AddAsset("example.embed", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAsset("example.embed", "hello.txt")
	Println(r.Value)
	// Output: hello
}

func ExampleGetAsset() {
	AddAsset("example.asset", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAsset("example.asset", "hello.txt")
	Println(r.Value)
	// Output: hello
}

func ExampleGetAssetBytes() {
	AddAsset("example.bytes", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAssetBytes("example.bytes", "hello.txt")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

func ExampleScanAssets() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-scan-example")
	defer fs.DeleteAll(dir)

	fs.Write(Path(dir, "main.go"), `package sample

import core "dappco.re/go"

func message() string {
	return core.GetAsset("assets", "message.txt").Value.(string)
}
`)

	r := ScanAssets([]string{Path(dir, "main.go")})
	pkgs := r.Value.([]ScannedPackage)
	Println(r.OK)
	Println(pkgs[0].PackageName)
	Println(pkgs[0].Assets[0].Group)
	Println(pkgs[0].Assets[0].Name)
	// Output:
	// true
	// sample
	// assets
	// message.txt
}

func ExampleGeneratePack() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-pack-example")
	defer fs.DeleteAll(dir)

	fs.Write(Path(dir, "assets", "message.txt"), "hello")
	fs.Write(Path(dir, "main.go"), `package sample

import core "dappco.re/go"

func message() string {
	return core.GetAsset("assets", "message.txt").Value.(string)
}
`)

	scanned := ScanAssets([]string{Path(dir, "main.go")}).Value.([]ScannedPackage)
	pack := GeneratePack(scanned[0])
	source := pack.Value.(string)
	Println(pack.OK)
	Println(Contains(source, `core.AddAsset("assets", "message.txt"`))
	// Output:
	// true
	// true
}

func ExampleMount() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-mount-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	r := Mount(DirFS(dir), "docs")
	emb := r.Value.(*Embed)
	Println(emb.BaseDirectory())
	Println(emb.ReadString("hello.txt").Value)
	// Output:
	// docs
	// hello
}

func ExampleMountEmbed() {
	_ = MountEmbed((&Embed{}).EmbedFS(), ".")
}

func ExampleEmbed_Open() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	r := emb.Open("hello.txt")
	Println(r.OK)
	CloseStream(r.Value)
	// Output: true
}

func ExampleEmbed_ReadDir() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	r := emb.ReadDir(".")
	Println(r.OK)
	// Output: true
}

func ExampleEmbed_ReadFile() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	r := emb.ReadFile("hello.txt")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

func ExampleEmbed_ReadString() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.ReadString("hello.txt").Value)
	// Output: hello
}

func ExampleEmbed_Sub() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "nested", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	sub := emb.Sub("nested").Value.(*Embed)
	Println(sub.ReadString("hello.txt").Value)
	// Output: hello
}

func ExampleEmbed_FS() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.FS() != nil)
	// Output: true
}

func ExampleEmbed_EmbedFS() {
	emb := &Embed{}
	_ = emb.EmbedFS()
}

func ExampleEmbed_BaseDirectory() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.BaseDirectory())
	// Output: docs
}

func ExampleExtract() {
	fs := (&Fs{}).New("/")
	source := fs.TempDir("core-extract-source")
	target := fs.TempDir("core-extract-target")
	defer fs.DeleteAll(source)
	defer fs.DeleteAll(target)

	fs.Write(Path(source, "README.md.tmpl"), "hello {{.Name}}")
	r := Extract(DirFS(source), target, map[string]string{"Name": "codex"})
	Println(r.OK)
	Println(fs.Read(Path(target, "README.md")).Value)
	// Output:
	// true
	// hello codex
}

func ExampleExtractOptions() {
	opts := ExtractOptions{
		TemplateFilters: []string{".gotmpl"},
		RenameFiles:     map[string]string{"README.md": "WELCOME.md"},
	}
	Println(opts.TemplateFilters)
	Println(opts.RenameFiles["README.md"])
	// Output:
	// [.gotmpl]
	// WELCOME.md
}
