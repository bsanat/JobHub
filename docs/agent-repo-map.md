# Карта репозитория для агента

Краткий указатель по репозиторию JobHub. Перед использованием проверяйте пути:
источником истины остаётся код.

## Фронтенд

- `frontend/src/pages/` — страницы маршрутов.
- `frontend/src/pages/EmployerApplicantsPage.tsx` — `/employer/applications`
  со всеми откликами владельца и `/employer/vacancies/:jobId/applicants`
  по вакансии; API и проверка владения — `handlers/candidates.go`
  и `candidates/postgres_marketplace.go` внутри `backend/internal/`.
- `frontend/src/pages/OnboardingPage.tsx` — опрос `/onboarding` после регистрации;
  `frontend/src/app/OnboardingGate.tsx` восстанавливает незавершённый опрос.
- `frontend/src/pages/LoginPage.tsx`, `frontend/src/pages/RegisterPage.tsx` and
  `frontend/src/pages/AuthCallbackPage.tsx` — email/password auth, обязательный
  Neon email OTP при регистрации и Google OAuth callback; API проверяет
  Neon EdDSA JWT до создания локального аккаунта.
- `frontend/src/pages/SavedPage.tsx` — единая страница `/saved` с вакансиями
  и компаниями; панели — `frontend/src/components/candidate/Saved{Jobs,Companies}Panel.tsx`.
  `/saved-companies` перенаправляет на `/saved?tab=companies`.
- `frontend/src/pages/StudentsPage.tsx` — публичная страница `/students` с вакансиями по фильтрам стажировки, без опыта и частичной занятости.
- `backend/cmd/import-public-jobs/` и `backend/internal/providers/publicapi/` — отключённый по умолчанию разовый импорт Arbeitnow/Lever; Greenhouse использует существующий адаптер. Настройка и ограничения — `docs/public-job-imports.md`.
- `backend/cmd/import-kz-company-jobs/`, `backend/internal/kzcompanies/` и
  `backend/config/kz-company-sources.json` — отключённый импорт официальных
  карьерных страниц компаний Казахстана; реестр и запуск — `docs/kz-company-imports.md`.
- `.github/workflows/kz-company-scheduled-import.yml` — почасовой запуск только
  для проверенных источников из переменной `KZ_IMPORT_AUTO_SOURCES`.
- `frontend/src/components/` — общие и функциональные компоненты.
- `frontend/src/components/jobs/SalaryFilter.tsx` — диапазон зарплаты в каталоге;
  параметры URL — `frontend/src/lib/jobSearchParams.ts`, SQL — `backend/sql/queries/jobs.sql`.
- `frontend/src/components/employer/EmployerWorkspaceLayout.tsx` and
  `frontend/src/components/candidate/CandidateWorkspaceLayout.tsx` — навигация
  кабинетов; их стили находятся в `frontend/src/styles/workspace.css`.
- `frontend/src/api/http/` — преобразование HTTP DTO в модели frontend.
- `frontend/src/api/models/` — доменные типы frontend.
- `frontend/src/i18n/locales/kk.ts` and `frontend/src/i18n/locales/ru.ts`
  содержат тексты действующего интерфейса.
- `frontend/src/styles/` — общие стили и стили каталога.

## Бэкенд

- `backend/cmd/api/` запускает API.
- `backend/internal/auth/onboarding.go` — ответы аккаунта и защита версий;
  `backend/internal/handlers/onboarding.go` — GET/PUT `/api/v1/auth/onboarding`.
- `backend/internal/auth/oauth.go` and `backend/internal/handlers/auth.go` —
  проверка Neon JWT и привязки Google account и создание JobHub session.
- `backend/cmd/telegram-publisher/` публикует подходящие вакансии в Telegram.
- `backend/internal/handlers/` обрабатывает HTTP-маршруты и запросы.
- `backend/internal/jobs/`, `backend/internal/candidates/`,
  `backend/internal/companies/` и `backend/internal/auth/` — доменная логика.
- `backend/internal/candidates/s3_storage.go` хранит приватные резюме
  в Neon Object Storage через SigV4 при serverless-развёртывании.
- `backend/internal/telegrampublisher/` форматирует сообщения, предотвращает
  дубликаты и управляет повторами.
- `backend/internal/handlers/telegram_publish.go` — защищённый endpoint
  однократного запуска publisher; после переноса на Neon публикация выключена.
- `backend/sql/queries/` — исходные запросы sqlc.
- `backend/migrations/` — миграции базы данных.
- `backend/internal/providers/` — адаптеры внешних источников вакансий.

## Документация и эксплуатация

- `README.md` — основной обзор проекта.
- `docs/` — архитектура, источники и рабочие инструкции.
- `Makefile` — запуск, миграции, импорт и демоданные.
- `.github/workflows/telegram-publisher.yml` — резервный запуск publisher вне Vercel.
- `.github/workflows/public-job-imports.yml` — выключенный по умолчанию
  расписанный запуск API-импортера.
- `scripts/supabase-telegram-publisher.sql` and
  `docs/telegram-publisher-webhook.md` описывают прежний Supabase webhook и cron; не применять к Neon.
