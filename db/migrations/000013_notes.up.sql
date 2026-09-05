-- MX4: notes — plain-text title + body free-form capture. No tags, no
-- pin/favourite/archive/trash, no category or other linkage (v1.md §15).

CREATE TABLE notes (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    title       text        NOT NULL,
    body        text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX notes_account_created_idx ON notes (account_id, created_at DESC);
