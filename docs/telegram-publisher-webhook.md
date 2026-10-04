# Публикация вакансий JobHub в Telegram

Документ описывает рабочую Production-схему. Сайт:
`https://jobhub-lovat.vercel.app`, канал: `@jobhu_test_jobs_2026`.
Docker и постоянно запущенный процесс на компьютере для этого не нужны.
Реальных значений секретов в репозитории нет.

## 1. Процесс по шагам

```text
Работодатель публикует вакансию, модерация её одобряет
  -> INSERT/UPDATE в jobhub.jobs (Supabase)
  -> вакансия впервые становится публичной: срабатывает триггер PostgreSQL
  -> pg_net асинхронно вызывает защищённый API на Vercel
  -> Go publisher выбирает одну подходящую, ещё не отправленную вакансию
  -> Telegram Bot API публикует сообщение в канале
  -> jobhub.telegram_publications сохраняет результат и message ID

pg_cron каждую минуту -> вызывает тот же API для пропущенных вакансий
  и повторов после известных ошибок
```

Триггер и cron не являются двумя независимыми алгоритмами публикации: они
используют один endpoint, запрос отбора и таблицу публикаций. Триггер быстро
ставит запрос в очередь, но доставка в Telegram асинхронна и не гарантирована
в ту же миллисекунду. Один вызов webhook обрабатывает одну вакансию. Старый
пятиминутный workflow GitHub Actions пока сохранён как дополнительный резерв.

## 2. Какие вакансии подходят

Общее условие публичности — функция `jobhub.job_is_public(job)`. Для
собственной вакансии (`source='jobhub'`) необходимы:

1. `publication_status='published'` и `moderation_status='approved'`.
2. Вакансия не удалена и срок её действия не истёк.
3. Компания имеет статус `active` и активного пользователя с ролью `owner`.

Импортированной вакансии нужны одобрение модерации, `source_status` со
значением `unknown` либо `active`, действующий `fresh_until` и отсутствие
удаления. Кроме того, publisher выбирает только включённые источники с
подтверждённым разрешением для Production. Демо-провайдер, демо-работодатель
и неосновная копия одинаковой вакансии исключаются. Режим
`TELEGRAM_PUBLISH_TEST_SOURCES` в Production не используется.

Webhook не получает ID или текст вакансии. Backend заново проверяет условия
непосредственно в базе. Поэтому ручной вызов webhook тоже не отправит
закрытую или неодобренную вакансию. Если статус компании/владельца изменится
без срабатывания триггера вакансии, её обнаружит очередной запуск cron.

## 3. Где хранятся токены и настройки

| Место | Имя | Назначение |
| --- | --- | --- |
| Vercel `jobhub` -> Settings -> Environment Variables -> Production | `TELEGRAM_BOT_TOKEN` (Secret) | Отправка через Telegram Bot API |
| Там же | `TELEGRAM_WEBHOOK_SECRET` (Secret, не менее 32 символов) | Защита вызова Supabase -> Vercel |
| Там же | `TELEGRAM_PUBLIC_CHANNEL_ID` | Отрицательный числовой ID канала |
| Там же | `TELEGRAM_PUBLIC_CHANNEL_USERNAME` | `jobhu_test_jobs_2026` |
| Там же | `JOBHUB_PUBLIC_BASE_URL` | `https://jobhub-lovat.vercel.app` |
| Там же | `DATABASE_URL` (Secret) | Подключение Go API к Production Supabase |
| Supabase `JobHub Production` -> Integrations -> Vault -> Secrets | `jobhub_telegram_publisher_url` | `https://jobhub-lovat.vercel.app/api/v1/internal/telegram-publish` |
| Там же | `jobhub_telegram_publisher_secret` | То же значение, что у Vercel `TELEGRAM_WEBHOOK_SECRET` |
| GitHub `juniorqazaq/JobHub` -> Settings -> Secrets and variables -> Actions | `TELEGRAM_BOT_TOKEN`, `TELEGRAM_PUBLISHER_DATABASE_URL` (Secrets) | Резервный workflow |
| Там же, Variables | `TELEGRAM_PUBLISHER_ENABLED`, `TELEGRAM_PUBLIC_CHANNEL_ID`, `TELEGRAM_PUBLIC_CHANNEL_USERNAME` | Включение резерва и выбор канала |

**Webhook secret и bot token — разные ключи.** Первый защищает внутренний
endpoint, второй даёт боту право отправлять сообщения. Не переносите их в
переменные frontend `VITE_*`, SQL Editor, коммиты или сообщения чата.
Изменение переменной Vercel требует нового Production deployment. При смене
ключа в Vault синхронно обновите его пару в Vercel.

В Production нельзя включать `TELEGRAM_PUBLISHER_DRY_RUN=true` и
`TELEGRAM_PUBLISH_WITHOUT_LINK=true`. В serverless API Vercel постоянный Go
цикл намеренно выключен: publisher вызывается через webhook. Параметр
`TELEGRAM_PUBLISHER_ENABLED` управляет самостоятельным worker и резервным
GitHub Actions workflow, а не регистрацией webhook. Маршрут webhook
регистрируется только при наличии `TELEGRAM_WEBHOOK_SECRET`.

## 4. Как работает webhook

В Production SQL Editor один раз был выполнен скрипт
[`scripts/supabase-telegram-publisher.sql`](../scripts/supabase-telegram-publisher.sql).
Он включает `pg_net`, `pg_cron`, создаёт функции, триггер для `jobhub.jobs`
и задачу `jobhub-telegram-publisher-retry`.

1. После `INSERT` или `UPDATE` полей `publication_status`,
   `moderation_status`, `source_status`, `fresh_until` триггер вызывает
   `jobhub.request_telegram_publication()`, если вакансия впервые стала
   публичной.
2. Функция читает URL и webhook secret из Vault. `pg_net` отправляет
   `POST /api/v1/internal/telegram-publish` с пустым телом `{}`. Текст
   вакансии и Telegram bot token этим запросом не передаются.
3. Go API проверяет заголовок `X-JobHub-Publisher-Secret` сравнением с
   постоянным временем выполнения. Неверный или отсутствующий ключ даёт
   `401`; ошибка publisher — `503`.
4. Если постановка запроса в очередь не удалась, триггер записывает warning,
   но не блокирует публикацию вакансии работодателем. Cron вызывает ту же
   функцию каждую минуту. Успешное выполнение SQL-задачи ещё не означает
   отправку в Telegram: проверяйте HTTP-ответ и запись публикации.

Код: [`backend/cmd/api/telegram_webhook.go`](../backend/cmd/api/telegram_webhook.go)
и [`backend/internal/handlers/telegram_publish.go`](../backend/internal/handlers/telegram_publish.go).

## 5. Формат сообщения и защита от повторов

Publisher выбирает самую старую подходящую вакансию. В таблице
`telegram_publications` действует ограничение `UNIQUE(job_id, channel_id)`.
Сначала создаётся заявка со статусом `pending`. После успешного ответа
Telegram сохраняются `sent`, `telegram_message_id` и `sent_at`. Ни триггер,
ни cron, ни GitHub Actions не выберут её повторно. Старые публикации уже
записаны в той же таблице, поэтому новая схема их не отправляет заново.

При известном сбое записывается `failed`. Повтор возможен после задержки,
обычно не более трёх засчитываемых попыток. Если результат отправки
неизвестен (например, timeout), запись остаётся `pending`: автоматический
повтор мог бы создать дубликат. Оператор должен сверить её с каналом.

Сообщение содержит название, компанию, город, видимую зарплату, формат,
опыт, короткое описание и ссылку на сайт. Часть подписей двуязычна;
HTML-текст экранируется. Формат ссылки:
`https://jobhub-lovat.vercel.app/jobs/<job-id>`. Для импортированной
вакансии добавляется указание источника.

## 6. Проверка и диагностика

Запросы ниже **только читают** Production-базу:

```sql
SELECT jobname, schedule, active FROM cron.job
WHERE jobname = 'jobhub-telegram-publisher-retry';

SELECT tgname FROM pg_trigger
WHERE tgrelid = 'jobhub.jobs'::regclass
  AND tgname = 'notify_telegram_on_publication' AND NOT tgisinternal;

SELECT j.jobname, d.status, d.return_message, d.start_time
FROM cron.job_run_details d JOIN cron.job j ON j.jobid = d.jobid
WHERE j.jobname = 'jobhub-telegram-publisher-retry'
ORDER BY d.start_time DESC LIMIT 5;

SELECT status_code, error_msg, created FROM net._http_response
ORDER BY created DESC LIMIT 5;

SELECT job_id, status, telegram_message_id, attempts, error_code, sent_at
FROM jobhub.telegram_publications ORDER BY created_at DESC LIMIT 10;
```

`cron.job_run_details.status='succeeded'` означает только выполнение SQL.
`net._http_response.status_code=200` подтверждает ответ Vercel, а
`telegram_publications.status='sent'` и `telegram_message_id` — сохранённую
публикацию. При установке 2026-10-04 были проверены один триггер, одна
активная cron-задача и HTTP 200; первый запуск отправил одну вакансию и
сохранил Telegram message ID `110`. Это не заменяет отдельную проверку
полного пути новой вакансии.

Ручной вызов защищённого endpoint (**может отправить настоящее сообщение**):

```sh
# Задайте WEBHOOK_SECRET безопасным способом как переменную окружения.
# Не вписывайте его значение в команду, историю shell или Git.
curl --fail-with-body -X POST \
  -H "X-JobHub-Publisher-Secret: ${WEBHOOK_SECRET}" \
  -H 'Content-Type: application/json' \
  -d '{}' \
  https://jobhub-lovat.vercel.app/api/v1/internal/telegram-publish
```

Пример ответа: `{"checked":1,"sent":1,"failed":0,"unknown":0}`.
`checked=0` означает отсутствие подходящей новой вакансии сейчас. Если
сообщение не пришло, проверьте публичность и права источника, историю cron,
статус `pg_net`, логи API Vercel и строку `telegram_publications`.
Статус `401` указывает на отсутствующий/несовпадающий webhook key, `503` —
на ошибку publisher (смотрите логи API). Не выводите в логи секреты,
полные заголовки или расшифрованные значения Vault.

## 7. Установка в новой среде и замена ключей

1. Проверьте таблицу `jobhub.telegram_publications` (миграция `000014`).
   Бот должен быть администратором канала с правом публикации.
2. Задайте переменные Vercel Production. Создайте отдельный случайный
   webhook key длиной не менее 32 символов и сохраните как Secret.
3. В Vault сохраните URL и **тот же** ключ в двух named secrets. Сделайте
   redeploy Vercel и проверьте работу API.
4. Только после этого выполните SQL-скрипт в Production SQL Editor.
5. Проверьте SQL-запросами выше, затем испытайте новую реальную вакансию:
   её страницу, единственное сообщение в канале и правильную ссылку.

Если webhook key раскрыт, замените **оба** значения в Vercel и Vault и
сделайте redeploy. Если раскрыт bot token, перевыпустите его через Telegram
BotFather и обновите Vercel и GitHub Actions Secrets. Решение об отключении
резервного GitHub Actions workflow принимайте отдельно после проверки нового
пути в Production.
