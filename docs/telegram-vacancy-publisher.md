# Publisher вакансий в Telegram

Publisher отправляет подходящие публичные вакансии в один канал JobHub.
Отдельная Go-команда находится в `backend/cmd/telegram-publisher` и может
работать с Supabase Postgres без Docker. Действующая Production-схема с
webhook описана в [отдельной инструкции](/docs/telegram-publisher-webhook).

## Правила отбора

- Собственные вакансии JobHub и внешние вакансии, для которых у источника
  подтверждено `production_permissions_confirmed = true`.
- Общая функция публичной видимости `jobhub.job_is_public(j)`.
- Демоданные работодателя `demo-employer@jobhub.local` и демо-провайдеры
  исключены; источники без прав повторного использования пропускаются.
- Одинаковые вакансии сравниваются по названию, компании, городу и описанию;
  сохраняется самая ранняя публичная запись.
- Каждая отправка записывается в `jobhub.telegram_publications` для защиты
  от дубликатов.
- Известные ошибки повторяются не ранее чем через минуту до лимита
  `TELEGRAM_PUBLISH_MAX_ATTEMPTS`; rate limit и временная потеря прав бота
  не расходуют лимит.
- Неопределённый исход остаётся `pending` для ручной сверки с каналом.

## Переменные окружения

| Имя | Назначение |
| --- | --- |
| `TELEGRAM_BOT_TOKEN` | Секрет бота, только на сервере |
| `TELEGRAM_PUBLIC_CHANNEL_ID` | Отрицательный числовой ID канала |
| `TELEGRAM_PUBLIC_CHANNEL_USERNAME` | Публичное имя канала; в режиме отправки бот проверяет соответствие ID |
| `TELEGRAM_PUBLISHER_ENABLED` | Разрешает отправку из самостоятельной Go-команды |
| `TELEGRAM_PUBLISHER_DRY_RUN` | Только печать сообщений, без отправки и записи `sent` |
| `TELEGRAM_PUBLISH_INTERVAL` | Интервал цикла; по умолчанию `60s` |
| `TELEGRAM_PUBLISH_BATCH_SIZE` | Вакансий за проход; по умолчанию `10` |
| `TELEGRAM_PUBLISH_MAX_ATTEMPTS` | Лимит учитываемых ошибок; по умолчанию `3` |
| `JOBHUB_PUBLIC_BASE_URL` | Публичный HTTPS-адрес для ссылок `/jobs/:id` |
| `TELEGRAM_PUBLISH_WITHOUT_LINK` | Временный тестовый режим до появления публичного URL; ссылки из текста удаляются |
| `TELEGRAM_PUBLISH_TEST_SOURCES` | Явный режим только для staging и тестового канала |
| `TELEGRAM_PUBLISH_TEST_CHANNEL_ID` | Должен совпадать с выбранным каналом при тестовых источниках |

Сообщения без ссылок разумно использовать только в тестовом канале: ссылка
на сайт не добавится к уже отправленному сообщению автоматически.

## Локальная проверка без отправки

Задайте `TELEGRAM_DEV_DATABASE_URL` на локальную PostgreSQL-базу разработки
с подходящим префиксом `jobhub_*_poc_*`. Dry-run отклоняет staging и
Production URL, даже если они находятся в `.env`.

```sh
make telegram-publisher-dev
```

## Разовый запуск с Supabase без Docker

**Команда может отправить настоящие сообщения.** Перед запуском проверьте
окружение, канал и записи о прежних публикациях.

```sh
set -a
. ./.env
set +a
cd backend
TELEGRAM_PUBLISHER_ENABLED=true TELEGRAM_PUBLISHER_DRY_RUN=false go run ./cmd/telegram-publisher --once
```

Без `--once` команда опрашивает базу постоянно. Для отдельно разрешённого
теста staging можно задать `TELEGRAM_PUBLISH_TEST_SOURCES=true` и
`TELEGRAM_PUBLISH_TEST_CHANNEL_ID`. Это не меняет права источника для
Production и отклоняется вне staging. Режим без ссылки задаётся через
`TELEGRAM_PUBLISH_WITHOUT_LINK=true`.

## Эксплуатация

- Боту нужны только права публикации в собственном публичном канале JobHub.
  При запуске отдельная команда сверяет username, числовой ID и право
  публикации до выбора вакансий.
- Существующие подходящие вакансии отправляются от старых к новым.
  Ограничение `UNIQUE(job_id, channel_id)` защищает от повторов при
  перезапусках и одновременной работе нескольких исполнителей.
- Запись `pending` после неопределённого ответа Telegram сначала сравните
  с каналом. У Bot API нет idempotency key для `sendMessage`.
- Примените миграцию `000014` до запуска publisher.
- Локальный API может запускать фоновой publisher, если он явно включён.
  На Vercel постоянный процесс не работает: основная схема использует
  webhook Supabase, а GitHub Actions — резервный разовый запуск.
- Между сообщениями в пакете выдерживается минимум 1,1 секунды. Ответ
  Telegram `429` останавливает пакет; `retry_after` учитывается при повторе.

## Резервное расписание GitHub Actions

`.github/workflows/telegram-publisher.yml` запускает команду `--once` каждые
пять минут на ветке `main`, если репозиторная переменная
`TELEGRAM_PUBLISHER_ENABLED=true`. GitHub может задержать или пропустить
запуск; точное время доставки не гарантируется.

Secrets: проверенный `PUBLIC_IMPORT_DATABASE_URL` той же Production-базы и
`TELEGRAM_BOT_TOKEN`. Variables: `PUBLIC_IMPORT_DATABASE_HOST`,
`PUBLIC_IMPORT_DATABASE_USER`, `TELEGRAM_PUBLIC_CHANNEL_ID` и
`TELEGRAM_PUBLIC_CHANNEL_USERNAME`. Workflow проверяет имя хоста и пользователя
перед подключением. Не записывайте секреты в файл workflow,
Variables или `VITE_*`. Домен ссылок — `https://jobhub-lovat.vercel.app`.
Один запуск отправляет до десяти вакансий. После отправки проверяйте логи
workflow и при необходимости вручную сверяйте `pending`.
