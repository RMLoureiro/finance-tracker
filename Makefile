GO ?= go

.PHONY: dev-server dev-web test build db-up

dev-server:
	cd server && $(GO) run ./cmd/server

dev-web:
	cd web && npm run dev

test:
	cd server && $(GO) test ./...

build:
	cd web && npm run build
	cd server && $(GO) build -o bin/server ./cmd/server

db-up:
	docker compose up -d db
