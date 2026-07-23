# DevKit

DevKit is a modular monorepo with three layers:

- **server/** — Go HTTP API (admin backend) with SQLite persistence
- **web/** — Vue 3 + TypeScript admin panel with Ant Design Vue
- **www/** — Next.js public-facing developer marketplace (App Router, SSR, SEO)

## Repository layout

```text
.
├── docs/       Architecture decision records
├── server/     Go 1.26 API and SQLite migrations
├── web/        Vue 3, Vite, TypeScript, and Ant Design Vue
├── www/        Next.js developer marketplace (App Router, Tailwind CSS)
├── AGENTS.md   Project specification and guidance
├── Makefile    Common development commands
└── .env.example
```

## Prerequisites

- Go 1.26 or newer
- Node.js 20.19+ or 22.12+ and npm (www requires Node 20+)
- `make`

The SQLite driver is pure Go (`modernc.org/sqlite`), so CGO and a system SQLite
installation are not required.

## Setup

Install dependencies and create a local environment file:

```sh
cd server && go mod download
cd ../web && npm install
cd ..
cp .env.example .env
```

The defaults work without a `.env` file. To use YAML configuration, copy
`server/config.example.yaml`, edit it, and set `DEVKIT_CONFIG` to its path.
Environment variables override YAML values.

## Development

### Marketplace (www — Next.js)

```sh
make dev-www
# → http://localhost:3000
```

### Admin Panel (web — Vue 3)

```sh
make dev    # Vue dev server
# → http://localhost:5173
```

### All services

```sh
make dev-all
# → API: http://localhost:8080
# → Admin: http://localhost:5173
# → Marketplace: http://localhost:3000
```

The Vite server proxies `/api` requests to the Go server. The Next.js dev server
can be configured via `www/.env`.

## Commands

Run `make help` to list the available commands:

```sh
make dev        # Start Vue admin dev server
make dev-www    # Start Next.js marketplace dev server
make dev-all    # Start all development servers
make build      # Build all applications
make test       # Run all tests
make lint       # Run all linters
make clean      # Remove generated build artifacts
```

## Configuration

### API (server)

| Variable | Default | Description |
| --- | --- | --- |
| `DEVKIT_CONFIG` | empty | YAML configuration file path |
| `DEVKIT_PORT` | `8080` | API listen port |
| `DEVKIT_DB_PATH` | `./devkit.db` | SQLite database path |

### Admin (web)

| Variable | Default | Description |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api/v1` | Browser-facing API base URL |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Vite dev proxy target |

### Marketplace (www)

| Variable | Default | Description |
| --- | --- | --- |
| `NEXT_PUBLIC_SITE_URL` | `http://localhost:3000` | Site URL for SEO metadata |

See [docs/architecture-decisions.md](docs/architecture-decisions.md) for the
architecture decision record placeholder.
