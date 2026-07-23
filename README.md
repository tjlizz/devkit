# DevKit

DevKit is a modular monorepo with a Go HTTP API, a Vue 3 + TypeScript web
application, and SQLite for local persistence.

## Repository layout

```text
.
├── docs/       Architecture decision records
├── server/     Go 1.26 API and SQLite migrations
├── web/        Vue 3, Vite, TypeScript, and Ant Design Vue
└── Makefile    Common development commands
```

## Prerequisites

- Go 1.26 or newer
- Node.js 20.19+ or 22.12+ and npm
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

Run the API and Vite development server together:

```sh
make dev
```

- Web application: http://localhost:5173
- API health endpoint: http://localhost:8080/api/v1/health

The Vite server proxies `/api` requests to the Go server during local
development.

## Commands

Run `make help` to list the available commands:

```sh
make dev        # Start the API and web development servers
make build      # Build the API binary and production web assets
make test       # Run Go and frontend tests
make lint       # Check Go formatting/vet and frontend lint
make clean      # Remove generated build and local database artifacts
```

Backend and frontend commands can also be run independently:

```sh
cd server
go test ./...
go vet ./...
go build ./...

cd ../web
npm run dev
npm run lint
npm run typecheck
npm run build
```

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `DEVKIT_CONFIG` | empty | Optional path to a YAML configuration file |
| `DEVKIT_PORT` | `8080` | API listen port |
| `DEVKIT_DB_PATH` | `./devkit.db` | SQLite database path |
| `VITE_API_BASE_URL` | `/api/v1` | Browser-facing API base URL |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Vite development proxy target |

See [docs/architecture-decisions.md](docs/architecture-decisions.md) for the
architecture decision record placeholder.
