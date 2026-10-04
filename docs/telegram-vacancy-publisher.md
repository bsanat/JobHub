# Telegram vacancy publisher

The Telegram vacancy publisher posts eligible public vacancies to one JobHub-owned public Telegram channel. It is a separate backend process at `backend/cmd/telegram-publisher` and can use Supabase Postgres directly without Docker.

MVP scope:

- posts native JobHub vacancies and external vacancies only when the source has `production_permissions_confirmed = true`;
- requires the existing shared public visibility predicate `jobhub.job_is_public(j)`;
- excludes the local/staging demo employer (`demo-employer@jobhub.local`) and demo providers so seeded examples are never published as real vacancies;
- skips external sources whose reuse permissions are not yet confirmed, even if their vacancies are visible on the site;
- stores each publication in `jobhub.telegram_publications` to prevent duplicates;
- keeps only the oldest public vacancy when title, company, location and description match exactly after basic normalization;
- retries confirmed failures after at least one minute, up to `TELEGRAM_PUBLISH_MAX_ATTEMPTS` for permanent failures; rate limits and temporary bot permission failures do not consume that limit;
- keeps uncertain sends in `pending` for manual reconciliation instead of risking a duplicate;
- supports `--dry-run` so local runs print messages without sending to Telegram or marking rows as sent.

Configuration:

- `TELEGRAM_BOT_TOKEN`: bot token, server-side only.
- `TELEGRAM_PUBLIC_CHANNEL_ID`: negative numeric channel ID, for example `-100...`.
- `TELEGRAM_PUBLIC_CHANNEL_USERNAME`: required public username in live mode; the bot verifies it matches the configured channel ID.
- `TELEGRAM_PUBLISHER_ENABLED`: set `true` to allow real sending.
- `TELEGRAM_PUBLISHER_DRY_RUN`: set `true` to print messages only.
- `TELEGRAM_PUBLISH_INTERVAL`: loop interval, default `60s`.
- `TELEGRAM_PUBLISH_BATCH_SIZE`: vacancies per iteration, default `10`.
- `TELEGRAM_PUBLISH_MAX_ATTEMPTS`: retry limit, default `3`.
- `JOBHUB_PUBLIC_BASE_URL`: required public HTTPS frontend base URL used for `/jobs/:id` links in live mode.
- `TELEGRAM_PUBLISH_WITHOUT_LINK`: temporary mode for a channel test before the site has a public URL; posts contain vacancy details without a broken localhost link, and URLs in the description are removed.
- `TELEGRAM_PUBLISH_TEST_SOURCES`: explicit staging-only opt-in to publish public imported vacancies whose production reuse permissions are not confirmed. Keep `false` for normal operation.
- `TELEGRAM_PUBLISH_TEST_CHANNEL_ID`: must match `TELEGRAM_PUBLIC_CHANNEL_ID` when test sources are enabled.

Linkless posts are best limited to a test channel: candidates cannot open the JobHub vacancy from those messages, and adding a public URL later does not automatically edit posts already sent.

Local dry run:

Set `TELEGRAM_DEV_DATABASE_URL` to a loopback PostgreSQL URL for a `jobhub_*_poc_*` development database. The dry run refuses staging and production URLs even if `.env` points at them.

```sh
make telegram-publisher-dev
```

Supabase one-shot backfill (no Docker):

```sh
set -a
. ./.env
set +a
cd backend
TELEGRAM_PUBLISHER_ENABLED=true TELEGRAM_PUBLISHER_DRY_RUN=false go run ./cmd/telegram-publisher --once
```

Omit `--once` to keep polling for new eligible vacancies. For an explicitly authorized staging test, set `TELEGRAM_PUBLISH_TEST_SOURCES=true` and `TELEGRAM_PUBLISH_TEST_CHANNEL_ID` to the chosen test channel ID. This does not change source production permissions and is rejected outside staging. Linkless posts use `TELEGRAM_PUBLISH_WITHOUT_LINK=true`.

Operational notes:

- Give the bot only the channel permission needed to post messages.
- The bot must be an administrator of the configured public channel with `can_post_messages`; startup verifies the username, numeric ID and permission before claiming vacancies.
- Existing eligible vacancies are posted oldest first in batches when publishing starts. The unique `(job_id, channel_id)` key prevents repeats across restarts and concurrent workers; the canonical-vacancy filter also avoids identical content stored under multiple job IDs.
- A `pending` record after an unknown Telegram outcome must be checked against the channel before an operator marks it sent or failed. Telegram's Bot API does not give this publisher an idempotency key for `sendMessage`.
- Keep the channel JobHub-owned.
- Do not expose `TELEGRAM_BOT_TOKEN` through any `VITE_*` variable.
- Run migrations before starting the publisher so `telegram_publications` exists.
- With `TELEGRAM_PUBLISHER_ENABLED=true`, the API process starts the publisher after checking migration 000014. External Compose passes this setting to the backend; no separate worker container is required.
- Local development keeps publishing disabled by default, even when the shared `.env` enables staging publishing. Set `TELEGRAM_LOCAL_PUBLISHER_ENABLED=true` only when intentionally testing live posts from the local Compose backend.
- Publisher claims one vacancy at a time and spaces messages by at least 1.1 seconds. Telegram 429 responses stop the batch and honor `retry_after` when it exceeds the normal loop interval.
