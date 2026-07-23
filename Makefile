SHELL := /bin/sh
.DEFAULT_GOAL := help

-include .env
export

.PHONY: help dev build build-server build-web test test-server test-web lint lint-server lint-web clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; print "DevKit commands:"} /^[a-zA-Z_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start the Go API and Vite development server
	@server_pid=""; web_pid=""; \
	cleanup() { \
		test -z "$$server_pid" || kill "$$server_pid" 2>/dev/null || true; \
		test -z "$$web_pid" || kill "$$web_pid" 2>/dev/null || true; \
	}; \
	trap cleanup EXIT INT TERM; \
	(cd server && go run ./cmd/server) & server_pid=$$!; \
	(cd web && npm run dev) & web_pid=$$!; \
	wait

build: build-server build-web ## Build the API and web application

build-server: ## Build the Go API binary
	@mkdir -p server/bin
	cd server && go build -o bin/devkit-server ./cmd/server

build-web: ## Build production web assets
	cd web && npm run build

test: test-server test-web ## Run all tests

test-server: ## Run Go tests
	cd server && go test ./...

test-web: ## Run frontend tests
	cd web && npm run test

lint: lint-server lint-web ## Run all linters

lint-server: ## Check Go formatting and run go vet
	@files=$$(find server -type f -name '*.go' -not -path '*/vendor/*'); \
	unformatted=$$(gofmt -l $$files); \
	test -z "$$unformatted" || { echo "Unformatted Go files:"; echo "$$unformatted"; exit 1; }
	cd server && go vet ./...

lint-web: ## Run frontend ESLint
	cd web && npm run lint

clean: ## Remove generated build and local database artifacts
	rm -rf server/bin web/dist
	rm -f server/*.db server/*.db-shm server/*.db-wal
	rm -f server/*.sqlite server/*.sqlite-shm server/*.sqlite-wal
	rm -f server/*.sqlite3 server/*.sqlite3-shm server/*.sqlite3-wal
