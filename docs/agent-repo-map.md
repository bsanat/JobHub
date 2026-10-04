# Agent repo map

This map is a compact locator for common JobHub work. Verify paths before using
it; the source code is authoritative.

## Frontend

- `frontend/src/pages/` contains route-level pages.
- `frontend/src/components/` contains shared and feature components.
- `frontend/src/api/http/` maps HTTP DTOs into frontend models.
- `frontend/src/api/models/` defines frontend domain shapes.
- `frontend/src/i18n/locales/kk.ts` and `frontend/src/i18n/locales/ru.ts`
  hold active UI copy.
- `frontend/src/styles/` contains global and marketplace styles.

## Backend

- `backend/cmd/api/` starts the API service.
- `backend/cmd/telegram-publisher/` posts eligible public vacancies from permitted sources to Telegram.
- `backend/internal/handlers/` owns HTTP routes and request/response handling.
- `backend/internal/jobs/`, `backend/internal/candidates/`,
  `backend/internal/companies/` and `backend/internal/auth/` own domain logic.
- `backend/internal/telegrampublisher/` owns Telegram publication formatting,
  duplicate tracking and retry orchestration.
- `backend/sql/queries/` contains sqlc query sources.
- `backend/migrations/` contains applied database migrations.
- `backend/internal/providers/` contains external vacancy provider adapters.

## Docs and operations

- `README.md` is the primary project overview.
- `docs/` contains architecture, provider and workflow notes.
- `Makefile` exposes local startup, migration, import and seed commands.
