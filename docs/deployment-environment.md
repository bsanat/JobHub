---
title: "Деплой и настройки окружения"
description: "Vercel services, environment variables, local/preview/production differences and Mintlify deployment notes."
---

# Деплой и настройки окружения

## Application deploy

`vercel.json` confirms a Vercel multi-service deployment:

```json
{
  "services": {
    "backend": { "root": "backend", "framework": "go" },
    "frontend": { "root": "frontend", "framework": "vite" }
  },
  "rewrites": [
    { "source": "/api/(.*)", "destination": { "service": "backend" } },
    { "source": "/(.*)", "destination": { "service": "frontend" } }
  ]
}
```

Frontend service has an internal rewrite to `/index.html`, which supports React Router routes.

## Environment differences

| Setting | Local | Preview | Production |
| --- | --- | --- | --- |
| Frontend API URL | Defaults to `http://localhost:8080/api/v1` | Not confirmed; likely Vercel public `/api/v1` if env not overridden | Defaults to `/api/v1` |
| Backend DB | `DATABASE_URL` from local `.env` | Not confirmed | Server env `DATABASE_URL` |
| Cookies | `Secure=false` unless `APP_ENV=production` | Depends on `APP_ENV` | `Secure=true` when `APP_ENV=production` |
| Telegram interval publisher | Can run outside Vercel if enabled | Not confirmed | Disabled inside Vercel API process when `VERCEL=1` |
| Telegram webhook | Works if secret and config are present | Not confirmed | Confirmed by code, but live env values are not visible in repo |

## Important variables

| Variable | Scope | Notes |
| --- | --- | --- |
| `APP_ENV` | backend | `development`, `test`, `staging`, `production` |
| `PORT` | backend | Default `8080` |
| `DATABASE_URL` | backend | PostgreSQL connection URL |
| `DATABASE_MIGRATION_URL` | backend migrations | Falls back to `DATABASE_URL` in config |
| `PGX_QUERY_EXEC_MODE` | backend | Must be compatible with Supabase pooler |
| `FRONTEND_ORIGIN` | backend | Trusted origin for mutations |
| `VITE_API_URL` | frontend | Build/runtime env for API base |
| `RESUME_STORAGE_DIR` | backend | Local resume storage |
| `SUPABASE_SECRET_KEY` | backend | Server-only storage key when using Supabase Storage |
| `TELEGRAM_WEBHOOK_SECRET` | backend | Enables internal publisher endpoint |
| `TELEGRAM_BOT_TOKEN` | publisher/backend | Server-only bot token |
| `JOBHUB_PUBLIC_BASE_URL` | publisher | Public job links in Telegram |

## Mintlify docs deploy

Mintlify reads `docs.json` from the default branch. This repo already has:

- `docs.json`
- `index.md`
- `docs/*.md`

Root URL should open `index.md`. If Mintlify shows 404 at `/`, check that `index` is the first navigation page and that `index.md` exists on the default branch.

Do not add Docusaurus dependencies for this docs site; the chosen target is Mintlify.

