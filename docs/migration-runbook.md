# Migration runbook

This runbook summarizes the safe path for local and remote migrations.

## Local migrations

```sh
make migrate-local
```

Local migrations target the Docker PostgreSQL service and use
`sslmode=disable`.

## Remote migrations

```sh
make migrate-version
make migrate-staging
```

Remote migrations require `APP_ENV=staging` and `DATABASE_MIGRATION_URL` pointed
at a direct or session connection on port `5432`.

## Rules

- Never run remote migrations through a transaction pooler on port `6543`.
- Never use `sslmode=disable` for Supabase.
- Do not rewrite applied migrations.
- Do not run down, reset or force operations against staging or production.
- Keep migration and application connections separate when possible.

See `backend/migrations/README.md` for table-level RLS requirements.
