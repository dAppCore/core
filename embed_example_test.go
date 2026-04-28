package core_test

import . "dappco.re/go"

// ExampleAddAsset adds an embedded asset through `AddAsset` for embedded asset packaging.
// Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleAddAsset() {
	AddAsset("example.embed", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAsset("example.embed", "hello.txt")
	Println(r.Value)
	// Output: hello
}

// ExampleGetAsset retrieves an embedded asset through `GetAsset` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleGetAsset() {
	AddAsset("example.asset", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAsset("example.asset", "hello.txt")
	Println(r.Value)
	// Output: hello
}

// ExampleGetAssetBytes retrieves embedded asset bytes through `GetAssetBytes` for embedded
// asset packaging. Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleGetAssetBytes() {
	AddAsset("example.bytes", "hello.txt", "H4sIAAAAAAAC/8pIzcnJBwQAAP//hqYQNgUAAAA=")
	r := GetAssetBytes("example.bytes", "hello.txt")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

// ExampleScanAssets scans embedded assets through `ScanAssets` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleGeneratePack generates an asset pack through `GeneratePack` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleMount mounts embedded assets through `Mount` for embedded asset packaging. Asset
// packing, mounting, and extraction stay declarative for consumers.
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

// ExampleMountEmbed mounts an embed filesystem through `MountEmbed` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleMountEmbed() {
	_ = MountEmbed((&Embed{}).EmbedFS(), ".")
}

// ExampleEmbed_Open opens a mounted asset for streaming reads from an embedded package.
// Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleEmbed_ReadDir lists directory entries through `Embed.ReadDir` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleEmbed_ReadFile reads a named file through `Embed.ReadFile` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleEmbed_ReadString reads text content through `Embed.ReadString` for embedded asset
// packaging. Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleEmbed_ReadString() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.ReadString("hello.txt").Value)
	// Output: hello
}

// ExampleEmbed_Sub opens a subdirectory through `Embed.Sub` for embedded asset packaging.
// Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleEmbed_FS exposes a filesystem through `Embed.FS` for embedded asset packaging.
// Asset packing, mounting, and extraction stay declarative for consumers.
func ExampleEmbed_FS() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.FS() != nil)
	// Output: true
}

// ExampleEmbed_EmbedFS exposes the underlying embed filesystem through `Embed.EmbedFS` for
// embedded asset packaging. Asset packing, mounting, and extraction stay declarative for
// consumers.
func ExampleEmbed_EmbedFS() {
	emb := &Embed{}
	_ = emb.EmbedFS()
}

// ExampleEmbed_BaseDirectory reports the embedded base directory through
// `Embed.BaseDirectory` for embedded asset packaging. Asset packing, mounting, and
// extraction stay declarative for consumers.
func ExampleEmbed_BaseDirectory() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-embed-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "docs", "hello.txt"), "hello")

	emb := Mount(DirFS(dir), "docs").Value.(*Embed)
	Println(emb.BaseDirectory())
	// Output: docs
}

// ExampleExtract extracts embedded files through `Extract` for embedded asset packaging.
// Asset packing, mounting, and extraction stay declarative for consumers.
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

// ExampleExtractOptions declares extraction options through `ExtractOptions` for embedded
// asset packaging. Asset packing, mounting, and extraction stay declarative for consumers.
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
