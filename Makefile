.PHONY: dev backend-dev dev-docker backend-dev-docker local-up local-down migrate-local migrate-staging migrate-version import-jooble seed-demo telegram-publisher-dev

dev:
	./scripts/dev-supabase.sh

backend-dev:
	./scripts/dev-supabase.sh backend

dev-docker:
	./scripts/dev-local.sh

backend-dev-docker:
	./scripts/dev-local.sh backend

local-up:
	docker compose --profile local up --build

local-down:
	docker compose --profile local down

migrate-local:
	docker compose --profile local run --rm migrate

migrate-staging:
	./scripts/migrate-remote.sh up

migrate-version:
	./scripts/migrate-remote.sh version

import-jooble:
	@set -a; . ./.env; set +a; cd backend && go run ./cmd/import-jooble

seed-demo:
	@set -a; . ./.env; set +a; cd backend && go run ./cmd/seed-demo

telegram-publisher-dev:
	@set -a; . ./.env; set +a; \
	if [ -z "$${TELEGRAM_DEV_DATABASE_URL:-}" ]; then echo "Set TELEGRAM_DEV_DATABASE_URL to a local development database" >&2; exit 1; fi; \
	cd backend && APP_ENV=development DATABASE_URL="$${TELEGRAM_DEV_DATABASE_URL}" DATABASE_MIGRATION_URL="$${TELEGRAM_DEV_DATABASE_URL}" go run ./cmd/telegram-publisher --dry-run --once
