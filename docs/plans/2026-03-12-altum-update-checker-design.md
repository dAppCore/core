# AltumCode Update Checker — Design

> **Note:** Layer 1 (version detection via PHP artisan) is implemented and documented at `docs/docs/php/packages/uptelligence.md`. Layer 2 (browser-automated downloads via Claude Code skill) is NOT yet implemented.

## Problem

Host UK runs 4 AltumCode SaaS products and 13 plugins across two marketplaces (CodeCanyon + LemonSqueezy). Checking for updates and downloading them is a manual process: ~50 clicks across two marketplace UIs, moving 16+ zip files, extracting to the right directories. This eats a morning of momentum every update cycle.

## Solution

Two-layer system: lightweight version detection (PHP artisan command) + browser-automated download (Claude Code skill).

## Architecture

```
Layer 1: Detection (core/php-uptelligence)
  artisan uptelligence:check-updates
  5 HTTP GETs, no auth, schedulable
  Compares remote vs deployed versions

Layer 2: Download (Claude Code skill)
  Playwright → LemonSqueezy (16 items)
  Claude in Chrome → CodeCanyon (2 items)
  Downloads zips to staging folder
  Extracts to saas/services/{product}/package/

Layer 3: Deploy (existing — manual)
  docker build → scp → deploy_saas.yml
  Human in the loop
```

## Layer 1: Version Detection

### Public Endpoints (no auth required)

| Endpoint | Returns |
|----------|---------|
| `GET https://66analytics.com/info.php` | `{"latest_release_version": "66.0.0", "latest_release_version_code": 6600}` |
| `GET https://66biolinks.com/info.php` | Same format |
| `GET https://66pusher.com/info.php` | Same format |
| `GET https://66socialproof.com/info.php` | Same format |
| `GET https://dev.altumcode.com/plugins-versions` | `{"affiliate": {"version": "2.0.1"}, "ultimate-blocks": {"version": "9.1.0"}, ...}` |

### Deployed Version Sources

- **Product version**: `PRODUCT_CODE` constant in deployed source `config.php`
- **Plugin versions**: `version` field in each plugin's `config.php` or `config.json`

### Artisan Command

`php artisan uptelligence:check-updates`

Output:
```
Product          Deployed    Latest     Status
──────────────────────────────────────────────
66analytics      65.0.0      66.0.0     UPDATE AVAILABLE
66biolinks       65.0.0      66.0.0     UPDATE AVAILABLE
66pusher         65.0.0      65.0.0     ✓ current
66socialproof    65.0.0      66.0.0     UPDATE AVAILABLE

Plugin           Deployed    Latest     Status
──────────────────────────────────────────────
affiliate        2.0.0       2.0.1      UPDATE AVAILABLE
ultimate-blocks  9.1.0       9.1.0      ✓ current
...
```

Lives in `core/php-uptelligence` as a scheduled check or on-demand command.

## Layer 2: Browser-Automated Download

### Claude Code Skill: `/update-altum`

Workflow:
1. Run version check (Layer 1) — show what needs updating
2. Ask for confirmation before downloading
3. Download from both marketplaces
4. Extract to staging directories
5. Report what changed

### Marketplace Access

**LemonSqueezy (Playwright)**
- Auth: Magic link email to `snider@lt.hn` — user taps on phone
- Flow per item: Navigate to order detail → click "Download" button
- 16 items across 2 pages of orders
- Session persists for the skill invocation

**CodeCanyon (Claude in Chrome)**
- Auth: Saved browser session cookies (user `snidered`)
- Flow per item: Click "Download" dropdown → "All files & documentation"
- 2 items on downloads page

### Product-to-Marketplace Mapping

| Product | CodeCanyon | LemonSqueezy |
|---------|-----------|--------------|
| 66biolinks | Regular licence | Extended licence (66biolinks custom, $359.28) |
| 66socialproof | Regular licence | — |
| 66analytics | — | Regular licence |
| 66pusher | — | Regular licence |

### Plugin Inventory (all LemonSqueezy)

| Plugin | Price | Applies To |
|--------|-------|------------|
| Pro Notifications | $58.80 | 66socialproof |
| Teams Plugin | $58.80 | All products |
| Push Notifications Plugin | $46.80 | All products |
| Ultimate Blocks | $32.40 | 66biolinks |
| Pro Blocks | $32.40 | 66biolinks |
| Payment Blocks | $32.40 | 66biolinks |
| Affiliate Plugin | $32.40 | All products |
| PWA Plugin | $25.20 | All products |
| Image Optimizer Plugin | $19.20 | All products |
| Email Shield Plugin | FREE | All products |
| Dynamic OG images plugin | FREE | 66biolinks |
| Offload & CDN Plugin | FREE | All products (gift from Altum) |

### Staging & Extraction

- Download to: `~/Code/lthn/saas/updates/YYYY-MM-DD/`
- Products extract to: `~/Code/lthn/saas/services/{product}/package/product/`
- Plugins extract to: `~/Code/lthn/saas/services/{product}/package/product/plugins/{plugin_id}/`

## LemonSqueezy Order UUIDs

Stable order URLs for direct navigation:

| Product | Order URL |
|---------|-----------|
| 66analytics | `/my-orders/2972471f-abac-4165-b78d-541b176de180` |

(Remaining UUIDs to be captured on first full run of the skill.)

## Out of Scope

- No auto-deploy to production (human runs `deploy_saas.yml`)
- No licence key handling or financial transactions
- No AltumCode Club membership management
- No Blesta updates (different vendor)
- No update SQL migration execution (handled by AltumCode's own update scripts)

## Key Technical Details

- AltumCode products use Unirest HTTP client for API calls
- Product `info.php` endpoints are public, no rate limiting observed
- Plugin versions endpoint (`dev.altumcode.com`) is also public
- Production Docker images have `/install/` and `/update/` directories stripped
- Updates require full Docker image rebuild and redeployment via Ansible
- CodeCanyon download URLs contain stable purchase UUIDs
- LemonSqueezy uses magic link auth (no password, email-based)
- Playwright can access LemonSqueezy; Claude in Chrome cannot (payment platform safety block)

## Workflow Summary

**Before**: Get email from AltumCode → log into 2 marketplaces → click through 18 products/plugins → download 16+ zips → extract to right directories → rebuild Docker images → deploy. Half a morning.

**After**: Run `artisan uptelligence:check-updates` → see what's behind → invoke `/update-altum` → tap magic link on phone → go make coffee → come back to staged files → `deploy_saas.yml`. 10 minutes of human time.
