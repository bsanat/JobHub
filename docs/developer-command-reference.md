# Developer command reference

This reference collects the commands that are most useful during JobHub local
development. It is intentionally short; keep detailed feature or provider notes
in the domain-specific documents under `docs/`.

## Local startup

| Task | Command | Notes |
| --- | --- | --- |
| Start the POC development stack | `make dev` | Uses the local POC database and does not modify staging Supabase data. |
| Start Docker services with an empty local database | `make local-up` | Builds and starts PostgreSQL, migrations and backend containers. |
| Stop Docker services | `make local-down` | Stops the local Compose profile. |
| Apply local migrations only | `make migrate-local` | Uses the local Docker PostgreSQL service. |
| Import Jooble POC vacancies | `make import-jooble` | Requires `JOOBLE_API_KEY` in `.env`; request count remains capped. |
| Seed demo data | `make seed-demo` | Loads local demo data through the Go command. |

## Frontend workflow

```sh
cd frontend
npm ci
npm run dev
```

The default Vite URL is `http://localhost:5173`. Keep `VITE_API_URL` pointed at
the local backend unless a task explicitly needs another environment.

## Backend workflow

```sh
docker compose --profile local up -d postgres
make migrate-local

set -a
. ./.env
set +a

cd backend
go run ./cmd/api
```

Do not run the Docker backend and the manual Go backend on the same port at the
same time.

## Verification

Backend checks:

```sh
cd backend
go test -race ./...
go vet ./...
go build ./...
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
```

Frontend checks:

```sh
cd frontend
npm run lint
npm run typecheck
npm run build
```

Repository checks:

```sh
git diff --check
git status
```

## Safety notes

- Keep `.env`, Supabase keys, provider API keys and real connection strings out
  of git.
- Use `sslmode=disable` only for local Docker PostgreSQL.
- Use `sslmode=require` for Supabase direct or session connections.
- Run remote migrations only with the guarded `make migrate-version` and
  `make migrate-staging` flow described in `README.md`.
