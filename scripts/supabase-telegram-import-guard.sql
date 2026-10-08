-- Apply once to an already configured Production webhook. This preserves the
-- native vacancy trigger and avoids one HTTP request per public API import row.
CREATE OR REPLACE FUNCTION jobhub.notify_telegram_on_publication()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM jobhub.job_sources s
        WHERE s.source = NEW.source
          AND s.provider IN ('arbeitnow', 'greenhouse', 'lever')
    ) THEN
        RETURN NEW;
    END IF;
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
