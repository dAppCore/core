# Scheduled Actions Design

## Goal

Allow CorePHP Actions to declare their own schedule via PHP 8.1 attributes, persist schedules to the database for runtime control, and auto-discover them during deploy — replacing the need for manual `routes/console.php` entries and enabling admin visibility.

## Architecture

**Attribute-driven, database-backed scheduling.** Actions declare defaults with `#[Scheduled]`. A sync command persists them to a `scheduled_actions` table. The scheduler reads the table at runtime. Admin panel provides visibility and control.

**Tech Stack:** PHP 8.1 attributes, Laravel Scheduler, Eloquent, existing CorePHP module scanner paths.

---

## Components

### 1. `#[Scheduled]` Attribute

**File:** `src/Core/Actions/Scheduled.php`

```php
#[Attribute(Attribute::TARGET_CLASS)]
class Scheduled
{
    public function __construct(
        public string $frequency,            // 'everyMinute', 'dailyAt:09:00', 'weeklyOn:1,09:00'
        public ?string $timezone = null,     // 'Europe/London' — null uses app default
        public bool $withoutOverlapping = true,
        public bool $runInBackground = true,
    ) {}
}
```

The `frequency` string maps to Laravel Schedule methods. Colon-separated arguments:
- `dailyAt:09:00` &rarr; `->dailyAt('09:00')`
- `weeklyOn:1,09:00` &rarr; `->weeklyOn(1, '09:00')`
- `everyMinute` &rarr; `->everyMinute()`
- `hourly` &rarr; `->hourly()`
- `monthlyOn:1,00:00` &rarr; `->monthlyOn(1, '00:00')`

### 2. `scheduled_actions` Table

```
scheduled_actions
├── id                    BIGINT PK
├── action_class          VARCHAR(255) UNIQUE  — fully qualified class name
├── frequency             VARCHAR(100)         — from attribute, admin-editable
├── timezone              VARCHAR(50) NULL
├── without_overlapping   BOOLEAN DEFAULT true
├── run_in_background     BOOLEAN DEFAULT true
├── is_enabled            BOOLEAN DEFAULT true — toggle in admin
├── last_run_at           TIMESTAMP NULL
├── next_run_at           TIMESTAMP NULL       — computed from frequency
├── created_at            TIMESTAMP
├── updated_at            TIMESTAMP
```

No tenant scoping — these are system-level platform schedules, not per-user.

### 3. `ScheduledAction` Model

**File:** `src/Core/Actions/ScheduledAction.php`

Eloquent model with:
- `scopeEnabled()` — where `is_enabled = true`
- `markRun()` — updates `last_run_at`, computes `next_run_at`
- `frequencyMethod()` / `frequencyArgs()` — parses `frequency` string

### 4. `ScheduledActionScanner`

**File:** `src/Core/Actions/ScheduledActionScanner.php`

Scans module paths for classes with `#[Scheduled]` attribute using `ReflectionClass::getAttributes()`.

Reuses the same scan paths as `ModuleScanner`:
- `app/Core`, `app/Mod`, `app/Website` (application)
- `src/Core`, `src/Mod` (framework)

Returns: `array<class-string, Scheduled>` — map of class name to attribute instance.

### 5. `schedule:sync` Command

**File:** `src/Core/Console/Commands/ScheduleSyncCommand.php`

```
php artisan schedule:sync
```

- Runs `ScheduledActionScanner`
- Upserts `scheduled_actions` rows:
  - **New classes** &rarr; insert with attribute defaults
  - **Removed classes** &rarr; set `is_enabled = false` (don't delete)
  - **Existing rows manually edited** &rarr; preserve the override (only overwrite if frequency matches the previous attribute default)
- Prints summary: `3 added, 1 disabled, 12 unchanged`
- Run during deploy/migration

### 6. `ScheduleServiceProvider`

**File:** `src/Core/Actions/ScheduleServiceProvider.php`

Registered in framework boot, console context only.

- Queries `scheduled_actions` where `is_enabled = true`
- For each row:
  ```php
  Schedule::call(fn () => $row->action_class::run())
      ->$frequencyMethod(...$frequencyArgs)
      ->withoutOverlapping()  // if set
      ->runInBackground()     // if set
      ->timezone($timezone)   // if set
  ```
- Updates `last_run_at` via `after()` callback

---

## Flow

### Deploy/Migration

```
artisan schedule:sync
    ├── ScheduledActionScanner scans #[Scheduled] attributes
    ├── Upsert scheduled_actions table
    └── Summary: "3 added, 1 disabled, 12 unchanged"
```

### Runtime (every minute)

```
artisan schedule:run
    └── ScheduleServiceProvider
            ├── Query scheduled_actions WHERE is_enabled = true
            ├── For each: Schedule::call(fn () => ActionClass::run())
            └── After each: update last_run_at, compute next_run_at
```

### Admin Panel (future, not MVP)

Table view of `scheduled_actions` with enable/disable toggle, frequency editing, last_run_at display.

---

## Usage Example

```php
<?php

declare(strict_types=1);

namespace Mod\Social\Actions;

use Core\Actions\Action;
use Core\Actions\Scheduled;

#[Scheduled(frequency: 'dailyAt:09:00', timezone: 'Europe/London')]
class PublishDiscordDigest
{
    use Action;

    public function handle(): void
    {
        // Gather yesterday's commits across repos
        // Summarise changes
        // Post to Discord webhook
    }
}
```

No Boot registration needed. No `routes/console.php` entry. The scanner discovers it, `schedule:sync` persists it, the scheduler runs it.

---

## Migration Strategy

- **Existing `routes/console.php` commands** stay as-is. No breaking changes.
- **New scheduled work** uses `#[Scheduled]` actions going forward.
- **Over time**, existing commands can be migrated to actions at natural touch points.

## First Consumers

- Discord daily digest (summarise repo changes, post to Lethean Discord)
- Social media scheduled posting triggers
- Image resizing queue triggers (VIP feature)
- AltumCode cron replacements (longer term — wget loops work for now)
- Sync operations (biolinks, analytics data, etc.)

## Non-Goals (MVP)

- Per-tenant scheduling (system-level only for now)
- Admin panel UI (just the table/model/command/provider)
- Caching scanner results (premature optimisation)
- Replacing existing `routes/console.php` entries (gradual migration)
