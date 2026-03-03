# lthn.sh Homelab Setup — Handover to Charon

> **For:** Charon (Linux homelab agent, 10.69.69.165)
> **From:** Virgil (macOS)
> **Date:** 3 Mar 2026

## Goal

Stand up the Host UK Laravel app on the Linux homelab as `hub.lthn.sh` — a private dev/ops hub away from production. This joins the existing homelab service mesh (ollama, qdrant, eaas) which is migrating from `*.lthn.lan` to `*.lthn.sh`.

## Domain Strategy

Clean separation between production and homelab:

| Zone | Purpose | Visibility |
|------|---------|------------|
| `*.lthn.sh` | Homelab — ML, agents, lab services | Internal only |
| `*.infra.lthn.sh` | Homelab — admin/infra tools | Internal only |
| `lthn.ai` | Production — portal, forge, API | Public |
| `lthn.io` | Production — landing, service mesh | Public |
| `leth.in` | Internal prod DNS (CoreDNS on noc) | Internal |
| `lthn.host` | Shared/collab — demos, community | Public when needed |

## Architecture

```
UniFi Gateway ──DNS──▶ 10.69.69.165 (Linux homelab)
                        ├── Traefik (TLS via real cert, *.lthn.sh + *.infra.lthn.sh)
                        ├── Laravel hub (FrankenPHP/Octane, port 80)
                        ├── MariaDB 11 (port 3306)
                        └── Redis/Dragonfly (port 6379)

Already running on 10.69.69.165:
  ollama   → Ollama (embeddings, LEM inference)
  qdrant   → Qdrant (vector search)
  eaas     → EaaS scoring API v0.2.0
```

**Hardware**: Ryzen 9, 128GB RAM, RX 7800 XT (AMD ROCm GPU)

## DNS Records — UniFi Gateway

No wildcard support on UniFi — each service needs an individual A record.
All point to `10.69.69.165`.

### Lab services (`*.lthn.sh`)

| Hostname | Service |
|----------|---------|
| `hub.lthn.sh` | Laravel admin hub |
| `lab.lthn.sh` | LEM Lab |
| `ollama.lthn.sh` | Ollama inference + embeddings |
| `qdrant.lthn.sh` | Qdrant vector search |
| `eaas.lthn.sh` | EaaS scoring API |

### Infrastructure (`*.infra.lthn.sh`)

| Hostname | Service |
|----------|---------|
| `traefik.infra.lthn.sh` | Traefik dashboard |
| `grafana.infra.lthn.sh` | Grafana |
| `prometheus.infra.lthn.sh` | Prometheus |
| `influx.infra.lthn.sh` | InfluxDB |
| `auth.infra.lthn.sh` | Authentik SSO |
| `portainer.infra.lthn.sh` | Portainer |
| `phpmyadmin.infra.lthn.sh` | phpMyAdmin |
| `maria.infra.lthn.sh` | MariaDB admin |
| `postgres.infra.lthn.sh` | PostgreSQL admin |
| `redis.infra.lthn.sh` | Redis admin |

## TLS Certificate

A real GoGetSSL cert covers all homelab domains. No self-signed certs, no TLS skip logic.

**SANs**: `lthn.sh`, `*.lthn.sh`, `*.infra.lthn.sh`
**Validity**: 3 Mar 2026 → 1 Jun 2026

### Deploy the cert to Traefik

Copy the cert + key to the homelab (from snider's Mac):

```bash
# From Mac
scp -P 4819 ~/Downloads/fullchain_lthn.sh.crt root@10.69.69.165:/opt/traefik/certs/lthn.sh.crt
scp -P 4819 ~/Downloads/lthn.sh.key root@10.69.69.165:/opt/traefik/certs/lthn.sh.key
```

Create a Traefik dynamic config file at `/opt/traefik/dynamic/lthn-sh-cert.yml`:

```yaml
tls:
  certificates:
    - certFile: /certs/lthn.sh.crt
      keyFile: /certs/lthn.sh.key
  stores:
    default:
      defaultCertificate:
        certFile: /certs/lthn.sh.crt
        keyFile: /certs/lthn.sh.key
```

Ensure Traefik's compose mounts the certs directory:

```yaml
volumes:
  - /opt/traefik/certs:/certs:ro
  - /opt/traefik/dynamic:/etc/traefik/dynamic:ro
```

And the static config watches for file provider changes:

```yaml
providers:
  file:
    directory: /etc/traefik/dynamic
    watch: true
```

## Step 1: Clone the Repo

```bash
mkdir -p /opt/services/lthn-sh
cd /opt/services/lthn-sh

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

Create `/opt/services/lthn-sh/docker-compose.yml`:

```yaml
services:
  app:
    build:
      context: ./app
      dockerfile: Dockerfile
    container_name: lthn-sh-hub
    restart: unless-stopped
    env_file: .env
    volumes:
      - app_storage:/app/storage/logs
    networks:
      - proxy
      - lthn-sh
    depends_on:
      mariadb:
        condition: service_healthy
    labels:
      traefik.enable: "true"
      traefik.http.routers.lthn-sh-hub.rule: "Host(`hub.lthn.sh`)"
      traefik.http.routers.lthn-sh-hub.entrypoints: websecure
      traefik.http.routers.lthn-sh-hub.tls: "true"
      traefik.http.services.lthn-sh-hub.loadbalancer.server.port: "80"
      traefik.docker.network: proxy

  mariadb:
    image: mariadb:11
    container_name: lthn-sh-db
    restart: unless-stopped
    environment:
      MARIADB_ROOT_PASSWORD: "${DB_ROOT_PASSWORD}"
      MARIADB_DATABASE: lthn_sh
      MARIADB_USER: lthn_sh
      MARIADB_PASSWORD: "${DB_PASSWORD}"
    volumes:
      - mariadb_data:/var/lib/mysql
    networks:
      - lthn-sh
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
  lthn-sh:
    driver: bridge
```

## Step 3: Create .env

Create `/opt/services/lthn-sh/.env`:

```bash
APP_NAME="LTHN Hub"
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=https://hub.lthn.sh
APP_DOMAIN=lthn.sh

TRUSTED_PROXIES=*

DB_CONNECTION=mariadb
DB_HOST=lthn-sh-db
DB_PORT=3306
DB_DATABASE=lthn_sh
DB_USERNAME=lthn_sh
DB_PASSWORD=changeme-generate-a-real-one
DB_ROOT_PASSWORD=changeme-generate-a-real-one

SESSION_DRIVER=redis
SESSION_DOMAIN=.lthn.sh
SESSION_SECURE_COOKIE=true

REDIS_CLIENT=predis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

CACHE_STORE=redis
QUEUE_CONNECTION=redis
BROADCAST_CONNECTION=log

OCTANE_SERVER=frankenphp

# OpenBrain — connects to homelab services
BRAIN_OLLAMA_URL=https://ollama.lthn.sh
BRAIN_QDRANT_URL=https://qdrant.lthn.sh
BRAIN_COLLECTION=openbrain
BRAIN_EMBEDDING_MODEL=embeddinggemma

# EaaS scorer
EAAS_URL=https://eaas.lthn.sh
```

Then generate the app key:
```bash
docker compose run --rm app php artisan key:generate
```

## Step 4: Update Existing Services

Existing services (ollama, qdrant, eaas) need their Traefik router rules updated from `*.lthn.lan` to `*.lthn.sh`. For each service, update the Host rule:

**Docker labels** (if using label-based routing):
```yaml
# Example: ollama
traefik.http.routers.ollama.rule: "Host(`ollama.lthn.sh`)"

# Example: qdrant
traefik.http.routers.qdrant.rule: "Host(`qdrant.lthn.sh`)"

# Example: eaas
traefik.http.routers.eaas.rule: "Host(`eaas.lthn.sh`)"
```

**Or** Traefik dynamic config files (if using file provider):
```yaml
# /opt/traefik/dynamic/ollama.yml
http:
  routers:
    ollama:
      rule: "Host(`ollama.lthn.sh`)"
      entryPoints: [websecure]
      tls: {}
      service: ollama
  services:
    ollama:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:11434"
```

Infrastructure tools follow the same pattern with `*.infra.lthn.sh` hostnames.

## Step 5: Build and Start

```bash
cd /opt/services/lthn-sh

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

# Test HTTP (from any machine with DNS configured)
curl -s https://hub.lthn.sh/ | head -20

# Check Horizon (queue workers)
curl -s https://hub.lthn.sh/horizon/api/stats

# Verify TLS is the real cert (not self-signed)
echo | openssl s_client -connect 10.69.69.165:443 -servername hub.lthn.sh 2>/dev/null | openssl x509 -noout -subject -dates
```

No `-k` flag needed — the cert is real and trusted.

## Embedding Model on GPU

The `embeddinggemma` model on ollama may be running on CPU. It's only ~256MB — should fit easily alongside whatever else is on the RX 7800 XT. Check with:

```bash
curl -s https://ollama.lthn.sh/api/ps
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

Reverb (WebSocket) is optional — skip it unless needed.

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

- **Embedding 500s on large sections**: Some very large plan sections (30KB+) cause Ollama to return 500. Not critical — only 4 out of 4,671 sections affected.
- **PHP 8.5**: The Dockerfile uses PHP 8.5. Make sure the base image supports it (`dunglas/frankenphp:1-php8.5-trixie`).

## Migration Checklist — lthn.lan → lthn.sh

When switching existing services from the old `*.lthn.lan` naming:

- [ ] Deploy cert + key to `/opt/traefik/certs/`
- [ ] Create `lthn-sh-cert.yml` dynamic config
- [ ] Update Traefik router rules (Host matchers) for all services
- [ ] Configure UniFi gateway DNS (all A records → 10.69.69.165)
- [ ] Test from Mac: `curl https://hub.lthn.sh/`
- [ ] Remove old `/etc/hosts` entries for `*.lthn.lan`
- [ ] Update `php-agentic/config.php` defaults to `*.lthn.sh`
- [ ] Update `Boot.php` — `verifySsl` can be `true` always (real certs)

## Future: Satellite Services

Once hub.lthn.sh is stable, the plan is to add Website/Service satellites:
- `eaas.lthn.ai` → Ethics scorer frontage (public)
- `models.lthn.ai` → Model data + HuggingFace info (public)
- `lab.lthn.sh` → LEM Lab (internal)
- Each feature gets its own subdomain via Website/Service pattern

These are just additional `Boot.php` modules with domain patterns — the multi-tenant modular monolith handles everything from one codebase.
