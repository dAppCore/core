Here is the technical documentation for the Core framework packages.

# Core Framework Documentation

## Package: pkg/log

### 1. Overview
`pkg/log` acts as the central observability and error handling primitive for the framework. It combines structured logging with a rich error type system (`Err`), allowing operational context (Operations, Codes) to travel with errors up the stack. It is designed to be used both standalone and as an injectable service within the Core framework.

### 2. Public API

**Error Types & Functions**
*   `type Err`: Struct implementing `error` with fields for `Op` (operation), `Msg` (message), `Err` (wrapped error), and `Code` (machine-readable code).
*   `func E(op, msg string, err error) error`: Creates a new error with operational context.
*   `func Wrap(err error, op, msg string) error`: Wraps an existing error, preserving existing codes if present.
*   `func WrapCode(err error, code, op, msg string) error`: Wraps an error and assigns a specific error code.
*   `func NewCode(code, msg string) error`: Creates a sentinel error with a code.
*   `func Is(err, target error) bool`: Wrapper for `errors.Is`.
*   `func As(err error, target any) bool`: Wrapper for `errors.As`.
*   `func Join(errs ...error) error`: Wrapper for `errors.Join`.
*   `func Op(err error) string`: Extracts the operation name from an error chain.
*   `func ErrCode(err error) string`: Extracts the error code from an error chain.
*   `func StackTrace(err error) []string`: Returns a slice of operations leading to the error.
*   `func LogError(err error, op, msg string) error`: Logs an error and returns it wrapped (reduces boilerplate).
*   `func LogWarn(err error, op, msg string) error`: Logs a warning and returns it wrapped.
*   `func Must(err error, op, msg string)`: Panics if error is not nil, logging it first.

**Logging Types & Functions**
*   `type Logger`: The main logging struct.
*   `type Level`: Integer type for log verbosity (`LevelQuiet` to `LevelDebug`).
*   `type Options`: Configuration struct for Logger (Level, Output, Rotation).
*   `type RotationOptions`: Config for log file rotation (Size, Age, Backups, Compression).
*   `func New(opts Options) *Logger`: Constructor.
*   `func Default() *Logger`: Returns the global default logger.
*   `func SetDefault(l *Logger)`: Sets the global default logger.
*   `func (l *Logger) Debug/Info/Warn/Error/Security(msg string, keyvals ...any)`: Leveled logging methods.

**Service Integration**
*   `type Service`: Wraps `Logger` for framework integration.
*   `func NewService(opts Options) func(*framework.Core) (any, error)`: Factory for dependency injection.
*   `type QueryLevel`, `type TaskSetLevel`: Message types for runtime management.

### 3. Internal Design
*   **Contextual Errors**: The `Err` struct forms a linked list via the `Err` field (inner error), allowing the reconstruction of a logical stack trace (`op` sequence) distinct from the runtime stack trace.
*   **Concurrency**: The `Logger` uses a `sync.RWMutex` to guard configuration and writes, ensuring thread safety.
*   **Rotation Strategy**: The `RotatingWriter` implements `io.WriteCloser`. It lazily opens files and checks size thresholds on every write, leveraging `pkg/io` to abstract the filesystem.
*   **Framework Integration**: The `Service` struct embeds `framework.ServiceRuntime`, utilizing the Actor pattern (Queries and Tasks) to allow dynamic log level adjustment at runtime without restarting the application.

### 4. Dependencies
*   `github.com/host-uk/core/pkg/io`: Used by `rotation.go` to handle file operations (renaming, deleting, writing) abstractly.
*   `github.com/host-uk/core/pkg/framework`: Used by `service.go` to hook into the application lifecycle and message bus.
*   Standard Lib: `errors`, `fmt`, `os`, `sync`, `time`.

### 5. Test Coverage Notes
*   **Error Unwrapping**: Verify `errors.Is` and `errors.As` work correctly through deep chains of `log.Err`.
*   **Logical Stack Traces**: Ensure `StackTrace()` returns the correct order of operations `["app.Run", "db.Query", "net.Dial"]`.
*   **Log Rotation**: Critical to test the boundary conditions of `MaxSize` and `MaxBackups` using a Mock Medium to avoid actual disk I/O.
*   **Concurrency**: Race detection on `Logger` when changing levels while logging is active.

### 6. Integration Points
*   **Application-wide**: This is the most imported package. All other packages should use `log.E` or `log.Wrap` instead of `fmt.Errorf` or `errors.New`.
*   **Core Framework**: The `Service` is designed to be passed to `core.New()`.

---

## Package: pkg/config

### 1. Overview
`pkg/config` provides 12-factor app configuration management. It layers configuration sources in a specific precedence (Environment > Config File > Defaults) and exposes them via a typed API or a dot-notation getter. It abstracts the underlying storage, allowing configs to be loaded from disk or memory.

### 2. Public API
*   `type Config`: The main configuration manager.
*   `type Option`: Functional option pattern for configuration.
*   `func New(opts ...Option) (*Config, error)`: Constructor.
*   `func LoadEnv(prefix string) map[string]any`: Helper to parse environment variables into a map.
*   `func (c *Config) Get(key string, out any) error`: Unmarshals a key (or root) into a struct.
*   `func (c *Config) Set(key string, v any) error`: Sets a value and persists it to storage.
*   `func (c *Config) LoadFile(m coreio.Medium, path string) error`: Merges a file into the current config.
*   `type Service`: Framework service wrapper for `Config`.
*   `func NewConfigService(c *core.Core) (any, error)`: Factory for dependency injection.

### 3. Internal Design
*   **Engine**: Uses `spf13/viper` as the underlying configuration engine for its merging and unmarshalling logic.
*   **Abstraction**: Unlike standard Viper usage, this package decouples the filesystem using `pkg/io.Medium`. This allows the config system to work in sandboxed environments or with mock filesystems.
*   **Persistence**: The `Set` method triggers an immediate write-back to the storage medium, making the config file the source of truth for runtime changes.
*   **Environment Mapping**: Automatically maps `CORE_CONFIG_FOO_BAR` to `foo.bar` using a `strings.Replacer`.

### 4. Dependencies
*   `github.com/spf13/viper`: Core logic for map merging and unmarshalling.
*   `gopkg.in/yaml.v3`: For marshalling data when saving.
*   `github.com/host-uk/core/pkg/io`: For reading/writing config files.
*   `github.com/host-uk/core/pkg/framework/core`: For service integration and error handling.

### 5. Test Coverage Notes
*   **Precedence**: Verify that Environment variables override File values.
*   **Persistence**: Test that `Set()` writes valid YAML back to the `Medium`.
*   **Type Safety**: Ensure `Get()` correctly unmarshals into complex structs and returns errors on type mismatches.

### 6. Integration Points
*   **Bootstrap**: Usually the first service initialized in `core.New()`.
*   **Service Configuration**: Other services (like `auth` or `log`) should inject `config.Service` to retrieve their startup settings.

---

## Package: pkg/io

### 1. Overview
`pkg/io` provides a filesystem abstraction layer (`Medium`). Its philosophy is to decouple business logic from the `os` package, facilitating easier testing (via mocks) and security (via sandboxing).

### 2. Public API
*   `type Medium`: Interface defining filesystem operations (`Read`, `Write`, `List`, `Stat`, `Open`, `Create`, `Delete`, `Rename`, etc.).
*   `var Local`: A pre-initialized `Medium` for the host root filesystem.
*   `func NewSandboxed(root string) (Medium, error)`: Returns a `Medium` restricted to a specific directory.
*   `type MockMedium`: In-memory implementation of `Medium` for testing.
*   `func NewMockMedium() *MockMedium`: Constructor for the mock.
*   **Helpers**: `Read`, `Write`, `Copy`, `EnsureDir`, `IsFile`, `ReadStream`, `WriteStream` (accept `Medium` as first arg).

### 3. Internal Design
*   **Interface Segregation**: The `Medium` interface mimics the capabilities of `os` and `io/fs` but bundles them into a single dependency.
*   **Mocking**: `MockMedium` uses `map[string]string` for files and `map[string]bool` for directories. It implements manual path logic to simulate filesystem behavior (e.g., verifying a directory is empty before deletion) without touching the disk.
*   **Sandboxing**: The `local` implementation (imported internally) enforces path scoping to prevent traversal attacks when using `NewSandboxed`.

### 4. Dependencies
*   Standard Lib: `io`, `io/fs`, `os`, `path/filepath`, `strings`, `time`.
*   `github.com/host-uk/core/pkg/io/local`: (Implied) The concrete implementation for OS disk access.

### 5. Test Coverage Notes
*   **Mock fidelity**: The `MockMedium` must behave exactly like the OS. E.g., `Rename` should fail if the source doesn't exist; `Delete` should fail if a directory is not empty.
*   **Sandboxing**: Verify that `..` traversal attempts in `NewSandboxed` cannot access files outside the root.

### 6. Integration Points
*   **Universal Dependency**: Used by `log` (rotation), `config` (loading), and `auth` (user DB).
*   **Testing**: Application code should accept `io.Medium` in constructors rather than using `os.Open` directly, enabling unit tests to use `NewMockMedium()`.

---

## Package: pkg/crypt

### 1. Overview
`pkg/crypt` provides "batteries-included," opinionated cryptographic primitives. It abstracts away the complexity of parameter selection (salt length, iteration counts, nonce generation) to prevent misuse of crypto algorithms.

### 2. Public API
*   **Hashing**: `HashPassword` (Argon2id), `VerifyPassword`, `HashBcrypt`, `VerifyBcrypt`.
*   **Symmetric**: `Encrypt`/`Decrypt` (ChaCha20-Poly1305), `EncryptAES`/`DecryptAES` (AES-GCM).
*   **KDF**: `DeriveKey` (Argon2), `DeriveKeyScrypt`, `HKDF`.
*   **Checksums**: `SHA256File`, `SHA512File`, `SHA256Sum`, `SHA512Sum`.
*   **HMAC**: `HMACSHA256`, `HMACSHA512`, `VerifyHMAC`.

### 3. Internal Design
*   **Safe Defaults**: Uses Argon2id for password hashing with tuned parameters (64MB memory, 3 iterations).
*   **Container Format**: Symmetric encryption functions return a concatenated byte slice: `[Salt (16b) | Nonce (Variable) | Ciphertext]`. This ensures the decryption function has everything it needs without separate state management.
*   **Randomness**: Automatically handles salt and nonce generation using `crypto/rand`.

### 4. Dependencies
*   `golang.org/x/crypto`: For Argon2, ChaCha20, HKDF, Scrypt.
*   Standard Lib: `crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256`.

### 5. Test Coverage Notes
*   **Interoperability**: Ensure `Encrypt` output can be read by `Decrypt`.
*   **Tamper Resistance**: manually modifying a byte in the ciphertext or nonce must result in a decryption failure (AuthTag check).
*   **Vectors**: Validate hashing against known test vectors where possible.

### 6. Integration Points
*   **Auth**: Heavily used by `pkg/auth` for password storage and potentially for encrypted user data.
*   **Data Protection**: Any service requiring data at rest encryption should use `crypt.Encrypt`.

---

## Package: pkg/auth

### 1. Overview
`pkg/auth` implements a persistent user identity system based on OpenPGP challenge-response authentication. It supports a unique "Air-Gapped" workflow where challenges and responses are exchanged via files, alongside standard online methods. It manages user lifecycles, sessions, and key storage.

### 2. Public API
*   `type Authenticator`: Main logic controller.
*   `type User`: User metadata struct.
*   `type Session`: Active session token struct.
*   `func New(m io.Medium, opts ...Option) *Authenticator`: Constructor.
*   `func (a *Authenticator) Register(username, password string) (*User, error)`: Creates new user and PGP keys.
*   `func (a *Authenticator) Login(userID, password string) (*Session, error)`: Password-based fallback login.
*   `func (a *Authenticator) CreateChallenge(userID string) (*Challenge, error)`: Starts PGP auth flow.
*   `func (a *Authenticator) ValidateResponse(userID string, signedNonce []byte) (*Session, error)`: Completes PGP auth flow.
*   `func (a *Authenticator) ValidateSession(token string) (*Session, error)`: Checks token validity.
*   `func (a *Authenticator) WriteChallengeFile(userID, path string) error`: For air-gapped flow.
*   `func (a *Authenticator) ReadResponseFile(userID, path string) (*Session, error)`: For air-gapped flow.

### 3. Internal Design
*   **Storage Layout**: Uses a flat-file database approach on `io.Medium`:
    *   `users/{id}.pub`: Public Key.
    *   `users/{id}.key`: Encrypted Private Key.
    *   `users/{id}.lthn`: Password Hash.
    *   `users/{id}.json`: Encrypted metadata.
*   **Identity**: User IDs are hashes of usernames to anonymize storage structure.
*   **Flow**:
    1.  Server generates random nonce.
    2.  Server encrypts nonce with User Public Key.
    3.  User decrypts nonce (client-side) and signs it.
    4.  Server validates signature against User Public Key.

### 4. Dependencies
*   `github.com/host-uk/core/pkg/io`: For user database storage.
*   `github.com/host-uk/core/pkg/crypt/lthn`: (Implied) Specific password hashing.
*   `github.com/host-uk/core/pkg/crypt/pgp`: (Implied) OpenPGP operations.
*   `github.com/host-uk/core/pkg/framework/core`: Error handling.

### 5. Test Coverage Notes
*   **Flow Verification**: Full integration test simulating a client: Register -> Get Challenge -> Decrypt/Sign (Mock Client) -> Validate -> Get Token.
*   **Security**: Ensure `server` user cannot be deleted. Ensure expired sessions are rejected.
*   **Persistence**: Ensure user data survives an `Authenticator` restart (i.e., data is actually written to medium).

### 6. Integration Points
*   **API Gateways**: HTTP handlers would call `ValidateSession` on every request.
*   **CLI Tools**: Would use `WriteChallengeFile`/`ReadResponseFile` for offline authentication.
