# Порядок выполнения миграций

Безопасный порядок локальных и удалённых миграций.

## Локальные миграции

```sh
make migrate-local
```

Локальные миграции работают с PostgreSQL в Docker и используют
`sslmode=disable`.

## Удалённые миграции

```sh
make migrate-version
make migrate-staging
```

Для удалённых миграций нужны `APP_ENV=staging` и `DATABASE_MIGRATION_URL` с
прямым или сессионным подключением через порт `5432`.

## Правила

- Не запускайте удалённые миграции через transaction pooler на порту `6543`.
- Не используйте `sslmode=disable` для Supabase.
- Не переписывайте применённые миграции.
- Не выполняйте down, reset и force для staging или Production.
- По возможности разделяйте подключения приложения и миграций.

Требования к RLS для таблиц описаны в `backend/migrations/README.md`.
