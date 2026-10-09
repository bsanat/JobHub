---
title: "Telegram publisher"
description: "Как JobHub публикует вакансии в Telegram, где хранит статусы and какие условия подтверждены кодом."
---

# Telegram publisher

Telegram publisher sends eligible vacancies to a Telegram channel and records status in `jobhub.telegram_publications`.

## Confirmed entrypoints

| Entrypoint | Source | Behavior |
| --- | --- | --- |
| Local/dev command | `make telegram-publisher-dev` | Runs `backend/cmd/telegram-publisher --dry-run --once` against `TELEGRAM_DEV_DATABASE_URL` |
| Backend interval publisher | `backend/cmd/api/telegram_publisher.go` | Starts only when `TELEGRAM_PUBLISHER_ENABLED` is true and `VERCEL != 1` |
| Optional webhook endpoint | `POST /api/v1/internal/telegram-publish` | Runs one batch when `TELEGRAM_WEBHOOK_SECRET` and the production cutoff are configured; it is not the scheduled Neon path |
| Retired Supabase SQL | `scripts/supabase-telegram-publisher.sql` | Historical only; do not run on Neon |
| GitHub Actions schedule | `.github/workflows/telegram-publisher.yml` | Every 5 minutes when GitHub variable `TELEGRAM_PUBLISHER_ENABLED` is true on `main`; GitHub can delay or skip runs |

## Required config

Use placeholders only:

```env
TELEGRAM_BOT_TOKEN=<server-only-token>
TELEGRAM_PUBLIC_CHANNEL_ID=-1000000000000
TELEGRAM_PUBLIC_CHANNEL_USERNAME=<channel_username_without_at>
TELEGRAM_PUBLISHER_ENABLED=true
TELEGRAM_PUBLISHER_DRY_RUN=false
TELEGRAM_PUBLISH_WITHOUT_LINK=false
TELEGRAM_PUBLISH_INTERVAL=60s
TELEGRAM_PUBLISH_BATCH_SIZE=10
TELEGRAM_PUBLISH_MAX_ATTEMPTS=3
TELEGRAM_PUBLISH_AFTER=<UTC-activation-time-in-RFC3339>
JOBHUB_PUBLIC_BASE_URL=https://example.com
TELEGRAM_WEBHOOK_SECRET=<long-random-secret>
```

`JOBHUB_PUBLIC_BASE_URL` must be a public HTTPS URL for real publishing. Localhost is rejected for real delivery.
Production delivery requires `TELEGRAM_PUBLISH_AFTER` in RFC3339 format. Set it to the UTC activation time so the first run does not send the existing backlog, then keep it fixed; moving it forward would skip unsent jobs and retries. A native employer draft created earlier is still eligible if it is published after this time; imported jobs use their first-seen time. The scheduled workflow reads the repository variable of the same name.

## Eligibility

Confirmed by `backend/internal/telegrampublisher/store.go` and tests:

- source is enabled;
- source has production permissions confirmed, unless explicit staging test mode is enabled;
- source provider is not `demo`;
- `jobhub.job_is_public(j)` is true;
- creator is not the demo employer email;
- canonical duplicate check does not find an earlier equivalent vacancy;
- no `sent` record exists for the same `job_id` and `channel_id`;
- failed rows may retry only below max attempts and after at least one minute.

The older MVP statement “only `source = jobhub`” is no longer fully accurate for the current code: permitted external sources can be eligible when source permission is confirmed.

## Status tracking

`jobhub.telegram_publications` stores:

- `job_id`
- `channel_id`
- `telegram_message_id`
- `status`: `pending`, `sent`, `failed`
- `error_code`
- `attempts`
- `sent_at`
- timestamps

Success marks status `sent` with `telegram_message_id` and `sent_at`. Known Telegram errors mark `failed`. Rate-limit errors stop the batch and do not count attempts in the tested case. Unknown delivery errors remain pending to avoid duplicate sends.

## Message format

`backend/internal/telegrampublisher/formatter.go` formats bilingual HTML:

- title;
- company;
- city/location;
- salary when visible;
- work mode;
- experience;
- employment type;
- short description;
- source attribution for non-JobHub sources;
- public vacancy link when `JOBHUB_PUBLIC_BASE_URL` is configured.

Tokens, real channel IDs and private channel names must not be included in documentation or screenshots.
