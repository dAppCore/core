# Studio Multimedia Pipeline — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CorePHP module (Studio) that orchestrates video remixing, content creation, and voice interaction by dispatching GPU work to homelab Docker services.

**Architecture:** Studio is a job orchestrator. LEM makes creative decisions (manifests), ffmpeg and GPU services execute mechanically. All GPU services are remote Docker containers on homelab, accessed over HTTP. The module follows the same `app/Mod/` patterns as the existing LEM module — Boot.php with lifecycle events, Actions with `use Action;` trait, Pest tests with Http::fake().

**Tech Stack:** Laravel 12, Livewire 3, Flux Pro UI, PostgreSQL, Redis queues, Ollama (LEM), Whisper, ffmpeg, Pest

**Design Doc:** `docs/plans/2026-03-08-studio-multimedia-pipeline-design.md`

**Scope:** Phase 1 (Foundation) + Phase 2 (Remix Pipeline) = April demo. Upload videos, enter brief, get remixed TikToks back.

---

## Reference Files

| File | Role |
|------|------|
| `app/Mod/Lem/Boot.php` | Module boot pattern (API + Console) |
| `app/Mod/Lem/Actions/ScoreContent.php` | Action trait + HTTP dispatch |
| `app/Mod/Lem/Routes/api.php` | API route registration |
| `app/Mod/Lem/Console/Health.php` | Artisan command pattern |
| `app/Mod/Lem/Tests/Feature/ScoringActionsTest.php` | Pest test with Http::fake() |
| `app/Mod/Hub/Boot.php` | Web routes + Livewire registration |
| `app/Mod/Hub/Livewire/HomepagePage.php` | Livewire page component |

All paths relative to `/Users/snider/Code/lab/host.uk.com/` unless noted (Hub is in `/Users/snider/Code/host-uk/agentic/`).

---

### Task 1: Module Config

**Files:**
- Create: `config/studio.php`

**Step 1: Create the config file**

```php
<?php

declare(strict_types=1);

return [
    'whisper' => [
        'url' => env('STUDIO_WHISPER_URL', 'http://studio-whisper:9100'),
        'model' => env('STUDIO_WHISPER_MODEL', 'large-v3-turbo'),
        'timeout' => (int) env('STUDIO_WHISPER_TIMEOUT', 120),
    ],

    'ollama' => [
        'url' => env('STUDIO_OLLAMA_URL', 'http://studio-ollama:11434'),
        'model' => env('STUDIO_OLLAMA_MODEL', 'lem-4b'),
        'timeout' => (int) env('STUDIO_OLLAMA_TIMEOUT', 60),
    ],

    'tts' => [
        'url' => env('STUDIO_TTS_URL', 'http://studio-tts:9200'),
        'voice' => env('STUDIO_TTS_VOICE', 'default'),
        'timeout' => (int) env('STUDIO_TTS_TIMEOUT', 60),
    ],

    'worker' => [
        'url' => env('STUDIO_WORKER_URL', 'http://studio-worker:9300'),
        'timeout' => (int) env('STUDIO_WORKER_TIMEOUT', 300),
    ],

    'storage' => [
        'disk' => env('STUDIO_STORAGE_DISK', 'local'),
        'assets_path' => env('STUDIO_ASSETS_PATH', 'studio/assets'),
        'renders_path' => env('STUDIO_RENDERS_PATH', 'studio/renders'),
    ],

    'templates' => [
        'tiktok-15s' => [
            'duration' => 15,
            'resolution' => '1080x1920',
            'fps' => 30,
        ],
        'tiktok-30s' => [
            'duration' => 30,
            'resolution' => '1080x1920',
            'fps' => 30,
        ],
        'tiktok-60s' => [
            'duration' => 60,
            'resolution' => '1080x1920',
            'fps' => 30,
        ],
    ],
];
```

**Step 2: Commit**

```bash
git add config/studio.php
git commit -m "feat(studio): add module config for GPU service endpoints"
```

---

### Task 2: Module Boot + Migration

**Files:**
- Create: `app/Mod/Studio/Boot.php`
- Create: `app/Mod/Studio/Migrations/2026_03_08_000001_create_studio_tables.php`

**Step 1: Create the migration**

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
        if (! Schema::hasTable('studio_assets')) {
            Schema::create('studio_assets', function (Blueprint $table) {
                $table->id();
                $table->foreignId('workspace_id')->nullable()->constrained()->nullOnDelete();
                $table->string('filename');
                $table->string('path');
                $table->string('mime_type', 64);
                $table->unsignedInteger('duration_ms')->nullable();
                $table->string('resolution', 16)->nullable();
                $table->unsignedBigInteger('file_size')->default(0);
                $table->json('tags')->nullable();
                $table->text('transcript')->nullable();
                $table->string('transcript_language', 8)->nullable();
                $table->timestamps();

                $table->index('workspace_id');
                $table->index('mime_type');
            });
        }

        if (! Schema::hasTable('studio_jobs')) {
            Schema::create('studio_jobs', function (Blueprint $table) {
                $table->id();
                $table->foreignId('workspace_id')->nullable()->constrained()->nullOnDelete();
                $table->string('type', 32);
                $table->string('status', 32)->default('pending');
                $table->json('input')->nullable();
                $table->json('manifest')->nullable();
                $table->json('output')->nullable();
                $table->text('error')->nullable();
                $table->timestamp('started_at')->nullable();
                $table->timestamp('completed_at')->nullable();
                $table->timestamps();

                $table->index(['workspace_id', 'status']);
                $table->index('type');
            });
        }
    }

    public function down(): void
    {
        Schema::disableForeignKeyConstraints();
        Schema::dropIfExists('studio_jobs');
        Schema::dropIfExists('studio_assets');
        Schema::enableForeignKeyConstraints();
    }
};
```

**Step 2: Create Boot.php**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio;

use Core\Events\ApiRoutesRegistering;
use Core\Events\ConsoleBooting;
use Core\Events\WebRoutesRegistering;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\ServiceProvider;

class Boot extends ServiceProvider
{
    protected string $moduleName = 'studio';

    public static array $listens = [
        ApiRoutesRegistering::class => 'onApiRoutes',
        ConsoleBooting::class => 'onConsole',
        WebRoutesRegistering::class => 'onWebRoutes',
    ];

    public function register(): void
    {
        $this->mergeConfigFrom(config_path('studio.php'), 'studio');
    }

    public function boot(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/Migrations');
    }

    public function onApiRoutes(ApiRoutesRegistering $event): void
    {
        $event->routes(fn () => Route::middleware('api')->group(__DIR__.'/Routes/api.php'));
    }

    public function onConsole(ConsoleBooting $event): void
    {
        // Commands registered in later tasks
    }

    public function onWebRoutes(WebRoutesRegistering $event): void
    {
        $event->views('studio', __DIR__.'/Views');
        $event->routes(fn () => require __DIR__.'/Routes/web.php');
    }
}
```

**Step 3: Create empty route stubs**

Create `app/Mod/Studio/Routes/api.php`:
```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;

Route::prefix('studio')->group(function () {
    // Asset and remix routes added in later tasks
});
```

Create `app/Mod/Studio/Routes/web.php`:
```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;

Route::prefix('studio')->group(function () {
    // Livewire pages added in later tasks
});
```

**Step 4: Run migration**

```bash
php artisan migrate
```

Expected: Two tables created (`studio_assets`, `studio_jobs`).

**Step 5: Commit**

```bash
git add app/Mod/Studio/ config/studio.php
git commit -m "feat(studio): scaffold module with Boot, migration, route stubs"
```

---

### Task 3: Asset Model

**Files:**
- Create: `app/Mod/Studio/Models/StudioAsset.php`
- Create: `app/Mod/Studio/Tests/Feature/StudioAssetTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Mod\Studio\Models\StudioAsset;

it('creates a studio asset with metadata', function () {
    $asset = StudioAsset::create([
        'filename' => 'summer-beach.mp4',
        'path' => 'studio/assets/summer-beach.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'resolution' => '1080x1920',
        'file_size' => 5242880,
        'tags' => ['summer', 'beach', 'lollipop'],
    ]);

    expect($asset->id)->toBeGreaterThan(0);
    expect($asset->filename)->toBe('summer-beach.mp4');
    expect($asset->tags)->toBe(['summer', 'beach', 'lollipop']);
    expect($asset->duration_ms)->toBe(15000);
});

it('scopes assets by mime type', function () {
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
    ]);
    StudioAsset::create([
        'filename' => 'thumb.jpg',
        'path' => 'studio/assets/thumb.jpg',
        'mime_type' => 'image/jpeg',
    ]);

    $videos = StudioAsset::videos()->get();
    expect($videos)->toHaveCount(1);
    expect($videos->first()->filename)->toBe('clip.mp4');
});

it('casts tags to array', function () {
    $asset = StudioAsset::create([
        'filename' => 'test.mp4',
        'path' => 'studio/assets/test.mp4',
        'mime_type' => 'video/mp4',
        'tags' => ['winter', 'office'],
    ]);

    $fresh = $asset->fresh();
    expect($fresh->tags)->toBeArray();
    expect($fresh->tags)->toContain('winter');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=StudioAssetTest
```

Expected: FAIL — class `Mod\Studio\Models\StudioAsset` not found.

**Step 3: Write the model**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Models;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

class StudioAsset extends Model
{
    protected $table = 'studio_assets';

    protected $fillable = [
        'workspace_id',
        'filename',
        'path',
        'mime_type',
        'duration_ms',
        'resolution',
        'file_size',
        'tags',
        'transcript',
        'transcript_language',
    ];

    protected $casts = [
        'duration_ms' => 'integer',
        'file_size' => 'integer',
        'tags' => 'array',
    ];

    public function scopeVideos(Builder $query): Builder
    {
        return $query->where('mime_type', 'like', 'video/%');
    }

    public function scopeImages(Builder $query): Builder
    {
        return $query->where('mime_type', 'like', 'image/%');
    }

    public function scopeAudio(Builder $query): Builder
    {
        return $query->where('mime_type', 'like', 'audio/%');
    }

    public function scopeTagged(Builder $query, string $tag): Builder
    {
        return $query->whereJsonContains('tags', $tag);
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=StudioAssetTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Models/ app/Mod/Studio/Tests/
git commit -m "feat(studio): add StudioAsset model with scopes and tag casting"
```

---

### Task 4: StudioJob Model

**Files:**
- Create: `app/Mod/Studio/Models/StudioJob.php`
- Create: `app/Mod/Studio/Tests/Feature/StudioJobTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Mod\Studio\Models\StudioJob;

it('creates a remix job with pending status', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'input' => ['brief' => 'summer lollipop TikTok, 15s, upbeat'],
    ]);

    expect($job->status)->toBe('pending');
    expect($job->type)->toBe('remix');
    expect($job->input['brief'])->toContain('summer');
});

it('transitions through status lifecycle', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'input' => ['brief' => 'test'],
    ]);

    $job->markStarted();
    expect($job->status)->toBe('processing');
    expect($job->started_at)->not->toBeNull();

    $job->markCompleted(['url' => '/renders/out.mp4']);
    expect($job->status)->toBe('completed');
    expect($job->completed_at)->not->toBeNull();
    expect($job->output['url'])->toBe('/renders/out.mp4');
});

it('marks job as failed with error message', function () {
    $job = StudioJob::create([
        'type' => 'transcribe',
        'input' => ['asset_id' => 1],
    ]);

    $job->markFailed('Whisper service unavailable');
    expect($job->status)->toBe('failed');
    expect($job->error)->toBe('Whisper service unavailable');
});

it('scopes by status', function () {
    StudioJob::create(['type' => 'remix', 'input' => [], 'status' => 'pending']);
    StudioJob::create(['type' => 'remix', 'input' => [], 'status' => 'completed']);

    expect(StudioJob::pending()->count())->toBe(1);
    expect(StudioJob::where('status', 'completed')->count())->toBe(1);
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=StudioJobTest
```

Expected: FAIL — class not found.

**Step 3: Write the model**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Models;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

class StudioJob extends Model
{
    protected $table = 'studio_jobs';

    protected $fillable = [
        'workspace_id',
        'type',
        'status',
        'input',
        'manifest',
        'output',
        'error',
        'started_at',
        'completed_at',
    ];

    protected $casts = [
        'input' => 'array',
        'manifest' => 'array',
        'output' => 'array',
        'started_at' => 'datetime',
        'completed_at' => 'datetime',
    ];

    public function markStarted(): void
    {
        $this->update([
            'status' => 'processing',
            'started_at' => now(),
        ]);
    }

    public function markCompleted(array $output): void
    {
        $this->update([
            'status' => 'completed',
            'output' => $output,
            'completed_at' => now(),
        ]);
    }

    public function markFailed(string $error): void
    {
        $this->update([
            'status' => 'failed',
            'error' => $error,
            'completed_at' => now(),
        ]);
    }

    public function scopePending(Builder $query): Builder
    {
        return $query->where('status', 'pending');
    }

    public function scopeProcessing(Builder $query): Builder
    {
        return $query->where('status', 'processing');
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=StudioJobTest
```

Expected: 4 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Models/StudioJob.php app/Mod/Studio/Tests/Feature/StudioJobTest.php
git commit -m "feat(studio): add StudioJob model with status lifecycle"
```

---

### Task 5: CatalogueAsset Action

**Files:**
- Create: `app/Mod/Studio/Actions/CatalogueAsset.php`
- Create: `app/Mod/Studio/Tests/Feature/CatalogueAssetTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\CatalogueAsset;
use Mod\Studio\Models\StudioAsset;

it('catalogues an uploaded video file', function () {
    Storage::fake('local');

    $file = UploadedFile::fake()->create('summer-beach.mp4', 5120, 'video/mp4');

    $asset = CatalogueAsset::run($file, ['summer', 'beach']);

    expect($asset)->toBeInstanceOf(StudioAsset::class);
    expect($asset->filename)->toBe('summer-beach.mp4');
    expect($asset->mime_type)->toBe('video/mp4');
    expect($asset->tags)->toBe(['summer', 'beach']);
    expect($asset->file_size)->toBeGreaterThan(0);
    Storage::disk('local')->assertExists($asset->path);
});

it('catalogues a file from an existing path', function () {
    Storage::fake('local');
    Storage::disk('local')->put('studio/assets/existing.mp4', 'video-content');

    $asset = CatalogueAsset::run(
        'studio/assets/existing.mp4',
        ['winter', 'office'],
    );

    expect($asset)->toBeInstanceOf(StudioAsset::class);
    expect($asset->filename)->toBe('existing.mp4');
    expect($asset->tags)->toBe(['winter', 'office']);
});

it('rejects unsupported mime types', function () {
    Storage::fake('local');

    $file = UploadedFile::fake()->create('readme.txt', 100, 'text/plain');

    CatalogueAsset::run($file);
})->throws(\InvalidArgumentException::class);
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=CatalogueAssetTest
```

Expected: FAIL — class not found.

**Step 3: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Models\StudioAsset;

class CatalogueAsset
{
    use Action;

    private const ALLOWED_MIME_PREFIXES = ['video/', 'image/', 'audio/'];

    /**
     * Catalogue an asset from an uploaded file or existing storage path.
     *
     * @param  UploadedFile|string  $source  Uploaded file or existing storage path
     * @param  array<string>  $tags
     */
    public function handle(UploadedFile|string $source, array $tags = []): StudioAsset
    {
        if ($source instanceof UploadedFile) {
            return $this->catalogueUpload($source, $tags);
        }

        return $this->cataloguePath($source, $tags);
    }

    private function catalogueUpload(UploadedFile $file, array $tags): StudioAsset
    {
        $this->validateMimeType($file->getMimeType() ?? '');

        $disk = config('studio.storage.disk', 'local');
        $assetsPath = config('studio.storage.assets_path', 'studio/assets');
        $storedPath = $file->store($assetsPath, $disk);

        return StudioAsset::create([
            'filename' => $file->getClientOriginalName(),
            'path' => $storedPath,
            'mime_type' => $file->getMimeType(),
            'file_size' => $file->getSize(),
            'tags' => $tags ?: null,
        ]);
    }

    private function cataloguePath(string $path, array $tags): StudioAsset
    {
        $disk = config('studio.storage.disk', 'local');
        $filename = basename($path);
        $mimeType = Storage::disk($disk)->mimeType($path) ?: 'application/octet-stream';
        $fileSize = Storage::disk($disk)->size($path);

        $this->validateMimeType($mimeType);

        return StudioAsset::create([
            'filename' => $filename,
            'path' => $path,
            'mime_type' => $mimeType,
            'file_size' => $fileSize,
            'tags' => $tags ?: null,
        ]);
    }

    private function validateMimeType(string $mimeType): void
    {
        foreach (self::ALLOWED_MIME_PREFIXES as $prefix) {
            if (str_starts_with($mimeType, $prefix)) {
                return;
            }
        }

        throw new \InvalidArgumentException("Unsupported mime type: {$mimeType}");
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=CatalogueAssetTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/CatalogueAsset.php app/Mod/Studio/Tests/Feature/CatalogueAssetTest.php
git commit -m "feat(studio): add CatalogueAsset action for file ingestion"
```

---

### Task 6: TranscribeAsset Action

**Files:**
- Create: `app/Mod/Studio/Actions/TranscribeAsset.php`
- Create: `app/Mod/Studio/Tests/Feature/TranscribeAssetTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\TranscribeAsset;
use Mod\Studio\Models\StudioAsset;

it('transcribes an asset via Whisper service', function () {
    Storage::fake('local');
    Storage::disk('local')->put('studio/assets/clip.mp4', 'video-content');

    $asset = StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
    ]);

    Http::fake([
        '*/transcribe' => Http::response([
            'text' => 'Hello, welcome to our summer collection.',
            'language' => 'en',
            'segments' => [
                ['start' => 0.0, 'end' => 2.5, 'text' => 'Hello, welcome to our summer collection.'],
            ],
        ]),
    ]);

    $result = TranscribeAsset::run($asset);

    expect($result->transcript)->toBe('Hello, welcome to our summer collection.');
    expect($result->transcript_language)->toBe('en');
});

it('returns null transcript on Whisper failure', function () {
    Storage::fake('local');
    Storage::disk('local')->put('studio/assets/clip.mp4', 'video-content');

    $asset = StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
    ]);

    Http::fake([
        '*/transcribe' => Http::response([], 500),
    ]);

    $result = TranscribeAsset::run($asset);

    expect($result->transcript)->toBeNull();
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=TranscribeAssetTest
```

Expected: FAIL — class not found.

**Step 3: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Models\StudioAsset;

class TranscribeAsset
{
    use Action;

    public function handle(StudioAsset $asset): StudioAsset
    {
        $whisperUrl = config('studio.whisper.url');
        $timeout = config('studio.whisper.timeout', 120);

        $disk = config('studio.storage.disk', 'local');
        $filePath = Storage::disk($disk)->path($asset->path);

        $response = Http::timeout($timeout)
            ->attach('file', file_get_contents($filePath), $asset->filename)
            ->post("{$whisperUrl}/transcribe", [
                'model' => config('studio.whisper.model', 'large-v3-turbo'),
            ]);

        if (! $response->successful()) {
            return $asset;
        }

        $data = $response->json();

        $asset->update([
            'transcript' => $data['text'] ?? null,
            'transcript_language' => $data['language'] ?? null,
        ]);

        return $asset->fresh();
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=TranscribeAssetTest
```

Expected: 2 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/TranscribeAsset.php app/Mod/Studio/Tests/Feature/TranscribeAssetTest.php
git commit -m "feat(studio): add TranscribeAsset action for Whisper STT"
```

---

### Task 7: Console Commands (catalogue + transcribe)

**Files:**
- Create: `app/Mod/Studio/Console/Catalogue.php`
- Create: `app/Mod/Studio/Console/Transcribe.php`
- Modify: `app/Mod/Studio/Boot.php` (register commands in `onConsole`)

**Step 1: Write the catalogue command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\CatalogueAsset;

class Catalogue extends Command
{
    protected $signature = 'studio:catalogue
        {path : Directory path to scan for assets}
        {--tags= : Comma-separated tags to apply to all assets}
        {--disk=local : Storage disk to use}';

    protected $description = 'Batch catalogue assets from a directory';

    public function handle(): int
    {
        $path = $this->argument('path');
        $tags = $this->option('tags')
            ? array_map('trim', explode(',', $this->option('tags')))
            : [];
        $disk = $this->option('disk');

        $files = Storage::disk($disk)->files($path);

        if (empty($files)) {
            $this->warn("No files found in {$path}");

            return Command::SUCCESS;
        }

        $this->info("Found ".count($files)." files in {$path}");
        $bar = $this->output->createProgressBar(count($files));
        $bar->start();

        $catalogued = 0;
        $skipped = 0;

        foreach ($files as $file) {
            try {
                CatalogueAsset::run($file, $tags);
                $catalogued++;
            } catch (\InvalidArgumentException) {
                $skipped++;
            }
            $bar->advance();
        }

        $bar->finish();
        $this->newLine();
        $this->info("Catalogued: {$catalogued} | Skipped: {$skipped}");

        return Command::SUCCESS;
    }
}
```

**Step 2: Write the transcribe command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Mod\Studio\Actions\TranscribeAsset;
use Mod\Studio\Models\StudioAsset;

class Transcribe extends Command
{
    protected $signature = 'studio:transcribe
        {--id= : Specific asset ID to transcribe}
        {--all : Transcribe all untranscribed assets}';

    protected $description = 'Transcribe assets via Whisper';

    public function handle(): int
    {
        if ($id = $this->option('id')) {
            $asset = StudioAsset::findOrFail($id);

            return $this->transcribeOne($asset);
        }

        if (! $this->option('all')) {
            $this->error('Specify --id=N or --all');

            return Command::FAILURE;
        }

        $assets = StudioAsset::videos()
            ->whereNull('transcript')
            ->get();

        if ($assets->isEmpty()) {
            $this->info('No untranscribed video assets found.');

            return Command::SUCCESS;
        }

        $this->info("Transcribing {$assets->count()} assets...");
        $bar = $this->output->createProgressBar($assets->count());
        $bar->start();

        $success = 0;

        foreach ($assets as $asset) {
            $result = TranscribeAsset::run($asset);
            if ($result->transcript) {
                $success++;
            }
            $bar->advance();
        }

        $bar->finish();
        $this->newLine();
        $this->info("Transcribed: {$success}/{$assets->count()}");

        return Command::SUCCESS;
    }

    private function transcribeOne(StudioAsset $asset): int
    {
        $this->info("Transcribing: {$asset->filename}");
        $result = TranscribeAsset::run($asset);

        if ($result->transcript) {
            $this->info("Transcript ({$result->transcript_language}): ".substr($result->transcript, 0, 200).'...');

            return Command::SUCCESS;
        }

        $this->error('Transcription failed.');

        return Command::FAILURE;
    }
}
```

**Step 3: Register commands in Boot.php**

Update the `onConsole` method in `app/Mod/Studio/Boot.php`:

```php
    public function onConsole(ConsoleBooting $event): void
    {
        $event->command(Console\Catalogue::class);
        $event->command(Console\Transcribe::class);
    }
```

**Step 4: Verify commands register**

```bash
php artisan list studio
```

Expected: Shows `studio:catalogue` and `studio:transcribe`.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Console/ app/Mod/Studio/Boot.php
git commit -m "feat(studio): add catalogue and transcribe artisan commands"
```

---

### Task 8: Asset API Routes + Controller

**Files:**
- Create: `app/Mod/Studio/Controllers/Api/AssetController.php`
- Create: `app/Mod/Studio/Tests/Feature/AssetApiTest.php`
- Modify: `app/Mod/Studio/Routes/api.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Models\StudioAsset;

it('lists assets via GET /api/studio/assets', function () {
    StudioAsset::create([
        'filename' => 'clip1.mp4',
        'path' => 'studio/assets/clip1.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
        'tags' => ['summer'],
    ]);

    $response = $this->getJson('/api/studio/assets');

    $response->assertOk();
    $response->assertJsonCount(1, 'data');
    $response->assertJsonPath('data.0.filename', 'clip1.mp4');
});

it('filters assets by tag via GET /api/studio/assets?tag=summer', function () {
    StudioAsset::create([
        'filename' => 'summer.mp4',
        'path' => 'studio/assets/summer.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
        'tags' => ['summer'],
    ]);
    StudioAsset::create([
        'filename' => 'winter.mp4',
        'path' => 'studio/assets/winter.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
        'tags' => ['winter'],
    ]);

    $response = $this->getJson('/api/studio/assets?tag=summer');

    $response->assertOk();
    $response->assertJsonCount(1, 'data');
    $response->assertJsonPath('data.0.filename', 'summer.mp4');
});

it('uploads an asset via POST /api/studio/assets', function () {
    Storage::fake('local');

    $file = UploadedFile::fake()->create('new-clip.mp4', 2048, 'video/mp4');

    $response = $this->postJson('/api/studio/assets', [
        'file' => $file,
        'tags' => ['test', 'upload'],
    ]);

    $response->assertCreated();
    $response->assertJsonPath('data.filename', 'new-clip.mp4');

    expect(StudioAsset::count())->toBe(1);
});

it('shows a single asset via GET /api/studio/assets/{id}', function () {
    $asset = StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
    ]);

    $response = $this->getJson("/api/studio/assets/{$asset->id}");

    $response->assertOk();
    $response->assertJsonPath('data.filename', 'clip.mp4');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=AssetApiTest
```

Expected: FAIL — 404 (no routes registered).

**Step 3: Write the controller**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Controllers\Api;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Mod\Studio\Actions\CatalogueAsset;
use Mod\Studio\Models\StudioAsset;

class AssetController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        $query = StudioAsset::query()->latest();

        if ($tag = $request->query('tag')) {
            $query->tagged($tag);
        }

        if ($type = $request->query('type')) {
            match ($type) {
                'video' => $query->videos(),
                'image' => $query->images(),
                'audio' => $query->audio(),
                default => null,
            };
        }

        return response()->json([
            'data' => $query->paginate(50),
        ]);
    }

    public function show(int $id): JsonResponse
    {
        $asset = StudioAsset::findOrFail($id);

        return response()->json(['data' => $asset]);
    }

    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'file' => 'required|file|max:512000',
            'tags' => 'nullable|array',
            'tags.*' => 'string|max:64',
        ]);

        $asset = CatalogueAsset::run(
            $request->file('file'),
            $validated['tags'] ?? [],
        );

        return response()->json(['data' => $asset], 201);
    }
}
```

**Step 4: Update API routes**

Replace `app/Mod/Studio/Routes/api.php`:

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Mod\Studio\Controllers\Api\AssetController;

Route::prefix('studio')->group(function () {
    Route::get('/assets', [AssetController::class, 'index'])->name('api.studio.assets.index');
    Route::get('/assets/{id}', [AssetController::class, 'show'])->name('api.studio.assets.show');
    Route::post('/assets', [AssetController::class, 'store'])->name('api.studio.assets.store');
});
```

**Step 5: Run test to verify it passes**

```bash
php artisan test --filter=AssetApiTest
```

Expected: 4 tests PASS.

**Step 6: Commit**

```bash
git add app/Mod/Studio/Controllers/ app/Mod/Studio/Routes/api.php app/Mod/Studio/Tests/Feature/AssetApiTest.php
git commit -m "feat(studio): add asset API endpoints (list, show, upload)"
```

---

### Task 9: GenerateManifest Action

**Files:**
- Create: `app/Mod/Studio/Actions/GenerateManifest.php`
- Create: `app/Mod/Studio/Tests/Feature/GenerateManifestTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Models\StudioAsset;
use Mod\Studio\Models\StudioJob;

it('generates a manifest from a brief via Ollama', function () {
    // Create some assets in the library
    StudioAsset::create([
        'filename' => 'beach-dance.mp4',
        'path' => 'studio/assets/beach-dance.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 30000,
        'resolution' => '1080x1920',
        'file_size' => 10240,
        'tags' => ['summer', 'beach', 'dance'],
    ]);
    StudioAsset::create([
        'filename' => 'lollipop-close.mp4',
        'path' => 'studio/assets/lollipop-close.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'resolution' => '1080x1920',
        'file_size' => 5120,
        'tags' => ['summer', 'lollipop'],
    ]);

    // LEM returns a valid manifest JSON
    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'template' => 'tiktok-15s',
                    'clips' => [
                        ['asset_id' => 1, 'start' => 3.2, 'end' => 8.1, 'order' => 1],
                        ['asset_id' => 2, 'start' => 0.0, 'end' => 5.5, 'order' => 2],
                    ],
                    'captions' => [
                        ['text' => 'Summer vibes only', 'at' => 0.5, 'duration' => 3, 'style' => 'bold-center'],
                    ],
                    'audio' => ['track' => 'original', 'fade_in' => 0.5],
                    'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
                ]),
            ],
        ]),
    ]);

    $job = GenerateManifest::run('summer lollipop TikTok, 15s, upbeat');

    expect($job)->toBeInstanceOf(StudioJob::class);
    expect($job->type)->toBe('remix');
    expect($job->manifest)->not->toBeNull();
    expect($job->manifest['template'])->toBe('tiktok-15s');
    expect($job->manifest['clips'])->toHaveCount(2);
    expect($job->status)->toBe('pending');
});

it('creates a failed job when Ollama returns invalid JSON', function () {
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 10000,
        'file_size' => 1024,
    ]);

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => 'Sure! Here is a remix plan: [not valid json]',
            ],
        ]),
    ]);

    $job = GenerateManifest::run('test brief');

    expect($job->status)->toBe('failed');
    expect($job->error)->toContain('Invalid manifest');
});

it('creates a failed job when Ollama is unreachable', function () {
    Http::fake([
        '*/api/chat' => Http::response([], 500),
    ]);

    $job = GenerateManifest::run('test brief');

    expect($job->status)->toBe('failed');
    expect($job->error)->toContain('LEM');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=GenerateManifestTest
```

Expected: FAIL — class not found.

**Step 3: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Mod\Studio\Models\StudioAsset;
use Mod\Studio\Models\StudioJob;

class GenerateManifest
{
    use Action;

    public function handle(string $brief, ?string $template = null): StudioJob
    {
        $job = StudioJob::create([
            'type' => 'remix',
            'input' => [
                'brief' => $brief,
                'template' => $template,
            ],
        ]);

        $assets = StudioAsset::videos()->get();

        if ($assets->isEmpty()) {
            $job->markFailed('No video assets in library');

            return $job;
        }

        $assetContext = $assets->map(fn (StudioAsset $a) => [
            'id' => $a->id,
            'filename' => $a->filename,
            'duration_ms' => $a->duration_ms,
            'resolution' => $a->resolution,
            'tags' => $a->tags,
            'transcript' => $a->transcript ? substr($a->transcript, 0, 200) : null,
        ])->toArray();

        $templates = config('studio.templates', []);

        $systemPrompt = $this->buildSystemPrompt($assetContext, $templates);

        $manifest = $this->queryLem($systemPrompt, $brief);

        if ($manifest === null) {
            $job->markFailed('LEM service unavailable');

            return $job;
        }

        if (! is_array($manifest) || ! isset($manifest['clips'])) {
            $job->markFailed('Invalid manifest JSON from LEM');

            return $job;
        }

        $job->update(['manifest' => $manifest]);

        return $job;
    }

    private function buildSystemPrompt(array $assets, array $templates): string
    {
        $assetJson = json_encode($assets, JSON_PRETTY_PRINT);
        $templateJson = json_encode($templates, JSON_PRETTY_PRINT);

        return <<<PROMPT
        You are a video remix editor. Given a creative brief and a library of video assets, produce a JSON manifest for the ffmpeg render worker.

        Available assets:
        {$assetJson}

        Available templates:
        {$templateJson}

        Respond with ONLY valid JSON matching this structure:
        {
          "template": "template-name",
          "clips": [{"asset_id": N, "start": float, "end": float, "order": N}],
          "captions": [{"text": "string", "at": float, "duration": float, "style": "bold-center"}],
          "audio": {"track": "original", "fade_in": float},
          "output": {"format": "mp4", "resolution": "WxH", "fps": N}
        }

        No markdown, no explanation. Only the JSON manifest.
        PROMPT;
    }

    private function queryLem(string $system, string $brief): ?array
    {
        $ollamaUrl = config('studio.ollama.url');
        $model = config('studio.ollama.model', 'lem-4b');
        $timeout = config('studio.ollama.timeout', 60);

        $response = Http::timeout($timeout)->post("{$ollamaUrl}/api/chat", [
            'model' => $model,
            'stream' => false,
            'format' => 'json',
            'messages' => [
                ['role' => 'system', 'content' => $system],
                ['role' => 'user', 'content' => $brief],
            ],
        ]);

        if (! $response->successful()) {
            return null;
        }

        $content = $response->json('message.content', '');

        $decoded = json_decode($content, true);

        return is_array($decoded) ? $decoded : null;
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=GenerateManifestTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/GenerateManifest.php app/Mod/Studio/Tests/Feature/GenerateManifestTest.php
git commit -m "feat(studio): add GenerateManifest action — brief to LEM to JSON manifest"
```

---

### Task 10: RenderManifest Action

**Files:**
- Create: `app/Mod/Studio/Actions/RenderManifest.php`
- Create: `app/Mod/Studio/Tests/Feature/RenderManifestTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Actions\RenderManifest;
use Mod\Studio\Models\StudioJob;

it('dispatches a manifest to the ffmpeg worker', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'input' => ['brief' => 'test'],
        'manifest' => [
            'template' => 'tiktok-15s',
            'clips' => [
                ['asset_id' => 1, 'start' => 0.0, 'end' => 5.0, 'order' => 1],
            ],
            'captions' => [],
            'audio' => ['track' => 'original'],
            'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
        ],
    ]);

    Http::fake([
        '*/render' => Http::response([
            'job_id' => 'ffmpeg-001',
            'status' => 'queued',
        ]),
    ]);

    $result = RenderManifest::run($job);

    expect($result->status)->toBe('processing');
    expect($result->started_at)->not->toBeNull();

    Http::assertSent(fn ($request) => $request->url() === config('studio.worker.url').'/render'
        && $request['manifest']['template'] === 'tiktok-15s'
    );
});

it('fails the job when worker is unreachable', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'input' => ['brief' => 'test'],
        'manifest' => ['clips' => []],
    ]);

    Http::fake([
        '*/render' => Http::response([], 500),
    ]);

    $result = RenderManifest::run($job);

    expect($result->status)->toBe('failed');
    expect($result->error)->toContain('Worker');
});

it('rejects jobs without a manifest', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'input' => ['brief' => 'test'],
    ]);

    $result = RenderManifest::run($job);

    expect($result->status)->toBe('failed');
    expect($result->error)->toContain('No manifest');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=RenderManifestTest
```

Expected: FAIL — class not found.

**Step 3: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Mod\Studio\Models\StudioJob;

class RenderManifest
{
    use Action;

    public function handle(StudioJob $job): StudioJob
    {
        if (empty($job->manifest)) {
            $job->markFailed('No manifest to render');

            return $job;
        }

        $workerUrl = config('studio.worker.url');
        $timeout = config('studio.worker.timeout', 300);

        $job->markStarted();

        $response = Http::timeout($timeout)->post("{$workerUrl}/render", [
            'job_id' => $job->id,
            'manifest' => $job->manifest,
            'callback_url' => route('api.studio.remix.callback', $job->id),
        ]);

        if (! $response->successful()) {
            $job->markFailed('Worker service unavailable');

            return $job;
        }

        return $job;
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=RenderManifestTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/RenderManifest.php app/Mod/Studio/Tests/Feature/RenderManifestTest.php
git commit -m "feat(studio): add RenderManifest action — dispatch to ffmpeg worker"
```

---

### Task 11: Remix API Routes + Controller

**Files:**
- Create: `app/Mod/Studio/Controllers/Api/RemixController.php`
- Create: `app/Mod/Studio/Tests/Feature/RemixApiTest.php`
- Modify: `app/Mod/Studio/Routes/api.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Models\StudioAsset;
use Mod\Studio\Models\StudioJob;

beforeEach(function () {
    // Seed a video asset so GenerateManifest has something to work with
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'file_size' => 1024,
        'tags' => ['test'],
    ]);
});

it('submits a remix brief via POST /api/studio/remix', function () {
    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'template' => 'tiktok-15s',
                    'clips' => [['asset_id' => 1, 'start' => 0, 'end' => 5, 'order' => 1]],
                    'captions' => [],
                    'audio' => ['track' => 'original'],
                    'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
                ]),
            ],
        ]),
    ]);

    $response = $this->postJson('/api/studio/remix', [
        'brief' => 'summer lollipop TikTok, 15s, upbeat',
    ]);

    $response->assertCreated();
    $response->assertJsonPath('data.type', 'remix');
    $response->assertJsonStructure(['data' => ['id', 'type', 'status', 'manifest']]);
});

it('polls remix status via GET /api/studio/remix/{id}', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'status' => 'processing',
        'input' => ['brief' => 'test'],
        'manifest' => ['clips' => []],
    ]);

    $response = $this->getJson("/api/studio/remix/{$job->id}");

    $response->assertOk();
    $response->assertJsonPath('data.status', 'processing');
});

it('receives worker callback via POST /api/studio/remix/{id}/callback', function () {
    $job = StudioJob::create([
        'type' => 'remix',
        'status' => 'processing',
        'input' => ['brief' => 'test'],
        'manifest' => ['clips' => []],
    ]);

    $response = $this->postJson("/api/studio/remix/{$job->id}/callback", [
        'status' => 'completed',
        'output' => [
            'url' => '/renders/remix-001.mp4',
            'duration_ms' => 15000,
            'file_size' => 8192000,
        ],
    ]);

    $response->assertOk();

    $job->refresh();
    expect($job->status)->toBe('completed');
    expect($job->output['url'])->toBe('/renders/remix-001.mp4');
});

it('validates brief is required', function () {
    $response = $this->postJson('/api/studio/remix', []);

    $response->assertUnprocessable();
    $response->assertJsonValidationErrors('brief');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=RemixApiTest
```

Expected: FAIL — 404.

**Step 3: Write the controller**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Controllers\Api;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Models\StudioJob;

class RemixController extends Controller
{
    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'brief' => 'required|string|max:2000',
            'template' => 'nullable|string|max:64',
        ]);

        $job = GenerateManifest::run(
            $validated['brief'],
            $validated['template'] ?? null,
        );

        return response()->json(['data' => $job], 201);
    }

    public function show(int $id): JsonResponse
    {
        $job = StudioJob::findOrFail($id);

        return response()->json(['data' => $job]);
    }

    public function callback(Request $request, int $id): JsonResponse
    {
        $job = StudioJob::findOrFail($id);

        $validated = $request->validate([
            'status' => 'required|string|in:completed,failed',
            'output' => 'nullable|array',
            'error' => 'nullable|string',
        ]);

        if ($validated['status'] === 'completed') {
            $job->markCompleted($validated['output'] ?? []);
        } else {
            $job->markFailed($validated['error'] ?? 'Unknown error');
        }

        return response()->json(['data' => $job]);
    }
}
```

**Step 4: Update API routes**

Replace `app/Mod/Studio/Routes/api.php`:

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Mod\Studio\Controllers\Api\AssetController;
use Mod\Studio\Controllers\Api\RemixController;

Route::prefix('studio')->group(function () {
    Route::get('/assets', [AssetController::class, 'index'])->name('api.studio.assets.index');
    Route::get('/assets/{id}', [AssetController::class, 'show'])->name('api.studio.assets.show');
    Route::post('/assets', [AssetController::class, 'store'])->name('api.studio.assets.store');

    Route::post('/remix', [RemixController::class, 'store'])->name('api.studio.remix.store');
    Route::get('/remix/{id}', [RemixController::class, 'show'])->name('api.studio.remix.show');
    Route::post('/remix/{id}/callback', [RemixController::class, 'callback'])->name('api.studio.remix.callback');
});
```

**Step 5: Run test to verify it passes**

```bash
php artisan test --filter=RemixApiTest
```

Expected: 4 tests PASS.

**Step 6: Commit**

```bash
git add app/Mod/Studio/Controllers/Api/RemixController.php app/Mod/Studio/Routes/api.php app/Mod/Studio/Tests/Feature/RemixApiTest.php
git commit -m "feat(studio): add remix API endpoints (submit, poll, callback)"
```

---

### Task 12: Remix Console Command

**Files:**
- Create: `app/Mod/Studio/Console/Remix.php`
- Modify: `app/Mod/Studio/Boot.php` (register command)

**Step 1: Write the command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Actions\RenderManifest;

class Remix extends Command
{
    protected $signature = 'studio:remix
        {brief : Creative brief for the remix}
        {--template= : Template name (e.g. tiktok-15s)}
        {--render : Immediately dispatch to ffmpeg worker}
        {--json : Output manifest as JSON}';

    protected $description = 'Generate a remix from a creative brief';

    public function handle(): int
    {
        $brief = $this->argument('brief');
        $template = $this->option('template');

        $this->info("Generating manifest for: {$brief}");

        $job = GenerateManifest::run($brief, $template);

        if ($job->status === 'failed') {
            $this->error("Failed: {$job->error}");

            return Command::FAILURE;
        }

        if ($this->option('json')) {
            $this->line(json_encode($job->manifest, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));

            return Command::SUCCESS;
        }

        $this->info("Job #{$job->id} created with manifest:");
        $this->line("  Template: {$job->manifest['template']}");
        $this->line("  Clips: ".count($job->manifest['clips'] ?? []));
        $this->line("  Captions: ".count($job->manifest['captions'] ?? []));

        if ($this->option('render')) {
            $this->info('Dispatching to render worker...');
            $job = RenderManifest::run($job);

            if ($job->status === 'failed') {
                $this->error("Render failed: {$job->error}");

                return Command::FAILURE;
            }

            $this->info("Rendering... Job #{$job->id} status: {$job->status}");
        }

        return Command::SUCCESS;
    }
}
```

**Step 2: Register in Boot.php**

Update the `onConsole` method:

```php
    public function onConsole(ConsoleBooting $event): void
    {
        $event->command(Console\Catalogue::class);
        $event->command(Console\Transcribe::class);
        $event->command(Console\Remix::class);
    }
```

**Step 3: Verify command registers**

```bash
php artisan list studio
```

Expected: Shows `studio:catalogue`, `studio:remix`, `studio:transcribe`.

**Step 4: Commit**

```bash
git add app/Mod/Studio/Console/Remix.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add studio:remix artisan command"
```

---

### Task 13: Livewire Asset Browser

**Files:**
- Create: `app/Mod/Studio/Livewire/AssetBrowserPage.php`
- Create: `app/Mod/Studio/Views/asset-browser.blade.php`
- Create: `app/Mod/Studio/Views/layouts/studio.blade.php`
- Modify: `app/Mod/Studio/Routes/web.php`
- Modify: `app/Mod/Studio/Boot.php` (register Livewire components)

**Step 1: Create the layout**

Create `app/Mod/Studio/Views/layouts/studio.blade.php`:

```blade
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ $title ?? 'Studio' }}</title>
    @vite(['resources/css/app.css', 'resources/js/app.js'])
    @livewireStyles
    @fluxStyles
</head>
<body class="min-h-screen bg-zinc-50 dark:bg-zinc-900">
    <nav class="border-b border-zinc-200 bg-white px-6 py-3 dark:border-zinc-700 dark:bg-zinc-800">
        <div class="flex items-center justify-between">
            <a href="{{ route('studio.assets') }}" class="text-lg font-semibold text-zinc-900 dark:text-white">
                <i class="fa-solid fa-film mr-2"></i>Studio
            </a>
            <div class="flex items-center gap-4">
                <a href="{{ route('studio.assets') }}" class="text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400">Assets</a>
                <a href="{{ route('studio.remix') }}" class="text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400">Remix</a>
            </div>
        </div>
    </nav>

    <main class="mx-auto max-w-7xl px-6 py-8">
        {{ $slot }}
    </main>

    @livewireScripts
    @fluxScripts
</body>
</html>
```

**Step 2: Create the Livewire component**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Livewire;

use Illuminate\Contracts\View\View;
use Livewire\Attributes\Layout;
use Livewire\Attributes\Title;
use Livewire\Attributes\Url;
use Livewire\Component;
use Livewire\WithPagination;
use Mod\Studio\Models\StudioAsset;

#[Layout('studio::layouts.studio')]
#[Title('Assets — Studio')]
class AssetBrowserPage extends Component
{
    use WithPagination;

    #[Url]
    public string $search = '';

    #[Url]
    public string $type = '';

    #[Url]
    public string $tag = '';

    public function updatedSearch(): void
    {
        $this->resetPage();
    }

    public function updatedType(): void
    {
        $this->resetPage();
    }

    public function updatedTag(): void
    {
        $this->resetPage();
    }

    public function render(): View
    {
        $query = StudioAsset::query()->latest();

        if ($this->search) {
            $query->where('filename', 'like', "%{$this->search}%");
        }

        if ($this->type) {
            match ($this->type) {
                'video' => $query->videos(),
                'image' => $query->images(),
                'audio' => $query->audio(),
                default => null,
            };
        }

        if ($this->tag) {
            $query->tagged($this->tag);
        }

        return view('studio::asset-browser', [
            'assets' => $query->paginate(24),
        ]);
    }
}
```

**Step 3: Create the Blade view**

Create `app/Mod/Studio/Views/asset-browser.blade.php`:

```blade
<div>
    <div class="mb-6 flex items-center justify-between">
        <h1 class="text-2xl font-bold text-zinc-900 dark:text-white">Asset Library</h1>
        <flux:button href="{{ route('studio.remix') }}" variant="primary">
            <i class="fa-solid fa-wand-magic-sparkles mr-1"></i> New Remix
        </flux:button>
    </div>

    <div class="mb-6 flex flex-wrap gap-3">
        <flux:input wire:model.live.debounce.300ms="search" placeholder="Search assets..." class="w-64" />

        <flux:select wire:model.live="type" class="w-40">
            <option value="">All Types</option>
            <option value="video">Video</option>
            <option value="image">Image</option>
            <option value="audio">Audio</option>
        </flux:select>

        <flux:input wire:model.live.debounce.300ms="tag" placeholder="Filter by tag..." class="w-48" />
    </div>

    @if($assets->isEmpty())
        <div class="rounded-lg border border-dashed border-zinc-300 p-12 text-center dark:border-zinc-600">
            <i class="fa-solid fa-photo-film mb-3 text-4xl text-zinc-400"></i>
            <p class="text-zinc-500 dark:text-zinc-400">No assets found. Upload some videos to get started.</p>
        </div>
    @else
        <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
            @foreach($assets as $asset)
                <div class="group rounded-lg border border-zinc-200 bg-white p-3 transition hover:shadow-md dark:border-zinc-700 dark:bg-zinc-800">
                    <div class="mb-2 flex h-24 items-center justify-center rounded bg-zinc-100 dark:bg-zinc-700">
                        @if(str_starts_with($asset->mime_type, 'video/'))
                            <i class="fa-solid fa-video text-2xl text-zinc-400"></i>
                        @elseif(str_starts_with($asset->mime_type, 'image/'))
                            <i class="fa-solid fa-image text-2xl text-zinc-400"></i>
                        @else
                            <i class="fa-solid fa-music text-2xl text-zinc-400"></i>
                        @endif
                    </div>
                    <p class="truncate text-sm font-medium text-zinc-900 dark:text-white" title="{{ $asset->filename }}">
                        {{ $asset->filename }}
                    </p>
                    <p class="text-xs text-zinc-500">
                        {{ $asset->duration_ms ? round($asset->duration_ms / 1000, 1) . 's' : '—' }}
                        · {{ number_format($asset->file_size / 1024 / 1024, 1) }}MB
                    </p>
                    @if($asset->tags)
                        <div class="mt-1 flex flex-wrap gap-1">
                            @foreach(array_slice($asset->tags, 0, 3) as $t)
                                <span class="rounded bg-zinc-100 px-1.5 py-0.5 text-xs text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300">{{ $t }}</span>
                            @endforeach
                        </div>
                    @endif
                    @if($asset->transcript)
                        <p class="mt-1 truncate text-xs text-green-600 dark:text-green-400" title="Transcribed">
                            <i class="fa-solid fa-check-circle"></i> Transcribed
                        </p>
                    @endif
                </div>
            @endforeach
        </div>

        <div class="mt-6">
            {{ $assets->links() }}
        </div>
    @endif
</div>
```

**Step 4: Register Livewire component and routes**

Update `onWebRoutes` in Boot.php:

```php
    public function onWebRoutes(WebRoutesRegistering $event): void
    {
        $event->views('studio', __DIR__.'/Views');
        $event->livewire('mod.studio.livewire.asset-browser-page', Livewire\AssetBrowserPage::class);
        $event->routes(fn () => require __DIR__.'/Routes/web.php');
    }
```

Update `app/Mod/Studio/Routes/web.php`:

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Mod\Studio\Livewire\AssetBrowserPage;

Route::prefix('studio')->group(function () {
    Route::get('/', AssetBrowserPage::class)->name('studio.assets');
});
```

**Step 5: Commit**

```bash
git add app/Mod/Studio/Livewire/ app/Mod/Studio/Views/ app/Mod/Studio/Routes/web.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add Livewire asset browser with search, type filter, tag filter"
```

---

### Task 14: Livewire Remix Form + Job Status

**Files:**
- Create: `app/Mod/Studio/Livewire/RemixPage.php`
- Create: `app/Mod/Studio/Views/remix.blade.php`
- Modify: `app/Mod/Studio/Routes/web.php`
- Modify: `app/Mod/Studio/Boot.php`

**Step 1: Create the Livewire component**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Livewire;

use Illuminate\Contracts\View\View;
use Livewire\Attributes\Layout;
use Livewire\Attributes\Title;
use Livewire\Component;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Actions\RenderManifest;
use Mod\Studio\Models\StudioJob;

#[Layout('studio::layouts.studio')]
#[Title('Remix — Studio')]
class RemixPage extends Component
{
    public string $brief = '';

    public string $template = 'tiktok-15s';

    public ?int $activeJobId = null;

    public function submit(): void
    {
        $this->validate([
            'brief' => 'required|string|max:2000',
            'template' => 'required|string|max:64',
        ]);

        $job = GenerateManifest::run($this->brief, $this->template);

        $this->activeJobId = $job->id;
        $this->brief = '';
    }

    public function render(): void
    {
        $this->dispatch('poll');
    }

    public function dispatchRender(): void
    {
        if (! $this->activeJobId) {
            return;
        }

        $job = StudioJob::find($this->activeJobId);

        if ($job && $job->manifest && $job->status === 'pending') {
            RenderManifest::run($job);
        }
    }

    public function render(): View
    {
        $job = $this->activeJobId ? StudioJob::find($this->activeJobId) : null;
        $templates = config('studio.templates', []);
        $recentJobs = StudioJob::latest()->limit(10)->get();

        return view('studio::remix', [
            'job' => $job,
            'templates' => $templates,
            'recentJobs' => $recentJobs,
        ]);
    }
}
```

**Step 2: Create the Blade view**

Create `app/Mod/Studio/Views/remix.blade.php`:

```blade
<div>
    <h1 class="mb-6 text-2xl font-bold text-zinc-900 dark:text-white">
        <i class="fa-solid fa-wand-magic-sparkles mr-2"></i>Remix
    </h1>

    <div class="grid gap-8 lg:grid-cols-2">
        {{-- Brief Form --}}
        <div>
            <form wire:submit="submit" class="space-y-4">
                <flux:textarea
                    wire:model="brief"
                    label="Creative Brief"
                    placeholder="summer lollipop TikTok, 15s, upbeat vibe, quick cuts..."
                    rows="4"
                />

                <flux:select wire:model="template" label="Template">
                    @foreach($templates as $name => $config)
                        <option value="{{ $name }}">{{ $name }} ({{ $config['duration'] }}s, {{ $config['resolution'] }})</option>
                    @endforeach
                </flux:select>

                <flux:button type="submit" variant="primary" wire:loading.attr="disabled">
                    <span wire:loading.remove>Generate Manifest</span>
                    <span wire:loading><i class="fa-solid fa-spinner fa-spin mr-1"></i> Generating...</span>
                </flux:button>
            </form>

            @error('brief')
                <p class="mt-2 text-sm text-red-600">{{ $message }}</p>
            @enderror
        </div>

        {{-- Active Job Status --}}
        <div>
            @if($job)
                <div class="rounded-lg border border-zinc-200 bg-white p-6 dark:border-zinc-700 dark:bg-zinc-800"
                     @if(in_array($job->status, ['pending', 'processing'])) wire:poll.2s @endif>
                    <div class="mb-4 flex items-center justify-between">
                        <h2 class="text-lg font-semibold text-zinc-900 dark:text-white">Job #{{ $job->id }}</h2>
                        <span @class([
                            'rounded-full px-3 py-1 text-xs font-medium',
                            'bg-yellow-100 text-yellow-800' => $job->status === 'pending',
                            'bg-blue-100 text-blue-800' => $job->status === 'processing',
                            'bg-green-100 text-green-800' => $job->status === 'completed',
                            'bg-red-100 text-red-800' => $job->status === 'failed',
                        ])>
                            {{ ucfirst($job->status) }}
                        </span>
                    </div>

                    @if($job->manifest)
                        <div class="mb-4 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                            <p><strong>Template:</strong> {{ $job->manifest['template'] ?? '—' }}</p>
                            <p><strong>Clips:</strong> {{ count($job->manifest['clips'] ?? []) }}</p>
                            <p><strong>Captions:</strong> {{ count($job->manifest['captions'] ?? []) }}</p>
                        </div>

                        @if($job->status === 'pending')
                            <flux:button wire:click="dispatchRender" variant="primary" size="sm">
                                <i class="fa-solid fa-play mr-1"></i> Render
                            </flux:button>
                        @endif

                        <details class="mt-4">
                            <summary class="cursor-pointer text-sm text-zinc-500">Manifest JSON</summary>
                            <pre class="mt-2 overflow-auto rounded bg-zinc-100 p-3 text-xs dark:bg-zinc-900">{{ json_encode($job->manifest, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) }}</pre>
                        </details>
                    @endif

                    @if($job->error)
                        <div class="mt-4 rounded bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-400">
                            {{ $job->error }}
                        </div>
                    @endif

                    @if($job->output)
                        <div class="mt-4">
                            <p class="text-sm font-medium text-green-700 dark:text-green-400">
                                <i class="fa-solid fa-check-circle mr-1"></i> Render complete
                            </p>
                            @if(isset($job->output['url']))
                                <a href="{{ $job->output['url'] }}" class="mt-2 inline-block text-sm text-blue-600 hover:underline">
                                    <i class="fa-solid fa-download mr-1"></i> Download
                                </a>
                            @endif
                        </div>
                    @endif
                </div>
            @endif
        </div>
    </div>

    {{-- Recent Jobs --}}
    @if($recentJobs->isNotEmpty())
        <div class="mt-12">
            <h2 class="mb-4 text-lg font-semibold text-zinc-900 dark:text-white">Recent Jobs</h2>
            <div class="overflow-hidden rounded-lg border border-zinc-200 dark:border-zinc-700">
                <table class="w-full text-sm">
                    <thead class="bg-zinc-50 dark:bg-zinc-800">
                        <tr>
                            <th class="px-4 py-2 text-left text-zinc-600 dark:text-zinc-300">ID</th>
                            <th class="px-4 py-2 text-left text-zinc-600 dark:text-zinc-300">Brief</th>
                            <th class="px-4 py-2 text-left text-zinc-600 dark:text-zinc-300">Status</th>
                            <th class="px-4 py-2 text-left text-zinc-600 dark:text-zinc-300">Created</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                        @foreach($recentJobs as $recentJob)
                            <tr class="bg-white dark:bg-zinc-800">
                                <td class="px-4 py-2 font-mono text-zinc-900 dark:text-white">#{{ $recentJob->id }}</td>
                                <td class="max-w-xs truncate px-4 py-2 text-zinc-600 dark:text-zinc-300">{{ $recentJob->input['brief'] ?? '—' }}</td>
                                <td class="px-4 py-2">
                                    <span @class([
                                        'rounded-full px-2 py-0.5 text-xs font-medium',
                                        'bg-yellow-100 text-yellow-800' => $recentJob->status === 'pending',
                                        'bg-blue-100 text-blue-800' => $recentJob->status === 'processing',
                                        'bg-green-100 text-green-800' => $recentJob->status === 'completed',
                                        'bg-red-100 text-red-800' => $recentJob->status === 'failed',
                                    ])>{{ $recentJob->status }}</span>
                                </td>
                                <td class="px-4 py-2 text-zinc-500">{{ $recentJob->created_at->diffForHumans() }}</td>
                            </tr>
                        @endforeach
                    </tbody>
                </table>
            </div>
        </div>
    @endif
</div>
```

**Step 3: Register component and route**

Update `onWebRoutes` in Boot.php — add the remix page:

```php
    public function onWebRoutes(WebRoutesRegistering $event): void
    {
        $event->views('studio', __DIR__.'/Views');
        $event->livewire('mod.studio.livewire.asset-browser-page', Livewire\AssetBrowserPage::class);
        $event->livewire('mod.studio.livewire.remix-page', Livewire\RemixPage::class);
        $event->routes(fn () => require __DIR__.'/Routes/web.php');
    }
```

Update `app/Mod/Studio/Routes/web.php`:

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Mod\Studio\Livewire\AssetBrowserPage;
use Mod\Studio\Livewire\RemixPage;

Route::prefix('studio')->group(function () {
    Route::get('/', AssetBrowserPage::class)->name('studio.assets');
    Route::get('/remix', RemixPage::class)->name('studio.remix');
});
```

**Step 4: Commit**

```bash
git add app/Mod/Studio/Livewire/RemixPage.php app/Mod/Studio/Views/remix.blade.php app/Mod/Studio/Routes/web.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add Livewire remix form with job status and polling"
```

---

### Task 15: Run Full Test Suite

**Step 1: Run all Studio tests**

```bash
php artisan test --filter=Studio
```

Expected: All tests pass (StudioAsset, StudioJob, CatalogueAsset, TranscribeAsset, GenerateManifest, RenderManifest, AssetApi, RemixApi).

**Step 2: Run lint**

```bash
./vendor/bin/pint --dirty
```

Expected: All files formatted.

**Step 3: Final commit if lint changed anything**

```bash
git add -A && git commit -m "style(studio): format code with Pint"
```

---

## Module Directory Structure (Final)

```
app/Mod/Studio/
├── Boot.php
├── Actions/
│   ├── CatalogueAsset.php
│   ├── GenerateManifest.php
│   ├── RenderManifest.php
│   └── TranscribeAsset.php
├── Console/
│   ├── Catalogue.php
│   ├── Remix.php
│   └── Transcribe.php
├── Controllers/
│   └── Api/
│       ├── AssetController.php
│       └── RemixController.php
├── Livewire/
│   ├── AssetBrowserPage.php
│   └── RemixPage.php
├── Migrations/
│   └── 2026_03_08_000001_create_studio_tables.php
├── Models/
│   ├── StudioAsset.php
│   └── StudioJob.php
├── Routes/
│   ├── api.php
│   └── web.php
├── Tests/
│   └── Feature/
│       ├── AssetApiTest.php
│       ├── CatalogueAssetTest.php
│       ├── GenerateManifestTest.php
│       ├── RemixApiTest.php
│       ├── RenderManifestTest.php
│       ├── StudioAssetTest.php
│       ├── StudioJobTest.php
│       └── TranscribeAssetTest.php
└── Views/
    ├── layouts/
    │   └── studio.blade.php
    ├── asset-browser.blade.php
    └── remix.blade.php

config/studio.php
```

## Phase 3-5 (Future — Not Implemented Here)

**Phase 3 — Voice & TTS**: Add `SynthesiseSpeech` action, `studio:voice` command, TTS service integration.

**Phase 4 — Visual Generation**: Add ComfyUI integration, `GenerateThumbnail` action, image overlay manifest support.

**Phase 5 — Production**: Authentik account for client, studio.lthn.ai deployment, 66analytics usage tracking.

These phases build on the same patterns established above — new Actions, new API routes, new Livewire pages. Same module, just more endpoints.
