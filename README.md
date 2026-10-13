<p align="center">
  <img src="frontend/public/brand/jobhub-logo.png" alt="Логотип JobHub" width="320">
</p>

# JobHub

JobHub — сайт вакансий для Казахстана. Соискатели ищут работу, сохраняют
вакансии, заполняют профиль и отправляют отклики. Работодатели управляют
компанией, публикуют вакансии и рассматривают кандидатов. Интерфейс
доступен на казахском и русском языках.

**Рабочий сайт:** [jobhub-lovat.vercel.app](https://jobhub-lovat.vercel.app/)

## Что работает

- Публичный каталог вакансий и компаний, поиск, сортировка и отдельные
  страницы вакансий и компаний.
- Регистрация соискателя и работодателя с подтверждением email через Supabase
  OTP, сессии в HttpOnly-cookie и защита изменяющих запросов CSRF-заголовком.
- Кабинет работодателя: создание, редактирование, публикация, приостановка,
  закрытие вакансий и просмотр откликов своей компании.
- Профиль соискателя, сохранённые вакансии, подписки на компании, PDF-резюме
  и отклики с сохранением версии резюме на момент отправки.
- Импортируемые вакансии и вакансии JobHub обрабатываются по разным правилам:
  на внешнюю вакансию отклик ведёт к источнику, на собственную — внутри сайта.
- Автоматическая публикация подходящих вакансий в Telegram-канал
  [@jobhu_test_jobs_2026](https://t.me/jobhu_test_jobs_2026) после
  одобрения и появления в публичном каталоге. Повторные сообщения
  блокируются записью в базе данных.

## Архитектура

| Компонент | Технологии | Production |
| --- | --- | --- |
| Frontend | React, TypeScript, Vite, TanStack Query, React Router, i18next | Сервис Vercel `/` |
| Backend | Go, Gin, pgx, sqlc | Сервис Vercel `/api/...` |
| Данные | PostgreSQL, миграции golang-migrate | Отдельный проект Supabase `JobHub Production` |
| Приватные PDF-резюме | Серверное хранилище | Закрытый bucket Supabase Storage `resumes` |
| Telegram publisher | PostgreSQL trigger, `pg_net`, `pg_cron`, Go webhook | Supabase вызывает защищённый endpoint Vercel |

Секреты находятся только в серверных переменных Vercel, Supabase Vault и,
для резервного GitHub Actions workflow, в GitHub Secrets. Frontend не
подключается к базе напрямую. Подробнее: [документация publisher](docs/telegram-publisher-webhook.md).

## Локальная разработка

Нужны Go 1.25+, Node.js 22.12+ и npm. Команда `make dev` запускает backend
и frontend напрямую против Supabase `DATABASE_URL` из `.env`; Docker для
обычной разработки не нужен. Перед локальным запуском создайте `.env`:

```sh
cp .env.example .env
make dev
```

`make dev` не применяет миграции автоматически и не загружает seed-данные:
схему Supabase обновляйте отдельно через защищённую команду
`make migrate-staging` после проверки target URL. Адреса:

- Сайт: <http://localhost:5173>
- Проверка API: <http://localhost:8080/api/v1/health>

Остановите процессы через `Ctrl+C`. Docker-режим для отдельной локальной
PostgreSQL-базы сохранён как `make dev-docker`; для остановки контейнеров
используйте `make local-down`. Не направляйте seed-команды и разрушающие тесты
на staging или Production Supabase.

Если backend уже запущен отдельно, не поднимайте второй процесс на том же
порту. Команды локальной разработки перечислены в
[справочнике](docs/developer-command-reference.md).

## Конфигурация Production

Проект Vercel развёрнут из репозитория как два сервиса (`frontend` и
`backend`). [`vercel.json`](vercel.json) направляет `/api/...` в Go backend,
а остальные публичные пути — во frontend. Backend использует отдельный
Production-проект Supabase. При изменении переменных Vercel требуется новый
deployment.

Главные серверные переменные Vercel Production:

| Переменная | Назначение |
| --- | --- |
| `DATABASE_URL` | Подключение Go API к Production Supabase |
| `PGX_QUERY_EXEC_MODE` | Режим pgx для используемого типа подключения |
| `FRONTEND_ORIGIN` | Разрешённый адрес frontend |
| `NEON_AUTH_BASE_URL`, `VITE_NEON_AUTH_URL` | Neon Auth: обязательное подтверждение email и Google OAuth |
| `AWS_ENDPOINT_URL_S3`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Приватные резюме в Neon Object Storage, bucket `uploads`; ключи только на сервере |
| `TELEGRAM_BOT_TOKEN` | Отправка сообщений через Telegram Bot API |
| `TELEGRAM_WEBHOOK_SECRET` | Проверка вызовов внутреннего publisher endpoint |
| `TELEGRAM_PUBLIC_CHANNEL_ID`, `TELEGRAM_PUBLIC_CHANNEL_USERNAME` | Канал публикации |
| `JOBHUB_PUBLIC_BASE_URL` | Основа ссылки на вакансию в сообщении |

Не копируйте реальные значения из `.env` в код, README или frontend
`VITE_*` переменные. Для миграций используйте отдельное
`DATABASE_MIGRATION_URL` и правила из [инструкции](docs/migration-runbook.md).
Переменные и шаги настройки Telegram подробно описаны в
[документации publisher](docs/telegram-publisher-webhook.md).

## Правила видимости вакансий

Вакансия JobHub видна публично, только если опубликована, одобрена
модерацией, не удалена и не просрочена, а компания и её владелец активны.
Импортируемые вакансии дополнительно зависят от состояния и прав
источника. Черновики, приостановленные и закрытые вакансии не попадают в
публичный каталог. Общие правила описаны в
[заметке о видимости](docs/job-visibility-rules.md).

## Проверки перед изменениями

```sh
cd backend
go test ./...
go vet ./...
go build ./...
```

```sh
cd frontend
npm ci
npm run lint
npm run typecheck
npm run build
```

После изменения SQL-запросов, конфигурации sqlc или используемых
генерируемым кодом типов запустите генерацию:

```sh
cd backend
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
```


## Документация

Навигация по темам: [docs/documentation-index.md](docs/documentation-index.md).
В ней есть инструкции по разработке, API, безопасности, источникам
вакансий, миграциям и Telegram publisher.
Не можешь guest заходить 

## Безопасность и ограничения

- `.env` и настоящие ключи не должны попадать в Git.
- Доступ к приватным резюме и откликам проверяет Go backend; скрытая кнопка
  в интерфейсе не является проверкой прав.
- Доступ к схеме `jobhub` через публичные ключи Supabase Data API не
  предоставляется.
- Для внешних источников нужны отдельно подтверждённые права на хранение,
  показ и повторную публикацию вакансий.
- В Production размер PDF ограничен лимитом запроса Vercel; локальный лимит
  может отличаться. Проверяйте действующие ограничения перед загрузкой.
