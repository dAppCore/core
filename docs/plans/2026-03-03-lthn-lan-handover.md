# lthn.lan Homelab Setup — Handover to Charon

> **For:** Charon (Linux homelab agent, 10.69.69.165)
> **From:** Virgil (macOS)
> **Date:** 3 Mar 2026

## Goal

Stand up the Host UK Laravel app on the Linux homelab as `lthn.lan` — a private dev/ops hub away from production. This joins the existing `.lan` service mesh (ollama.lan, qdrant.lan, eaas.lan).

## What lthn.lan Is

The personal admin hub — issues, dashboards, agent coordination. NOT production-facing. Runs the same codebase as lthn.ai but configured for homelab use with access to local AI services.

## Architecture

```
Mac (snider) ──hosts file──▶ lthn.lan (10.69.69.165)
                             ├── Traefik (TLS termination, self-signed)
                             ├── Laravel app (FrankenPHP/Octane, port 80)
                             ├── MariaDB 11 (port 3306)
                             └── Redis/Dragonfly (port 6379)

Already running on 10.69.69.165:
  ollama.lan  → Ollama (embeddings, LEM inference)
  qdrant.lan  → Qdrant (vector search)
  eaas.lan    → EaaS scoring API v0.2.0
```

## Prerequisites

These should already exist on the machine:

- Docker (or Podman) with Traefik v3.6+ running
- External Docker network `proxy` for Traefik
- SSH key for forge.lthn.ai (port 2223) — needed for `composer install`

If Traefik isn't set up yet, see the existing `.lan` services for the pattern.

## Step 1: Clone the Repo

```bash
mkdir -p /opt/services/lthn-lan
cd /opt/services/lthn-lan

# Clone via forge SSH
git clone ssh://git@forge.lthn.ai:2223/lthn/hostuk.git app
cd app
```

**SSH config** for forge (add to `~/.ssh/config` if not present):
```
Host forge.lthn.ai
  HostName forge.lthn.ai
  Port 2223
  User git
  IdentityFile ~/.ssh/id_ed25519
```

## Step 2: Create docker-compose.yml

Create `/opt/services/lthn-lan/docker-compose.yml`:

```yaml
services:
  app:
    build:
      context: ./app
      dockerfile: Dockerfile
    container_name: lthn-lan
    restart: unless-stopped
    env_file: .env
    volumes:
      - app_storage:/app/storage/logs
    networks:
      - proxy
      - lthn-lan
    depends_on:
      mariadb:
        condition: service_healthy

  mariadb:
    image: mariadb:11
    container_name: lthn-lan-db
    restart: unless-stopped
    environment:
      MARIADB_ROOT_PASSWORD: "${DB_ROOT_PASSWORD}"
      MARIADB_DATABASE: lthn_lan
      MARIADB_USER: lthn_lan
      MARIADB_PASSWORD: "${DB_PASSWORD}"
    volumes:
      - mariadb_data:/var/lib/mysql
    networks:
      - lthn-lan
    healthcheck:
      test: ["CMD", "healthcheck.sh", "--connect", "--innodb_initialized"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  app_storage:
  mariadb_data:

networks:
  proxy:
    external: true
  lthn-lan:
    driver: bridge
```

## Step 3: Create .env

Create `/opt/services/lthn-lan/.env`:

```bash
APP_NAME="LTHN Hub"
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=https://lthn.lan
APP_DOMAIN=lthn.lan

TRUSTED_PROXIES=*

DB_CONNECTION=mariadb
DB_HOST=lthn-lan-db
DB_PORT=3306
DB_DATABASE=lthn_lan
DB_USERNAME=lthn_lan
DB_PASSWORD=changeme-generate-a-real-one
DB_ROOT_PASSWORD=changeme-generate-a-real-one

SESSION_DRIVER=redis
SESSION_DOMAIN=.lthn.lan
SESSION_SECURE_COOKIE=true

REDIS_CLIENT=predis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

CACHE_STORE=redis
QUEUE_CONNECTION=redis
BROADCAST_CONNECTION=log

OCTANE_SERVER=frankenphp

# OpenBrain — connects to existing .lan services
BRAIN_OLLAMA_URL=https://ollama.lan
BRAIN_QDRANT_URL=https://qdrant.lan
BRAIN_COLLECTION=openbrain
BRAIN_EMBEDDING_MODEL=embeddinggemma

# EaaS scorer
EAAS_URL=https://eaas.lan
```

Then generate the app key:
```bash
docker compose run --rm app php artisan key:generate
```

## Step 4: Traefik Labels

Add these labels to the `app` service in docker-compose.yml (or use a Traefik dynamic config file):

```yaml
labels:
  traefik.enable: "true"
  traefik.http.routers.lthn-lan.rule: "Host(`lthn.lan`)"
  traefik.http.routers.lthn-lan.entrypoints: websecure
  traefik.http.routers.lthn-lan.tls: "true"
  traefik.http.services.lthn-lan.loadbalancer.server.port: "80"
  traefik.docker.network: proxy
```

Note: For `.lan` domains, Traefik uses self-signed certs (no Let's Encrypt — not a real TLD). The same pattern as ollama.lan/qdrant.lan/eaas.lan.

## Step 5: Build and Start

```bash
cd /opt/services/lthn-lan

# Build the image (amd64)
docker compose build

# Start DB first, wait for healthy
docker compose up -d mariadb
docker compose logs -f mariadb  # wait for "ready for connections"

# Start the app
docker compose up -d app
docker compose logs -f app
```

The `laravel-entrypoint.sh` inside the container will:
1. Clear caches
2. Run `php artisan migrate --force`
3. Run `php artisan db:seed --force`
4. Start Octane on port 80

## Step 6: Verify

```bash
# Check container health
docker compose ps

# Check migrations ran
docker compose logs app | grep -i migration

# Test HTTP (from the machine itself)
curl -sk https://lthn.lan/ | head -20

# Check Horizon (queue workers)
curl -sk https://lthn.lan/horizon/api/stats
```

## Step 7: /etc/hosts on Mac

Already done by snider:
```
10.69.69.165  lthn.lan
```

## Embedding Model on GPU

The `embeddinggemma` model on ollama.lan appears to be running on CPU. It's only ~256MB — should fit easily alongside whatever else is on the RX 7800 XT. Check with:

```bash
# On the Linux machine
curl -sk https://ollama.lan/api/ps
```

If it shows CPU, try pulling it fresh or restarting Ollama — it should auto-detect the GPU.

## What's Inside the App

The app container runs 4 supervised processes:

| Process | What | Port |
|---------|------|------|
| Octane | FrankenPHP HTTP server | 80 |
| Horizon | Queue workers + dashboard | — |
| Scheduler | Cron loop (`schedule:run`) | — |
| Redis | In-container cache/session/queue | 6379 |

Reverb (WebSocket) is optional for lthn.lan — skip it unless needed.

## Key Artisan Commands

```bash
# Enter the container
docker compose exec app bash

# Seed OpenBrain with all knowledge (4,671 sections)
php artisan brain:ingest --workspace=1 --fresh --source=all

# Quick memory-only seed
php artisan brain:seed-memory --workspace=1

# Test recall
php artisan tinker
>>> app(\Core\Mod\Agentic\Services\BrainService::class)->recall('Traefik setup', 5, [], 1)
```

## Composer Packages

All `core/php-*` packages come from forge.lthn.ai via SSH VCS repos. The Dockerfile handles `composer install` during build. If you need to update packages:

```bash
docker compose exec app composer update
```

SSH key must be available inside the container for forge access. The Dockerfile should handle this via build args or SSH agent forwarding.

## Known Issues

- **Self-signed TLS**: All `.lan` domains use self-signed certs. The app's `BrainService` auto-detects `.lan` URLs and skips verification. Browsers will warn — just accept.
- **Embedding 500s on large sections**: Some very large plan sections (30KB+) cause Ollama to return 500. Not critical — only 4 out of 4,671 sections affected.
- **PHP 8.5**: The Dockerfile uses PHP 8.5. Make sure the base image supports it (`dunglas/frankenphp:1-php8.5-trixie`).

## Future: Satellite Services

Once lthn.lan is stable, the plan is to add Website/Service satellites:
- `eaas.lthn.ai` → Ethics scorer frontage
- `models.lthn.ai` → Model data + HuggingFace info
- Each feature (LEM.Lab etc) → its own subdomain via Website/Service pattern

These are just additional `Boot.php` modules with domain patterns — the multi-tenant modular monolith handles everything from one codebase.
