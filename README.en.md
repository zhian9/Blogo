# Blogo

A self-hosted, production-deployed **blog content platform (CMS)** — Go backend with two React frontends (admin console and public site).

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql)](https://www.mysql.com)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

> Live: <https://blogo.cloud>　Admin: <https://admin.blogo.cloud>
> (Self-hosted single-site deployment — not multi-tenant)
>
> 简体中文文档见 [README.md](README.md)

---

## What is this

Blogo is a blog content platform built from scratch. The backend handles content management, users and permissions, comment moderation, traffic analytics and audit logging; the frontend is split into two independent apps — an **admin console** for authors/administrators and a **public site** for readers.

The design goal is a content platform one person can maintain long-term: a single process to deploy, configuration-driven from the admin console, and permissions plus content visibility driven by data so that copywriting and operations can change without touching code.

Scale reference: roughly **24k lines** of Go on the backend, **150+** REST routes, **25** database tables, and about **15k lines** of TypeScript across the two frontends.

## Features

**Content management**

- **Articles**: Markdown authoring with pre-rendered HTML, categories / tags / pinning, drafts and publishing, view counts
- **Project showcase**: project cards (cover, one-line summary, tech stack, highlights, repository and demo links) designed for a portfolio use case
- **Supporting content**: categories, tags, standalone pages, friend links, media assets
- **Comments**: for both articles and projects, guest commenting, moderation (approve / reject / delete)

**Users and permissions**

- Registration with email activation, login with image captcha, user profile, follow/followers, favorites and likes
- **RBAC**: user → role → menu → API resource, with policies generated from database mappings and hot-reloaded at runtime
- **Content visibility**: articles and projects support public / private / specific-users-only

**Operations and observability**

- **Traffic analytics**: page views, unique visitors and unique IPs, charted in the dashboard for the last 1 / 7 / 30 days
- **Audit log**: records operator, source IP, module, action, request path, status code and error detail, searchable by multiple filters
- **Monitoring**: Prometheus metrics (request counts, latency histograms) with Grafana dashboards
- **Site configuration**: homepage copy, about-page copy and tech stack, contact email and more are all editable from the admin console and take effect on the next page load

**Infrastructure**

- Image uploads to Cloudflare R2 object storage, falling back to local disk when not configured
- Email: account activation, subscriptions and notifications delivered by an asynchronous mail worker
- Security: JWT authentication (backed by Redis for revocation), per-IP and per-user rate limiting, asynchronous audit logging

## Tech stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25, Gin, GORM 2.0, Casbin 2.0, Google Wire (DI), Viper (config), zap (logging) |
| **Storage** | MySQL 8.0, Redis 7 (cache / captcha / token store / rate limiting / visit counters) |
| **Admin console** | React 19, TypeScript, Vite, Ant Design 6, Redux Toolkit + RTK Query, ECharts |
| **Public site** | React 19, TypeScript, Vite, Ant Design 6, TailwindCSS, Zustand, TanStack Query |
| **Deployment & monitoring** | Docker Compose, Nginx, Cloudflare, Prometheus, Grafana |

## Architecture

**Modular monolith (not microservices)**: a single Gin process split into two business modules — `rbac` (auth, users, roles, menus, permissions) and `blog` (content, interactions, analytics) — decoupled through interfaces and dependency injection.

**Layers**: `API` (HTTP request/response) → `BIZ` (business orchestration) → `DAL` (data access) → `schema` (models and forms), wired at startup by Google Wire (`make wire` regenerates the code).

**Authentication flow**: JWT (backed by Redis, supporting logout and active revocation) → an interceptor parses the token and injects the identity into `context` → Casbin authorizes request by request. Business code only reads identity from `context` and never trusts a user identifier supplied in request parameters.

**Hot-reloading permission policies**: role→menu→resource mappings live in the database; the service generates a Casbin policy file at startup and reloads it when menus or roles change (signalled through Redis plus periodic comparison), so policies take effect without a restart.

**Performance and consistency trade-offs**

- Visibility filtering for list endpoints is pushed down to SQL (`public OR own OR explicitly shared`), and the "specific users" case is validated in batches to avoid N+1 queries
- Traffic counting happens in Redis first (idempotent set-based deduplication with TTL), then is written back to the statistics table once per day, avoiding a database write per page view
- Analytics queries return dates in ascending order with missing days filled in, so the chart timeline stays continuous and the frontend does no gap filling
- Site configuration and single-page content change often, so the frontend uses a short cache with refetch-on-focus to guarantee "edit in admin, refresh the site, see it immediately"

## Project layout

```
Blogo/
├── blogo-server/               Go backend
│   ├── cmd/                    CLI entrypoint (cobra: start / stop / version)
│   ├── internal/
│   │   ├── bootstrap/          startup wiring (HTTP server, middleware, graceful shutdown)
│   │   ├── config/             config loading (YAML + .env expansion)
│   │   ├── mods/               business modules
│   │   │   ├── rbac/           auth, users, roles, menus, operation logs
│   │   │   └── blog/           articles, projects, comments, analytics, media, subscriptions
│   │   ├── utility/prom/       Prometheus metrics
│   │   └── wirex/              Wire DI container (wire_gen.go is generated)
│   └── pkg/                    reusable packages (cachex / jwtx / logging / middleware / ossx …)
├── blogo-admin/                React admin console (:5174)
├── blogo-web/                  React public site (:5173)
├── configs/server/             backend configs (dev / prod / test + Casbin model and policy)
├── deploy/
│   ├── compose/                docker compose (full stack / monitoring)
│   ├── docker/                 Dockerfile (builds) and Dockerfile.nobuild (prebuilt binary)
│   ├── nginx/                  reverse proxy and site configs
│   ├── prometheus/ grafana/    monitoring config and dashboards
│   └── scripts/                server-side deploy.sh / rollback.sh
├── docs/                       API docs (Swagger) and project notes
├── scripts/                    local development and build scripts
└── Makefile                    engineering command entrypoint
```

## Getting started

### Prerequisites

- Go 1.25+
- Node.js 20+ (22 recommended)
- MySQL 8.0 and Redis 7
- (Optional) Air for hot reload: `go install github.com/air-verse/air@latest`

### 1. Configure

```bash
cd blogo-server
cp .env.example .env
# Adjust as needed: database DSN, Redis password, JWT signing key, admin password hash, etc.
```

### 2. Prepare the database

Only an empty database is required (for example `blogo`), with `DB_DSN` in `.env` pointing at it. On first startup the service **creates all tables automatically** and seeds the initial data: the admin account, base roles, the menu tree, default pages and site settings (controlled by `auto_migrate: true`).

### 3. Run (three terminals)

```bash
# backend on :8040
make server              # equivalent to: cd blogo-server && air

# public site on :5173
make web

# admin console on :5174
make admin
```

> Without Air, run the backend directly: `cd blogo-server && go run . start -d ../configs/server -c dev`.

### 4. Access

| Service | URL | Notes |
|---|---|---|
| Backend API | http://localhost:8040 | `/health` for health checks |
| Public site | http://localhost:5173 | Vite proxies `/api` and `/uploads` to 8040 |
| Admin console | http://localhost:5174 | Username is `ROOT_USERNAME`; the password matches `ROOT_PASSWORD_HASH` |
| Swagger | http://localhost:8040/swagger/index.html | **Disabled by default**; set `disable_swagger` to `false` to enable |

## Deployment

Production runs on a low-spec cloud server, so **compilation happens locally and the server only runs artifacts**:

```bash
# 1. Cross-compile a static Linux binary locally
cd blogo-server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "-w -s -X main.VERSION=v1.0.0" -o blogo .

# 2. Upload artifacts (binary + both frontend dists) as .new to avoid partial rollouts
scp blogo-server/blogo  user@server:/path/Blogo/blogo-server/blogo.new
scp -r blogo-web/dist   user@server:/path/Blogo/blogo-web/dist.new
scp -r blogo-admin/dist user@server:/path/Blogo/blogo-admin/dist.new

# 3. Deploy on the server (automatic backup + health check + auto rollback on failure)
cd /path/Blogo
bash deploy/scripts/deploy.sh --dry-run    # dry run first
bash deploy/scripts/deploy.sh              # real deployment
```

Notes:

- **Configs and frontend assets are bind-mounted** (`configs/server` into the container, `dist` into Nginx), so config changes and frontend releases need **no image rebuild** — only a binary change does (`docker compose build`)
- The image is built from `deploy/docker/Dockerfile.nobuild`, which only copies the locally compiled binary; nothing is compiled on the server
- **Rollback**: `bash deploy/scripts/rollback.sh` lists available backups and restores one, including the previous binary, image tag, frontend dists, configs and a database dump
- Nginx + Cloudflare + HTTPS serve the site; public site, admin console and monitoring are split across domains, with monitoring behind Basic Auth
- After a frontend release, purge the Cloudflare cache (`vite build` replaces hashed asset filenames)

## Common commands

| Command | Description |
|---|---|
| `make server` / `make web` / `make admin` | Run the backend (Air hot reload), public site and admin console |
| `make build` | Build backend and both frontends (`build-server` / `build-web` / `build-admin` individually) |
| `make docker` | Build the backend Docker image |
| `make wire` | Regenerate Wire dependency injection code |
| `make swagger` | Generate Swagger docs |
| `make clean` | Clean build artifacts |
| `make help` | List all commands |

## Configuration

| Path | Description |
|---|---|
| `configs/server/dev/`, `prod/`, `test/` | Per-environment configs (server, storage, middleware, logging) |
| `configs/server/rbac_model.conf` | Casbin RBAC model definition |
| `configs/server/gen_rbac_policy.csv` | Policy file generated from roles, menus and resources |
| `blogo-server/.env` | Environment variables (**not committed**, see `.env.example`) |

| Key variable | Description |
|---|---|
| `DB_DSN` | MySQL connection string |
| `REDIS_PASSWORD` | Redis password |
| `ROOT_USERNAME` / `ROOT_PASSWORD_HASH` | Admin account and bcrypt password hash (synced to the database on every startup) |
| `JWT_SIGNING_KEY` | JWT signing key (at least 32 characters) |
| `SITE_URL` / `CORS_ALLOW_ORIGINS` | Site URL and CORS allowlist |
| `R2_*` | Cloudflare R2 object storage (leave empty to store files locally) |

Configuration precedence: struct defaults → YAML file → `.env` environment variables → enforced values in code.

## Known limitations and roadmap

Being explicit about what is not done yet:

- **No unit tests**: only ad-hoc manual regression (the deployment script runs API smoke checks). Priority is tests for permissions, content visibility and analytics counting
- **Notification endpoints need tighter authorization**: `/api/v1/notifications` currently sits in the auth allowlist and reads the user identifier from request parameters; it should derive identity from the token and verify ownership
- **The "response latency" chart in the dashboard shows placeholder data** and still needs real latency collection
- **Single-tenant deployment**: no multi-tenancy isolation or billing; turning this into a product would require a tenant model, tenant-level configuration and a quota system

## License

MIT — see [LICENSE](LICENSE)
