# Справочник команд разработчика

Краткий список команд для локальной разработки JobHub. Подробности о функциях
и провайдерах находятся в отдельных документах `docs/`.

## Локальный запуск

| Задача | Команда | Примечание |
| --- | --- | --- |
| Запустить стек разработки | `make dev` | Локальная POC-база; данные staging Supabase не меняются. |
| Запустить Docker-сервисы с пустой базой | `make local-up` | Собирает и запускает PostgreSQL, миграции и backend. |
| Остановить Docker-сервисы | `make local-down` | Останавливает локальный профиль Compose. |
| Применить локальные миграции | `make migrate-local` | Использует локальный PostgreSQL в Docker. |
| Импортировать POC-вакансии Jooble | `make import-jooble` | Нужен `JOOBLE_API_KEY` в `.env`; число запросов ограничено. |
| Загрузить демоданные | `make seed-demo` | Локальные данные через Go-команду. |

## Работа с frontend

```sh
cd frontend
npm ci
npm run dev
```

Адрес Vite по умолчанию — `http://localhost:5173`. Для обычной разработки
`VITE_API_URL` должен указывать на локальный backend.

## Работа с backend

```sh
docker compose --profile local up -d postgres
make migrate-local

set -a
. ./.env
set +a

cd backend
go run ./cmd/api
```

Не запускайте backend из Docker и Go-процесс вручную на одном порту одновременно.

## Проверки

Backend:

```sh
cd backend
go test -race ./...
go vet ./...
go build ./...
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
```

Frontend:

```sh
cd frontend
npm run lint
npm run typecheck
npm run build
```

Репозиторий:

```sh
git diff --check
git status
```

## Правила безопасности

- Не добавляйте в Git `.env`, ключи Supabase и провайдеров, реальные строки подключения.
- Используйте `sslmode=disable` только для локального PostgreSQL.
- Для прямых и сессионных подключений к Supabase нужен `sslmode=require`.
- Удалённые миграции выполняйте только защищёнными командами
  `make migrate-version` и `make migrate-staging` по инструкции в `README.md`.
