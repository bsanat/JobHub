CREATE TABLE jobhub.account_onboarding (
    user_id uuid PRIMARY KEY REFERENCES jobhub.users(id) ON DELETE CASCADE,
    persona text NOT NULL DEFAULT '',
    persona_other text NOT NULL DEFAULT '',
    goal text NOT NULL DEFAULT '',
    city_id text NOT NULL DEFAULT '',
    profession text NOT NULL DEFAULT '',
    referral text NOT NULL DEFAULT '',
    referral_other text NOT NULL DEFAULT '',
    step integer NOT NULL DEFAULT 1 CHECK (step BETWEEN 1 AND 5),
    version bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (persona IN ('', 'school_student', 'university_student', 'graduate', 'unemployed', 'other')),
    CHECK (goal IN ('', 'job', 'internship', 'browsing')),
    CHECK (referral IN ('', 'telegram', 'instagram', 'google', 'friend', 'university', 'other', 'skip')),
    CHECK (char_length(persona_other) <= 200 AND char_length(referral_other) <= 200 AND char_length(profession) <= 200),
    CHECK (city_id = '' OR city_id IN ('astana', 'almaty', 'shymkent', 'karaganda', 'aktobe', 'taraz', 'pavlodar', 'oskemen', 'semey', 'atyrau', 'kostanay', 'kyzylorda', 'aktau', 'oral', 'petropavl', 'turkistan')),
    CHECK (completed_at IS NULL OR (
        step = 5 AND persona <> '' AND goal <> '' AND city_id <> '' AND length(btrim(profession)) > 0
        AND (persona <> 'other' OR length(btrim(persona_other)) > 0)
        AND (referral <> 'other' OR length(btrim(referral_other)) > 0)
    ))
);

ALTER TABLE jobhub.account_onboarding ENABLE ROW LEVEL SECURITY;
COMMENT ON TABLE jobhub.account_onboarding IS 'Private registration questionnaire. Only new registrations get a row; existing accounts are not enrolled. Access is through the authenticated Go API only.';
