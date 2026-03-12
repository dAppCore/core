# Scheduled Actions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add attribute-driven, database-backed Action scheduling to CorePHP so Actions can declare `#[Scheduled]` and be auto-discovered, persisted, and run by the Laravel scheduler.

**Architecture:** PHP 8.1 `#[Scheduled]` attribute on Action classes. `ScheduledActionScanner` discovers them via reflection. `schedule:sync` command persists to `scheduled_actions` table. `ScheduleServiceProvider` reads the table and wires into Laravel's `Schedule` at runtime.

**Tech Stack:** PHP 8.1 attributes, Laravel 12 Scheduler, Eloquent, Orchestra Testbench, PHPUnit

---

## Context

**Repository:** `/Users/snider/Code/core/php` (`forge.lthn.ai/core/php`)

**Existing files you'll interact with:**
- `src/Core/Actions/Action.php` — existing trait with `run()` helper
- `src/Core/Actions/Actionable.php` — existing optional interface
- `src/Core/ModuleScanner.php` — scans `Boot.php` for `$listens` (reference for scan pattern)
- `src/Core/LifecycleEventProvider.php` — orchestrates lifecycle events
- `src/Core/Front/Cli/Boot.php` — fires `ConsoleBooting`, processes commands
- `src/Core/Console/Boot.php` — registers framework commands via `ConsoleBooting`
- `database/migrations/2024_01_01_000001_create_activity_log_table.php` — migration naming convention
- `tests/TestCase.php` — base test class (Orchestra Testbench, SQLite :memory:)
- `tests/Fixtures/` — test fixtures directory

**Test command:** `cd /Users/snider/Code/core/php && composer test` (runs `vendor/bin/phpunit`)
**Single test:** `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter test_name`
**Lint:** `cd /Users/snider/Code/core/php && composer pint`

**Coding standards:**
- `declare(strict_types=1);` in every PHP file
- UK English (e.g. `synchronise` not `synchronize` in comments — but method names follow Laravel convention)
- EUPL-1.2 licence header on framework files
- Type hints on all parameters and return types

---

### Task 1: Create the `#[Scheduled]` Attribute

**Files:**
- Create: `src/Core/Actions/Scheduled.php`
- Test: `tests/Feature/ScheduledAttributeTest.php`

**Step 1: Write the failing test**

Create `tests/Feature/ScheduledAttributeTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\Scheduled;
use PHPUnit\Framework\TestCase;
use ReflectionClass;

class ScheduledAttributeTest extends TestCase
{
    public function test_attribute_can_be_instantiated_with_frequency(): void
    {
        $attr = new Scheduled(frequency: 'dailyAt:09:00');

        $this->assertSame('dailyAt:09:00', $attr->frequency);
        $this->assertNull($attr->timezone);
        $this->assertTrue($attr->withoutOverlapping);
        $this->assertTrue($attr->runInBackground);
    }

    public function test_attribute_accepts_all_parameters(): void
    {
        $attr = new Scheduled(
            frequency: 'weeklyOn:1,09:00',
            timezone: 'Europe/London',
            withoutOverlapping: false,
            runInBackground: false,
        );

        $this->assertSame('weeklyOn:1,09:00', $attr->frequency);
        $this->assertSame('Europe/London', $attr->timezone);
        $this->assertFalse($attr->withoutOverlapping);
        $this->assertFalse($attr->runInBackground);
    }

    public function test_attribute_targets_class_only(): void
    {
        $ref = new ReflectionClass(Scheduled::class);
        $attrs = $ref->getAttributes(\Attribute::class);

        $this->assertNotEmpty($attrs);
        $instance = $attrs[0]->newInstance();
        $this->assertSame(\Attribute::TARGET_CLASS, $instance->flags);
    }

    public function test_attribute_can_be_read_from_class(): void
    {
        $ref = new ReflectionClass(ScheduledAttributeTest_Stub::class);
        $attrs = $ref->getAttributes(Scheduled::class);

        $this->assertCount(1, $attrs);
        $instance = $attrs[0]->newInstance();
        $this->assertSame('everyMinute', $instance->frequency);
    }
}

#[Scheduled(frequency: 'everyMinute')]
class ScheduledAttributeTest_Stub
{
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledAttributeTest`
Expected: FAIL — `Class "Core\Actions\Scheduled" not found`

**Step 3: Write minimal implementation**

Create `src/Core/Actions/Scheduled.php`:

```php
<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Actions;

use Attribute;

/**
 * Mark an Action class for scheduled execution.
 *
 * The frequency string maps to Laravel Schedule methods:
 * - 'everyMinute' → ->everyMinute()
 * - 'dailyAt:09:00' → ->dailyAt('09:00')
 * - 'weeklyOn:1,09:00' → ->weeklyOn(1, '09:00')
 * - 'hourly' → ->hourly()
 * - 'monthlyOn:1,00:00' → ->monthlyOn(1, '00:00')
 *
 * Usage:
 *   #[Scheduled(frequency: 'dailyAt:09:00', timezone: 'Europe/London')]
 *   class PublishDigest
 *   {
 *       use Action;
 *       public function handle(): void { ... }
 *   }
 *
 * Discovered by ScheduledActionScanner, persisted to scheduled_actions table
 * via `php artisan schedule:sync`, and executed by ScheduleServiceProvider.
 */
#[Attribute(Attribute::TARGET_CLASS)]
class Scheduled
{
    public function __construct(
        public string $frequency,
        public ?string $timezone = null,
        public bool $withoutOverlapping = true,
        public bool $runInBackground = true,
    ) {}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledAttributeTest`
Expected: OK (4 tests, 4 assertions)

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/php
git add src/Core/Actions/Scheduled.php tests/Feature/ScheduledAttributeTest.php
git commit -m "feat(actions): add #[Scheduled] attribute for Action classes"
```

---

### Task 2: Create the `ScheduledAction` Model and Migration

**Files:**
- Create: `src/Core/Actions/ScheduledAction.php`
- Create: `database/migrations/2024_01_01_000002_create_scheduled_actions_table.php`
- Test: `tests/Feature/ScheduledActionModelTest.php`

**Step 1: Write the failing test**

Create `tests/Feature/ScheduledActionModelTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\ScheduledAction;
use Core\Tests\TestCase;
use Illuminate\Foundation\Testing\RefreshDatabase;

class ScheduledActionModelTest extends TestCase
{
    use RefreshDatabase;

    protected function defineDatabaseMigrations(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/../../database/migrations');
    }

    public function test_model_can_be_created(): void
    {
        $action = ScheduledAction::create([
            'action_class' => 'App\\Actions\\TestAction',
            'frequency' => 'dailyAt:09:00',
            'timezone' => 'Europe/London',
            'without_overlapping' => true,
            'run_in_background' => true,
            'is_enabled' => true,
        ]);

        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'App\\Actions\\TestAction',
            'frequency' => 'dailyAt:09:00',
        ]);
    }

    public function test_enabled_scope(): void
    {
        ScheduledAction::create([
            'action_class' => 'App\\Actions\\Enabled',
            'frequency' => 'hourly',
            'is_enabled' => true,
        ]);
        ScheduledAction::create([
            'action_class' => 'App\\Actions\\Disabled',
            'frequency' => 'hourly',
            'is_enabled' => false,
        ]);

        $enabled = ScheduledAction::enabled()->get();
        $this->assertCount(1, $enabled);
        $this->assertSame('App\\Actions\\Enabled', $enabled->first()->action_class);
    }

    public function test_frequency_method_parses_simple_frequency(): void
    {
        $action = new ScheduledAction(['frequency' => 'everyMinute']);
        $this->assertSame('everyMinute', $action->frequencyMethod());
        $this->assertSame([], $action->frequencyArgs());
    }

    public function test_frequency_method_parses_frequency_with_args(): void
    {
        $action = new ScheduledAction(['frequency' => 'dailyAt:09:00']);
        $this->assertSame('dailyAt', $action->frequencyMethod());
        $this->assertSame(['09:00'], $action->frequencyArgs());
    }

    public function test_frequency_method_parses_multiple_args(): void
    {
        $action = new ScheduledAction(['frequency' => 'weeklyOn:1,09:00']);
        $this->assertSame('weeklyOn', $action->frequencyMethod());
        $this->assertSame(['1', '09:00'], $action->frequencyArgs());
    }

    public function test_mark_run_updates_last_run_at(): void
    {
        $action = ScheduledAction::create([
            'action_class' => 'App\\Actions\\Runnable',
            'frequency' => 'hourly',
            'is_enabled' => true,
        ]);

        $this->assertNull($action->last_run_at);

        $action->markRun();
        $action->refresh();

        $this->assertNotNull($action->last_run_at);
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledActionModelTest`
Expected: FAIL — class/table not found

**Step 3: Write the migration**

Create `database/migrations/2024_01_01_000002_create_scheduled_actions_table.php`:

```php
<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('scheduled_actions', function (Blueprint $table) {
            $table->id();
            $table->string('action_class')->unique();
            $table->string('frequency', 100);
            $table->string('timezone', 50)->nullable();
            $table->boolean('without_overlapping')->default(true);
            $table->boolean('run_in_background')->default(true);
            $table->boolean('is_enabled')->default(true);
            $table->timestamp('last_run_at')->nullable();
            $table->timestamp('next_run_at')->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('scheduled_actions');
    }
};
```

**Step 4: Write the model**

Create `src/Core/Actions/ScheduledAction.php`:

```php
<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Actions;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

/**
 * Represents a scheduled action persisted in the database.
 *
 * @property int $id
 * @property string $action_class
 * @property string $frequency
 * @property string|null $timezone
 * @property bool $without_overlapping
 * @property bool $run_in_background
 * @property bool $is_enabled
 * @property \Illuminate\Support\Carbon|null $last_run_at
 * @property \Illuminate\Support\Carbon|null $next_run_at
 * @property \Illuminate\Support\Carbon $created_at
 * @property \Illuminate\Support\Carbon $updated_at
 */
class ScheduledAction extends Model
{
    protected $fillable = [
        'action_class',
        'frequency',
        'timezone',
        'without_overlapping',
        'run_in_background',
        'is_enabled',
        'last_run_at',
        'next_run_at',
    ];

    protected function casts(): array
    {
        return [
            'without_overlapping' => 'boolean',
            'run_in_background' => 'boolean',
            'is_enabled' => 'boolean',
            'last_run_at' => 'datetime',
            'next_run_at' => 'datetime',
        ];
    }

    /**
     * Scope to only enabled actions.
     */
    public function scopeEnabled(Builder $query): Builder
    {
        return $query->where('is_enabled', true);
    }

    /**
     * Parse the frequency string and return the method name.
     *
     * 'dailyAt:09:00' → 'dailyAt'
     * 'everyMinute' → 'everyMinute'
     */
    public function frequencyMethod(): string
    {
        return explode(':', $this->frequency, 2)[0];
    }

    /**
     * Parse the frequency string and return the arguments.
     *
     * 'dailyAt:09:00' → ['09:00']
     * 'weeklyOn:1,09:00' → ['1', '09:00']
     * 'everyMinute' → []
     */
    public function frequencyArgs(): array
    {
        $parts = explode(':', $this->frequency, 2);

        if (! isset($parts[1])) {
            return [];
        }

        return explode(',', $parts[1]);
    }

    /**
     * Record that this action has just run.
     */
    public function markRun(): void
    {
        $this->update(['last_run_at' => now()]);
    }
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledActionModelTest`
Expected: OK (6 tests, 6 assertions)

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/php
git add src/Core/Actions/ScheduledAction.php database/migrations/2024_01_01_000002_create_scheduled_actions_table.php tests/Feature/ScheduledActionModelTest.php
git commit -m "feat(actions): add ScheduledAction model and migration"
```

---

### Task 3: Create the `ScheduledActionScanner`

**Files:**
- Create: `src/Core/Actions/ScheduledActionScanner.php`
- Create: `tests/Feature/ScheduledActionScannerTest.php`
- Create: `tests/Fixtures/Mod/Scheduled/Actions/EveryMinuteAction.php`
- Create: `tests/Fixtures/Mod/Scheduled/Actions/DailyAction.php`
- Create: `tests/Fixtures/Mod/Scheduled/Actions/NotScheduledAction.php`

**Step 1: Create test fixtures**

Create `tests/Fixtures/Mod/Scheduled/Actions/EveryMinuteAction.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Fixtures\Mod\Scheduled\Actions;

use Core\Actions\Action;
use Core\Actions\Scheduled;

#[Scheduled(frequency: 'everyMinute')]
class EveryMinuteAction
{
    use Action;

    public function handle(): string
    {
        return 'ran';
    }
}
```

Create `tests/Fixtures/Mod/Scheduled/Actions/DailyAction.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Fixtures\Mod\Scheduled\Actions;

use Core\Actions\Action;
use Core\Actions\Scheduled;

#[Scheduled(frequency: 'dailyAt:09:00', timezone: 'Europe/London', withoutOverlapping: false)]
class DailyAction
{
    use Action;

    public function handle(): string
    {
        return 'daily';
    }
}
```

Create `tests/Fixtures/Mod/Scheduled/Actions/NotScheduledAction.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Fixtures\Mod\Scheduled\Actions;

use Core\Actions\Action;

class NotScheduledAction
{
    use Action;

    public function handle(): string
    {
        return 'not scheduled';
    }
}
```

**Step 2: Write the failing test**

Create `tests/Feature/ScheduledActionScannerTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\Scheduled;
use Core\Actions\ScheduledActionScanner;
use Core\Tests\Fixtures\Mod\Scheduled\Actions\DailyAction;
use Core\Tests\Fixtures\Mod\Scheduled\Actions\EveryMinuteAction;
use PHPUnit\Framework\TestCase;

class ScheduledActionScannerTest extends TestCase
{
    private ScheduledActionScanner $scanner;

    protected function setUp(): void
    {
        parent::setUp();
        $this->scanner = new ScheduledActionScanner();
    }

    public function test_scan_discovers_scheduled_actions(): void
    {
        $results = $this->scanner->scan([
            dirname(__DIR__).'/Fixtures/Mod/Scheduled',
        ]);

        $this->assertArrayHasKey(EveryMinuteAction::class, $results);
        $this->assertArrayHasKey(DailyAction::class, $results);
    }

    public function test_scan_ignores_non_scheduled_actions(): void
    {
        $results = $this->scanner->scan([
            dirname(__DIR__).'/Fixtures/Mod/Scheduled',
        ]);

        $classes = array_keys($results);
        foreach ($classes as $class) {
            $this->assertStringNotContainsString('NotScheduled', $class);
        }
    }

    public function test_scan_returns_attribute_instances(): void
    {
        $results = $this->scanner->scan([
            dirname(__DIR__).'/Fixtures/Mod/Scheduled',
        ]);

        $attr = $results[EveryMinuteAction::class];
        $this->assertInstanceOf(Scheduled::class, $attr);
        $this->assertSame('everyMinute', $attr->frequency);
    }

    public function test_scan_preserves_attribute_parameters(): void
    {
        $results = $this->scanner->scan([
            dirname(__DIR__).'/Fixtures/Mod/Scheduled',
        ]);

        $attr = $results[DailyAction::class];
        $this->assertSame('dailyAt:09:00', $attr->frequency);
        $this->assertSame('Europe/London', $attr->timezone);
        $this->assertFalse($attr->withoutOverlapping);
    }

    public function test_scan_handles_empty_directory(): void
    {
        $results = $this->scanner->scan(['/nonexistent/path']);
        $this->assertEmpty($results);
    }
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledActionScannerTest`
Expected: FAIL — `Class "Core\Actions\ScheduledActionScanner" not found`

**Step 4: Write minimal implementation**

Create `src/Core/Actions/ScheduledActionScanner.php`:

```php
<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Actions;

use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use ReflectionClass;

/**
 * Scans directories for Action classes with the #[Scheduled] attribute.
 *
 * Unlike ModuleScanner (which scans Boot.php files), this scanner finds
 * any PHP class with the #[Scheduled] attribute in the given directories.
 *
 * It uses PHP's native reflection to read attributes — no file parsing.
 *
 * @see Scheduled The attribute this scanner discovers
 * @see ModuleScanner Similar pattern for Boot.php discovery
 */
class ScheduledActionScanner
{
    /**
     * Scan directories for classes with #[Scheduled] attribute.
     *
     * @param  array<string>  $paths  Directories to scan recursively
     * @return array<class-string, Scheduled>  Map of class name to attribute instance
     */
    public function scan(array $paths): array
    {
        $results = [];

        foreach ($paths as $path) {
            if (! is_dir($path)) {
                continue;
            }

            $iterator = new RecursiveIteratorIterator(
                new RecursiveDirectoryIterator($path, RecursiveDirectoryIterator::SKIP_DOTS)
            );

            foreach ($iterator as $file) {
                if ($file->getExtension() !== 'php') {
                    continue;
                }

                $class = $this->classFromFile($file->getPathname());

                if ($class === null || ! class_exists($class)) {
                    continue;
                }

                $attribute = $this->extractScheduled($class);

                if ($attribute !== null) {
                    $results[$class] = $attribute;
                }
            }
        }

        return $results;
    }

    /**
     * Extract the #[Scheduled] attribute from a class.
     */
    private function extractScheduled(string $class): ?Scheduled
    {
        try {
            $ref = new ReflectionClass($class);
            $attrs = $ref->getAttributes(Scheduled::class);

            if (empty($attrs)) {
                return null;
            }

            return $attrs[0]->newInstance();
        } catch (\ReflectionException) {
            return null;
        }
    }

    /**
     * Derive fully qualified class name from a PHP file.
     *
     * Reads the file's namespace declaration and class name.
     */
    private function classFromFile(string $file): ?string
    {
        $contents = file_get_contents($file);

        if ($contents === false) {
            return null;
        }

        $namespace = null;
        $class = null;

        foreach (token_get_all($contents) as $i => $token) {
            if (! is_array($token)) {
                continue;
            }

            if ($token[0] === T_NAMESPACE) {
                $namespace = $this->extractAfterToken($contents, $token[2]);
            }

            if ($token[0] === T_CLASS) {
                $class = $this->extractClassName($contents, $token[2]);
                break;
            }
        }

        if ($class === null) {
            return null;
        }

        return $namespace !== null ? "{$namespace}\\{$class}" : $class;
    }

    /**
     * Extract the namespace or class name string after a token.
     */
    private function extractAfterToken(string $contents, int $line): ?string
    {
        // Re-tokenise to get position-aware extraction
        $tokens = token_get_all($contents);
        $capture = false;
        $parts = [];

        foreach ($tokens as $token) {
            if (is_array($token) && $token[0] === T_NAMESPACE) {
                $capture = true;

                continue;
            }

            if ($capture) {
                if (is_array($token) && in_array($token[0], [T_NAME_QUALIFIED, T_STRING, T_NS_SEPARATOR], true)) {
                    $parts[] = $token[1];
                } elseif ($token === ';' || $token === '{') {
                    break;
                }
            }
        }

        return ! empty($parts) ? implode('', $parts) : null;
    }

    /**
     * Extract class name from tokens after T_CLASS.
     */
    private function extractClassName(string $contents, int $line): ?string
    {
        $tokens = token_get_all($contents);
        $nextIsClass = false;

        foreach ($tokens as $token) {
            if (is_array($token) && $token[0] === T_CLASS) {
                $nextIsClass = true;

                continue;
            }

            if ($nextIsClass && is_array($token)) {
                if ($token[0] === T_WHITESPACE) {
                    continue;
                }
                if ($token[0] === T_STRING) {
                    return $token[1];
                }

                return null;
            }
        }

        return null;
    }
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduledActionScannerTest`
Expected: OK (5 tests, 5 assertions)

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/php
git add src/Core/Actions/ScheduledActionScanner.php tests/Feature/ScheduledActionScannerTest.php tests/Fixtures/Mod/Scheduled/
git commit -m "feat(actions): add ScheduledActionScanner — discovers #[Scheduled] classes"
```

---

### Task 4: Create the `schedule:sync` Command

**Files:**
- Create: `src/Core/Console/Commands/ScheduleSyncCommand.php`
- Modify: `src/Core/Console/Boot.php:27` — register the new command
- Test: `tests/Feature/ScheduleSyncCommandTest.php`

**Step 1: Write the failing test**

Create `tests/Feature/ScheduleSyncCommandTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\ScheduledAction;
use Core\Actions\ScheduledActionScanner;
use Core\Tests\TestCase;
use Illuminate\Foundation\Testing\RefreshDatabase;

class ScheduleSyncCommandTest extends TestCase
{
    use RefreshDatabase;

    protected function defineDatabaseMigrations(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/../../database/migrations');
    }

    protected function defineEnvironment($app): void
    {
        parent::defineEnvironment($app);

        // Point scanner at test fixtures
        $app['config']->set('core.scheduled_action_paths', [
            __DIR__.'/../Fixtures/Mod/Scheduled',
        ]);
    }

    public function test_sync_inserts_new_scheduled_actions(): void
    {
        $this->artisan('schedule:sync')
            ->assertSuccessful();

        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'frequency' => 'everyMinute',
            'is_enabled' => true,
        ]);

        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\DailyAction',
            'frequency' => 'dailyAt:09:00',
            'timezone' => 'Europe/London',
        ]);
    }

    public function test_sync_disables_removed_actions(): void
    {
        // Pre-populate with an action that no longer exists
        ScheduledAction::create([
            'action_class' => 'App\\Actions\\RemovedAction',
            'frequency' => 'hourly',
            'is_enabled' => true,
        ]);

        $this->artisan('schedule:sync')
            ->assertSuccessful();

        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'App\\Actions\\RemovedAction',
            'is_enabled' => false,
        ]);
    }

    public function test_sync_preserves_manually_edited_frequency(): void
    {
        // Pre-populate with a manually edited action
        ScheduledAction::create([
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'frequency' => 'hourly', // Manually changed from everyMinute
            'is_enabled' => true,
        ]);

        $this->artisan('schedule:sync')
            ->assertSuccessful();

        // Should preserve the manual edit
        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'frequency' => 'hourly',
        ]);
    }

    public function test_sync_is_idempotent(): void
    {
        $this->artisan('schedule:sync')->assertSuccessful();
        $this->artisan('schedule:sync')->assertSuccessful();

        $count = ScheduledAction::where(
            'action_class',
            'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction'
        )->count();

        $this->assertSame(1, $count);
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduleSyncCommandTest`
Expected: FAIL — command not registered

**Step 3: Write the command**

Create `src/Core/Console/Commands/ScheduleSyncCommand.php`:

```php
<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Console\Commands;

use Core\Actions\ScheduledAction;
use Core\Actions\ScheduledActionScanner;
use Illuminate\Console\Command;

/**
 * Sync #[Scheduled] attribute declarations to the database.
 *
 * Scans configured paths for Action classes with the #[Scheduled] attribute
 * and upserts them into the scheduled_actions table. Run during deploy/migration.
 */
class ScheduleSyncCommand extends Command
{
    protected $signature = 'schedule:sync';

    protected $description = 'Sync #[Scheduled] action attributes to the database';

    public function handle(ScheduledActionScanner $scanner): int
    {
        $paths = config('core.scheduled_action_paths', [
            app_path('Core'),
            app_path('Mod'),
            app_path('Website'),
        ]);

        // Also scan framework paths
        $frameworkSrc = dirname(__DIR__, 3);
        $paths[] = $frameworkSrc.'/Core';
        $paths[] = $frameworkSrc.'/Mod';

        $discovered = $scanner->scan($paths);

        $added = 0;
        $disabled = 0;
        $unchanged = 0;

        // Upsert discovered actions
        foreach ($discovered as $class => $attribute) {
            $existing = ScheduledAction::where('action_class', $class)->first();

            if ($existing) {
                $unchanged++;

                continue;
            }

            ScheduledAction::create([
                'action_class' => $class,
                'frequency' => $attribute->frequency,
                'timezone' => $attribute->timezone,
                'without_overlapping' => $attribute->withoutOverlapping,
                'run_in_background' => $attribute->runInBackground,
                'is_enabled' => true,
            ]);

            $added++;
        }

        // Disable actions no longer in codebase
        $discoveredClasses = array_keys($discovered);
        $stale = ScheduledAction::where('is_enabled', true)
            ->whereNotIn('action_class', $discoveredClasses)
            ->get();

        foreach ($stale as $action) {
            $action->update(['is_enabled' => false]);
            $disabled++;
        }

        $this->info("Schedule sync complete: {$added} added, {$disabled} disabled, {$unchanged} unchanged.");

        return Command::SUCCESS;
    }
}
```

**Step 4: Register the command in Console/Boot.php**

Modify `src/Core/Console/Boot.php` — add the new command after line 32:

Add `$event->command(Commands\ScheduleSyncCommand::class);` to the `onConsole` method.

**Step 5: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduleSyncCommandTest`
Expected: OK (4 tests, 4+ assertions)

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/php
git add src/Core/Console/Commands/ScheduleSyncCommand.php src/Core/Console/Boot.php tests/Feature/ScheduleSyncCommandTest.php
git commit -m "feat(actions): add schedule:sync command — persists #[Scheduled] to database"
```

---

### Task 5: Create the `ScheduleServiceProvider`

**Files:**
- Create: `src/Core/Actions/ScheduleServiceProvider.php`
- Modify: `src/Core/Front/Cli/Boot.php:38` — register the provider
- Test: `tests/Feature/ScheduleServiceProviderTest.php`

**Step 1: Write the failing test**

Create `tests/Feature/ScheduleServiceProviderTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\ScheduledAction;
use Core\Actions\ScheduleServiceProvider;
use Core\Tests\TestCase;
use Illuminate\Console\Scheduling\Schedule;
use Illuminate\Foundation\Testing\RefreshDatabase;

class ScheduleServiceProviderTest extends TestCase
{
    use RefreshDatabase;

    protected function defineDatabaseMigrations(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/../../database/migrations');
    }

    protected function getPackageProviders($app): array
    {
        return array_merge(parent::getPackageProviders($app), [
            ScheduleServiceProvider::class,
        ]);
    }

    public function test_provider_registers_enabled_actions_with_scheduler(): void
    {
        ScheduledAction::create([
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'frequency' => 'everyMinute',
            'is_enabled' => true,
        ]);

        ScheduledAction::create([
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\DailyAction',
            'frequency' => 'dailyAt:09:00',
            'timezone' => 'Europe/London',
            'is_enabled' => false,
        ]);

        // Re-boot the provider to pick up the new rows
        $provider = new ScheduleServiceProvider($this->app);
        $provider->boot();

        $schedule = $this->app->make(Schedule::class);
        $events = $schedule->events();

        // Should have at least the enabled action
        $this->assertNotEmpty($events);
    }

    public function test_provider_skips_when_table_does_not_exist(): void
    {
        // Drop the table
        \Illuminate\Support\Facades\Schema::dropIfExists('scheduled_actions');

        // Should not throw
        $provider = new ScheduleServiceProvider($this->app);
        $provider->boot();

        $this->assertTrue(true);
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduleServiceProviderTest`
Expected: FAIL — class not found

**Step 3: Write the provider**

Create `src/Core/Actions/ScheduleServiceProvider.php`:

```php
<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Actions;

use Illuminate\Console\Scheduling\Schedule;
use Illuminate\Support\Facades\Schema;
use Illuminate\Support\ServiceProvider;

/**
 * Reads scheduled_actions table and wires enabled actions into Laravel's scheduler.
 *
 * This provider runs in console context only. It queries the database for enabled
 * scheduled actions and registers them with the Laravel Schedule facade.
 *
 * The scheduled_actions table is populated by the `schedule:sync` command,
 * which discovers #[Scheduled] attributes on Action classes.
 */
class ScheduleServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        if (! $this->app->runningInConsole()) {
            return;
        }

        // Guard against table not existing (pre-migration)
        if (! Schema::hasTable('scheduled_actions')) {
            return;
        }

        $this->app->booted(function () {
            $schedule = $this->app->make(Schedule::class);

            $actions = ScheduledAction::enabled()->get();

            foreach ($actions as $action) {
                $class = $action->action_class;

                if (! class_exists($class)) {
                    continue;
                }

                $event = $schedule->call(function () use ($class, $action) {
                    $class::run();
                    $action->markRun();
                });

                // Apply frequency
                $method = $action->frequencyMethod();
                $args = $action->frequencyArgs();
                $event->{$method}(...$args);

                // Apply options
                if ($action->without_overlapping) {
                    $event->withoutOverlapping();
                }

                if ($action->run_in_background) {
                    $event->runInBackground();
                }

                if ($action->timezone) {
                    $event->timezone($action->timezone);
                }
            }
        });
    }
}
```

**Step 4: Register the provider in Front/Cli/Boot.php**

Modify `src/Core/Front/Cli/Boot.php` — in the `boot()` method, before the `runningInConsole` check, register the provider:

Add `$this->app->register(\Core\Actions\ScheduleServiceProvider::class);` inside the `boot()` method, after the `runningInConsole` guard (line 35).

**Step 5: Run test to verify it passes**

Run: `cd /Users/snider/Code/core/php && vendor/bin/phpunit --filter ScheduleServiceProviderTest`
Expected: OK (2 tests, 2 assertions)

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/php
git add src/Core/Actions/ScheduleServiceProvider.php src/Core/Front/Cli/Boot.php tests/Feature/ScheduleServiceProviderTest.php
git commit -m "feat(actions): add ScheduleServiceProvider — wires DB-backed actions into scheduler"
```

---

### Task 6: Integration Test — End to End

**Files:**
- Create: `tests/Feature/ScheduledActionsIntegrationTest.php`

**Step 1: Write the integration test**

Create `tests/Feature/ScheduledActionsIntegrationTest.php`:

```php
<?php

declare(strict_types=1);

namespace Core\Tests\Feature;

use Core\Actions\ScheduledAction;
use Core\Actions\ScheduleServiceProvider;
use Core\Tests\TestCase;
use Illuminate\Console\Scheduling\Schedule;
use Illuminate\Foundation\Testing\RefreshDatabase;

class ScheduledActionsIntegrationTest extends TestCase
{
    use RefreshDatabase;

    protected function defineDatabaseMigrations(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/../../database/migrations');
    }

    protected function defineEnvironment($app): void
    {
        parent::defineEnvironment($app);
        $app['config']->set('core.scheduled_action_paths', [
            __DIR__.'/../Fixtures/Mod/Scheduled',
        ]);
    }

    protected function getPackageProviders($app): array
    {
        return array_merge(parent::getPackageProviders($app), [
            ScheduleServiceProvider::class,
        ]);
    }

    public function test_full_flow_scan_sync_schedule(): void
    {
        // Step 1: Sync discovers and persists
        $this->artisan('schedule:sync')->assertSuccessful();

        // Step 2: Verify rows exist
        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'is_enabled' => true,
        ]);

        // Step 3: Provider registers with scheduler
        $provider = new ScheduleServiceProvider($this->app);
        $provider->boot();

        // Force booted callbacks
        // (In real app, $this->app->booted() fires automatically)
    }

    public function test_disabled_action_not_scheduled(): void
    {
        $this->artisan('schedule:sync')->assertSuccessful();

        // Disable one
        ScheduledAction::where('action_class', 'like', '%EveryMinute%')
            ->update(['is_enabled' => false]);

        $enabled = ScheduledAction::enabled()->get();
        $classes = $enabled->pluck('action_class')->toArray();

        $this->assertNotContains(
            'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            $classes
        );
    }

    public function test_resync_after_disable_does_not_reenable(): void
    {
        $this->artisan('schedule:sync')->assertSuccessful();

        // Admin disables an action
        ScheduledAction::where('action_class', 'like', '%EveryMinute%')
            ->update(['is_enabled' => false]);

        // Re-sync (deploy)
        $this->artisan('schedule:sync')->assertSuccessful();

        // Should still be disabled (existing row preserved)
        $this->assertDatabaseHas('scheduled_actions', [
            'action_class' => 'Core\\Tests\\Fixtures\\Mod\\Scheduled\\Actions\\EveryMinuteAction',
            'is_enabled' => false,
        ]);
    }
}
```

**Step 2: Run all tests**

Run: `cd /Users/snider/Code/core/php && composer test`
Expected: All tests pass (existing + new)

**Step 3: Lint**

Run: `cd /Users/snider/Code/core/php && composer pint`

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/php
git add tests/Feature/ScheduledActionsIntegrationTest.php
git commit -m "test(actions): add integration tests for scheduled actions flow"
```

---

### Task 7: Push to Forge

**Step 1: Run full test suite one final time**

Run: `cd /Users/snider/Code/core/php && composer test`
Expected: All tests pass

**Step 2: Push**

```bash
cd /Users/snider/Code/core/php
git push origin main
```

---

## File Summary

| File | Action |
|------|--------|
| `src/Core/Actions/Scheduled.php` | Create — PHP attribute |
| `src/Core/Actions/ScheduledAction.php` | Create — Eloquent model |
| `src/Core/Actions/ScheduledActionScanner.php` | Create — reflection-based scanner |
| `src/Core/Actions/ScheduleServiceProvider.php` | Create — wires DB rows into Laravel Schedule |
| `src/Core/Console/Commands/ScheduleSyncCommand.php` | Create — `artisan schedule:sync` |
| `src/Core/Console/Boot.php` | Modify — register ScheduleSyncCommand |
| `src/Core/Front/Cli/Boot.php` | Modify — register ScheduleServiceProvider |
| `database/migrations/..._create_scheduled_actions_table.php` | Create — migration |
| `tests/Feature/ScheduledAttributeTest.php` | Create — attribute tests |
| `tests/Feature/ScheduledActionModelTest.php` | Create — model tests |
| `tests/Feature/ScheduledActionScannerTest.php` | Create — scanner tests |
| `tests/Feature/ScheduleSyncCommandTest.php` | Create — sync command tests |
| `tests/Feature/ScheduleServiceProviderTest.php` | Create — provider tests |
| `tests/Feature/ScheduledActionsIntegrationTest.php` | Create — end-to-end tests |
| `tests/Fixtures/Mod/Scheduled/Actions/*.php` | Create — 3 fixture classes |
