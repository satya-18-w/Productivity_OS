-- M1 Phase 3: the account and session tables. Every later table hangs off accounts.

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE accounts (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         citext      NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    timezone      text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    token        text        PRIMARY KEY,
    account_id   uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_account_id_idx ON sessions (account_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
