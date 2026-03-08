# Studio: Multimedia Pipeline Design

**Date:** 8 March 2026
**Status:** Approved

## Goal

Local AI multimedia pipeline for video remixing, content creation, and voice interaction. Runs as a CorePHP service (Studio) dispatching GPU work to homelab infrastructure. First client: OF agency remixing existing footage into TikTok-ready variants.

## Architecture

Studio is a job orchestrator. LEM handles creative decisions (smart layer), ffmpeg and GPU services handle execution (dumb layer). LEM never touches video frames — it produces JSON manifests that the execution layer consumes mechanically.

```
Studio (CorePHP, lthn.ai/lthn.sh)
  ├── Livewire UI (studio.lthn.ai)
  ├── Artisan Commands (CLI)
  └── API Routes (/api/studio/*)
        │
        ▼
  Studio Actions (RemixVideo, GenerateManifest, etc.)
        │
  Redis Job Queue
        │
        ├── Ollama (LEM fleet) ─── Creative decisions, scripts, captions
        ├── Whisper Service ────── Transcribe source footage, STT
        ├── TTS Service ────────── Voiceover generation
        ├── ffmpeg Worker ──────── Render manifests to video
        └── ComfyUI (Phase 2) ─── Image gen, thumbnails, overlays
```

All GPU services are Docker containers on the homelab (or any GPU server). Studio dispatches over HTTP. No local GPU dependency — remote-first from day one.

## Library & Cataloguing

Source material catalogued across three stores:

- **PG** (`studio_assets`): Metadata — filename, duration, resolution, tags (season/theme/mood), workspace
- **Qdrant**: Vector embeddings from Whisper transcripts + CLIP image embeddings (phase 2). Semantic search
- **Filesystem**: Raw files on homelab storage, PG references paths
- **.md catalogue files**: Human-readable collection descriptions, style guides, brand notes. LEM reads as context

Query flow:
```
Brief ("summer lollipop TikTok, 15s, upbeat")
  → LEM queries PG for tagged assets
  → LEM queries Qdrant for semantic matches
  → LEM reads collection .md for style context
  → LEM outputs manifest JSON
```

## Manifest Format

LEM produces, ffmpeg consumes. No AI in execution.

```json
{
  "template": "tiktok-15s",
  "clips": [
    {"asset_id": 42, "start": 3.2, "end": 8.1, "order": 1},
    {"asset_id": 17, "start": 0.0, "end": 5.5, "order": 2}
  ],
  "captions": [
    {"text": "Summer vibes only", "at": 0.5, "duration": 3, "style": "bold-center"}
  ],
  "audio": {"track": "original", "fade_in": 0.5},
  "output": {"format": "mp4", "resolution": "1080x1920", "fps": 30}
}
```

Variants: LEM produces multiple manifests from the same brief. Worker renders each independently.

## GPU Services (Homelab)

| Service | Container | Port | Model | Purpose |
|---------|-----------|------|-------|---------|
| Ollama | studio-ollama | 11434 | LEM fleet | Creative decisions, scripts, captions |
| Whisper | studio-whisper | 9100 | whisper-large-v3-turbo | Transcribe footage, STT |
| TTS | studio-tts | 9200 | Kokoro/Parler | Voiceover generation |
| ffmpeg Worker | studio-worker | — | n/a | Queue consumer, renders manifests |
| ComfyUI | studio-comfyui | 8188 | Flux/SD3.5 | Image gen, thumbnails (Phase 2) |

Shared with existing homelab: noc-net Docker network, Traefik, PG, Qdrant. Each service exposes REST, Studio POSTs work and gets callbacks.

Deployment: Ansible playbook per service, ROCm Docker images for GPU services.

## CorePHP Module

`app/Mod/Studio/` — same patterns as LEM module.

**Actions:**
- `CatalogueAsset::run()` — ingest, extract metadata, generate embeddings
- `GenerateManifest::run()` — brief + library → LEM → manifest JSON
- `RenderManifest::run()` — dispatch to ffmpeg worker
- `TranscribeAsset::run()` — send to Whisper, store transcript
- `SynthesiseSpeech::run()` — send to TTS, return audio

**Artisan commands:**
- `studio:catalogue` — batch ingest directory
- `studio:remix` — brief in, rendered videos out
- `studio:transcribe` — batch transcribe library

**API routes** (`/api/studio/*`):
- `POST /remix` — submit brief, get job ID
- `GET /remix/{id}` — poll status, get output URLs
- `POST /assets` — upload/catalogue
- `GET /assets` — search library

**Livewire UI:**
- Asset browser with tag/search
- Remix form — pick assets or let LEM choose, enter brief, select template
- Job status + preview
- Download/share

**Config:** `config/studio.php` — GPU endpoints, templates, Qdrant collection, storage paths.

## Phased Delivery

### Phase 1 — Foundation (before April)
- Studio module scaffolding (actions, routes, commands)
- Asset cataloguing (upload, PG metadata, Whisper transcripts)
- Whisper service on homelab
- `studio:transcribe` end to end
- Basic Livewire asset browser

### Phase 2 — Remix Pipeline
- Manifest format finalised
- LEM integration via Ollama (brief → manifest)
- ffmpeg worker on homelab
- `studio:remix` CLI + API
- Livewire remix form + job status

### Phase 3 — Voice & TTS
- TTS service on homelab (Kokoro)
- Voice interface: Whisper STT → LEM → TTS
- Voiceover generation for scripts

### Phase 4 — Visual Generation
- ComfyUI on homelab with Flux/SD3.5
- Thumbnail generation
- Image overlays in manifests
- Video generation via Wan2.1 (experimental)

### Phase 5 — Production
- Full library from agency
- Authentik account for client
- studio.lthn.ai live
- Usage tracking via 66analytics

Phase 1 + 2 = April demo. Upload videos, enter brief, get remixed TikToks back.

## Key Decisions

- **Smart/dumb separation**: LEM produces prompts and manifests (creative), ffmpeg executes (mechanical). Value is in the creative layer.
- **Remote-first GPU**: All inference on homelab/GPU server, never local. Easy to scale to cloud later.
- **Manifest-driven**: JSON contract between LEM and execution. Either side can evolve independently.
- **Same Action pattern**: CLI and API call identical actions. UI is just a thin Livewire layer.
- **Existing infra**: PG, Redis, Qdrant, Ollama, Traefik, Authentik — all already deployed.
