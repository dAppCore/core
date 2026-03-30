# Core API Contract (RFC)

This file is the canonical API catalog for `dappco.re/go/core`.

- Exports follow the `core.Result` contract (`{Value any, OK bool}`) for outcomes.
- Inputs are passed as `core.Options` collections of `core.Option` key-value pairs.
- All method and function examples below compile against current repository behavior.

## 1) Core construction and top-level bootstrap types

### `type CoreOption func(*Core) Result`

```go
func customOption(*core.Core) core.Result
```

### `func New(opts ...CoreOption) *Core`

```go
c := core.New(
    core.WithOption("name", "agent"),
    core.WithService(process.Register),
    core.WithServiceLock(),
)
```

### `func WithOptions(opts Options) CoreOption`

```go
c := core.New(core.WithOptions(core.NewOptions(
    core.Option{Key: "name", Value: "agent"},
)))
```

### `func WithService(factory func(*Core) Result) CoreOption`

```go
c := core.New(core.WithService(process.Register))
```

### `func WithName(name string, factory func(*Core) Result) CoreOption`

```go
c := core.New(core.WithName("process", process.Register))
```

### `func WithOption(key string, value any) CoreOption`

```go
c := core.New(
    core.WithOption("name", "agent"),
    core.WithOption("debug", true),
)
```

### `func WithServiceLock() CoreOption`

```go
c := core.New(
    core.WithService(auth.Register),
    core.WithServiceLock(),
)
```

### `type ServiceFactory func() Result`

```go
type ServiceFactory func() core.Result
```

### `func NewWithFactories(app any, factories map[string]ServiceFactory) Result`

```go
r := core.NewWithFactories(nil, map[string]core.ServiceFactory{
    "audit": func() core.Result {
        return core.Result{Value: mySvc(core.NewServiceRuntime(core.Core, core.MyOptions{})), OK: true}
    },
})
if !r.OK { panic(r.Value) }
```

### `func NewRuntime(app any) Result`

```go
r := core.NewRuntime(nil)
runtime := r.Value.(*core.Runtime)
```

### `type Runtime struct`

- `app any`
- `Core *Core`

#### `func (r *Runtime) ServiceName() string`

```go
name := rt.ServiceName() // "Core"
```

#### `func (r *Runtime) ServiceStartup(ctx context.Context, options any) Result`

```go
r := rt.ServiceStartup(context.Background(), nil)
```

#### `func (r *Runtime) ServiceShutdown(ctx context.Context) Result`

```go
r := rt.ServiceShutdown(context.Background())
```

## 2) Core accessors and lifecycle

### `type Core struct`

#### `func (c *Core) Options() *Options`

```go
name := c.Options().String("name")
```

#### `func (c *Core) App() *App`

```go
appName := c.App().Name
```

#### `func (c *Core) Data() *Data`

```go
r := c.Data().ReadString("brain/prompt.md")
```

#### `func (c *Core) Drive() *Drive`

```go
r := c.Drive().Get("api")
```

#### `func (c *Core) Fs() *Fs`

```go
r := c.Fs().Write("status.json", "ok")
```

#### `func (c *Core) Config() *Config`

```go
c.Config().Set("debug", true)
```

#### `func (c *Core) Error() *ErrorPanic`

```go
defer c.Error().Recover()
```

#### `func (c *Core) Log() *ErrorLog`

```go
_ = c.Log().Info("boot")
```

#### `func (c *Core) Cli() *Cli`

```go
c.Cli().SetBanner(func(_ *core.Cli) string { return "agent" })
```

#### `func (c *Core) IPC() *Ipc`

```go
c.RegisterAction(func(_ *core.Core, _ core.Message) core.Result { return core.Result{OK: true} })
```

#### `func (c *Core) I18n() *I18n`

```go
_ = c.I18n().Language()
```

#### `func (c *Core) Env(key string) string`

```go
home := c.Env("DIR_HOME")
```

#### `func (c *Core) Context() context.Context`

```go
goCtx := c.Context()
```

#### `func (c *Core) Core() *Core`

```go
self := c.Core()
```

#### `func (c *Core) RunE() error`

```go
if err := c.RunE(); err != nil { /* handle */ }
```

#### `func (c *Core) Run()`

```go
c.Run()
```

#### `func (c *Core) ACTION(msg Message) Result`

```go
c.ACTION(core.ActionServiceStartup{})
```

#### `func (c *Core) QUERY(q Query) Result`

```go
r := c.QUERY(core.NewOptions(core.Option{Key: "name", Value: "agent"}))
```

#### `func (c *Core) QUERYALL(q Query) Result`

```go
r := c.QUERYALL(core.NewOptions())
```

#### `func (c *Core) LogError(err error, op, msg string) Result`

```go
_ = c.LogError(err, "core.Start", "startup failed")
```

#### `func (c *Core) LogWarn(err error, op, msg string) Result`

```go
_ = c.LogWarn(err, "agent.Check", "warning")
```

#### `func (c *Core) Must(err error, op, msg string)`

```go
c.Must(err, "core.op", "must hold")
```

#### `func (c *Core) RegistryOf(name string) *Registry[any]`

```go
svcNames := c.RegistryOf("services").Names()
```

## 3) Service and lifecycle discovery APIs

### `type Service struct`

```go
type Service struct {
    Name string
    Instance any
    Options Options
    OnStart  func() Result
    OnStop   func() Result
    OnReload func() Result
}
```

### `func (c *Core) Service(name string, service ...Service) Result`

```go
_ = c.Service("cache", core.Service{
    OnStart: func() core.Result { return core.Result{OK: true} },
    OnStop:  func() core.Result { return core.Result{OK: true} },
})
```

### `func (c *Core) RegisterService(name string, instance any) Result`

```go
_ = c.RegisterService("process", processSvc)
```

### `func (c *Core) Services() []string`

```go
names := c.Services()
```

### `func (c *Core) Lock(name string) *Lock`

```go
l := c.Lock("agent")
l.Mutex.Lock()
defer l.Mutex.Unlock()
```

### `func (c *Core) LockEnable(name ...string)`

```go
c.LockEnable()
```

### `func (c *Core) LockApply(name ...string)`

```go
c.LockApply()
```

### `func (c *Core) Startables() Result`

```go
r := c.Startables()
```

### `func (c *Core) Stoppables() Result`

```go
r := c.Stoppables()
```

### `func (c *Core) ServiceStartup(ctx context.Context, options any) Result`

```go
r := c.ServiceStartup(context.Background(), nil)
```

### `func (c *Core) ServiceShutdown(ctx context.Context) Result`

```go
r := c.ServiceShutdown(context.Background())
```

### `type ServiceRuntime[T any]`

```go
sr := core.NewServiceRuntime(c, MyServiceOptions{})
```

#### `func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T]`

```go
sr := core.NewServiceRuntime(c, core.MyOpts{})
```

#### `func (r *ServiceRuntime[T]) Core() *Core`

```go
c := sr.Core()
```

#### `func (r *ServiceRuntime[T]) Options() T`

```go
opts := sr.Options()
```

#### `func (r *ServiceRuntime[T]) Config() *Config`

```go
cfg := sr.Config()
```

### `type ServiceRegistry struct`

`*Registry[*Service]` + `lockEnabled bool`.

### `func ServiceFor[T any](c *Core, name string) (T, bool)`

```go
svc, ok := core.ServiceFor[*myService](c, "myservice")
```

### `func MustServiceFor[T any](c *Core, name string) T`

```go
svc := core.MustServiceFor[*myService](c, "myservice")
```

### `type Startable interface { OnStartup(context.Context) Result }`

### `type Stoppable interface { OnShutdown(context.Context) Result }`

## 4) Actions, Tasks, and message-driven capability primitives

### `type ActionHandler func(context.Context, Options) Result`

```go
var h ActionHandler = func(ctx context.Context, opts core.Options) core.Result { return core.Result{OK: true} }
```

### `type Action struct`

```go
a := c.Action("process.run")
```

#### `func (a *Action) Run(ctx context.Context, opts Options) Result`

```go
r := a.Run(ctx, core.NewOptions(core.Option{Key: "command", Value: "go test"}))
```

#### `func (a *Action) Exists() bool`

```go
if c.Action("process.run").Exists() { /* invoke */ }
```

### `func (c *Core) Action(name string, handler ...ActionHandler) *Action`

```go
c.Action("agent.echo", func(_ context.Context, opts core.Options) core.Result {
    return core.Result{Value: opts.String("msg"), OK: true}
})
```

### `func (c *Core) Actions() []string`

```go
names := c.Actions()
```

### `type Step struct`

`Action`, `With`, `Async`, `Input`.

### `type Task struct`

`Name`, `Description`, `Steps`.

#### `func (t *Task) Run(ctx context.Context, c *Core, opts Options) Result`

```go
r := c.Task("deploy").Run(ctx, c, core.NewOptions())
```

### `func (c *Core) Task(name string, def ...Task) *Task`

```go
c.Task("deploy", core.Task{Steps: []core.Step{{Action: "agent.echo", Async: false}}})
```

### `func (c *Core) Tasks() []string`

```go
for _, n := range c.Tasks() {}
```

### `type Message any`

```go
c.ACTION(core.ActionTaskStarted{TaskIdentifier: "t1", Action: "agent.echo"})
```

### `type Query any`

```go
type statusQuery struct{}
_ = statusQuery{}
```

### `type QueryHandler func(*Core, Query) Result`

### `func (c *Core) Query(q Query) Result`

```go
r := c.Query(core.Query("ping"))
```

### `func (c *Core) QueryAll(q Query) Result`

```go
r := c.QueryAll(core.Query("ping"))
```

### `func (c *Core) RegisterQuery(handler QueryHandler)`

```go
c.RegisterQuery(func(_ *core.Core, _ core.Query) core.Result { return core.Result{Value: "pong", OK: true} })
```

### `func (c *Core) RegisterAction(handler func(*Core, Message) Result)`

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result { return core.Result{OK: true} })
```

### `func (c *Core) RegisterActions(handlers ...func(*Core, Message) Result)`

```go
c.RegisterActions(h1, h2)
```

### `func (c *Core) RemoteAction(name string, ctx context.Context, opts Options) Result`

```go
r := c.RemoteAction("charon:agent.status", ctx, core.NewOptions())
```

### `type ActionServiceStartup struct{}`
### `type ActionServiceShutdown struct{}`

### `type ActionTaskStarted struct`

`TaskIdentifier`, `Action`, `Options`.

### `type ActionTaskProgress struct`

`TaskIdentifier`, `Action`, `Progress`, `Message`.

### `type ActionTaskCompleted struct`

`TaskIdentifier`, `Action`, `Result`.

## 5) Remote API and stream transport

### `type Stream interface`

```go
type Stream interface {
    Send(data []byte) error
    Receive() ([]byte, error)
    Close() error
}
```

### `type StreamFactory func(handle *DriveHandle) (Stream, error)`

```go
var f core.StreamFactory = func(h *core.DriveHandle) (core.Stream, error) { return nil, nil }
```

### `type API struct`

#### `func (c *Core) API() *API`

```go
_ = c.API()
```

#### `func (a *API) RegisterProtocol(scheme string, factory StreamFactory)`

```go
c.API().RegisterProtocol("mcp", mcpStreamFactory)
```

#### `func (a *API) Stream(name string) Result`

```go
r := c.API().Stream("api")
```

#### `func (a *API) Call(endpoint, action string, opts Options) Result`

```go
r := c.API().Call("api", "agent.status", core.NewOptions())
```

#### `func (a *API) Protocols() []string`

```go
names := c.API().Protocols()
```

## 6) Command and CLI layer

### `type CliOptions struct{}`
### `type Cli struct`

#### `func CliRegister(c *Core) Result`

```go
_ = core.CliRegister(c)
```

#### `func (cl *Cli) Print(format string, args ...any)`

```go
cl.Print("starting %s", "agent")
```

#### `func (cl *Cli) SetOutput(w io.Writer)`

```go
cl.SetOutput(os.Stderr)
```

#### `func (cl *Cli) Run(args ...string) Result`

```go
r := cl.Run("deploy", "to", "homelab")
```

#### `func (cl *Cli) PrintHelp()`

```go
cl.PrintHelp()
```

#### `func (cl *Core) SetBanner(fn func(*Cli) string)`

```go
cl.SetBanner(func(_ *core.Cli) string { return "agent-cli" })
```

#### `func (cl *Cli) Banner() string`

```go
label := cl.Banner()
```

### `type CommandAction func(Options) Result`

### `type Command struct`

```go
c.Command("deploy", core.Command{Action: func(opts core.Options) core.Result {
    return core.Result{Value: "ok", OK: true}
}})
```

#### `func (cmd *Command) I18nKey() string`

```go
key := c.Command("deploy/to/homelab").Value.(*core.Command).I18nKey()
```

#### `func (cmd *Command) Run(opts Options) Result`

```go
r := cmd.Run(core.NewOptions(core.Option{Key: "name", Value: "x"}))
```

#### `func (cmd *Command) IsManaged() bool`

```go
if cmd.IsManaged() { /* managed lifecycle */ }
```

### `type CommandRegistry struct`

`*Registry[*Command]`.

#### `func (c *Core) Command(path string, command ...Command) Result`

```go
c.Command("agent/deploy", core.Command{
    Action: func(opts core.Options) core.Result { return core.Result{OK: true} },
})
```

#### `func (c *Core) Commands() []string`

```go
paths := c.Commands()
```

### `type Command` fields
`Name`, `Description`, `Path`, `Action`, `Managed`, `Flags`, `Hidden`, internal `commands`.

## 7) Subsystems: App, Data, Drive, Fs, I18n, Process

### `type App struct`

#### `func (a App) New(opts Options) App`

```go
app := (core.App{}).New(core.NewOptions(
    core.Option{Key: "name", Value: "agent"},
))
```

#### `func (a App) Find(filename, name string) Result`

```go
r := core.App{}.Find("go", "Go")
```

### `type Data struct`

#### `func (d *Data) New(opts Options) Result`

```go
r := c.Data().New(core.NewOptions(
    core.Option{Key: "name", Value: "brain"},
    core.Option{Key: "source", Value: brainFS},
    core.Option{Key: "path", Value: "prompts"},
))
```

#### `func (d *Data) ReadFile(path string) Result`

```go
r := c.Data().ReadFile("brain/readme.md")
```

#### `func (d *Data) ReadString(path string) Result`

```go
r := c.Data().ReadString("brain/readme.md")
```

#### `func (d *Data) List(path string) Result`

```go
r := c.Data().List("brain/templates")
```

#### `func (d *Data) ListNames(path string) Result`

```go
r := c.Data().ListNames("brain/templates")
```

#### `func (d *Data) Extract(path, targetDir string, templateData any) Result`

```go
r := c.Data().Extract("brain/template", "/tmp/ws", map[string]string{"Name": "demo"})
```

#### `func (d *Data) Mounts() []string`

```go
for _, m := range c.Data().Mounts() {}
```

### `type DriveHandle struct`

`Name`, `Transport`, `Options`.

### `type Drive struct`

#### `func (d *Drive) New(opts Options) Result`

```go
r := c.Drive().New(core.NewOptions(
    core.Option{Key: "name", Value: "mcp"},
    core.Option{Key: "transport", Value: "mcp://localhost:1234"},
))
```

### `type Fs struct`

#### `func (m *Fs) New(root string) *Fs`

```go
fs := (&core.Fs{}).New("/tmp")
```

#### `func (m *Fs) NewUnrestricted() *Fs`

```go
fs := (&core.Fs{}).NewUnrestricted()
```

#### `func (m *Fs) Root() string`

```go
root := c.Fs().Root()
```

#### `func (m *Fs) Read(p string) Result`

```go
r := c.Fs().Read("status.txt")
```

#### `func (m *Fs) Write(p, content string) Result`

```go
_ = c.Fs().Write("status.txt", "ok")
```

#### `func (m *Fs) WriteMode(p, content string, mode os.FileMode) Result`

```go
_ = c.Fs().WriteMode("secret", "x", 0600)
```

#### `func (m *Fs) TempDir(prefix string) string`

```go
tmp := c.Fs().TempDir("agent-")
```

#### `func (m *Fs) WriteAtomic(p, content string) Result`

```go
_ = c.Fs().WriteAtomic("status.json", "{\"ok\":true}")
```

#### `func (m *Fs) EnsureDir(p string) Result`

```go
_ = c.Fs().EnsureDir("cache")
```

#### `func (m *Fs) IsDir(p string) bool`

```go
ok := c.Fs().IsDir("cache")
```

#### `func (m *Fs) IsFile(p string) bool`

```go
ok := c.Fs().IsFile("go.mod")
```

#### `func (m *Fs) Exists(p string) bool`

```go
if c.Fs().Exists("go.mod") {}
```

#### `func (m *Fs) List(p string) Result`

```go
r := c.Fs().List(".")
```

#### `func (m *Fs) Stat(p string) Result`

```go
r := c.Fs().Stat("go.mod")
```

#### `func (m *Fs) Open(p string) Result`

```go
r := c.Fs().Open("go.mod")
```

#### `func (m *Fs) Create(p string) Result`

```go
r := c.Fs().Create("notes.txt")
```

#### `func (m *Fs) Append(p string) Result`

```go
r := c.Fs().Append("notes.txt")
```

#### `func (m *Fs) Rename(oldPath, newPath string) Result`

```go
_ = c.Fs().Rename("a.txt", "b.txt")
```

#### `func (m *Fs) Delete(p string) Result`

```go
_ = c.Fs().Delete("tmp.txt")
```

#### `func (m *Fs) DeleteAll(p string) Result`

```go
_ = c.Fs().DeleteAll("tmpdir")
```

#### `func (m *Fs) ReadStream(path string) Result`

```go
r := c.Fs().ReadStream("go.mod")
```

#### `func (m *Fs) WriteStream(path string) Result`

```go
r := c.Fs().WriteStream("go.mod")
```

### Package functions in `fs.go`

### `func DirFS(dir string) fs.FS`

```go
fsys := core.DirFS("/tmp")
```

### `func ReadAll(reader any) Result`

```go
r := core.ReadAll(c.Fs().ReadStream("go.mod").Value)
```

### `func WriteAll(writer any, content string) Result`

```go
r := core.WriteAll(c.Fs().WriteStream("out.txt").Value, "value")
```

### `func CloseStream(v any)`

```go
core.CloseStream(handle)
```

### `type I18n struct`

#### `func (i *I18n) AddLocales(mounts ...*Embed)`

```go
c.I18n().AddLocales(emb)
```

#### `func (i *I18n) Locales() Result`

```go
r := c.I18n().Locales()
```

#### `func (i *I18n) SetTranslator(t Translator)`

```go
c.I18n().SetTranslator(translator)
```

#### `func (i *I18n) Translator() Result`

```go
r := c.I18n().Translator()
```

#### `func (i *I18n) Translate(messageID string, args ...any) Result`

```go
r := c.I18n().Translate("cmd.deploy.description")
```

#### `func (i *I18n) SetLanguage(lang string) Result`

```go
r := c.I18n().SetLanguage("de")
```

#### `func (i *I18n) Language() string`

```go
lang := c.I18n().Language()
```

#### `func (i *I18n) AvailableLanguages() []string`

```go
langs := c.I18n().AvailableLanguages()
```

### `type Process struct`

#### `func (c *Core) Process() *Process`

```go
p := c.Process()
```

#### `func (p *Process) Run(ctx context.Context, command string, args ...string) Result`

```go
r := c.Process().Run(ctx, "git", "status")
```

#### `func (p *Process) RunIn(ctx context.Context, dir, command string, args ...string) Result`

```go
r := c.Process().RunIn(ctx, "/tmp", "go", "test", "./...")
```

#### `func (p *Process) RunWithEnv(ctx context.Context, dir string, env []string, command string, args ...string) Result`

```go
r := c.Process().RunWithEnv(ctx, "/", []string{"CI=true"}, "go", "version")
```

#### `func (p *Process) Start(ctx context.Context, opts Options) Result`

```go
r := c.Process().Start(ctx, core.NewOptions(core.Option{Key: "command", Value: "sleep"}))
```

#### `func (p *Process) Kill(ctx context.Context, opts Options) Result`

```go
r := c.Process().Kill(ctx, core.NewOptions(core.Option{Key: "id", Value: "1234"))
```

#### `func (p *Process) Exists() bool`

```go
if c.Process().Exists() {}
```

## 8) Task/background execution and progress

### `func (c *Core) PerformAsync(action string, opts Options) Result`

```go
r := c.PerformAsync("agent.dispatch", core.NewOptions(core.Option{Key: "id", Value: 1}))
```

### `func (c *Core) Progress(taskID string, progress float64, message string, action string)`

```go
c.Progress(taskID, 0.5, "halfway", "agent.dispatch")
```

### `func (c *Core) RegisterAction(handler func(*Core, Message) Result)`

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
    _ = msg
    return core.Result{OK: true}
})
```

## 9) Logging and output

### `type Level int`

```go
const (
    core.LevelQuiet Level = iota
    core.LevelError
    core.LevelWarn
    core.LevelInfo
    core.LevelDebug
)
```

#### `func (l Level) String() string`

```go
s := core.LevelInfo.String()
```

### `type Log struct`

#### `func NewLog(opts LogOptions) *Log`

```go
logger := core.NewLog(core.LogOptions{Level: core.LevelDebug})
```

#### `func (l *Log) SetLevel(level Level)`

```go
logger.SetLevel(core.LevelWarn)
```

#### `func (l *Log) Level() Level`

```go
lvl := logger.Level()
```

#### `func (l *Log) SetOutput(w io.Writer)`

```go
logger.SetOutput(os.Stderr)
```

#### `func (l *Log) SetRedactKeys(keys ...string)`

```go
logger.SetRedactKeys("token", "password")
```

#### `func (l *Log) Debug(msg string, keyvals ...any)`

```go
logger.Debug("booted", "pid", 123)
```

#### `func (l *Log) Info(msg string, keyvals ...any)`

```go
logger.Info("agent started")
```

#### `func (l *Log) Warn(msg string, keyvals ...any)`

```go
logger.Warn("disk nearly full")
```

#### `func (l *Log) Error(msg string, keyvals ...any)`

```go
logger.Error("failed to bind", "err", err)
```

#### `func (l *Log) Security(msg string, keyvals ...any)`

```go
logger.Security("sandbox escape", "attempt", path)
```

### `type ErrorLog struct`

#### `func (el *ErrorLog) Error(err error, op, msg string) Result`

```go
r := c.Log().Error(err, "core.Run", "startup failed")
```

#### `func (el *ErrorLog) Warn(err error, op, msg string) Result`

```go
r := c.Log().Warn(err, "warn", "soft error")
```

#### `func (el *ErrorLog) Must(err error, op, msg string)`

```go
c.Log().Must(err, "core.maybe", "must hold")
```

### `type ErrorPanic struct`

#### `func (h *ErrorPanic) Recover()`

```go
defer c.Error().Recover()
```

#### `func (h *ErrorPanic) SafeGo(fn func())`

```go
c.Error().SafeGo(func() { panic("boom") })
```

#### `func (h *ErrorPanic) Reports(n int) Result`

```go
r := c.Error().Reports(3)
```

### `type LogErr struct`

#### `func NewLogErr(log *Log) *LogErr`

```go
le := core.NewLogErr(core.Default())
```

#### `func (le *LogErr) Log(err error)`

```go
le.Log(err)
```

### `type LogPanic struct`

#### `func NewLogPanic(log *Log) *LogPanic`

```go
lp := core.NewLogPanic(core.Default())
```

#### `func (lp *LogPanic) Recover()`

```go
defer lp.Recover()
```

### Package-level logger helpers

`Default`, `SetDefault`, `SetLevel`, `SetRedactKeys`, `Debug`, `Info`, `Warn`, `Error`, `Security`.

### `func Default() *Log`

```go
l := core.Default()
```

### `func SetDefault(l *Log)`

```go
core.SetDefault(core.NewLog(core.LogOptions{Level: core.LevelDebug}))
```

### `func SetLevel(level Level)`

```go
core.SetLevel(core.LevelInfo)
```

### `func SetRedactKeys(keys ...string)`

```go
core.SetRedactKeys("password")
```

### `func Debug(msg string, keyvals ...any)`

```go
core.Debug("start")
```

### `func Info(msg string, keyvals ...any)`

```go
core.Info("ready")
```

### `func Warn(msg string, keyvals ...any)`

```go
core.Warn("high load")
```

### `func Error(msg string, keyvals ...any)`

```go
core.Error("failure", "err", err)
```

### `func Security(msg string, keyvals ...any)`

```go
core.Security("policy", "event", "denied")
```

### `type LogOptions struct`

`Level`, `Output`, `Rotation`, `RedactKeys`.

### `type RotationLogOptions struct`

`Filename`, `MaxSize`, `MaxAge`, `MaxBackups`, `Compress`.

### `var RotationWriterFactory func(RotationLogOptions) io.WriteCloser`

```go
core.RotationWriterFactory = myFactory
```

### `func Username() string`

```go
u := core.Username()
```

## 10) Error model and diagnostics

### `type Err struct`

`Operation`, `Message`, `Cause`, `Code`.

#### `func (e *Err) Error() string`

```go
_ = err.Error()
```

#### `func (e *Err) Unwrap() error`

```go
_ = errors.Unwrap(err)
```

### `func E(op, msg string, err error) error`

```go
r := core.E("core.Run", "startup failed", err)
```

### `func Wrap(err error, op, msg string) error`

```go
r := core.Wrap(err, "api.Call", "request failed")
```

### `func WrapCode(err error, code, op, msg string) error`

```go
r := core.WrapCode(err, "NOT_AUTHORIZED", "api.Call", "forbidden")
```

### `func NewCode(code, msg string) error`

```go
_ = core.NewCode("VALIDATION", "invalid input")
```

### `func NewError(text string) error`

```go
_ = core.NewError("boom")
```

### `func ErrorJoin(errs ...error) error`

```go
err := core.ErrorJoin(e1, e2, e3)
```

### `func ErrorMessage(err error) string`

```go
msg := core.ErrorMessage(err)
```

### `func ErrorCode(err error) string`

```go
code := core.ErrorCode(err)
```

### `func Operation(err error) string`

```go
op := core.Operation(err)
```

### `func Root(err error) error`

```go
root := core.Root(err)
```

### `func AllOperations(err error) iter.Seq[string]`

```go
for op := range core.AllOperations(err) { fmt.Println(op) }
```

### `func StackTrace(err error) []string`

```go
stack := core.StackTrace(err)
```

### `func FormatStackTrace(err error) string`

```go
fmt.Println(core.FormatStackTrace(err))
```

### `func As(err error, target any) bool`

```go
var ee *core.Err
_ = core.As(err, &ee)
```

### `func Is(err, target error) bool`

```go
_ = core.Is(err, io.EOF)
```

### `type CrashReport struct`

### `type CrashSystem struct`

## 11) Asset packing and embed helpers

### `type AssetGroup struct`
### `type AssetRef struct`
### `type ScannedPackage struct`

### `func AddAsset(group, name, data string)`

```go
core.AddAsset("g", "n", "payload")
```

### `func GetAsset(group, name string) Result`

```go
r := core.GetAsset("g", "n")
```

### `func GetAssetBytes(group, name string) Result`

```go
r := core.GetAssetBytes("g", "n")
```

### `func ScanAssets(filenames []string) Result`

```go
r := core.ScanAssets([]string{"main.go"})
```

### `func GeneratePack(pkg ScannedPackage) Result`

```go
r := core.GeneratePack(scanned)
```

### `func Mount(fsys fs.FS, basedir string) Result`

```go
r := core.Mount(appFS, "assets")
```

### `func MountEmbed(efs embed.FS, basedir string) Result`

```go
r := core.MountEmbed(efs, "assets")
```

### `func Extract(fsys fs.FS, targetDir string, data any, opts ...ExtractOptions) Result`

```go
r := core.Extract(embeds, "/tmp/out", map[string]string{"Name": "demo"})
```

### `type ExtractOptions struct`
`TemplateFilters []string`, `IgnoreFiles map[string]struct{}`, `RenameFiles map[string]string`.

### `type Embed struct`

#### `func (s *Embed) Open(name string) Result`

```go
r := emb.Open("readme.md")
```

#### `func (s *Embed) ReadDir(name string) Result`

```go
r := emb.ReadDir("templates")
```

#### `func (s *Embed) ReadFile(name string) Result`

```go
r := emb.ReadFile("readme.md")
```

#### `func (s *Embed) ReadString(name string) Result`

```go
r := emb.ReadString("readme.md")
```

#### `func (s *Embed) Sub(subDir string) Result`

```go
r := emb.Sub("assets")
```

#### `func (s *Embed) FS() fs.FS`

```go
fsys := emb.FS()
```

#### `func (s *Embed) EmbedFS() embed.FS`

```go
efs := emb.EmbedFS()
```

#### `func (s *Embed) BaseDirectory() string`

```go
base := emb.BaseDirectory()
```

## 12) Configuration and primitives

### `type Option struct`

```go
opt := core.Option{Key: "name", Value: "agent"}
```

### `type ConfigVar[T any]`

#### `func NewConfigVar[T any](val T) ConfigVar[T]`

```go
v := core.NewConfigVar("blue")
```

#### `func (v *ConfigVar[T]) Get() T`

```go
val := v.Get()
```

#### `func (v *ConfigVar[T]) Set(val T)`

```go
v.Set("red")
```

#### `func (v *ConfigVar[T]) IsSet() bool`

```go
if v.IsSet() {}
```

#### `func (v *ConfigVar[T]) Unset()`

```go
v.Unset()
```

### `type ConfigOptions struct`
`Settings map[string]any`, `Features map[string]bool`.

### `type Config struct`

#### `func (e *Config) New() *Config`

```go
cfg := (&core.Config{}).New()
```

#### `func (e *Config) Set(key string, val any)`

```go
cfg.Set("port", 8080)
```

#### `func (e *Config) Get(key string) Result`

```go
r := cfg.Get("port")
```

#### `func (e *Config) String(key string) string`

```go
v := cfg.String("port")
```

#### `func (e *Config) Int(key string) int`

```go
v := cfg.Int("retries")
```

#### `func (e *Config) Bool(key string) bool`

```go
v := cfg.Bool("debug")
```

#### `func (e *Config) Enable(feature string)`

```go
e.Enable("tracing")
```

#### `func (e *Config) Disable(feature string)`

```go
e.Disable("tracing")
```

#### `func (e *Config) Enabled(feature string) bool`

```go
if e.Enabled("tracing") {}
```

#### `func (e *Config) EnabledFeatures() []string`

```go
features := e.EnabledFeatures()
```

### `func ConfigGet[T any](e *Config, key string) T`

```go
limit := core.ConfigGet[int](cfg, "retries")
```

### `type Options struct`

#### `func NewOptions(items ...Option) Options`

```go
opts := core.NewOptions(core.Option{Key: "name", Value: "x"})
```

#### `func (o *Options) Set(key string, value any)`

```go
opts.Set("debug", true)
```

#### `func (o Options) Get(key string) Result`

```go
r := opts.Get("debug")
```

#### `func (o Options) Has(key string) bool`

```go
if opts.Has("debug") {}
```

#### `func (o Options) String(key string) string`

```go
v := opts.String("name")
```

#### `func (o Options) Int(key string) int`

```go
v := opts.Int("port")
```

#### `func (o Options) Bool(key string) bool`

```go
b := opts.Bool("debug")
```

#### `func (o Options) Len() int`

```go
n := opts.Len()
```

#### `func (o Options) Items() []Option`

```go
all := opts.Items()
```

### `type Result struct`

#### `func (r Result) Result(args ...any) Result`

```go
r := core.Result{}.Result(file, err)
```

#### `func (r Result) New(args ...any) Result`

```go
r := core.Result{}.New(file, nil)
```

#### `func (r Result) Get() Result`

```go
v := r.Get()
```

### `func JSONMarshal(v any) Result`

```go
r := core.JSONMarshal(map[string]string{"k": "v"})
```

### `func JSONMarshalString(v any) string`

```go
s := core.JSONMarshalString(map[string]any{"k": "v"})
```

### `func JSONUnmarshal(data []byte, target any) Result`

```go
r := core.JSONUnmarshal([]byte(`{"x":1}`), &cfg)
```

### `func JSONUnmarshalString(s string, target any) Result`

```go
r := core.JSONUnmarshalString(`{"x":1}`, &cfg)
```

## 13) Registry primitive

### `type Registry[T any] struct`

#### `func NewRegistry[T any]() *Registry[T]`

```go
r := core.NewRegistry[*core.Service]()
```

#### `func (r *Registry[T]) Set(name string, item T) Result`

```go
r.Set("process", svc)
```

#### `func (r *Registry[T]) Get(name string) Result`

```go
got := r.Get("process")
```

#### `func (r *Registry[T]) Has(name string) bool`

```go
if r.Has("process") {}
```

#### `func (r *Registry[T]) Names() []string`

```go
for _, n := range r.Names() {}
```

#### `func (r *Registry[T]) List(pattern string) []T`

```go
vals := r.List("process.*")
```

#### `func (r *Registry[T]) Each(fn func(string, T))`

```go
r.Each(func(name string, v *core.Service) {})
```

#### `func (r *Registry[T]) Len() int`

```go
n := r.Len()
```

#### `func (r *Registry[T]) Delete(name string) Result`

```go
_ = r.Delete("legacy")
```

#### `func (r *Registry[T]) Disable(name string) Result`

```go
_ = r.Disable("legacy")
```

#### `func (r *Registry[T]) Enable(name string) Result`

```go
_ = r.Enable("legacy")
```

#### `func (r *Registry[T]) Disabled(name string) bool`

```go
if r.Disabled("legacy") {}
```

#### `func (r *Registry[T]) Lock()`

```go
r.Lock()
```

#### `func (r *Registry[T]) Locked() bool`

```go
locked := r.Locked()
```

#### `func (r *Registry[T]) Seal()`

```go
r.Seal()
```

#### `func (r *Registry[T]) Sealed() bool`

```go
if r.Sealed() {}
```

#### `func (r *Registry[T]) Open()`

```go
r.Open()
```

## 14) Entitlement and security policy hooks

### `type Entitlement struct`

`Allowed`, `Unlimited`, `Limit`, `Used`, `Remaining`, `Reason`.

### `func (e Entitlement) NearLimit(threshold float64) bool`

```go
if e.NearLimit(0.8) {}
```

### `func (e Entitlement) UsagePercent() float64`

```go
pct := e.UsagePercent()
```

### `type EntitlementChecker func(action string, quantity int, ctx context.Context) Entitlement`
### `type UsageRecorder func(action string, quantity int, ctx context.Context)`

#### `func (c *Core) Entitled(action string, quantity ...int) Entitlement`

```go
e := c.Entitled("process.run")
```

#### `func (c *Core) SetEntitlementChecker(checker EntitlementChecker)`

```go
c.SetEntitlementChecker(myChecker)
```

#### `func (c *Core) RecordUsage(action string, quantity ...int)`

```go
c.RecordUsage("process.run")
```

#### `func (c *Core) SetUsageRecorder(recorder UsageRecorder)`

```go
c.SetUsageRecorder(myRecorder)
```

## 15) Generic and helper collections

### `type Array[T comparable]`

#### `func NewArray[T comparable](items ...T) *Array[T]`

```go
arr := core.NewArray("a", "b")
```

#### `func (s *Array[T]) Add(values ...T)`

```go
arr.Add("c")
```

#### `func (s *Array[T]) AddUnique(values ...T)`

```go
arr.AddUnique("a")
```

#### `func (s *Array[T]) Contains(val T) bool`

```go
_ = arr.Contains("a")
```

#### `func (s *Array[T]) Filter(fn func(T) bool) Result`

```go
r := arr.Filter(func(v string) bool { return v != "x" })
```

#### `func (s *Array[T]) Each(fn func(T))`

```go
arr.Each(func(v string) {})
```

#### `func (s *Array[T]) Remove(val T)`

```go
arr.Remove("b")
```

#### `func (s *Array[T]) Deduplicate()`

```go
arr.Deduplicate()
```

#### `func (s *Array[T]) Len() int`

```go
n := arr.Len()
```

#### `func (s *Array[T]) Clear()`

```go
arr.Clear()
```

#### `func (s *Array[T]) AsSlice() []T`

```go
vals := arr.AsSlice()
```

## 16) String and path utility API

### String helpers
- `HasPrefix`, `HasSuffix`, `TrimPrefix`, `TrimSuffix`, `Contains`, `Split`, `SplitN`, `Join`, `Replace`, `Lower`, `Upper`, `Trim`, `RuneCount`, `NewBuilder`, `NewReader`, `Sprint`, `Sprintf`, `Concat`.

```go
core.Join("/", "deploy", "to", "homelab")
core.Concat("cmd.", "deploy", ".description")
```

### Path helpers
- `Path`, `PathBase`, `PathDir`, `PathExt`, `PathIsAbs`, `JoinPath`, `CleanPath`, `PathGlob`.

```go
core.Path("Code", "agent")
core.PathIsAbs("/tmp")
```

#### `func JoinPath(segments ...string) string`

```go
p := core.JoinPath("workspace", "deploy", "main.go")
```

### Generic I/O helpers
- `Arg`, `ArgString`, `ArgInt`, `ArgBool`, `ParseFlag`, `FilterArgs`, `IsFlag`, `ID`, `ValidateName`, `SanitisePath`.

```go
id := core.ID()
name := core.ValidateName("agent").Value
```

### Package-level functions in this section

#### `func Print(w io.Writer, format string, args ...any)`

```go
core.Print(os.Stderr, "hello %s", "world")
```

#### `func Println(args ...any)`

```go
core.Println("ready")
```

#### `func Arg(index int, args ...any) Result`

```go
r := core.Arg(0, "x", 1)
```

#### `func ArgString(index int, args ...any) string`

```go
v := core.ArgString(0, "x", 1)
```

#### `func ArgInt(index int, args ...any) int`

```go
v := core.ArgInt(1, "x", 42)
```

#### `func ArgBool(index int, args ...any) bool`

```go
v := core.ArgBool(2, true)
```

#### `func IsFlag(arg string) bool`

```go
ok := core.IsFlag("--debug")
```

#### `func ParseFlag(arg string) (string, string, bool)`

```go
key, val, ok := core.ParseFlag("--name=agent")
```

#### `func FilterArgs(args []string) []string`

```go
clean := core.FilterArgs(os.Args[1:])
```

#### `func ID() string`

```go
id := core.ID()
```

#### `func ValidateName(name string) Result`

```go
r := core.ValidateName("deploy")
```

#### `func SanitisePath(path string) string`

```go
safe := core.SanitisePath("../../etc")
```

#### `func DirFS(dir string) fs.FS`

```go
root := core.DirFS("/tmp")
```

## 17) System and environment helpers

### `func Env(key string) string`

```go
home := core.Env("DIR_HOME")
```

### `func EnvKeys() []string`

```go
keys := core.EnvKeys()
```

### `func Username() string`

```go
who := core.Username()
```

## 18) Remaining interfaces and types

### `type ErrorSink interface`

`Error(msg string, keyvals ...any)` and `Warn(msg string, keyvals ...any)`.

### `type Stream interface`

`Send`, `Receive`, `Close` as above.

### `type Translator interface`

`Translate`, `SetLanguage`, `Language`, `AvailableLanguages`.

### `type LocaleProvider interface`

`Locales() *Embed`.

### `type Runtime helpers`

- `type Ipc` (named fields, no exported methods)
- `type Lock struct{ Name string; Mutex *sync.RWMutex }`
- `type SysInfo struct{ values map[string]string }`

### `type Ipc struct`

```go
ipc := c.IPC()
```

### `type Lock struct`

```go
guard := c.Lock("core-bootstrap")
guard.Mutex.Lock()
defer guard.Mutex.Unlock()
```

### `type SysInfo struct`

```go
// SysInfo values are accessed through Env()/EnvKeys(); direct type construction is not required.
home := core.Env("DIR_HOME")
keys := core.EnvKeys()
```

## 19) Legacy behavior notes

- `Service` lifecycle callbacks are the DTO fields `OnStart` and `OnStop`.
- `Startable`/`Stoppable` are the interface contracts and map to `OnStartup(context.Context)` / `OnShutdown(context.Context)`.
- Registry iteration (`Names`, `List`, `Each`, `Services`, `Startables`, `Stoppables`) is insertion-order based via `Registry.order`.
