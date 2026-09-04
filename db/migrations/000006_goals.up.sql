-- M5: goals — a title, optional description, optional target date, and a manually
-- set four-state progress. Not linked to any other entity (v1.md §10).

CREATE TABLE goals (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    title       text        NOT NULL,
    description text,
    target_date date,
    progress    text        NOT NULL DEFAULT 'NOT_STARTED'
                            CHECK (progress IN ('NOT_STARTED', 'IN_PROGRESS', 'ACHIEVED', 'ABANDONED')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX goals_account_created_idx ON goals (account_id, created_at DESC);
