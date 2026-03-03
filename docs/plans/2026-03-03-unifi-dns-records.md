# UniFi Gateway DNS Records — lthn.sh

> **For:** snider (UniFi gateway config)
> **Date:** 3 Mar 2026

## Overview

UniFi gateway doesn't support wildcard DNS. Each service needs an individual A record pointing to the homelab box (10.69.69.165).

All covered by a single GoGetSSL cert with SANs: `lthn.sh`, `*.lthn.sh`, `*.infra.lthn.sh`.

## Lab Services — `*.lthn.sh`

| Hostname | IP | Service |
|----------|-----|---------|
| `hub.lthn.sh` | 10.69.69.165 | Laravel admin hub |
| `lab.lthn.sh` | 10.69.69.165 | LEM Lab |
| `ollama.lthn.sh` | 10.69.69.165 | Ollama inference + embeddings |
| `qdrant.lthn.sh` | 10.69.69.165 | Qdrant vector search |
| `eaas.lthn.sh` | 10.69.69.165 | EaaS scoring API |

## Infrastructure — `*.infra.lthn.sh`

| Hostname | IP | Service |
|----------|-----|---------|
| `traefik.infra.lthn.sh` | 10.69.69.165 | Traefik dashboard |
| `grafana.infra.lthn.sh` | 10.69.69.165 | Grafana |
| `prometheus.infra.lthn.sh` | 10.69.69.165 | Prometheus |
| `influx.infra.lthn.sh` | 10.69.69.165 | InfluxDB |
| `auth.infra.lthn.sh` | 10.69.69.165 | Authentik SSO |
| `portainer.infra.lthn.sh` | 10.69.69.165 | Portainer |
| `phpmyadmin.infra.lthn.sh` | 10.69.69.165 | phpMyAdmin |
| `maria.infra.lthn.sh` | 10.69.69.165 | MariaDB admin |
| `postgres.infra.lthn.sh` | 10.69.69.165 | PostgreSQL admin |
| `redis.infra.lthn.sh` | 10.69.69.165 | Redis admin |

## Bare domain

| Hostname | IP | Service |
|----------|-----|---------|
| `lthn.sh` | 10.69.69.165 | Redirects to `hub.lthn.sh` (or landing page) |

## Total: 16 A records

All pointing to the same IP. Add more as new services come online.

## After UniFi Config

Once DNS is live, remove the old `/etc/hosts` entries on Mac:

```
# REMOVE these lines from /etc/hosts:
10.69.69.165 ollama.lthn.lan
10.69.69.165 qdrant.lthn.lan
10.69.69.165 eaas.lthn.lan
10.69.69.165 lthn.lan
10.69.69.165 traefik.lthn.lan
10.69.69.165 blesta.lthn.lan
10.69.69.165 auth.lthn.lan
10.69.69.165 phpmyadmin.lthn.lan
10.69.69.165 portainer.lthn.lan
10.69.69.165 grafana.lthn.lan
10.69.69.165 lab.lthn.lan
10.69.69.165 prometheus.lthn.lan
10.69.69.165 influx.lthn.lan
10.69.69.165 maria.lthn.lan
10.69.69.165 postgres.lthn.lan
10.69.69.165 redis.lthn.lan
```

Test resolution:
```bash
# Should resolve to 10.69.69.165 via UniFi DNS
dig hub.lthn.sh @<unifi-gateway-ip> +short

# Test each service
for h in hub lab ollama qdrant eaas; do
  echo -n "$h.lthn.sh → "; dig $h.lthn.sh +short
done
```

## Notes

- UniFi gateway DNS serves these records to all LAN clients automatically
- No public DNS records exist for `lthn.sh` — the zone in CloudNS is empty (used only for ACME DNS-01 cert validation)
- The Mac, the Linux homelab, and any other LAN device will all resolve via UniFi
- Charon's CoreDNS on the Linux box can coexist — it handles `leth.in` (prod internal), UniFi handles `lthn.sh` (homelab)
