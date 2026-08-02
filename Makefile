GO ?= go

.PHONY: dev-server dev-web test build db-up

dev-db:
	cd server && $(GO) run ./cmd/devdb

dev-server:
	cd server && DATABASE_URL=$${DATABASE_URL:-postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable} $(GO) run ./cmd/server

dev-web:
	cd web && npm run dev

test:
	cd server && $(GO) test ./...

build:
	cd web && npm run build
	cd server && $(GO) build -o bin/server ./cmd/server

db-up:
	docker compose up -d db
