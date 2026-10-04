---
title: "О проекте"
description: "Что делает JobHub, кто им пользуется, какие технологии применяются и где лежат ключевые части кода."
---

# О проекте

JobHub — веб-приложение для поиска и публикации вакансий в Казахстане. По коду подтверждены три основные роли:

- **Гость** смотрит публичные вакансии и компании.
- **Соискатель** ведёт профиль, загружает PDF-резюме, сохраняет вакансии и компании, отправляет отклики.
- **Работодатель** управляет профилем компании, вакансиями и откликами на свои вакансии.
- **Администратор** имеет защищённый frontend-маршрут `/admin` и backend-проверку `/api/v1/admin/me`; полноценные admin CRUD-функции в проверенных handler-ах не подтверждены.

## Технологии

| Слой | Подтверждено кодом |
| --- | --- |
| Frontend | React `19.3.0`, React Router `7.18.4`, TypeScript `6.0.3`, Vite `8.3.0`, TanStack Query `5.103.2`, i18next `26.4.2`, axios `1.20.0`, Tailwind CSS `4.3.3`, lucide-react `1.47.0` |
| Backend | Go `1.25.0`, Gin `1.11.0`, pgx `5.7.6`, golang-migrate `4.18.3`, sqlc-generated database layer |
| Database | PostgreSQL schema `jobhub`, migrations in `backend/migrations` |
| Storage | Local filesystem by default; Supabase Storage adapter for resumes exists in `backend/internal/candidates/supabase_storage.go` |
| Deploy | Vercel multi-service project: `frontend` serves `/`, `backend` serves `/api/...` |
| Telegram | Go publisher, Telegram Bot API client, optional Supabase trigger/cron SQL and GitHub Actions fallback |
| Documentation | Mintlify config in `docs.json`; source pages live in `index.md` and `docs/*.md` |

## Структура репозитория

| Путь | Назначение |
| --- | --- |
| `frontend/src/App.tsx` | React routes and role-protected pages |
| `frontend/src/pages/` | Page-level components |
| `frontend/src/components/` | Shared UI and feature components |
| `frontend/src/api/` | HTTP repositories, DTO mappers and domain models |
| `frontend/src/lib/env.ts` | Frontend runtime config; default API URL is `/api/v1` in production and `http://localhost:8080/api/v1` in development |
| `backend/cmd/api/` | API server entrypoint and Telegram webhook registration |
| `backend/internal/handlers/` | Gin route handlers and auth/role middleware |
| `backend/internal/jobs/` | Vacancy read/write store logic |
| `backend/internal/candidates/` | Candidate profile, resume, saved jobs and applications logic |
| `backend/internal/companies/` | Company profiles and follows |
| `backend/internal/auth/` | Users, sessions, cookies and CSRF |
| `backend/internal/telegrampublisher/` | Vacancy eligibility, duplicate protection and Telegram delivery flow |
| `backend/internal/providers/` | External providers and Telegram client |
| `backend/sql/queries/` | sqlc source queries |
| `backend/migrations/` | Database schema migrations |
| `scripts/` | Local dev and operational scripts |
| `.github/workflows/telegram-publisher.yml` | Scheduled Telegram publisher fallback |
| `docs/` | Project documentation and runbooks |

Папка `documents` в текущем рабочем дереве не найдена. Существующие материалы лежат в `docs/`; они сохранены и связаны с этим Mintlify-сайтом.

## Команды разработки

Команды подтверждены `Makefile`, `frontend/package.json` and backend Go layout.

| Команда | Где запускать | Что делает |
| --- | --- | --- |
| `make dev` | root | Запускает `scripts/dev-supabase.sh` для разработки против Supabase из `.env` |
| `make backend-dev` | root | Запускает backend-часть через `scripts/dev-supabase.sh backend` |
| `make dev-docker` | root | Запускает старый локальный PostgreSQL flow через `scripts/dev-local.sh` |
| `make local-up` | root | Поднимает локальный Docker Compose profile `local` |
| `make local-down` | root | Останавливает локальный Docker Compose profile `local` |
| `make migrate-local` | root | Запускает миграции в локальном Docker Compose окружении |
| `make migrate-staging` | root | Запускает `scripts/migrate-remote.sh up`; использовать только после проверки целевого `DATABASE_MIGRATION_URL` |
| `make telegram-publisher-dev` | root | Запускает publisher в dry-run против `TELEGRAM_DEV_DATABASE_URL` |
| `npm run dev` | `frontend` | Vite dev server |
| `npm run build` | `frontend` | `tsc -b` and `vite build` |
| `npm run lint` | `frontend` | ESLint |
| `go test ./...` | `backend` | Backend tests; известная историческая проблема с packages without non-test files должна проверяться по текущему состоянию |

## Существующая документация

Полезные исходные материалы:

- [Индекс документации](/docs/documentation-index)
- [Карта репозитория](/docs/agent-repo-map)
- [Команды разработки](/docs/developer-command-reference)
- [Telegram publisher](/docs/telegram-vacancy-publisher)
- [Telegram webhook](/docs/telegram-publisher-webhook)
- [API notes](/docs/api-route-notes)
- [Migration runbook](/docs/migration-runbook)

Часть старых документов пересекается по темам: `telegram-vacancy-publisher.md`, `telegram-publisher-webhook.md`, `backend-verification.md`, `testing-matrix.md`, `release-readiness.md`. Они не удалены; новые страницы дают карту фактического состояния и ссылаются на них как на детали.
