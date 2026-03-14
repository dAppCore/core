# AltumCode Update Checker Implementation Plan

> **Note:** Layer 1 (Tasks 1-2, 4: version checking + seeder + sync command) is implemented and documented at `docs/docs/php/packages/uptelligence.md`. Task 3 (Claude Code browser skill for Layer 2 downloads) is NOT yet implemented.

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add AltumCode product + plugin version checking to uptelligence, and create a Claude Code skill for browser-automated downloads from LemonSqueezy and CodeCanyon.

**Architecture:** Extend the existing `VendorUpdateCheckerService` to handle `PLATFORM_ALTUM` vendors via 5 public HTTP endpoints. Seed the vendors table with all 4 products and 13 plugins. Create a Claude Code plugin skill that uses Playwright (LemonSqueezy) and Chrome (CodeCanyon) to download updates.

**Tech Stack:** PHP 8.4, Laravel, Pest, Claude Code plugins (Playwright MCP + Chrome MCP)

---

### Task 1: Add AltumCode check to VendorUpdateCheckerService

**Files:**
- Modify: `/Users/snider/Code/core/php-uptelligence/Services/VendorUpdateCheckerService.php`
- Test: `/Users/snider/Code/core/php-uptelligence/tests/Unit/AltumCodeCheckerTest.php`

**Step 1: Write the failing test**

Create `/Users/snider/Code/core/php-uptelligence/tests/Unit/AltumCodeCheckerTest.php`:

```php
<?php

declare(strict_types=1);

use Core\Mod\Uptelligence\Models\Vendor;
use Core\Mod\Uptelligence\Services\VendorUpdateCheckerService;
use Illuminate\Support\Facades\Http;

beforeEach(function () {
    $this->service = app(VendorUpdateCheckerService::class);
});

it('checks altum product version via info.php', function () {
    Http::fake([
        'https://66analytics.com/info.php' => Http::response([
            'latest_release_version' => '66.0.0',
            'latest_release_version_code' => 6600,
        ]),
    ]);

    $vendor = Vendor::factory()->create([
        'slug' => '66analytics',
        'name' => '66analytics',
        'source_type' => Vendor::SOURCE_LICENSED,
        'plugin_platform' => Vendor::PLATFORM_ALTUM,
        'current_version' => '65.0.0',
        'is_active' => true,
    ]);

    $result = $this->service->checkVendor($vendor);

    expect($result['status'])->toBe('success')
        ->and($result['current'])->toBe('65.0.0')
        ->and($result['latest'])->toBe('66.0.0')
        ->and($result['has_update'])->toBeTrue();
});

it('reports no update when altum product is current', function () {
    Http::fake([
        'https://66analytics.com/info.php' => Http::response([
            'latest_release_version' => '65.0.0',
            'latest_release_version_code' => 6500,
        ]),
    ]);

    $vendor = Vendor::factory()->create([
        'slug' => '66analytics',
        'name' => '66analytics',
        'source_type' => Vendor::SOURCE_LICENSED,
        'plugin_platform' => Vendor::PLATFORM_ALTUM,
        'current_version' => '65.0.0',
        'is_active' => true,
    ]);

    $result = $this->service->checkVendor($vendor);

    expect($result['has_update'])->toBeFalse();
});

it('checks altum plugin versions via plugins-versions endpoint', function () {
    Http::fake([
        'https://dev.altumcode.com/plugins-versions' => Http::response([
            'affiliate' => ['version' => '2.0.1'],
            'teams' => ['version' => '3.0.0'],
        ]),
    ]);

    $vendor = Vendor::factory()->create([
        'slug' => 'altum-plugin-affiliate',
        'name' => 'Affiliate Plugin',
        'source_type' => Vendor::SOURCE_PLUGIN,
        'plugin_platform' => Vendor::PLATFORM_ALTUM,
        'current_version' => '2.0.0',
        'is_active' => true,
    ]);

    $result = $this->service->checkVendor($vendor);

    expect($result['status'])->toBe('success')
        ->and($result['latest'])->toBe('2.0.1')
        ->and($result['has_update'])->toBeTrue();
});

it('handles altum info.php timeout gracefully', function () {
    Http::fake([
        'https://66analytics.com/info.php' => Http::response('', 500),
    ]);

    $vendor = Vendor::factory()->create([
        'slug' => '66analytics',
        'name' => '66analytics',
        'source_type' => Vendor::SOURCE_LICENSED,
        'plugin_platform' => Vendor::PLATFORM_ALTUM,
        'current_version' => '65.0.0',
        'is_active' => true,
    ]);

    $result = $this->service->checkVendor($vendor);

    expect($result['status'])->toBe('error')
        ->and($result['has_update'])->toBeFalse();
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=AltumCodeChecker`
Expected: FAIL — altum vendors still hit `skipCheck()`

**Step 3: Write minimal implementation**

In `/Users/snider/Code/core/php-uptelligence/Services/VendorUpdateCheckerService.php`, modify `checkVendor()` to route altum vendors:

```php
public function checkVendor(Vendor $vendor): array
{
    $result = match (true) {
        $this->isAltumPlatform($vendor) && $vendor->isLicensed() => $this->checkAltumProduct($vendor),
        $this->isAltumPlatform($vendor) && $vendor->isPlugin() => $this->checkAltumPlugin($vendor),
        $vendor->isOss() && $this->isGitHubUrl($vendor->git_repo_url) => $this->checkGitHub($vendor),
        $vendor->isOss() && $this->isGiteaUrl($vendor->git_repo_url) => $this->checkGitea($vendor),
        default => $this->skipCheck($vendor),
    };

    // ... rest unchanged
}
```

Add the three new methods:

```php
/**
 * Check if vendor is on the AltumCode platform.
 */
protected function isAltumPlatform(Vendor $vendor): bool
{
    return $vendor->plugin_platform === Vendor::PLATFORM_ALTUM;
}

/**
 * AltumCode product info endpoint mapping.
 */
protected function getAltumProductInfoUrl(Vendor $vendor): ?string
{
    $urls = [
        '66analytics' => 'https://66analytics.com/info.php',
        '66biolinks' => 'https://66biolinks.com/info.php',
        '66pusher' => 'https://66pusher.com/info.php',
        '66socialproof' => 'https://66socialproof.com/info.php',
    ];

    return $urls[$vendor->slug] ?? null;
}

/**
 * Check an AltumCode product for updates via its info.php endpoint.
 */
protected function checkAltumProduct(Vendor $vendor): array
{
    $url = $this->getAltumProductInfoUrl($vendor);
    if (! $url) {
        return $this->errorResult("No info.php URL mapped for {$vendor->slug}");
    }

    try {
        $response = Http::timeout(5)->get($url);

        if (! $response->successful()) {
            return $this->errorResult("AltumCode info.php returned {$response->status()}");
        }

        $data = $response->json();
        $latestVersion = $data['latest_release_version'] ?? null;

        if (! $latestVersion) {
            return $this->errorResult('No version in info.php response');
        }

        return $this->buildResult(
            vendor: $vendor,
            latestVersion: $this->normaliseVersion($latestVersion),
            releaseInfo: [
                'version_code' => $data['latest_release_version_code'] ?? null,
                'source' => $url,
            ]
        );
    } catch (\Exception $e) {
        return $this->errorResult("AltumCode check failed: {$e->getMessage()}");
    }
}

/**
 * Check an AltumCode plugin for updates via the central plugins-versions endpoint.
 */
protected function checkAltumPlugin(Vendor $vendor): array
{
    try {
        $allPlugins = $this->getAltumPluginVersions();

        if ($allPlugins === null) {
            return $this->errorResult('Failed to fetch AltumCode plugin versions');
        }

        // Extract the plugin_id from the vendor slug (strip 'altum-plugin-' prefix)
        $pluginId = str_replace('altum-plugin-', '', $vendor->slug);

        if (! isset($allPlugins[$pluginId])) {
            return $this->errorResult("Plugin '{$pluginId}' not found in AltumCode registry");
        }

        $latestVersion = $allPlugins[$pluginId]['version'] ?? null;

        return $this->buildResult(
            vendor: $vendor,
            latestVersion: $this->normaliseVersion($latestVersion),
            releaseInfo: ['source' => 'dev.altumcode.com/plugins-versions']
        );
    } catch (\Exception $e) {
        return $this->errorResult("AltumCode plugin check failed: {$e->getMessage()}");
    }
}

/**
 * Fetch all AltumCode plugin versions (cached for 1 hour within a check run).
 */
protected ?array $altumPluginVersionsCache = null;

protected function getAltumPluginVersions(): ?array
{
    if ($this->altumPluginVersionsCache !== null) {
        return $this->altumPluginVersionsCache;
    }

    $response = Http::timeout(5)->get('https://dev.altumcode.com/plugins-versions');

    if (! $response->successful()) {
        return null;
    }

    $this->altumPluginVersionsCache = $response->json();

    return $this->altumPluginVersionsCache;
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=AltumCodeChecker`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/php-uptelligence
git add Services/VendorUpdateCheckerService.php tests/Unit/AltumCodeCheckerTest.php
git commit -m "feat: add AltumCode product + plugin version checking

Extends VendorUpdateCheckerService to check AltumCode products via
their info.php endpoints and plugins via dev.altumcode.com/plugins-versions.
No auth required — all endpoints are public.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 2: Seed AltumCode vendors

**Files:**
- Create: `/Users/snider/Code/core/php-uptelligence/database/seeders/AltumCodeVendorSeeder.php`
- Test: `/Users/snider/Code/core/php-uptelligence/tests/Unit/AltumCodeVendorSeederTest.php`

**Step 1: Write the failing test**

Create `/Users/snider/Code/core/php-uptelligence/tests/Unit/AltumCodeVendorSeederTest.php`:

```php
<?php

declare(strict_types=1);

use Core\Mod\Uptelligence\Models\Vendor;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

it('seeds 4 altum products', function () {
    $this->artisan('db:seed', ['--class' => 'Core\\Mod\\Uptelligence\\Database\\Seeders\\AltumCodeVendorSeeder']);

    expect(Vendor::where('source_type', Vendor::SOURCE_LICENSED)
        ->where('plugin_platform', Vendor::PLATFORM_ALTUM)
        ->count()
    )->toBe(4);
});

it('seeds 13 altum plugins', function () {
    $this->artisan('db:seed', ['--class' => 'Core\\Mod\\Uptelligence\\Database\\Seeders\\AltumCodeVendorSeeder']);

    expect(Vendor::where('source_type', Vendor::SOURCE_PLUGIN)
        ->where('plugin_platform', Vendor::PLATFORM_ALTUM)
        ->count()
    )->toBe(13);
});

it('is idempotent', function () {
    $this->artisan('db:seed', ['--class' => 'Core\\Mod\\Uptelligence\\Database\\Seeders\\AltumCodeVendorSeeder']);
    $this->artisan('db:seed', ['--class' => 'Core\\Mod\\Uptelligence\\Database\\Seeders\\AltumCodeVendorSeeder']);

    expect(Vendor::where('plugin_platform', Vendor::PLATFORM_ALTUM)->count())->toBe(17);
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=AltumCodeVendorSeeder`
Expected: FAIL — seeder class not found

**Step 3: Write minimal implementation**

Create `/Users/snider/Code/core/php-uptelligence/database/seeders/AltumCodeVendorSeeder.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Mod\Uptelligence\Database\Seeders;

use Core\Mod\Uptelligence\Models\Vendor;
use Illuminate\Database\Seeder;

class AltumCodeVendorSeeder extends Seeder
{
    public function run(): void
    {
        $products = [
            ['slug' => '66analytics', 'name' => '66analytics', 'vendor_name' => 'AltumCode', 'current_version' => '65.0.0'],
            ['slug' => '66biolinks', 'name' => '66biolinks', 'vendor_name' => 'AltumCode', 'current_version' => '65.0.0'],
            ['slug' => '66pusher', 'name' => '66pusher', 'vendor_name' => 'AltumCode', 'current_version' => '65.0.0'],
            ['slug' => '66socialproof', 'name' => '66socialproof', 'vendor_name' => 'AltumCode', 'current_version' => '65.0.0'],
        ];

        foreach ($products as $product) {
            Vendor::updateOrCreate(
                ['slug' => $product['slug']],
                [
                    ...$product,
                    'source_type' => Vendor::SOURCE_LICENSED,
                    'plugin_platform' => Vendor::PLATFORM_ALTUM,
                    'is_active' => true,
                ]
            );
        }

        $plugins = [
            ['slug' => 'altum-plugin-affiliate', 'name' => 'Affiliate Plugin', 'current_version' => '2.0.0'],
            ['slug' => 'altum-plugin-push-notifications', 'name' => 'Push Notifications Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-teams', 'name' => 'Teams Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-pwa', 'name' => 'PWA Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-image-optimizer', 'name' => 'Image Optimizer Plugin', 'current_version' => '3.1.0'],
            ['slug' => 'altum-plugin-email-shield', 'name' => 'Email Shield Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-dynamic-og-images', 'name' => 'Dynamic OG Images Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-offload', 'name' => 'Offload & CDN Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-payment-blocks', 'name' => 'Payment Blocks Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-ultimate-blocks', 'name' => 'Ultimate Blocks Plugin', 'current_version' => '9.1.0'],
            ['slug' => 'altum-plugin-pro-blocks', 'name' => 'Pro Blocks Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-pro-notifications', 'name' => 'Pro Notifications Plugin', 'current_version' => '1.0.0'],
            ['slug' => 'altum-plugin-aix', 'name' => 'AIX Plugin', 'current_version' => '1.0.0'],
        ];

        foreach ($plugins as $plugin) {
            Vendor::updateOrCreate(
                ['slug' => $plugin['slug']],
                [
                    ...$plugin,
                    'vendor_name' => 'AltumCode',
                    'source_type' => Vendor::SOURCE_PLUGIN,
                    'plugin_platform' => Vendor::PLATFORM_ALTUM,
                    'is_active' => true,
                ]
            );
        }
    }
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=AltumCodeVendorSeeder`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/php-uptelligence
git add database/seeders/AltumCodeVendorSeeder.php tests/Unit/AltumCodeVendorSeederTest.php
git commit -m "feat: seed AltumCode vendors — 4 products + 13 plugins

Idempotent seeder using updateOrCreate. Products are SOURCE_LICENSED,
plugins are SOURCE_PLUGIN, all PLATFORM_ALTUM. Version numbers will
need updating to match actual deployed versions.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 3: Create Claude Code plugin skill for downloads

**Files:**
- Create: `/Users/snider/.claude/plugins/altum-updater/plugin.json`
- Create: `/Users/snider/.claude/plugins/altum-updater/skills/update-altum.md`

**Step 1: Create plugin manifest**

Create `/Users/snider/.claude/plugins/altum-updater/plugin.json`:

```json
{
  "name": "altum-updater",
  "description": "Download AltumCode product and plugin updates from LemonSqueezy and CodeCanyon",
  "version": "0.1.0",
  "skills": [
    {
      "name": "update-altum",
      "path": "skills/update-altum.md",
      "description": "Download AltumCode product and plugin updates from marketplaces. Use when the user mentions updating AltumCode products, downloading from LemonSqueezy or CodeCanyon, or running the update checker."
    }
  ]
}
```

**Step 2: Create skill file**

Create `/Users/snider/.claude/plugins/altum-updater/skills/update-altum.md`:

```markdown
---
name: update-altum
description: Download AltumCode product and plugin updates from LemonSqueezy and CodeCanyon
---

# AltumCode Update Downloader

## Overview

Downloads updated AltumCode products and plugins from two marketplaces:
- **LemonSqueezy** (Playwright): 66analytics, 66pusher, 66biolinks (extended), 13 plugins
- **CodeCanyon** (Claude in Chrome): 66biolinks (regular), 66socialproof

## Pre-flight

1. Run `php artisan uptelligence:check-updates --vendor=66analytics` (or check all) to see what needs updating
2. Show the user the version comparison table
3. Ask which products/plugins to download

## LemonSqueezy Download Flow (Playwright)

LemonSqueezy uses magic link auth. The user will need to tap the link on their phone.

1. Navigate to `https://app.lemonsqueezy.com/my-orders`
2. If on login page, fill email `snider@lt.hn` and click Sign In
3. Tell user: "Magic link sent — tap the link on your phone"
4. Wait for redirect to orders page
5. For each product/plugin that needs updating:
   a. Click the product link on the orders page (paginated — 10 per page, 2 pages)
   b. In the order detail, find the "Download" button under "Files"
   c. Click Download — file saves to default downloads folder
6. Move downloaded zips to staging: `~/Code/lthn/saas/updates/YYYY-MM-DD/`

### LemonSqueezy Product Names (as shown on orders page)

| Our Name | LemonSqueezy Order Name |
|----------|------------------------|
| 66analytics | "66analytics - Regular License" |
| 66pusher | "66pusher - Regular License" |
| 66biolinks (extended) | "66biolinks custom" |
| Affiliate Plugin | "Affiliate Plugin" |
| Push Notifications Plugin | "Push Notifications Plugin" |
| Teams Plugin | "Teams Plugin" |
| PWA Plugin | "PWA Plugin" |
| Image Optimizer Plugin | "Image Optimizer Plugin" |
| Email Shield Plugin | "Email Shield Plugin" |
| Dynamic OG Images | "Dynamic OG images plugin" |
| Offload & CDN | "Offload & CDN Plugin" |
| Payment Blocks | "Payment Blocks - 66biolinks plugin" |
| Ultimate Blocks | "Ultimate Blocks - 66biolinks plugin" |
| Pro Blocks | "Pro Blocks - 66biolinks plugin" |
| Pro Notifications | "Pro Notifications - 66socialproof plugin" |
| AltumCode Club | "The AltumCode Club" |

## CodeCanyon Download Flow (Claude in Chrome)

CodeCanyon uses saved browser session cookies (user: snidered).

1. Navigate to `https://codecanyon.net/downloads`
2. Dismiss cookie banner if present (click "Reject all")
3. For 66socialproof:
   a. Find "66socialproof" Download button
   b. Click the dropdown arrow
   c. Click "All files & documentation"
4. For 66biolinks:
   a. Find "66biolinks" Download button (scroll down)
   b. Click the dropdown arrow
   c. Click "All files & documentation"
5. Move downloaded zips to staging

### CodeCanyon Download URLs (stable)

- 66socialproof: `/user/snidered/download_purchase/8d8ef4c1-5add-4eba-9a89-4261a9c87e0b`
- 66biolinks: `/user/snidered/download_purchase/38d79f4e-19cd-480a-b068-4332629b5206`

## Post-Download

1. List all zips in staging folder
2. For each product zip:
   - Extract to `~/Code/lthn/saas/services/{product}/package/product/`
3. For each plugin zip:
   - Extract to the correct product's `plugins/{plugin_id}/` directory
   - Note: Some plugins apply to multiple products (affiliate, teams, etc.)
4. Show summary of what was updated
5. Remind user: "Files staged. Run `deploy_saas.yml` when ready to deploy."

## Important Notes

- Never make purchases or enter financial information
- LemonSqueezy session expires — if Playwright gets a login page mid-flow, re-trigger magic link
- CodeCanyon session depends on Chrome cookies — if logged out, tell user to log in manually
- The AltumCode Club subscription is NOT a downloadable product — skip it
- Plugin `aix` may not appear on LemonSqueezy (bundled with products) — skip if not found
```

**Step 3: Verify plugin loads**

Run: `claude` in a new terminal, then type `/update-altum` to verify the skill is discovered.

**Step 4: Commit**

```bash
cd /Users/snider/.claude/plugins/altum-updater
git init
git add plugin.json skills/update-altum.md
git commit -m "feat: altum-updater Claude Code plugin — marketplace download skill

Playwright for LemonSqueezy, Chrome for CodeCanyon. Includes full
product/plugin mapping and download flow documentation.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 4: Sync deployed plugin versions from source

**Files:**
- Create: `/Users/snider/Code/core/php-uptelligence/Console/SyncAltumVersionsCommand.php`
- Modify: `/Users/snider/Code/core/php-uptelligence/Boot.php` (register command)
- Test: `/Users/snider/Code/core/php-uptelligence/tests/Unit/SyncAltumVersionsCommandTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

it('reads product version from saas service config', function () {
    $this->artisan('uptelligence:sync-altum-versions', ['--dry-run' => true])
        ->assertExitCode(0);
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=SyncAltumVersions`
Expected: FAIL — command not found

**Step 3: Write minimal implementation**

Create `/Users/snider/Code/core/php-uptelligence/Console/SyncAltumVersionsCommand.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Mod\Uptelligence\Console;

use Core\Mod\Uptelligence\Models\Vendor;
use Illuminate\Console\Command;

/**
 * Sync deployed AltumCode product/plugin versions from local source files.
 *
 * Reads PRODUCT_CODE from each product's source and plugin versions
 * from config.php files, then updates the vendors table.
 */
class SyncAltumVersionsCommand extends Command
{
    protected $signature = 'uptelligence:sync-altum-versions
                            {--dry-run : Show what would be updated without writing}
                            {--path= : Base path to saas services (default: ~/Code/lthn/saas/services)}';

    protected $description = 'Sync deployed AltumCode product and plugin versions from source files';

    protected array $productPaths = [
        '66analytics' => '66analytics/package/product',
        '66biolinks' => '66biolinks/package/product',
        '66pusher' => '66pusher/package/product',
        '66socialproof' => '66socialproof/package/product',
    ];

    public function handle(): int
    {
        $basePath = $this->option('path')
            ?? env('SAAS_SERVICES_PATH', base_path('../lthn/saas/services'));
        $dryRun = $this->option('dry-run');

        $this->info('Syncing AltumCode versions from source...');
        $this->newLine();

        $updates = [];

        // Sync product versions
        foreach ($this->productPaths as $slug => $relativePath) {
            $productPath = rtrim($basePath, '/') . '/' . $relativePath;
            $version = $this->readProductVersion($productPath);

            if ($version) {
                $updates[] = $this->syncVendorVersion($slug, $version, $dryRun);
            } else {
                $this->warn("  Could not read version for {$slug} at {$productPath}");
            }
        }

        // Sync plugin versions — read from biolinks as canonical source
        $biolinkPluginsPath = rtrim($basePath, '/') . '/66biolinks/package/product/plugins';
        if (is_dir($biolinkPluginsPath)) {
            foreach (glob($biolinkPluginsPath . '/*/config.php') as $configFile) {
                $pluginId = basename(dirname($configFile));
                $version = $this->readPluginVersion($configFile);

                if ($version) {
                    $slug = "altum-plugin-{$pluginId}";
                    $updates[] = $this->syncVendorVersion($slug, $version, $dryRun);
                }
            }
        }

        // Output table
        $this->table(
            ['Vendor', 'Old Version', 'New Version', 'Status'],
            array_filter($updates)
        );

        if ($dryRun) {
            $this->warn('Dry run — no changes written.');
        }

        return self::SUCCESS;
    }

    protected function readProductVersion(string $productPath): ?string
    {
        // Read version from app/init.php or similar — look for PRODUCT_VERSION define
        $initFile = $productPath . '/app/init.php';
        if (! file_exists($initFile)) {
            return null;
        }

        $content = file_get_contents($initFile);
        if (preg_match("/define\('PRODUCT_VERSION',\s*'([^']+)'\)/", $content, $matches)) {
            return $matches[1];
        }

        return null;
    }

    protected function readPluginVersion(string $configFile): ?string
    {
        if (! file_exists($configFile)) {
            return null;
        }

        $content = file_get_contents($configFile);

        // PHP config format: 'version' => '2.0.0'
        if (preg_match("/'version'\s*=>\s*'([^']+)'/", $content, $matches)) {
            return $matches[1];
        }

        return null;
    }

    protected function syncVendorVersion(string $slug, string $version, bool $dryRun): ?array
    {
        $vendor = Vendor::where('slug', $slug)->first();
        if (! $vendor) {
            return [$slug, '(not in DB)', $version, 'SKIPPED'];
        }

        $oldVersion = $vendor->current_version;
        if ($oldVersion === $version) {
            return [$slug, $oldVersion, $version, 'current'];
        }

        if (! $dryRun) {
            $vendor->update(['current_version' => $version]);
        }

        return [$slug, $oldVersion ?? '(none)', $version, $dryRun ? 'WOULD UPDATE' : 'UPDATED'];
    }
}
```

Register in Boot.php — add to `onConsole()`:

```php
$event->command(Console\SyncAltumVersionsCommand::class);
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php-uptelligence && composer test -- --filter=SyncAltumVersions`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/php-uptelligence
git add Console/SyncAltumVersionsCommand.php Boot.php tests/Unit/SyncAltumVersionsCommandTest.php
git commit -m "feat: sync deployed AltumCode versions from source files

Reads PRODUCT_VERSION from product init.php and plugin versions from
config.php files. Updates uptelligence_vendors table so check-updates
knows what's actually deployed.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 5: End-to-end verification

**Step 1: Seed vendors on local dev**

```bash
cd /Users/snider/Code/lab/host.uk.com
php artisan db:seed --class="Core\Mod\Uptelligence\Database\Seeders\AltumCodeVendorSeeder"
```

**Step 2: Sync actual deployed versions**

```bash
php artisan uptelligence:sync-altum-versions --path=/Users/snider/Code/lthn/saas/services
```

**Step 3: Run the update check**

```bash
php artisan uptelligence:check-updates
```

Expected: Table showing current vs latest versions for all 17 AltumCode vendors.

**Step 4: Test the skill**

Open a new Claude Code session and run `/update-altum` to verify the skill loads and shows the workflow.

**Step 5: Commit any fixes**

```bash
git add -A && git commit -m "fix: adjustments from end-to-end testing"
```
