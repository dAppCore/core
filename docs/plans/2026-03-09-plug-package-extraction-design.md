# Plug Package Extraction Design

**Goal:** Extract `app/Plug/*` categories from the Laravel app into 8 independent `core/php-plug-*` packages on forge, restoring the original package split that was flattened during the GitHub to Forge migration. Move shared contracts into `core/php` as the framework base.

**Pattern:** Same as CoreGO — `core/php` is the foundation framework (like `core/go`), with domain packages depending on it.

---

## Current State

**Framework (`core/php` `src/Plug/`)** — 7 files:
- `Boot.php` — Registry singleton registration
- `Registry.php` — Auto-discovery, capability checking
- `Response.php` — Standardised operation response
- `Concern/BuildsResponse.php` — Response builder trait
- `Concern/ManagesTokens.php` — OAuth token management trait
- `Concern/UsesHttp.php` — HTTP client helpers trait
- `Enum/Status.php` — OK, UNAUTHORIZED, RATE_LIMITED, etc.

**App (`app/Plug/`)** — 140 files across 10 categories:

| Category | Providers | Files |
|----------|-----------|-------|
| Contract | 8 interfaces (Authenticable, Postable, Deletable, Readable, Commentable, Listable, MediaUploadable, Refreshable) | 8 |
| Social | LinkedIn, Meta, Pinterest, Reddit, TikTok, Twitter, VK, YouTube | ~48 |
| Web3 | Bluesky, Farcaster, Lemmy, Mastodon, Nostr, Threads | ~28 |
| Content | Devto, Hashnode, Medium, Wordpress | ~18 |
| Chat | Discord, Slack, Telegram | ~14 |
| Business | GoogleMyBusiness | ~5 |
| Cdn | Bunny (Purge, Stats, CdnManager) + contracts | 5 |
| Storage | Bunny (Browse, Delete, Download, Upload, VBucket, StorageManager) + contracts | 8 |
| Stock | Unsplash (Search, Photo, Collection, Download, Exception, Jobs) | 6 |

## Namespace Alignment

The app currently uses `Plug\` namespace while core/php uses `Core\Plug\`. This extraction aligns everything under `Core\Plug\*` to match the framework convention.

## Target State

### 1. Contracts move into `core/php`

`src/Plug/Contract/` gains 8 interfaces from the app:

```
core/php/src/Plug/Contract/
├── Authenticable.php
├── Commentable.php
├── Deletable.php
├── Listable.php
├── MediaUploadable.php
├── Postable.php
├── Readable.php
└── Refreshable.php
```

Namespace: `Core\Plug\Contract\` (aligned with framework).

### 2. Eight new packages on forge

Each package has its own repo at `forge.lthn.ai/core/php-plug-{name}`.

```
core/php-plug-social/
├── composer.json       # requires core/php
├── CLAUDE.md
└── src/
    ├── LinkedIn/{Auth,Post,Delete,Media,Pages,Read}.php
    ├── Meta/{Auth,Post,Delete,Media,Pages,Read}.php
    ├── Pinterest/{Auth,Post,Delete,Media,Boards,Read}.php
    ├── Reddit/{Auth,Post,Delete,Media,Read,Subreddits}.php
    ├── TikTok/{Auth,Post,Read}.php
    ├── Twitter/{Auth,Post,Delete,Media,Read}.php
    ├── VK/{Auth,Post,Delete,Media,Groups,Read}.php
    └── YouTube/{Auth,Post,Delete,Comment,Read}.php
```

PSR-4 mapping per package:

| Package | Composer Name | Namespace | Autoload |
|---------|--------------|-----------|----------|
| `core/php-plug-social` | `core/php-plug-social` | `Core\Plug\Social\` | `src/` |
| `core/php-plug-web3` | `core/php-plug-web3` | `Core\Plug\Web3\` | `src/` |
| `core/php-plug-content` | `core/php-plug-content` | `Core\Plug\Content\` | `src/` |
| `core/php-plug-chat` | `core/php-plug-chat` | `Core\Plug\Chat\` | `src/` |
| `core/php-plug-business` | `core/php-plug-business` | `Core\Plug\Business\` | `src/` |
| `core/php-plug-cdn` | `core/php-plug-cdn` | `Core\Plug\Cdn\` | `src/` |
| `core/php-plug-storage` | `core/php-plug-storage` | `Core\Plug\Storage\` | `src/` |
| `core/php-plug-stock` | `core/php-plug-stock` | `Core\Plug\Stock\` | `src/` |

Cdn and Storage packages include their own sub-contracts (`Core\Plug\Cdn\Contract\*`, `Core\Plug\Storage\Contract\*`) since those are domain-specific (Purgeable, HasStats, Browseable, Uploadable, etc.) rather than shared Plug contracts.

### 3. Registry update

`Registry::discover()` currently scans `__DIR__` for category subdirectories. After extraction, providers live in composer-installed paths. Two options:

**Chosen:** Each package registers its providers via a service provider that calls `Registry::register()`. The Registry gains a `register(string $identifier, array $meta)` method. No filesystem scanning needed.

### 4. App cleanup

`app/Plug/` is deleted entirely:
- `Boot.php` — redundant (core/php has one)
- `Registry.php` — redundant (core/php has one)
- `Response.php` — redundant (core/php has one)
- `Contract/` — moved to core/php
- All category dirs — moved to packages

The app's `composer.json` gains the 8 new packages as dependencies.

## Dependency Graph

```
core/php (framework)
├── src/Plug/Contract/*        ← shared interfaces
├── src/Plug/Registry.php      ← provider registry
├── src/Plug/Response.php      ← standardised response
├── src/Plug/Concern/*         ← shared traits
└── src/Plug/Enum/Status.php   ← status enum

core/php-plug-social    ─┐
core/php-plug-web3      ─┤
core/php-plug-content   ─┤
core/php-plug-chat      ─┤ all depend on core/php
core/php-plug-business  ─┤
core/php-plug-cdn       ─┤
core/php-plug-storage   ─┤
core/php-plug-stock     ─┘
```

## Namespace Mapping

All provider code is renamed from `Plug\*` to `Core\Plug\*`:

```php
// Before (in app): Plug\Social\Twitter\Post
// After (in package): Core\Plug\Social\Twitter\Post

// Before: use Plug\Contract\Postable;
// After:  use Core\Plug\Contract\Postable;

// Before: use Plug\Concern\UsesHttp;
// After:  use Core\Plug\Concern\UsesHttp;
```

The app's `"Plug\\" => "app/Plug/"` autoload entry is removed entirely.

## Composer Repository Config

Each package uses forge's Composer repository:

```json
{
    "type": "vcs",
    "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-social.git"
}
```
