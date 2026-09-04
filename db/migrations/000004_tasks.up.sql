-- M3 Phase 1: tasks and the fixed four-column board (v1.md §7, §8).

CREATE TABLE tasks (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    title       text        NOT NULL,
    description text,
    due_date    date,
    state       text        NOT NULL CHECK (state IN ('BACKLOG', 'TODO', 'IN_PROGRESS', 'DONE')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Every state change, including creation (from_state NULL). Drives "tasks that
-- entered DONE in a range" (reviews, reports).
CREATE TABLE task_transitions (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    uuid        NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    account_id uuid        NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    from_state text,
    to_state   text        NOT NULL,
    at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tasks_account_created_idx ON tasks (account_id, created_at DESC);
CREATE INDEX task_transitions_account_state_at_idx ON task_transitions (account_id, to_state, at);
