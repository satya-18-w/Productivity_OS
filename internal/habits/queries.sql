-- name: CreateHabit :one
INSERT INTO habits (account_id, name)
VALUES ($1, $2)
RETURNING id, name, archived_at, created_at;

-- name: SetHabitArchived :execrows
UPDATE habits
SET archived_at = sqlc.arg(archived_at)
WHERE account_id = $1 AND id = $2;

-- name: HabitBelongsToAccount :one
SELECT count(*)
FROM habits
WHERE account_id = $1 AND id = $2;

-- name: ListActiveHabits :many
SELECT id, name, archived_at, created_at
FROM habits
WHERE account_id = $1 AND archived_at IS NULL
ORDER BY lower(name), created_at;

-- name: ListArchivedHabits :many
SELECT id, name, archived_at, created_at
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
