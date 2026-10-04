# Карта репозитория для агента

Краткий указатель по репозиторию JobHub. Перед использованием проверяйте пути:
источником истины остаётся код.

## Фронтенд

- `frontend/src/pages/` — страницы маршрутов.
- `frontend/src/components/` — общие и функциональные компоненты.
- `frontend/src/api/http/` — преобразование HTTP DTO в модели frontend.
- `frontend/src/api/models/` — доменные типы frontend.
- `frontend/src/i18n/locales/kk.ts` and `frontend/src/i18n/locales/ru.ts`
  содержат тексты действующего интерфейса.
- `frontend/src/styles/` — общие стили и стили каталога.

## Бэкенд

- `backend/cmd/api/` запускает API.
- `backend/cmd/telegram-publisher/` публикует подходящие вакансии в Telegram.
- `backend/internal/handlers/` обрабатывает HTTP-маршруты и запросы.
- `backend/internal/jobs/`, `backend/internal/candidates/`,
  `backend/internal/companies/` и `backend/internal/auth/` — доменная логика.
- `backend/internal/candidates/supabase_storage.go` хранит приватные резюме
  в Supabase Storage при serverless-развёртывании.
- `backend/internal/telegrampublisher/` форматирует сообщения, предотвращает
  дубликаты и управляет повторами.
- `backend/internal/handlers/telegram_publish.go` — защищённый endpoint
  однократного запуска publisher по уведомлению Supabase.
- `backend/sql/queries/` — исходные запросы sqlc.
- `backend/migrations/` — миграции базы данных.
- `backend/internal/providers/` — адаптеры внешних источников вакансий.

## Документация и эксплуатация

- `README.md` — основной обзор проекта.
- `docs/` — архитектура, источники и рабочие инструкции.
- `Makefile` — запуск, миграции, импорт и демоданные.
- `.github/workflows/telegram-publisher.yml` — резервный запуск publisher вне Vercel.
- `scripts/supabase-telegram-publisher.sql` and
  `docs/telegram-publisher-webhook.md` описывают Production webhook и cron.
