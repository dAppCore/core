# OpenBrain — Usage Guide

> **For:** All agents (Virgil, Charon, Athena, Darbs)
> **Date:** 3 Mar 2026

## What Is OpenBrain

A shared vector-indexed knowledge store for all agents. Memories from Claude Code sessions, implementation plans, repo conventions, and research notes are embedded as vectors and stored in Qdrant for semantic search. MariaDB holds the metadata for filtering.

Think of it as long-term memory that any agent can query: "What do we know about Traefik setup?" returns the most relevant memories regardless of which agent wrote them or when.

## Architecture

```
Agent ──remember()──▶ BrainService
                       ├── Ollama (embed text → 768d vector)
                       ├── Qdrant (store/search vectors)
                       └── MariaDB (metadata, filtering, audit)

Agent ──recall()────▶ BrainService
                       ├── Ollama (embed query → 768d vector)
                       ├── Qdrant (cosine similarity search)
                       └── MariaDB (hydrate full memory records)
```

| Service | URL | What |
|---------|-----|------|
| Ollama | `https://ollama.lthn.lan` | Embedding model (`embeddinggemma`, 768 dimensions) |
| Qdrant | `https://qdrant.lthn.lan` | Vector storage + cosine similarity search |
| MariaDB | `lthn-lan-db:3306` | `brain_memories` table (workspace-scoped) |
| Laravel | `https://lthn.lan` | BrainService, artisan commands, MCP tools |

All `.lan` services use self-signed TLS behind Traefik. The app auto-skips TLS verification for `.lan` URLs.

## Seeding Knowledge

### Full Archive Ingest

Imports everything from the local filesystem into OpenBrain:

```bash
php artisan brain:ingest --workspace=1 --fresh --source=all
```

| Flag | What |
|------|------|
| `--workspace=1` | Workspace ID (required, scopes all memories) |
| `--fresh` | Clears the Qdrant collection first (removes duplicates) |
| `--source=all` | Ingest from all 4 source types |
| `--dry-run` | Preview without storing anything |
| `--agent=virgil` | Tag memories with the ingesting agent's name |
| `--code-path=~/Code` | Override the root code directory |

### Source Types

| Source | What It Scans | Files | Confidence |
|--------|--------------|-------|------------|
| `memory` | `~/.claude/projects/*/memory/*.md` | ~27 | 0.8 |
| `plans` | `docs/plans/` across all repos + `~/.claude/plans/` | ~139 | 0.6 |
| `claude-md` | `CLAUDE.md` files across all repos | ~93 | 0.9 |
| `tasks` | `~/Code/host-uk/core/tasks/*.md` | ~41 | 0.5 |

Run a single source type:
```bash
php artisan brain:ingest --workspace=1 --source=plans --dry-run
```

### Quick Memory-Only Seed

For just the Claude Code memory files (faster, ~388 sections):

```bash
php artisan brain:seed-memory --workspace=1
```

### Go Tool (Standalone)

If the Laravel app isn't available, use the Go brain-seed tool:

```bash
cd ~/Code/go-ai
go run cmd/brain-seed/main.go \
  --ollama=https://ollama.lthn.lan \
  --qdrant=https://qdrant.lthn.lan \
  --collection=openbrain \
  --model=embeddinggemma
```

Add `--plans` to also scan `docs/plans/` directories. Add `--dry-run` to preview.

## Querying Memories

### Via Artisan Tinker

```bash
php artisan tinker
```

```php
$brain = app(\Core\Mod\Agentic\Services\BrainService::class);

// Semantic search — returns most relevant memories
$results = $brain->recall('Ansible production servers', 5, [], 1);

// With filters
$results = $brain->recall('deployment', 5, [
    'type' => 'plan',           // Filter by memory type
    'project' => 'core',        // Filter by project
    'agent_id' => 'virgil',     // Filter by who wrote it
    'min_confidence' => 0.7,    // Minimum confidence score
], 1);

// Results structure
// $results['memories'] — array of memory records
// $results['scores']   — array of similarity scores (0.0–1.0)
```

### Via MCP Tools

Agents with MCP access can use these tools directly:

```
brain_remember  — Store a new memory
brain_recall    — Semantic search for memories
brain_forget    — Remove a specific memory
brain_list      — List memories with filters
```

These are registered in the Agentic module's `onMcpTools` handler.

### Via Direct Qdrant API

For debugging or bulk operations:

```bash
# Collection stats
curl -sk https://qdrant.lthn.lan/collections/openbrain | python3 -m json.tool

# Raw vector search (embed query first via Ollama)
VECTOR=$(curl -sk https://ollama.lthn.lan/api/embeddings \
  -d '{"model":"embeddinggemma","prompt":"Traefik setup"}' \
  | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin)['embedding']))")

curl -sk https://qdrant.lthn.lan/collections/openbrain/points/search \
  -H 'Content-Type: application/json' \
  -d "{\"vector\":$VECTOR,\"limit\":5,\"with_payload\":true}" \
  | python3 -m json.tool
```

## Storing New Memories

### Via BrainService (PHP)

```php
$brain = app(\Core\Mod\Agentic\Services\BrainService::class);

$brain->remember([
    'workspace_id' => 1,
    'agent_id' => 'charon',
    'type' => 'decision',              // architecture, convention, decision, bug, plan, research, observation
    'content' => "The heading\n\nThe full content of the memory.",
    'tags' => ['source:manual', 'project:go-ai'],
    'project' => 'go-ai',
    'confidence' => 0.8,               // 0.0–1.0
]);
```

### Via MCP (Agent Tool Call)

```json
{
  "tool": "brain_remember",
  "arguments": {
    "content": "Dragonfly on the homelab listens on 6379, no auth required for LAN access.",
    "type": "architecture",
    "tags": ["homelab", "redis", "dragonfly"],
    "confidence": 0.9
  }
}
```

## Memory Types

| Type | When to Use |
|------|------------|
| `architecture` | System design, infrastructure, service topology |
| `convention` | Coding standards, naming rules, workflow patterns |
| `decision` | Choices made and why (domain strategy, tech selection) |
| `bug` | Bugs found, fixes applied, lessons learned |
| `plan` | Implementation plans, roadmaps, task breakdowns |
| `research` | Findings, analysis, RFC summaries |
| `observation` | General notes that don't fit other categories |

## Tags

Tags are freeform strings. The ingest command auto-tags with:
- `source:{type}` — where it came from (memory, plans, claude-md, tasks)
- `project:{name}` — which repo/project it relates to

Add your own tags for filtering:
- `homelab`, `production`, `deprecated`
- `go`, `php`, `infra`
- Agent names: `virgil`, `charon`, `athena`

## Maintenance

### Re-ingest After Changes

When memory files are updated, re-run with `--fresh` to avoid duplicates:

```bash
php artisan brain:ingest --workspace=1 --fresh --source=memory
```

### Check Collection Health

```bash
curl -sk https://qdrant.lthn.lan/collections/openbrain | \
  python3 -c "import sys,json; r=json.load(sys.stdin)['result']; print(f'Points: {r[\"points_count\"]}, Status: {r[\"status\"]}')"
```

### Delete a Specific Memory

```php
$brain->forget($memoryId);  // Removes from both Qdrant and MariaDB
```

### Full Reset

```bash
php artisan brain:ingest --workspace=1 --fresh --source=all
```

This deletes the Qdrant collection and re-creates it with fresh data.

## Current State (3 Mar 2026)

- **4,598 points** in the `openbrain` collection
- **300 source files** across 4 source types
- **Embedding model**: `embeddinggemma` (768d, ~135ms per embedding on GPU)
- **Collection status**: green
- **4 failed sections**: Oversized plan sections that exceeded Ollama's context window — not critical
