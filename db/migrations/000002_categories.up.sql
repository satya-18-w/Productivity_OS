-- M2 Phase 2: flat, user-defined category labels for time blocks (v1.md §2).

CREATE TABLE categories (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    name        text        NOT NULL,
    archived_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- One active category per name per account (case-insensitive). Archived rows do
-- not block a new active category of the same name.
CREATE UNIQUE INDEX categories_active_name_uniq
    ON categories (account_id, lower(name))
    WHERE archived_at IS NULL;

CREATE INDEX categories_account_id_idx ON categories (account_id);
