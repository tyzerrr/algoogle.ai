SHELL := /bin/sh

COMPOSE ?= docker compose

.PHONY: help up start down stop restart logs ps build test test-api test-web health clean

help:
	@printf '%s\n' 'Available targets:'
	@printf '  %-12s %s\n' 'make up' 'Build and start API + Web in the background'
	@printf '  %-12s %s\n' 'make down' 'Stop API + Web containers'
	@printf '  %-12s %s\n' 'make restart' 'Restart API + Web'
	@printf '  %-12s %s\n' 'make logs' 'Follow Docker Compose logs'
	@printf '  %-12s %s\n' 'make ps' 'Show container status'
	@printf '  %-12s %s\n' 'make health' 'Check API health endpoint'
	@printf '  %-12s %s\n' 'make test' 'Run API and Web checks'

up start:
	$(COMPOSE) up --build -d
	@printf '\n%s\n' 'Servers are starting:'
	@printf '  Web: %s\n' 'http://localhost:3000'
	@printf '  API: %s\n' 'http://localhost:8000'
	@printf '  Health: %s\n\n' 'http://localhost:8000/health'
	@printf '  NeoVim sync files: %s\n\n' './workspace'
	$(COMPOSE) ps

down stop:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up --build -d
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

build:
	$(COMPOSE) build

health:
	@curl -fsS http://localhost:8000/health
	@printf '\n'

test: test-api test-web

test-api:
	cd apps/api && CGO_ENABLED=0 go test ./...

test-web:
	cd apps/web && NODE_OPTIONS=--no-deprecation corepack pnpm install --frozen-lockfile
	cd apps/web && corepack pnpm typecheck
	cd apps/web && corepack pnpm test
	cd apps/web && NEXT_TELEMETRY_DISABLED=1 corepack pnpm build

clean:
	$(COMPOSE) down -v
