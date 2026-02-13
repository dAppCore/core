# Codex Task: Core App — FrankenPHP Native Desktop App

## Context

You are working on `cmd/core-app/` inside the `host-uk/core` Go monorepo. This is a **working** native desktop application that embeds the PHP runtime (FrankenPHP) inside a Wails v3 window. A single 53MB binary runs Laravel 12 with Livewire 4, Octane worker mode, and SQLite — no Docker, no php-fpm, no nginx, no external dependencies.

**It already builds and runs.** Your job is to refine, not rebuild.

## Architecture

```
Wails v3 WebView (native window)
    |
    | AssetOptions.Handler → http.Handler
    v
FrankenPHP (CGO, PHP 8.4 ZTS runtime)
    |
    | ServeHTTP() → Laravel public/index.php
    v
Laravel 12 (Octane worker mode, 2 workers)
    ├── Livewire 4 (server-rendered reactivity)
    ├── SQLite (~/Library/Application Support/core-app/)
    └── Native Bridge (localhost HTTP API for PHP→Go calls)
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Wails app entry, system tray, window config |
| `handler.go` | PHPHandler — FrankenPHP init, Octane worker mode, try_files URL resolution |
| `embed.go` | `//go:embed all:laravel` + extraction to temp dir |
| `env.go` | Persistent data dir, .env generation, APP_KEY management |
| `app_service.go` | Wails service bindings (version, data dir, window management) |
| `native_bridge.go` | PHP→Go HTTP bridge on localhost (random port) |
| `laravel/` | Full Laravel 12 skeleton (vendor excluded from git, built via `composer install`) |

## Build Requirements

- **PHP 8.4 ZTS**: `brew install shivammathur/php/php@8.4-zts`
- **Go 1.25+** with CGO enabled
- **Build tags**: `-tags nowatcher` (FrankenPHP's watcher needs libwatcher-c, skip it)
- **ZTS php-config**: Must use `/opt/homebrew/opt/php@8.4-zts/bin/php-config` (NOT the default php-config which may point to non-ZTS PHP)

```bash
# Install Laravel deps (one-time)
cd laravel && composer install --no-dev --optimize-autoloader

# Build
ZTS_PHP_CONFIG=/opt/homebrew/opt/php@8.4-zts/bin/php-config
CGO_ENABLED=1 \
CGO_CFLAGS="$($ZTS_PHP_CONFIG --includes)" \
CGO_LDFLAGS="-L/opt/homebrew/opt/php@8.4-zts/lib $($ZTS_PHP_CONFIG --ldflags) $($ZTS_PHP_CONFIG --libs)" \
go build -tags nowatcher -o ../../bin/core-app .
```

## Known Patterns & Gotchas

1. **FrankenPHP can't serve from embed.FS** — must extract to temp dir, symlink `storage/` to persistent data dir
2. **WithWorkers API (v1.5.0)**: `WithWorkers(name, fileName string, num int, env map[string]string, watch []string)` — 5 positional args, NOT variadic
3. **Worker mode needs Octane**: Workers point at `vendor/laravel/octane/bin/frankenphp-worker.php` with `APP_BASE_PATH` and `FRANKENPHP_WORKER=1` env vars
4. **Paths with spaces**: macOS `~/Library/Application Support/` has a space — ALL .env values with paths MUST be quoted
5. **URL resolution**: FrankenPHP doesn't auto-resolve `/` → `/index.php` — the Go handler implements try_files logic
6. **Auto-migration**: `AppServiceProvider::boot()` runs `migrate --force` wrapped in try/catch (must not fail during composer operations)
7. **Vendor dir**: Excluded from git (`.gitignore`), built at dev time via `composer install`, embedded by `//go:embed all:laravel` at build time

## Coding Standards

- **UK English**: colour, organisation, centre
- **PHP**: `declare(strict_types=1)` in every file, full type hints, PSR-12 via Pint
- **Go**: Standard Go conventions, error wrapping with `fmt.Errorf("context: %w", err)`
- **License**: EUPL-1.2
- **Testing**: Pest syntax for PHP (not PHPUnit)

## Tasks for Codex

### Priority 1: Code Quality
- [ ] Review all Go files for error handling consistency
- [ ] Ensure handler.go's try_files logic handles edge cases (double slashes, encoded paths, path traversal)
- [ ] Add Go tests for PHPHandler URL resolution (unit tests, no FrankenPHP needed)
- [ ] Add Go tests for env.go (resolveDataDir, writeEnvFile, loadOrGenerateAppKey)

### Priority 2: Laravel Polish
- [ ] Add `config/octane.php` with FrankenPHP server config
- [ ] Update welcome view to show migration status (table count from SQLite)
- [ ] Add a second Livewire component (e.g., todo list) to prove full CRUD with SQLite
- [ ] Add proper error page views (404, 500) styled to match the dark theme

### Priority 3: Build Hardening
- [ ] Verify the Taskfile.yml tasks work end-to-end (`task app:setup && task app:composer && task app:build`)
- [ ] Add `.gitignore` entries for build artifacts (`bin/core-app`, temp dirs)
- [ ] Ensure `go.work` and `go.mod` are consistent

## CRITICAL WARNINGS

- **DO NOT push to GitHub** — GitHub remotes have been removed deliberately. The host-uk org is flagged.
- **DO NOT add GitHub as a remote** — Forge (forge.lthn.io / git.lthn.ai) is the source of truth.
- **DO NOT modify files outside `cmd/core-app/`** — This is a workspace module, keep changes scoped.
- **DO NOT remove the `-tags nowatcher` build flag** — It will fail without libwatcher-c.
- **DO NOT change the PHP-ZTS path** — It must be the ZTS variant, not the default Homebrew PHP.
