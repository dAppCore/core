# Plug Package Extraction Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract `app/Plug/*` from the Laravel app into 8 independent `core/php-plug-*` repos on forge, with contracts moving to `core/php` and all namespaces aligned to `Core\Plug\*`.

**Architecture:** The app's `Plug\` namespace was a flattened copy of what should be `Core\Plug\*` in the framework. Contracts (shared interfaces) go into `core/php`. Each category (Social, Web3, etc.) becomes its own composer package. The app's `app/Plug/` directory is deleted entirely.

**Tech Stack:** PHP 8.2+, Composer, Laravel, PSR-4 autoloading, Forge SSH repos

---

### Task 1: Move contracts into `core/php`

**Context:** The 8 shared interfaces that all providers implement currently live in `app/Plug/Contract/` with `Plug\Contract\` namespace. They belong in the framework at `core/php/src/Plug/Contract/` with `Core\Plug\Contract\` namespace.

**Files:**
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Authenticable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Commentable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Deletable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Listable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/MediaUploadable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Postable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Readable.php`
- Create: `/Users/snider/Code/core/php/src/Plug/Contract/Refreshable.php`
- Modify: `/Users/snider/Code/core/php/src/Plug/Registry.php` — add `register()` method

**Step 1:** Copy contracts from app to core/php

```bash
mkdir -p /Users/snider/Code/core/php/src/Plug/Contract
cp /Users/snider/Code/lab/host.uk.com/app/Plug/Contract/*.php /Users/snider/Code/core/php/src/Plug/Contract/
```

**Step 2:** Update namespace in all 8 contracts from `Plug\Contract` → `Core\Plug\Contract` and `use Plug\Response` → `use Core\Plug\Response`

```bash
cd /Users/snider/Code/core/php/src/Plug/Contract
sed -i '' 's/namespace Plug\\Contract;/namespace Core\\Plug\\Contract;/' *.php
sed -i '' 's/use Plug\\Response;/use Core\\Plug\\Response;/' *.php
```

**Step 3:** Add `register()` method to Registry so packages can self-register instead of relying on filesystem scanning

In `/Users/snider/Code/core/php/src/Plug/Registry.php`, add after the `discover()` method:

```php
/**
 * Register a provider programmatically.
 *
 * Used by plug packages to self-register without filesystem scanning.
 */
public function register(string $identifier, string $category, string $name, string $namespace): void
{
    $this->providers[$identifier] = [
        'category' => $category,
        'name' => $name,
        'namespace' => $namespace,
        'path' => null,
    ];
}
```

**Step 4:** Verify

```bash
cd /Users/snider/Code/core/php
grep -r "namespace Core\\Plug\\Contract" src/Plug/Contract/
# Should show 8 files with Core\Plug\Contract namespace
```

**Step 5:** Commit

```bash
cd /Users/snider/Code/core/php
git add src/Plug/Contract/ src/Plug/Registry.php
git commit -m "feat(plug): add shared contracts and Registry::register()"
```

---

### Task 2: Create `core/php-plug-social`

**Context:** The largest package — 8 social media providers (LinkedIn, Meta, Pinterest, Reddit, TikTok, Twitter, VK, YouTube), ~48 files. This is the template for all other packages.

**Files:**
- Source: `/Users/snider/Code/lab/host.uk.com/app/Plug/Social/` (all files)
- Create: `/Users/snider/Code/core/php-plug-social/` (new repo)

**Step 1:** Create repo on forge

```bash
# Create repo via Forgejo API
curl -X POST "https://forge.lthn.ai/api/v1/orgs/core/repos" \
  -H "Authorization: token $(cat ~/.config/forge/token)" \
  -H "Content-Type: application/json" \
  -d '{"name":"php-plug-social","description":"Social media provider integrations (Twitter, Meta, LinkedIn, etc.)","private":true,"auto_init":true}'
```

**Step 2:** Clone and set up directory structure

```bash
cd /Users/snider/Code/core
git clone ssh://git@forge.lthn.ai:2223/core/php-plug-social.git
cd php-plug-social
mkdir -p src
```

**Step 3:** Create `composer.json`

```json
{
    "name": "core/php-plug-social",
    "description": "Social media provider integrations for the Plug framework",
    "type": "library",
    "license": "EUPL-1.2",
    "require": {
        "php": "^8.2",
        "core/php": "^1.0"
    },
    "autoload": {
        "psr-4": {
            "Core\\Plug\\Social\\": "src/"
        }
    },
    "minimum-stability": "dev",
    "prefer-stable": true,
    "repositories": [
        {
            "type": "vcs",
            "url": "ssh://git@forge.lthn.ai:2223/core/php.git"
        }
    ]
}
```

**Step 4:** Copy provider files

```bash
cp -r /Users/snider/Code/lab/host.uk.com/app/Plug/Social/* src/
```

**Step 5:** Update namespaces in all files — two changes per file:

1. `namespace Plug\Social\{Provider}` → `namespace Core\Plug\Social\{Provider}`
2. `use Plug\{...}` → `use Core\Plug\{...}`

```bash
cd /Users/snider/Code/core/php-plug-social
# Update namespace declarations
find src -name "*.php" -exec sed -i '' 's/namespace Plug\\Social\\/namespace Core\\Plug\\Social\\/' {} \;
# Update use statements for framework classes
find src -name "*.php" -exec sed -i '' 's/use Plug\\Concern\\/use Core\\Plug\\Concern\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Contract\\/use Core\\Plug\\Contract\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Response;/use Core\\Plug\\Response;/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Enum\\/use Core\\Plug\\Enum\\/' {} \;
# Update internal cross-references (e.g., new Media in Twitter\Post)
find src -name "*.php" -exec sed -i '' 's/use Plug\\Social\\/use Core\\Plug\\Social\\/' {} \;
```

**Step 6:** Verify namespaces are correct

```bash
grep -r "namespace " src/ | head -20
# Every line should show Core\Plug\Social\{Provider}

grep -r "use Plug\\\\" src/
# Should return nothing — all should be Core\Plug\*
```

**Step 7:** Commit and push

```bash
cd /Users/snider/Code/core/php-plug-social
git add .
git commit -m "feat: extract social providers from app/Plug/Social"
git push origin main
```

---

### Task 3: Create `core/php-plug-web3`

**Same pattern as Task 2.** 6 providers (Bluesky, Farcaster, Lemmy, Mastodon, Nostr, Threads), ~28 files.

**Step 1:** Create repo on forge (name: `php-plug-web3`, description: "Decentralised/Web3 provider integrations")

**Step 2:** Clone, create `composer.json` (same template, change name/description/namespace to `Core\\Plug\\Web3\\`)

**Step 3:** Copy files

```bash
cp -r /Users/snider/Code/lab/host.uk.com/app/Plug/Web3/* src/
```

**Step 4:** Update namespaces

```bash
find src -name "*.php" -exec sed -i '' 's/namespace Plug\\Web3\\/namespace Core\\Plug\\Web3\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Concern\\/use Core\\Plug\\Concern\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Contract\\/use Core\\Plug\\Contract\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Response;/use Core\\Plug\\Response;/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Enum\\/use Core\\Plug\\Enum\\/' {} \;
find src -name "*.php" -exec sed -i '' 's/use Plug\\Web3\\/use Core\\Plug\\Web3\\/' {} \;
```

**Step 5:** Verify, commit, push

---

### Task 4: Create `core/php-plug-content`

**Same pattern.** 4 providers (Devto, Hashnode, Medium, Wordpress), ~18 files.

**Namespace:** `Core\\Plug\\Content\\`

**Copy:** `app/Plug/Content/*` → `src/`

**Sed replacements:** Same pattern — `Plug\Content\` → `Core\Plug\Content\`, plus framework imports.

---

### Task 5: Create `core/php-plug-chat`

**Same pattern.** 3 providers (Discord, Slack, Telegram), ~14 files.

**Namespace:** `Core\\Plug\\Chat\\`

**Copy:** `app/Plug/Chat/*` → `src/`

**Note:** Chat providers use `Postable` but not `ManagesTokens` in some cases (Slack, Discord use webhook-style auth). Verify `use` statements are correct after sed.

---

### Task 6: Create `core/php-plug-business`

**Same pattern.** 1 provider (GoogleMyBusiness), ~5 files.

**Namespace:** `Core\\Plug\\Business\\`

**Copy:** `app/Plug/Business/*` → `src/`

---

### Task 7: Create `core/php-plug-cdn`

**Different from social/web3 — includes its own domain-specific contracts.**

**Files:**
- Source: `/Users/snider/Code/lab/host.uk.com/app/Plug/Cdn/`
- Includes: `Contract/HasStats.php`, `Contract/Purgeable.php`, `CdnManager.php`, `Bunny/Purge.php`, `Bunny/Stats.php`

**Namespace:** `Core\\Plug\\Cdn\\`

**Additional sed:** `use Plug\\Cdn\\Contract\\` → `use Core\\Plug\\Cdn\\Contract\\`

---

### Task 8: Create `core/php-plug-storage`

**Same pattern as CDN — includes domain-specific contracts.**

**Files:**
- Source: `/Users/snider/Code/lab/host.uk.com/app/Plug/Storage/`
- Includes: `Contract/{Browseable,Deletable,Downloadable,Uploadable}.php`, `StorageManager.php`, `Bunny/{Browse,Delete,Download,Upload,VBucket}.php`

**Namespace:** `Core\\Plug\\Storage\\`

**Additional sed:** `use Plug\\Storage\\Contract\\` → `use Core\\Plug\\Storage\\Contract\\`

---

### Task 9: Create `core/php-plug-stock`

**Same pattern.** 1 provider (Unsplash), ~6 files. Has a Jobs subdirectory.

**Namespace:** `Core\\Plug\\Stock\\`

**Copy:** `app/Plug/Stock/*` → `src/`

**Additional sed:** `use Plug\\Stock\\` → `use Core\\Plug\\Stock\\` (for internal cross-references like `TriggerDownload` → `Download`)

---

### Task 10: Clean up the Laravel app

**Context:** After all packages are extracted, remove the flattened code from the app and wire up the new packages.

**Files:**
- Delete: `/Users/snider/Code/lab/host.uk.com/app/Plug/` (entire directory)
- Modify: `/Users/snider/Code/lab/host.uk.com/composer.json` — add 8 new requires + repos, remove `Plug\\` autoload
- Modify: `/Users/snider/Code/lab/host.uk.com/app/Boot.php` — remove `Plug\Boot` reference

**Step 1:** Remove `app/Plug/` entirely

```bash
rm -rf /Users/snider/Code/lab/host.uk.com/app/Plug
```

**Step 2:** Update `composer.json`

Remove from `autoload.psr-4`:
```json
"Plug\\": "app/Plug/",
```

Add to `require`:
```json
"core/php-plug-social": "dev-main",
"core/php-plug-web3": "dev-main",
"core/php-plug-content": "dev-main",
"core/php-plug-chat": "dev-main",
"core/php-plug-business": "dev-main",
"core/php-plug-cdn": "dev-main",
"core/php-plug-storage": "dev-main",
"core/php-plug-stock": "dev-main"
```

Add to `repositories`:
```json
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-social.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-web3.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-content.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-chat.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-business.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-cdn.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-storage.git"},
{"type": "vcs", "url": "ssh://git@forge.lthn.ai:2223/core/php-plug-stock.git"}
```

**Step 3:** Update `app/Boot.php` — remove the `Plug\Boot::class` provider registration

**Step 4:** Run composer update

```bash
cd /Users/snider/Code/lab/host.uk.com
composer update
```

**Step 5:** Clear caches and verify

```bash
php artisan config:clear
php artisan cache:clear
php artisan route:clear
```

**Step 6:** Commit

```bash
cd /Users/snider/Code/lab/host.uk.com
git add -A
git commit -m "refactor: replace app/Plug with core/php-plug-* packages"
```

---

### Task 11: Push `core/php` contracts and verify

**Step 1:** Push core/php with new contracts

```bash
cd /Users/snider/Code/core/php
git push origin main
```

**Step 2:** Verify all packages resolve

```bash
cd /Users/snider/Code/lab/host.uk.com
composer update --dry-run 2>&1 | head -30
```

**Step 3:** Run the app

```bash
# Visit lthn.test in browser — should load without errors
php artisan tinker --execute="app('plug.registry')->identifiers()"
```
