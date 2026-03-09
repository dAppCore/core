# Studio Phases 3-5 — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the Studio module with TTS voice pipeline, visual generation via ComfyUI, and production deployment at studio.lthn.ai — enabling content flywheel for remaking existing footage at higher quality standards.

**Architecture:** Same smart/dumb pattern as Phases 1-2. New Actions follow TranscribeAsset's HTTP dispatch pattern. TTS service (Kokoro) and ComfyUI are Docker containers on homelab accessed over HTTP. Production deployment via Ansible. Content flywheel adds batch queue processing for high-volume remix work.

**Tech Stack:** Laravel 12, Livewire 3, Flux Pro, Pest, Ansible, Docker (ROCm), Kokoro TTS, ComfyUI, Traefik, Authentik, 66analytics

**Existing module:** `/Users/snider/Code/lab/host.uk.com/app/Mod/Studio/` (36 tests, 145 assertions)

**Design doc:** `docs/plans/2026-03-08-studio-multimedia-pipeline-design.md`

---

## Phase 3 — Voice & TTS

### Task 1: SynthesiseSpeech Action

**Files:**
- Create: `app/Mod/Studio/Actions/SynthesiseSpeech.php`
- Create: `app/Mod/Studio/Tests/Feature/SynthesiseSpeechTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\SynthesiseSpeech;

it('synthesises speech from text via TTS service', function () {
    Storage::fake('local');

    Http::fake([
        '*/synthesise' => Http::response('fake-audio-bytes', 200, [
            'Content-Type' => 'audio/wav',
        ]),
    ]);

    $result = SynthesiseSpeech::run('Hello, welcome to our summer collection.');

    expect($result)->toBeArray();
    expect($result['path'])->toContain('studio/renders/');
    expect($result['mime_type'])->toBe('audio/wav');
    expect($result['duration_ms'])->toBeNull();
    Storage::disk('local')->assertExists($result['path']);
});

it('synthesises with custom voice', function () {
    Storage::fake('local');

    Http::fake([
        '*/synthesise' => Http::response('fake-audio-bytes', 200, [
            'Content-Type' => 'audio/wav',
        ]),
    ]);

    $result = SynthesiseSpeech::run('Test text', voice: 'sarah');

    expect($result)->toBeArray();

    Http::assertSent(fn ($request) => $request['voice'] === 'sarah');
});

it('returns null on TTS failure', function () {
    Http::fake([
        '*/synthesise' => Http::response([], 500),
    ]);

    $result = SynthesiseSpeech::run('Test text');

    expect($result)->toBeNull();
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=SynthesiseSpeechTest
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
use Illuminate\Support\Str;

class SynthesiseSpeech
{
    use Action;

    /**
     * @return array{path: string, mime_type: string, duration_ms: int|null}|null
     */
    public function handle(string $text, ?string $voice = null): ?array
    {
        $ttsUrl = config('studio.tts.url');
        $timeout = config('studio.tts.timeout', 60);
        $defaultVoice = config('studio.tts.voice', 'default');

        $response = Http::timeout($timeout)
            ->post("{$ttsUrl}/synthesise", [
                'text' => $text,
                'voice' => $voice ?? $defaultVoice,
            ]);

        if (! $response->successful()) {
            return null;
        }

        $disk = config('studio.storage.disk', 'local');
        $rendersPath = config('studio.storage.renders_path', 'studio/renders');
        $filename = 'tts-'.Str::random(12).'.wav';
        $path = "{$rendersPath}/{$filename}";

        Storage::disk($disk)->put($path, $response->body());

        return [
            'path' => $path,
            'mime_type' => $response->header('Content-Type', 'audio/wav'),
            'duration_ms' => null,
        ];
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=SynthesiseSpeechTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/SynthesiseSpeech.php app/Mod/Studio/Tests/Feature/SynthesiseSpeechTest.php
git commit -m "feat(studio): add SynthesiseSpeech action for TTS"
```

---

### Task 2: GenerateVoiceover Action

**Files:**
- Create: `app/Mod/Studio/Actions/GenerateVoiceover.php`
- Create: `app/Mod/Studio/Tests/Feature/GenerateVoiceoverTest.php`

This action chains LEM (script generation) → TTS (speech synthesis). Given a brief and optional context, LEM writes a voiceover script, then TTS renders it to audio.

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\GenerateVoiceover;
use Mod\Studio\Models\StudioJob;

it('generates a voiceover from a brief', function () {
    Storage::fake('local');

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'script' => 'Welcome to our summer collection. Sun, sand, and style.',
                    'voice' => 'default',
                    'pace' => 'medium',
                ]),
            ],
        ]),
        '*/synthesise' => Http::response('fake-audio-bytes', 200, [
            'Content-Type' => 'audio/wav',
        ]),
    ]);

    $job = GenerateVoiceover::run('upbeat summer voiceover, 10 seconds');

    expect($job)->toBeInstanceOf(StudioJob::class);
    expect($job->type)->toBe('voiceover');
    expect($job->status)->toBe('completed');
    expect($job->output)->toHaveKeys(['script', 'audio_path']);
    expect($job->output['script'])->toContain('summer');
});

it('fails when LEM returns invalid script JSON', function () {
    Http::fake([
        '*/api/chat' => Http::response([
            'message' => ['content' => 'not valid json'],
        ]),
    ]);

    $job = GenerateVoiceover::run('test brief');

    expect($job->status)->toBe('failed');
    expect($job->error)->toContain('script');
});

it('fails when TTS service is unavailable', function () {
    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'script' => 'Test voiceover text.',
                    'voice' => 'default',
                ]),
            ],
        ]),
        '*/synthesise' => Http::response([], 500),
    ]);

    $job = GenerateVoiceover::run('test brief');

    expect($job->status)->toBe('failed');
    expect($job->error)->toContain('TTS');
});
```

**Step 2: Run test to verify it fails**

```bash
php artisan test --filter=GenerateVoiceoverTest
```

**Step 3: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Mod\Studio\Models\StudioJob;

class GenerateVoiceover
{
    use Action;

    public function handle(string $brief, ?string $voice = null): StudioJob
    {
        $job = StudioJob::create([
            'type' => 'voiceover',
            'input' => [
                'brief' => $brief,
                'voice' => $voice,
            ],
        ]);

        $script = $this->generateScript($brief);

        if ($script === null) {
            $job->markFailed('LEM service unavailable');

            return $job;
        }

        if (! is_array($script) || ! isset($script['script'])) {
            $job->markFailed('Invalid script JSON from LEM');

            return $job;
        }

        $audio = SynthesiseSpeech::run(
            $script['script'],
            voice: $voice ?? $script['voice'] ?? null,
        );

        if ($audio === null) {
            $job->markFailed('TTS service unavailable');

            return $job;
        }

        $job->markCompleted([
            'script' => $script['script'],
            'voice' => $script['voice'] ?? config('studio.tts.voice'),
            'audio_path' => $audio['path'],
            'audio_mime' => $audio['mime_type'],
        ]);

        return $job;
    }

    private function generateScript(string $brief): ?array
    {
        $ollamaUrl = config('studio.ollama.url');
        $model = config('studio.ollama.model', 'lem-4b');
        $timeout = config('studio.ollama.timeout', 60);

        $response = Http::timeout($timeout)->post("{$ollamaUrl}/api/chat", [
            'model' => $model,
            'stream' => false,
            'format' => 'json',
            'messages' => [
                ['role' => 'system', 'content' => 'You are a voiceover scriptwriter. Given a brief, write a natural voiceover script. Respond with ONLY valid JSON: {"script": "the voiceover text", "voice": "default", "pace": "medium"}. No markdown.'],
                ['role' => 'user', 'content' => $brief],
            ],
        ]);

        if (! $response->successful()) {
            return null;
        }

        $content = $response->json('message.content', '');

        return is_string($content) ? json_decode($content, true) : null;
    }
}
```

**Step 4: Run test to verify it passes**

```bash
php artisan test --filter=GenerateVoiceoverTest
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add app/Mod/Studio/Actions/GenerateVoiceover.php app/Mod/Studio/Tests/Feature/GenerateVoiceoverTest.php
git commit -m "feat(studio): add GenerateVoiceover action — LEM script + TTS synthesis"
```

---

### Task 3: Voice Console Command + API

**Files:**
- Create: `app/Mod/Studio/Console/Voice.php`
- Create: `app/Mod/Studio/Controllers/Api/VoiceController.php`
- Create: `app/Mod/Studio/Tests/Feature/VoiceApiTest.php`
- Modify: `app/Mod/Studio/Routes/api.php`
- Modify: `app/Mod/Studio/Boot.php`

**Step 1: Write the failing API test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;

it('submits a voiceover brief via POST /studio/voice', function () {
    Storage::fake('local');

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'script' => 'Test voiceover.',
                    'voice' => 'default',
                ]),
            ],
        ]),
        '*/synthesise' => Http::response('audio-bytes', 200, [
            'Content-Type' => 'audio/wav',
        ]),
    ]);

    $response = $this->postJson('/studio/voice', [
        'brief' => 'upbeat summer voiceover',
    ]);

    $response->assertCreated();
    $response->assertJsonPath('data.type', 'voiceover');
    $response->assertJsonPath('data.status', 'completed');
});

it('polls voiceover status via GET /studio/voice/{id}', function () {
    $job = \Mod\Studio\Models\StudioJob::create([
        'type' => 'voiceover',
        'status' => 'completed',
        'input' => ['brief' => 'test'],
        'output' => ['script' => 'Hello', 'audio_path' => 'studio/renders/test.wav'],
    ]);

    $response = $this->getJson("/studio/voice/{$job->id}");

    $response->assertOk();
    $response->assertJsonPath('data.output.script', 'Hello');
});

it('validates brief is required for voiceover', function () {
    $response = $this->postJson('/studio/voice', []);

    $response->assertUnprocessable();
});
```

**Step 2: Write the controller**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Controllers\Api;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Mod\Studio\Actions\GenerateVoiceover;
use Mod\Studio\Models\StudioJob;

class VoiceController extends Controller
{
    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'brief' => 'required|string|max:2000',
            'voice' => 'nullable|string|max:64',
        ]);

        $job = GenerateVoiceover::run(
            $validated['brief'],
            $validated['voice'] ?? null,
        );

        return response()->json(['data' => $job], 201);
    }

    public function show(int $id): JsonResponse
    {
        $job = StudioJob::findOrFail($id);

        return response()->json(['data' => $job]);
    }
}
```

**Step 3: Write the console command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Mod\Studio\Actions\GenerateVoiceover;

class Voice extends Command
{
    protected $signature = 'studio:voice
        {brief : Description of the voiceover}
        {--voice= : Voice to use}
        {--json : Output as JSON}';

    protected $description = 'Generate a voiceover from a brief';

    public function handle(): int
    {
        $brief = $this->argument('brief');
        $voice = $this->option('voice');

        $this->info("Generating voiceover: {$brief}");

        $job = GenerateVoiceover::run($brief, $voice);

        if ($job->status === 'failed') {
            $this->error("Failed: {$job->error}");

            return Command::FAILURE;
        }

        if ($this->option('json')) {
            $this->line(json_encode($job->output, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));

            return Command::SUCCESS;
        }

        $this->info("Job #{$job->id} completed");
        $this->line("  Script: {$job->output['script']}");
        $this->line("  Audio: {$job->output['audio_path']}");

        return Command::SUCCESS;
    }
}
```

**Step 4: Update routes and Boot.php**

Add to `api.php` inside the studio prefix group:
```php
Route::post('/voice', [VoiceController::class, 'store'])->name('api.studio.voice.store');
Route::get('/voice/{id}', [VoiceController::class, 'show'])->name('api.studio.voice.show');
```

Add to Boot.php `onConsole()`:
```php
$event->command(Console\Voice::class);
```

**Step 5: Run tests**

```bash
php artisan test --filter=Studio
```

Expected: All tests pass.

**Step 6: Commit**

```bash
git add app/Mod/Studio/Console/Voice.php app/Mod/Studio/Controllers/Api/VoiceController.php app/Mod/Studio/Tests/Feature/VoiceApiTest.php app/Mod/Studio/Routes/api.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add voice API, CLI command, and voiceover generation"
```

---

### Task 4: Manifest Voiceover Support

**Files:**
- Modify: `app/Mod/Studio/Actions/GenerateManifest.php`
- Create: `app/Mod/Studio/Tests/Feature/ManifestVoiceoverTest.php`

Extend the manifest format to support a `voiceover` field. When present, the remix pipeline generates TTS audio and mixes it into the rendered video.

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Models\StudioAsset;

it('accepts manifests with voiceover field', function () {
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'file_size' => 1024,
    ]);

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'template' => 'tiktok-15s',
                    'clips' => [['asset_id' => 1, 'start' => 0, 'end' => 5, 'order' => 1]],
                    'captions' => [],
                    'voiceover' => [
                        'script' => 'Summer is here!',
                        'voice' => 'default',
                        'volume' => 0.8,
                    ],
                    'audio' => ['track' => 'voiceover', 'fade_in' => 0.5],
                    'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
                ]),
            ],
        ]),
    ]);

    $job = GenerateManifest::run('summer TikTok with voiceover');

    expect($job->manifest)->toHaveKey('voiceover');
    expect($job->manifest['voiceover']['script'])->toBe('Summer is here!');
    expect($job->manifest['audio']['track'])->toBe('voiceover');
});
```

**Step 2: Update system prompt in GenerateManifest**

Add voiceover to the manifest schema in the system prompt:

```
"voiceover": {"script": "text to speak", "voice": "default", "volume": 0.8}  (optional)
```

And update audio track options: `"track": "original" or "voiceover" or "music"`

**Step 3: Run tests**

```bash
php artisan test --filter=Studio
```

Expected: All tests pass (new + existing).

**Step 4: Commit**

```bash
git add app/Mod/Studio/Actions/GenerateManifest.php app/Mod/Studio/Tests/Feature/ManifestVoiceoverTest.php
git commit -m "feat(studio): extend manifest format with voiceover support"
```

---

## Phase 4 — Visual Generation

### Task 5: GenerateImage Action

**Files:**
- Create: `app/Mod/Studio/Actions/GenerateImage.php`
- Create: `app/Mod/Studio/Tests/Feature/GenerateImageTest.php`

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\GenerateImage;

it('generates an image via ComfyUI', function () {
    Storage::fake('local');

    Http::fake([
        '*/prompt' => Http::response([
            'prompt_id' => 'comfy-001',
        ]),
        '*/history/comfy-001' => Http::response([
            'comfy-001' => [
                'outputs' => [
                    'node_id' => [
                        'images' => [
                            ['filename' => 'output.png', 'subfolder' => '', 'type' => 'output'],
                        ],
                    ],
                ],
            ],
        ]),
        '*/view*' => Http::response('fake-png-bytes', 200, [
            'Content-Type' => 'image/png',
        ]),
    ]);

    $result = GenerateImage::run(
        prompt: 'summer beach sunset, vibrant colours, TikTok thumbnail style',
        width: 1080,
        height: 1920,
    );

    expect($result)->toBeArray();
    expect($result['path'])->toContain('studio/renders/');
    expect($result['prompt_id'])->toBe('comfy-001');
    Storage::disk('local')->assertExists($result['path']);
});

it('returns null when ComfyUI is unreachable', function () {
    Http::fake([
        '*/prompt' => Http::response([], 500),
    ]);

    $result = GenerateImage::run(prompt: 'test prompt');

    expect($result)->toBeNull();
});
```

**Step 2: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;

class GenerateImage
{
    use Action;

    /**
     * @return array{path: string, prompt_id: string, mime_type: string}|null
     */
    public function handle(
        string $prompt,
        int $width = 1024,
        int $height = 1024,
        ?string $workflow = null,
    ): ?array {
        $comfyUrl = config('studio.comfyui.url', 'http://studio-comfyui:8188');
        $timeout = config('studio.comfyui.timeout', 300);

        $workflowData = $this->buildWorkflow($prompt, $width, $height, $workflow);

        $response = Http::timeout($timeout)
            ->post("{$comfyUrl}/prompt", [
                'prompt' => $workflowData,
            ]);

        if (! $response->successful()) {
            return null;
        }

        $promptId = $response->json('prompt_id');

        if (! $promptId) {
            return null;
        }

        $imageData = $this->pollForResult($comfyUrl, $promptId, $timeout);

        if ($imageData === null) {
            return null;
        }

        $disk = config('studio.storage.disk', 'local');
        $rendersPath = config('studio.storage.renders_path', 'studio/renders');
        $filename = 'img-'.Str::random(12).'.png';
        $path = "{$rendersPath}/{$filename}";

        Storage::disk($disk)->put($path, $imageData);

        return [
            'path' => $path,
            'prompt_id' => $promptId,
            'mime_type' => 'image/png',
        ];
    }

    private function buildWorkflow(string $prompt, int $width, int $height, ?string $workflow): array
    {
        // Default Flux txt2img workflow
        return [
            '1' => [
                'class_type' => 'CheckpointLoaderSimple',
                'inputs' => ['ckpt_name' => config('studio.comfyui.model', 'flux1-dev.safetensors')],
            ],
            '2' => [
                'class_type' => 'CLIPTextEncode',
                'inputs' => ['text' => $prompt, 'clip' => ['1', 1]],
            ],
            '3' => [
                'class_type' => 'EmptyLatentImage',
                'inputs' => ['width' => $width, 'height' => $height, 'batch_size' => 1],
            ],
            '4' => [
                'class_type' => 'KSampler',
                'inputs' => [
                    'model' => ['1', 0],
                    'positive' => ['2', 0],
                    'negative' => ['5', 0],
                    'latent_image' => ['3', 0],
                    'seed' => random_int(0, PHP_INT_MAX),
                    'steps' => 20,
                    'cfg' => 7.0,
                    'sampler_name' => 'euler',
                    'scheduler' => 'normal',
                ],
            ],
            '5' => [
                'class_type' => 'CLIPTextEncode',
                'inputs' => ['text' => 'blurry, low quality, watermark, text', 'clip' => ['1', 1]],
            ],
            '6' => [
                'class_type' => 'VAEDecode',
                'inputs' => ['samples' => ['4', 0], 'vae' => ['1', 2]],
            ],
            '7' => [
                'class_type' => 'SaveImage',
                'inputs' => ['images' => ['6', 0], 'filename_prefix' => 'studio'],
            ],
        ];
    }

    private function pollForResult(string $comfyUrl, string $promptId, int $timeout): ?string
    {
        $deadline = time() + min($timeout, 300);

        while (time() < $deadline) {
            $history = Http::timeout(10)->get("{$comfyUrl}/history/{$promptId}");

            if (! $history->successful()) {
                usleep(2_000_000);
                continue;
            }

            $data = $history->json("{$promptId}.outputs");

            if ($data === null) {
                usleep(2_000_000);
                continue;
            }

            // Find first image output
            foreach ($data as $nodeOutput) {
                if (isset($nodeOutput['images'][0])) {
                    $img = $nodeOutput['images'][0];
                    $imageResponse = Http::timeout(30)
                        ->get("{$comfyUrl}/view", [
                            'filename' => $img['filename'],
                            'subfolder' => $img['subfolder'] ?? '',
                            'type' => $img['type'] ?? 'output',
                        ]);

                    if ($imageResponse->successful()) {
                        return $imageResponse->body();
                    }
                }
            }

            usleep(2_000_000);
        }

        return null;
    }
}
```

**Step 3: Add ComfyUI config**

Add to `config/studio.php`:
```php
'comfyui' => [
    'url' => env('STUDIO_COMFYUI_URL', 'http://studio-comfyui:8188'),
    'model' => env('STUDIO_COMFYUI_MODEL', 'flux1-dev.safetensors'),
    'timeout' => (int) env('STUDIO_COMFYUI_TIMEOUT', 300),
],
```

**Step 4: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Actions/GenerateImage.php app/Mod/Studio/Tests/Feature/GenerateImageTest.php config/studio.php
git commit -m "feat(studio): add GenerateImage action for ComfyUI integration"
```

---

### Task 6: GenerateThumbnail Action

**Files:**
- Create: `app/Mod/Studio/Actions/GenerateThumbnail.php`
- Create: `app/Mod/Studio/Tests/Feature/GenerateThumbnailTest.php`

Chains LEM (prompt generation from brief) → ComfyUI (image generation). Creates a StudioAsset for the generated thumbnail.

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Mod\Studio\Actions\GenerateThumbnail;
use Mod\Studio\Models\StudioAsset;

it('generates a thumbnail from a brief via LEM + ComfyUI', function () {
    Storage::fake('local');

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'prompt' => 'vibrant summer beach scene, golden hour, professional photography',
                    'negative' => 'blurry, low quality',
                ]),
            ],
        ]),
        '*/prompt' => Http::response(['prompt_id' => 'thumb-001']),
        '*/history/thumb-001' => Http::response([
            'thumb-001' => [
                'outputs' => [
                    'node' => [
                        'images' => [
                            ['filename' => 'out.png', 'subfolder' => '', 'type' => 'output'],
                        ],
                    ],
                ],
            ],
        ]),
        '*/view*' => Http::response('png-bytes', 200, [
            'Content-Type' => 'image/png',
        ]),
    ]);

    $job = GenerateThumbnail::run('summer beach TikTok thumbnail');

    expect($job->status)->toBe('completed');
    expect($job->output)->toHaveKey('asset_id');
    expect(StudioAsset::find($job->output['asset_id']))->not->toBeNull();
});

it('fails when LEM returns invalid prompt', function () {
    Http::fake([
        '*/api/chat' => Http::response([
            'message' => ['content' => 'not json'],
        ]),
    ]);

    $job = GenerateThumbnail::run('test');

    expect($job->status)->toBe('failed');
});
```

**Step 2: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Illuminate\Support\Facades\Http;
use Mod\Studio\Models\StudioAsset;
use Mod\Studio\Models\StudioJob;

class GenerateThumbnail
{
    use Action;

    public function handle(string $brief, int $width = 1080, int $height = 1920): StudioJob
    {
        $job = StudioJob::create([
            'type' => 'thumbnail',
            'input' => ['brief' => $brief, 'width' => $width, 'height' => $height],
        ]);

        $promptData = $this->generatePrompt($brief);

        if ($promptData === null || ! isset($promptData['prompt'])) {
            $job->markFailed('Invalid prompt from LEM');

            return $job;
        }

        $job->markStarted();

        $image = GenerateImage::run(
            prompt: $promptData['prompt'],
            width: $width,
            height: $height,
        );

        if ($image === null) {
            $job->markFailed('ComfyUI generation failed');

            return $job;
        }

        $asset = StudioAsset::create([
            'filename' => basename($image['path']),
            'path' => $image['path'],
            'mime_type' => $image['mime_type'],
            'file_size' => 0,
            'tags' => ['thumbnail', 'generated'],
        ]);

        $job->markCompleted([
            'asset_id' => $asset->id,
            'image_path' => $image['path'],
            'prompt_used' => $promptData['prompt'],
        ]);

        return $job;
    }

    private function generatePrompt(string $brief): ?array
    {
        $ollamaUrl = config('studio.ollama.url');
        $model = config('studio.ollama.model', 'lem-4b');
        $timeout = config('studio.ollama.timeout', 60);

        $response = Http::timeout($timeout)->post("{$ollamaUrl}/api/chat", [
            'model' => $model,
            'stream' => false,
            'format' => 'json',
            'messages' => [
                ['role' => 'system', 'content' => 'You are a thumbnail prompt designer. Given a brief, create an image generation prompt for a TikTok thumbnail. Respond with ONLY valid JSON: {"prompt": "detailed image description", "negative": "things to avoid"}. No markdown.'],
                ['role' => 'user', 'content' => $brief],
            ],
        ]);

        if (! $response->successful()) {
            return null;
        }

        $content = $response->json('message.content', '');

        return is_string($content) ? json_decode($content, true) : null;
    }
}
```

**Step 3: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Actions/GenerateThumbnail.php app/Mod/Studio/Tests/Feature/GenerateThumbnailTest.php
git commit -m "feat(studio): add GenerateThumbnail action — LEM prompt + ComfyUI generation"
```

---

### Task 7: Thumbnail Console Command + API

**Files:**
- Create: `app/Mod/Studio/Console/Thumbnail.php`
- Create: `app/Mod/Studio/Controllers/Api/ImageController.php`
- Create: `app/Mod/Studio/Tests/Feature/ImageApiTest.php`
- Modify: `app/Mod/Studio/Routes/api.php`
- Modify: `app/Mod/Studio/Boot.php`

**Step 1: Write the failing API test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;

it('generates a thumbnail via POST /studio/images/thumbnail', function () {
    Storage::fake('local');

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode(['prompt' => 'test thumbnail', 'negative' => '']),
            ],
        ]),
        '*/prompt' => Http::response(['prompt_id' => 'thumb-api']),
        '*/history/thumb-api' => Http::response([
            'thumb-api' => [
                'outputs' => [
                    'n' => ['images' => [['filename' => 'out.png', 'subfolder' => '', 'type' => 'output']]],
                ],
            ],
        ]),
        '*/view*' => Http::response('png-bytes', 200, ['Content-Type' => 'image/png']),
    ]);

    $response = $this->postJson('/studio/images/thumbnail', [
        'brief' => 'summer beach thumbnail',
    ]);

    $response->assertCreated();
    $response->assertJsonPath('data.type', 'thumbnail');
});

it('validates brief is required for thumbnail', function () {
    $response = $this->postJson('/studio/images/thumbnail', []);

    $response->assertUnprocessable();
});
```

**Step 2: Write controller**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Controllers\Api;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Mod\Studio\Actions\GenerateThumbnail;

class ImageController extends Controller
{
    public function thumbnail(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'brief' => 'required|string|max:2000',
            'width' => 'nullable|integer|min:256|max:4096',
            'height' => 'nullable|integer|min:256|max:4096',
        ]);

        $job = GenerateThumbnail::run(
            $validated['brief'],
            $validated['width'] ?? 1080,
            $validated['height'] ?? 1920,
        );

        return response()->json(['data' => $job], 201);
    }
}
```

**Step 3: Write console command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Mod\Studio\Actions\GenerateThumbnail;

class Thumbnail extends Command
{
    protected $signature = 'studio:thumbnail
        {brief : Description of the thumbnail}
        {--width=1080 : Image width}
        {--height=1920 : Image height}';

    protected $description = 'Generate a thumbnail via LEM + ComfyUI';

    public function handle(): int
    {
        $brief = $this->argument('brief');
        $width = (int) $this->option('width');
        $height = (int) $this->option('height');

        $this->info("Generating thumbnail: {$brief}");

        $job = GenerateThumbnail::run($brief, $width, $height);

        if ($job->status === 'failed') {
            $this->error("Failed: {$job->error}");

            return Command::FAILURE;
        }

        $this->info("Job #{$job->id} completed");
        $this->line("  Image: {$job->output['image_path']}");
        $this->line("  Prompt: {$job->output['prompt_used']}");

        return Command::SUCCESS;
    }
}
```

**Step 4: Update routes + Boot.php**

Add to `api.php`:
```php
Route::post('/images/thumbnail', [ImageController::class, 'thumbnail'])->name('api.studio.images.thumbnail');
```

Add to Boot.php `onConsole()`:
```php
$event->command(Console\Thumbnail::class);
```

**Step 5: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Console/Thumbnail.php app/Mod/Studio/Controllers/Api/ImageController.php app/Mod/Studio/Tests/Feature/ImageApiTest.php app/Mod/Studio/Routes/api.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add thumbnail CLI command and image API endpoint"
```

---

### Task 8: Manifest Image Overlay Support

**Files:**
- Modify: `app/Mod/Studio/Actions/GenerateManifest.php`
- Create: `app/Mod/Studio/Tests/Feature/ManifestImageTest.php`

Extend the manifest format with an `overlays` field for image overlays on video.

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Actions\GenerateManifest;
use Mod\Studio\Models\StudioAsset;

it('accepts manifests with image overlays', function () {
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'file_size' => 1024,
    ]);

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'template' => 'tiktok-15s',
                    'clips' => [['asset_id' => 1, 'start' => 0, 'end' => 15, 'order' => 1]],
                    'captions' => [],
                    'overlays' => [
                        ['type' => 'image', 'asset_id' => 2, 'at' => 3.0, 'duration' => 5, 'position' => 'top-right', 'opacity' => 0.8],
                    ],
                    'audio' => ['track' => 'original'],
                    'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
                ]),
            ],
        ]),
    ]);

    $job = GenerateManifest::run('video with product logo overlay');

    expect($job->manifest)->toHaveKey('overlays');
    expect($job->manifest['overlays'])->toHaveCount(1);
    expect($job->manifest['overlays'][0]['type'])->toBe('image');
});
```

**Step 2: Update system prompt in GenerateManifest**

Add overlays to the manifest schema:
```
"overlays": [{"type": "image", "asset_id": N, "at": float, "duration": float, "position": "top-right", "opacity": 0.8}]  (optional)
```

**Step 3: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Actions/GenerateManifest.php app/Mod/Studio/Tests/Feature/ManifestImageTest.php
git commit -m "feat(studio): extend manifest format with image overlay support"
```

---

### Task 9: Batch Remix Command (Content Flywheel)

**Files:**
- Create: `app/Mod/Studio/Actions/BatchRemix.php`
- Create: `app/Mod/Studio/Console/BatchRemix.php`
- Create: `app/Mod/Studio/Tests/Feature/BatchRemixTest.php`
- Modify: `app/Mod/Studio/Boot.php`

This enables the content flywheel — feed a directory of source videos + a brief, generate remixed variants in batch.

**Step 1: Write the failing test**

```php
<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Http;
use Mod\Studio\Actions\BatchRemix;
use Mod\Studio\Models\StudioAsset;
use Mod\Studio\Models\StudioJob;

it('creates multiple remix jobs from a batch brief', function () {
    StudioAsset::create([
        'filename' => 'clip1.mp4',
        'path' => 'studio/assets/clip1.mp4',
        'mime_type' => 'video/mp4',
        'duration_ms' => 15000,
        'file_size' => 1024,
        'tags' => ['summer'],
    ]);

    Http::fake([
        '*/api/chat' => Http::response([
            'message' => [
                'content' => json_encode([
                    'template' => 'tiktok-15s',
                    'clips' => [['asset_id' => 1, 'start' => 0, 'end' => 5, 'order' => 1]],
                    'captions' => [['text' => 'Test', 'at' => 0, 'duration' => 3, 'style' => 'bold-center']],
                    'audio' => ['track' => 'original'],
                    'output' => ['format' => 'mp4', 'resolution' => '1080x1920', 'fps' => 30],
                ]),
            ],
        ]),
    ]);

    $jobs = BatchRemix::run(
        brief: 'summer TikTok series, upbeat',
        variants: 3,
    );

    expect($jobs)->toHaveCount(3);
    expect($jobs[0])->toBeInstanceOf(StudioJob::class);
    expect($jobs[0]->type)->toBe('remix');
});

it('limits variants to maximum of 10', function () {
    StudioAsset::create([
        'filename' => 'clip.mp4',
        'path' => 'studio/assets/clip.mp4',
        'mime_type' => 'video/mp4',
        'file_size' => 1024,
    ]);

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

    $jobs = BatchRemix::run(brief: 'test', variants: 20);

    expect($jobs)->toHaveCount(10);
});
```

**Step 2: Write the action**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Actions;

use Core\Actions\Action;
use Mod\Studio\Models\StudioJob;

class BatchRemix
{
    use Action;

    private const MAX_VARIANTS = 10;

    /**
     * @return array<StudioJob>
     */
    public function handle(string $brief, int $variants = 3, ?string $template = null): array
    {
        $variants = min($variants, self::MAX_VARIANTS);
        $jobs = [];

        for ($i = 0; $i < $variants; $i++) {
            $variantBrief = $brief . " (variant " . ($i + 1) . " of {$variants} — make each unique)";
            $jobs[] = GenerateManifest::run($variantBrief, $template);
        }

        return $jobs;
    }
}
```

**Step 3: Write the console command**

```php
<?php

declare(strict_types=1);

namespace Mod\Studio\Console;

use Illuminate\Console\Command;
use Mod\Studio\Actions\BatchRemix;

class BatchRemixCommand extends Command
{
    protected $signature = 'studio:batch-remix
        {brief : Creative brief for all variants}
        {--variants=3 : Number of remix variants to generate}
        {--template= : Template name}
        {--render : Dispatch all to render worker}';

    protected $description = 'Generate multiple remix variants from a single brief';

    public function handle(): int
    {
        $brief = $this->argument('brief');
        $variants = (int) $this->option('variants');
        $template = $this->option('template');

        $this->info("Generating {$variants} remix variants...");

        $jobs = BatchRemix::run($brief, $variants, $template);

        $succeeded = 0;
        $failed = 0;

        foreach ($jobs as $job) {
            if ($job->status === 'failed') {
                $this->warn("  Job #{$job->id}: FAILED — {$job->error}");
                $failed++;
            } else {
                $clips = count($job->manifest['clips'] ?? []);
                $this->info("  Job #{$job->id}: {$clips} clips");
                $succeeded++;

                if ($this->option('render')) {
                    \Mod\Studio\Actions\RenderManifest::run($job);
                    $this->line("    → Dispatched to render worker");
                }
            }
        }

        $this->newLine();
        $this->info("Generated: {$succeeded} | Failed: {$failed}");

        return $failed === count($jobs) ? Command::FAILURE : Command::SUCCESS;
    }
}
```

**Step 4: Register in Boot.php**

```php
$event->command(Console\BatchRemixCommand::class);
```

**Step 5: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Actions/BatchRemix.php app/Mod/Studio/Console/BatchRemixCommand.php app/Mod/Studio/Tests/Feature/BatchRemixTest.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add batch remix for content flywheel — multiple variants from one brief"
```

---

## Phase 5 — Production

### Task 10: Livewire Voice + Thumbnail Pages

**Files:**
- Create: `app/Mod/Studio/Livewire/VoicePage.php`
- Create: `app/Mod/Studio/Views/voice.blade.php`
- Create: `app/Mod/Studio/Livewire/ThumbnailPage.php`
- Create: `app/Mod/Studio/Views/thumbnail.blade.php`
- Modify: `app/Mod/Studio/Routes/web.php`
- Modify: `app/Mod/Studio/Boot.php`
- Modify: `app/Mod/Studio/Views/layouts/studio.blade.php` (add nav links)

Add UI pages for voiceover generation and thumbnail generation. Follow the same RemixPage pattern — form on left, job status on right, recent jobs below.

**VoicePage:** Brief textarea, voice select, submit → GenerateVoiceover. Show script + audio player on completion.

**ThumbnailPage:** Brief textarea, width/height inputs, submit → GenerateThumbnail. Show generated image on completion.

**Nav bar update:** Add Voice and Thumbnail links alongside Assets and Remix.

**Routes:**
```php
Route::get('/voice', VoicePage::class)->name('studio.voice');
Route::get('/thumbnails', ThumbnailPage::class)->name('studio.thumbnails');
```

**Step 1: Create both Livewire components and views following existing RemixPage pattern**

**Step 2: Update layout nav bar**

**Step 3: Register components in Boot.php, add routes to web.php**

**Step 4: Run tests, commit**

```bash
php artisan test --filter=Studio
git add app/Mod/Studio/Livewire/ app/Mod/Studio/Views/ app/Mod/Studio/Routes/web.php app/Mod/Studio/Boot.php
git commit -m "feat(studio): add Livewire voice and thumbnail pages"
```

---

### Task 11: Homelab Docker Services (Ansible)

**Files:**
- Create: `/Users/snider/Code/DevOps/playbooks/deploy_studio_tts.yml`
- Create: `/Users/snider/Code/DevOps/playbooks/deploy_studio_comfyui.yml`
- Create: `/Users/snider/Code/DevOps/playbooks/deploy_studio_worker.yml`

These Ansible playbooks deploy the Studio GPU services to the homelab. Follow the existing `deploy_eaas.yml` pattern.

**TTS playbook** (`deploy_studio_tts.yml`):
- Docker container: `ghcr.io/remsky/kokoro-fastapi:latest` (or ROCm variant)
- Port: 9200
- Network: noc-net
- Traefik route: `tts.lthn.sh` → studio-tts:9200
- Volume: model cache

**ComfyUI playbook** (`deploy_studio_comfyui.yml`):
- Docker container: ComfyUI with ROCm support
- Port: 8188
- Network: noc-net
- Traefik route: `comfyui.lthn.sh` → studio-comfyui:8188
- Volumes: models, output, custom_nodes

**ffmpeg Worker playbook** (`deploy_studio_worker.yml`):
- Docker container: custom image with ffmpeg + Python Flask/FastAPI
- Port: 9300
- Network: noc-net
- Volumes: shared asset/render storage
- Callback support: POSTs to Studio API when renders complete

**Step 1: Write all three playbooks following deploy_eaas.yml pattern**

**Step 2: Test with `--check` mode**

```bash
cd /Users/snider/Code/DevOps
ansible-playbook playbooks/deploy_studio_tts.yml --check -l loc-dev-edge-01
```

**Step 3: Commit in DevOps repo**

```bash
cd /Users/snider/Code/DevOps
git add playbooks/deploy_studio_*.yml
git commit -m "feat(studio): add Ansible playbooks for TTS, ComfyUI, and ffmpeg worker"
```

---

### Task 12: Production Deployment Config

**Files:**
- Modify: `/Users/snider/Code/lab/host.uk.com/.env.example`
- Create: `/Users/snider/Code/DevOps/playbooks/deploy_studio.yml`

Add Studio-specific environment variables to `.env.example` and create a deployment playbook for the Studio module itself (updating .env on the homelab, restarting the app).

**Environment variables:**
```env
STUDIO_WHISPER_URL=https://whisper.lthn.sh
STUDIO_OLLAMA_URL=https://ollama.lthn.sh
STUDIO_TTS_URL=https://tts.lthn.sh
STUDIO_WORKER_URL=http://studio-worker:9300
STUDIO_COMFYUI_URL=https://comfyui.lthn.sh
STUDIO_COMFYUI_MODEL=flux1-dev.safetensors
STUDIO_STORAGE_DISK=local
STUDIO_ASSETS_PATH=studio/assets
STUDIO_RENDERS_PATH=studio/renders
```

**Step 1: Update .env.example**

**Step 2: Create deploy_studio.yml Ansible playbook**

**Step 3: Commit**

```bash
git add .env.example
git commit -m "feat(studio): add production environment variables"
```

---

### Task 13: Run Full Test Suite + Lint

**Step 1: Run all Studio tests**

```bash
php artisan test --filter=Studio
```

Expected: All tests pass (Phase 1-2 tests + Phase 3-5 tests).

**Step 2: Run lint**

```bash
./vendor/bin/pint --dirty
```

**Step 3: Verify commands**

```bash
php artisan list studio
```

Expected: studio:catalogue, studio:remix, studio:transcribe, studio:voice, studio:thumbnail, studio:batch-remix

**Step 4: Final commit if lint changed anything**

```bash
git add -A && git commit -m "style(studio): format code with Pint"
```

---

## Final Module Structure

```
app/Mod/Studio/
├── Boot.php
├── Actions/
│   ├── BatchRemix.php            (Phase 4 — content flywheel)
│   ├── CatalogueAsset.php        (Phase 1)
│   ├── GenerateImage.php         (Phase 4)
│   ├── GenerateManifest.php      (Phase 2, extended Phase 3+4)
│   ├── GenerateThumbnail.php     (Phase 4)
│   ├── GenerateVoiceover.php     (Phase 3)
│   ├── RenderManifest.php        (Phase 2)
│   ├── SynthesiseSpeech.php      (Phase 3)
│   └── TranscribeAsset.php       (Phase 1)
├── Console/
│   ├── BatchRemixCommand.php     (Phase 4)
│   ├── Catalogue.php             (Phase 1)
│   ├── Remix.php                 (Phase 2)
│   ├── Thumbnail.php             (Phase 4)
│   ├── Transcribe.php            (Phase 1)
│   └── Voice.php                 (Phase 3)
├── Controllers/Api/
│   ├── AssetController.php       (Phase 1)
│   ├── ImageController.php       (Phase 4)
│   ├── RemixController.php       (Phase 2)
│   └── VoiceController.php       (Phase 3)
├── Livewire/
│   ├── AssetBrowserPage.php      (Phase 1)
│   ├── RemixPage.php             (Phase 2)
│   ├── ThumbnailPage.php         (Phase 5)
│   └── VoicePage.php             (Phase 5)
├── Migrations/
│   └── 2026_03_08_000001_create_studio_tables.php
├── Models/
│   ├── StudioAsset.php
│   └── StudioJob.php
├── Routes/
│   ├── api.php
│   └── web.php
├── Tests/Feature/
│   ├── AssetApiTest.php
│   ├── BatchRemixTest.php
│   ├── CatalogueAssetTest.php
│   ├── GenerateImageTest.php
│   ├── GenerateManifestTest.php
│   ├── GenerateThumbnailTest.php
│   ├── GenerateVoiceoverTest.php
│   ├── ImageApiTest.php
│   ├── ManifestImageTest.php
│   ├── ManifestVoiceoverTest.php
│   ├── RemixApiTest.php
│   ├── RenderManifestTest.php
│   ├── StudioAssetTest.php
│   ├── StudioJobTest.php
│   ├── SynthesiseSpeechTest.php
│   ├── TranscribeAssetTest.php
│   └── VoiceApiTest.php
└── Views/
    ├── layouts/studio.blade.php
    ├── asset-browser.blade.php
    ├── remix.blade.php
    ├── thumbnail.blade.php
    └── voice.blade.php

config/studio.php (extended with comfyui section)
```
