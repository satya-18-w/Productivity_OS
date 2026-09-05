-- name: CreateHabit :one
INSERT INTO habits (account_id, name, category_id, target)
VALUES ($1, $2, $3, $4)
RETURNING id, name, archived_at, category_id, target, created_at;

-- name: SetHabitArchived :execrows
UPDATE habits
SET archived_at = sqlc.arg(archived_at)
WHERE account_id = $1 AND id = $2;

-- name: SetHabitCategory :execrows
UPDATE habits
SET category_id = $3
WHERE account_id = $1 AND id = $2;

-- name: UpdateHabitFields :one
-- Full replace of name + target (MX3 Phase 1) — the habit's only other editable
-- fields; category has its own endpoint (SetHabitCategory, ADR-0009).
UPDATE habits
SET name = $3, target = $4
WHERE account_id = $1 AND id = $2
RETURNING id, name, archived_at, category_id, target, created_at;

-- name: HabitBelongsToAccount :one
SELECT count(*)
FROM habits
WHERE account_id = $1 AND id = $2;

-- name: ListActiveHabits :many
SELECT id, name, archived_at, category_id, target, created_at
FROM habits
WHERE account_id = $1 AND archived_at IS NULL
ORDER BY lower(name), created_at;

-- name: ListArchivedHabits :many
SELECT id, name, archived_at, category_id, target, created_at
FROM habits
WHERE account_id = $1 AND archived_at IS NOT NULL
ORDER BY lower(name), created_at;

-- name: MarkCompletion :exec
INSERT INTO habit_completions (habit_id, account_id, on_date)
VALUES ($1, $2, $3)
ON CONFLICT (habit_id, on_date) DO NOTHING;

-- name: UnmarkCompletion :exec
DELETE FROM habit_completions
WHERE habit_id = $1 AND on_date = $2;

-- name: ListCompletionDatesSince :many
SELECT on_date
FROM habit_completions
WHERE habit_id = $1 AND on_date >= sqlc.arg(since)
ORDER BY on_date;

-- name: CountHabitsByCategory :many
-- Active habits only — archived ones are already hidden from the habits list, so
-- they should not inflate a category's shown count.
SELECT category_id, count(*) AS total
FROM habits
WHERE account_id = $1 AND category_id IS NOT NULL AND archived_at IS NULL
GROUP BY category_id;

-- name: ListAllHabits :many
-- Active and archived, raw records, for M8 export completeness.
SELECT id, name, archived_at, category_id, target, created_at
FROM habits
WHERE account_id = $1
ORDER BY lower(name), created_at;

-- name: ListAllCompletions :many
-- Every completion of every habit (active and archived), for M8 export
-- completeness.
SELECT habit_id, on_date
FROM habit_completions
WHERE account_id = $1
ORDER BY habit_id, on_date;

-- name: CompletionCountsInRange :many
-- Active and archived habits alike, so a past week's history is complete even for
-- a habit archived since (v1.md, M6 decision). A habit with zero completions in
-- the range still appears, with total = 0.
SELECT h.id AS habit_id, h.name,
       count(hc.id) FILTER (WHERE hc.on_date BETWEEN sqlc.arg(from_date) AND sqlc.arg(to_date)) AS total
FROM habits h
LEFT JOIN habit_completions hc ON hc.habit_id = h.id
WHERE h.account_id = $1
GROUP BY h.id, h.name
ORDER BY lower(h.name);

-- name: HabitHistory :many
-- One row per (habit, completion-in-range) pair; a habit with no completions in
-- range still appears once, with on_date NULL (R2, docs/left.md Phase 6 heatmap).
SELECT h.id AS habit_id, h.name, h.archived_at, hc.on_date
FROM habits h
LEFT JOIN habit_completions hc
    ON hc.habit_id = h.id AND hc.on_date BETWEEN sqlc.arg(from_date) AND sqlc.arg(to_date)
WHERE h.account_id = $1
ORDER BY lower(h.name), hc.on_date;
