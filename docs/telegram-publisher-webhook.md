# Telegram publisher on Supabase

The Vercel backend exposes `POST /api/v1/internal/telegram-publish`. It is
available only when `TELEGRAM_WEBHOOK_SECRET` is configured, and requires the
matching `X-JobHub-Publisher-Secret` header. It runs the existing publisher for
one vacancy per request. `telegram_publications` remains the source of truth
for deduplication, Telegram message IDs, attempts, and unknown outcomes.

## Production setup

1. In Vercel **Production only**, set `TELEGRAM_WEBHOOK_SECRET` to a random
   32-character-or-longer value. Also set `TELEGRAM_BOT_TOKEN`,
   `TELEGRAM_PUBLIC_CHANNEL_ID`, `TELEGRAM_PUBLIC_CHANNEL_USERNAME`, and
   `JOBHUB_PUBLIC_BASE_URL=https://jobhub-lovat.vercel.app`. Do not set
   `TELEGRAM_PUBLISHER_DRY_RUN=true` or `TELEGRAM_PUBLISH_WITHOUT_LINK=true`.
   Redeploy and confirm the normal site remains healthy.
2. In **JobHub Production** Supabase Vault, create named secrets
   `jobhub_telegram_publisher_url` with value
   `https://jobhub-lovat.vercel.app/api/v1/internal/telegram-publish` and
   `jobhub_telegram_publisher_secret` with the same random secret as Vercel.
   Do not put the secret in SQL history or the repository.
3. In that project's SQL Editor, run
   `scripts/supabase-telegram-publisher.sql`. This enables `pg_net` and
   `pg_cron`, creates an asynchronous trigger for newly public jobs, and
   schedules a one-minute reconciliation for missed notifications and retries.
   Run this only after the endpoint and Vault secrets are ready.
4. Verify a new public native vacancy: inspect its site detail, the `sent`
   row and `telegram_message_id` in `jobhub.telegram_publications`, and its
   channel link. Confirm a later reconciliation does not produce another post.

The webhook sends no vacancy payload. The backend queries the same approved,
public, permitted-source eligibility rules as the existing publisher. Cron
is a fallback, not a second independent publication path. A pending row with
an unknown Telegram outcome is deliberately not retried automatically, since
Telegram may have accepted the message; reconcile such a row manually.

For diagnostics, inspect Supabase Cron job history, `net._http_response`,
Vercel backend logs, and `jobhub.telegram_publications`. Do not print secrets
or full request headers in logs. The old GitHub Actions publisher should be
disabled only after this path has been proven in production.
