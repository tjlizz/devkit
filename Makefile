SHELL := /bin/sh
.DEFAULT_GOAL := help

-include .env
export

.PHONY: help dev dev-all build build-server build-web build-www test test-server test-web test-www lint lint-server lint-web lint-www clean clean-www

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; print "DevKit commands:"} /^[a-zA-Z_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start the Vue admin dev server
	cd web && npm run dev

dev-www: ## Start the Next.js marketplace dev server
	cd www && npm run dev

dev-all: ## Start all dev servers
	@server_pid=""; web_pid=""; www_pid=""; \
	cleanup() { \
		test -z "$$server_pid" || kill "$$server_pid" 2>/dev/null || true; \
		test -z "$$web_pid" || kill "$$web_pid" 2>/dev/null || true; \
		test -z "$$www_pid" || kill "$$www_pid" 2>/dev/null || true; \
	}; \
	trap cleanup EXIT INT TERM; \
	(cd server && go run ./cmd/server) & server_pid=$$!; \
	(cd web && npm run dev) & web_pid=$$!; \
	(cd www && npm run dev) & www_pid=$$!; \
	wait

build: build-server build-web build-www ## Build all applications

build-server: ## Build the Go API binary
	@mkdir -p server/bin
	cd server && go build -o bin/devkit-server ./cmd/server

build-web: ## Build production web assets
	cd web && npm run build

build-www: ## Build the Next.js marketplace for production
	cd www && npm run build

test: test-server test-web test-www ## Run all tests

test-server: ## Run Go tests
	cd server && go test ./...

test-web: ## Run frontend tests
	cd web && npm run test

test-www: ## Run Next.js marketplace tests
	cd www && npm run test || true

lint: lint-server lint-web lint-www ## Run all linters

lint-server: ## Check Go formatting and run go vet
	@files=$$(find server -type f -name '*.go' -not -path '*/vendor/*'); \
	unformatted=$$(gofmt -l $$files); \
	test -z "$$unformatted" || { echo "Unformatted Go files:"; echo "$$unformatted"; exit 1; }
	cd server && go vet ./...

lint-web: ## Run frontend ESLint
	cd web && npm run lint

lint-www: ## Run Next.js marketplace lint
	cd www && npm run lint || true

clean: clean-www ## Remove generated build and local database artifacts
	rm -rf server/bin web/dist
	rm -f web/*.tsbuildinfo
	rm -f server/*.db server/*.db-shm server/*.db-wal
	rm -f server/*.sqlite server/*.sqlite-shm server/*.sqlite-wal
	rm -f server/*.sqlite3 server/*.sqlite3-shm server/*.sqlite3-wal

clean-www: ## Remove Next.js build output
	rm -rf www/.next www/out
