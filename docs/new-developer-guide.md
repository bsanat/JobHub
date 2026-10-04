---
title: "Инструкция для нового разработчика"
description: "Пошаговый вход в проект JobHub: что читать, как запустить, что проверить и где не трогать production."
---

# Инструкция для нового разработчика

## Быстрый маршрут

1. Прочитайте [О проекте](/docs/project-overview) и [Архитектуру системы](/docs/system-architecture).
2. Проверьте локальные требования: Go `1.25+`, Node.js, npm и доступ к Supabase PostgreSQL для backend.
3. Создайте `.env` из `.env.example`. Не копируйте реальные production secrets в документацию, issue, PR или скриншоты.
4. Для локального full-stack запуска используйте `make dev`.
5. Frontend работает через typed repositories in `frontend/src/api/`; браузер не подключается напрямую к Supabase.
6. Backend routes находятся в `backend/internal/handlers/`; SQL source queries in `backend/sql/queries/`; migrations in `backend/migrations/`.
7. Перед изменениями смотрите `git status --short --branch`, чтобы не смешать свою работу с чужими изменениями.

## Локальный запуск

```sh
cp .env.example .env
make dev
```

Подтвержденные локальные адреса:

- Frontend: `http://localhost:5173`
- Backend health: `http://localhost:8080/api/v1/health`

Если вы запускаете frontend отдельно:

```sh
cd frontend
npm run dev
```

Если запускаете backend отдельно, используйте `make backend-dev` или убедитесь,
что `DATABASE_URL`, `FRONTEND_ORIGIN` and `PORT` заданы корректно.

## Переменные окружения

| Переменная | Где используется | Назначение |
| --- | --- | --- |
| `DATABASE_URL` | backend | PostgreSQL connection URL |
| `DATABASE_MIGRATION_URL` | backend migrate scripts | URL для миграций; если пустой, backend config берёт `DATABASE_URL` |
| `PGX_QUERY_EXEC_MODE` | backend | `cache_statement`, `exec` or `simple_protocol`; для Supabase transaction pooler on port `6543` нужен не `cache_statement` |
| `FRONTEND_ORIGIN` | backend CORS/origin checks | Trusted origin для state-changing requests |
| `VITE_API_URL` | frontend | Base API URL; default `/api/v1` in production, `http://localhost:8080/api/v1` in development |
| `RESUME_STORAGE_DIR` | backend | Local resume storage fallback |
| `SUPABASE_SECRET_KEY` | backend storage adapter | Server-only key for Supabase Storage; не подтверждено как обязательный для всех окружений |
| `TELEGRAM_*` | publisher | Telegram bot, channel and publishing settings |
| `JOBHUB_PUBLIC_BASE_URL` | publisher | Public URL used in Telegram vacancy links |
| `TELEGRAM_WEBHOOK_SECRET` | API | Enables protected webhook endpoint when present |

Не документируйте реальные значения `.env`. В примерах используйте placeholders.

## Что безопасно менять сначала

| Задача | Где начинать |
| --- | --- |
| Frontend page text or layout | `frontend/src/pages`, `frontend/src/components`, `frontend/src/i18n/locales/kk.ts`, `frontend/src/i18n/locales/ru.ts` |
| API response mapping | `frontend/src/api/http/*`, backend handler DTOs |
| Vacancy visibility | `backend/internal/jobs`, `backend/sql/queries/jobs.sql`, migration function `jobhub.job_is_public` |
| Employer vacancy flow | `EmployerVacancyFormPage.tsx`, `HttpEmployerJobsRepository.ts`, `backend/internal/handlers/employer_jobs.go` |
| Candidate profile/resume | `frontend/src/pages/ProfilePage.tsx`, `ResumePage.tsx`, `backend/internal/handlers/candidates.go` |
| Telegram publishing | `backend/internal/telegrampublisher`, `backend/cmd/api/telegram_webhook.go`, `scripts/supabase-telegram-publisher.sql` |

## Safety rules

- Не запускайте migrations against production without explicit approval and verified target URL.
- Не направляйте seed/demo scripts на production Supabase.
- Не выводите `.env`, tokens, database URLs or private resume content в logs, screenshots или docs.
- Для JobHub UI сейчас поддерживаются только `kk` и `ru`; новые UI тексты добавляйте в оба locale-файла.
- Не удаляйте старые docs: помечайте устаревшее и связывайте с новой страницей.
