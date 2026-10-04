-- Run once in the JobHub Production Supabase SQL Editor after the Vercel
-- endpoint and both Vault secrets described in docs/telegram-publisher-webhook.md exist.
CREATE SCHEMA IF NOT EXISTS extensions;
CREATE EXTENSION IF NOT EXISTS pg_net WITH SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS pg_cron;

CREATE OR REPLACE FUNCTION jobhub.request_telegram_publication()
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    endpoint text;
    token text;
BEGIN
    SELECT decrypted_secret INTO endpoint
    FROM vault.decrypted_secrets
    WHERE name = 'jobhub_telegram_publisher_url';

    SELECT decrypted_secret INTO token
    FROM vault.decrypted_secrets
    WHERE name = 'jobhub_telegram_publisher_secret';

    IF endpoint IS NULL OR endpoint NOT LIKE 'https://%'
       OR token IS NULL OR length(token) < 32 THEN
        RAISE EXCEPTION 'JobHub Telegram publisher Vault configuration is missing';
    END IF;

    PERFORM net.http_post(
        url := endpoint,
        headers := jsonb_build_object(
            'Content-Type', 'application/json',
            'X-JobHub-Publisher-Secret', token
        ),
        body := '{}'::jsonb,
        timeout_milliseconds := 20000
    );
END;
$$;

REVOKE ALL ON FUNCTION jobhub.request_telegram_publication() FROM PUBLIC;

CREATE OR REPLACE FUNCTION jobhub.notify_telegram_on_publication()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
    IF NOT jobhub.job_is_public(NEW) THEN
        RETURN NEW;
    END IF;
    IF TG_OP = 'UPDATE' THEN
        IF jobhub.job_is_public(OLD) THEN
            RETURN NEW;
        END IF;
    END IF;

    PERFORM jobhub.request_telegram_publication();
    RETURN NEW;
EXCEPTION WHEN OTHERS THEN
    RAISE WARNING 'JobHub Telegram notification could not be queued (SQLSTATE %)', SQLSTATE;
    RETURN NEW;
END;
$$;

REVOKE ALL ON FUNCTION jobhub.notify_telegram_on_publication() FROM PUBLIC;

DROP TRIGGER IF EXISTS notify_telegram_on_publication ON jobhub.jobs;
CREATE TRIGGER notify_telegram_on_publication
AFTER INSERT OR UPDATE OF publication_status, moderation_status, source_status, fresh_until
ON jobhub.jobs
FOR EACH ROW
EXECUTE FUNCTION jobhub.notify_telegram_on_publication();

SELECT cron.schedule(
    'jobhub-telegram-publisher-retry',
    '* * * * *',
    'SELECT jobhub.request_telegram_publication()'
);
