-- M2 Phase 3: planned and actual time blocks (v1.md §3, §4). A block has a start
-- instant, an end instant strictly after it, an optional category, and a fixed
-- kind. Blocks may overlap and may span midnight.

CREATE TABLE time_blocks (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    kind        text        NOT NULL CHECK (kind IN ('planned', 'actual')),
    starts_at   timestamptz NOT NULL,
    ends_at     timestamptz NOT NULL,
    category_id uuid        REFERENCES categories (id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT time_blocks_end_after_start CHECK (ends_at > starts_at)
);

CREATE INDEX time_blocks_account_starts_idx ON time_blocks (account_id, starts_at);
CREATE INDEX time_blocks_category_idx ON time_blocks (category_id);
