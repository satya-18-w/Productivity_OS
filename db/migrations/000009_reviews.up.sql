-- M6 Phase 2: daily and weekly reviews (v1.md §11, §12). Answers are a whole
-- {prompt_key: text} map, saved and loaded together — never queried per-key.

CREATE TABLE daily_reviews (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    on_date     date        NOT NULL,
    answers     jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT daily_reviews_account_date_uniq UNIQUE (account_id, on_date)
);

CREATE TABLE weekly_reviews (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    iso_year    integer     NOT NULL,
    iso_week    integer     NOT NULL CHECK (iso_week BETWEEN 1 AND 53),
    answers     jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT weekly_reviews_account_week_uniq UNIQUE (account_id, iso_year, iso_week)
);

CREATE INDEX daily_reviews_account_idx ON daily_reviews (account_id);
CREATE INDEX weekly_reviews_account_idx ON weekly_reviews (account_id);
