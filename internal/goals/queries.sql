-- name: CreateGoal :one
INSERT INTO goals (account_id, title, description, target_date, category_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, target_date, progress, category_id, created_at, updated_at;

-- name: UpdateGoalFields :execrows
UPDATE goals
SET title = $3, description = $4, target_date = $5, category_id = $6, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: UpdateGoalProgress :execrows
UPDATE goals
SET progress = $3, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: DeleteGoal :execrows
DELETE FROM goals WHERE account_id = $1 AND id = $2;

-- name: ListGoals :many
SELECT id, title, description, target_date, progress, category_id, created_at, updated_at
FROM goals
WHERE account_id = $1
ORDER BY created_at DESC, id;

-- name: CountGoalsByCategory :many
SELECT category_id, count(*) AS total
FROM goals
WHERE account_id = $1 AND category_id IS NOT NULL
GROUP BY category_id;

-- name: CountAssignableGoal :one
SELECT count(*)
FROM goals
WHERE account_id = $1 AND id = $2;
