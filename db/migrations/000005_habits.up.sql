-- M4 Phase 1: daily habits and their completion history (v1.md §9).

CREATE TABLE habits (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    name        text        NOT NULL,
    archived_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- One row per (habit, calendar date) that is completed. Absence = not completed.
-- Archiving a habit never deletes these (Q11).
CREATE TABLE habit_completions (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id   uuid        NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    account_id uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    on_date    date        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    UNIQUE (habit_id, on_date)
);

CREATE INDEX habits_account_archived_idx ON habits (account_id, archived_at);
CREATE INDEX habit_completions_habit_date_idx ON habit_completions (habit_id, on_date);
