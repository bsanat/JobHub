CREATE TABLE jobhub.telegram_publications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id uuid NOT NULL REFERENCES jobhub.jobs(id) ON DELETE CASCADE,
    channel_id bigint NOT NULL,
    telegram_message_id bigint,
    status text NOT NULL DEFAULT 'pending',
    error_code text,
    attempts integer NOT NULL DEFAULT 0,
    sent_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (job_id, channel_id),
    CONSTRAINT telegram_publications_status_valid CHECK (status IN ('pending', 'sent', 'failed')),
    CONSTRAINT telegram_publications_attempts_nonnegative CHECK (attempts >= 0),
    CONSTRAINT telegram_publications_sent_message_required CHECK (
        status <> 'sent' OR (telegram_message_id IS NOT NULL AND sent_at IS NOT NULL)
    )
);

ALTER TABLE jobhub.telegram_publications ENABLE ROW LEVEL SECURITY;

CREATE INDEX telegram_publications_status_updated_idx
    ON jobhub.telegram_publications (status, updated_at);

CREATE INDEX telegram_publications_channel_sent_idx
    ON jobhub.telegram_publications (channel_id, sent_at DESC)
    WHERE status = 'sent';
