-- name: CreateGoal :one
INSERT INTO goals (account_id, title, description, target_date)
VALUES ($1, $2, $3, $4)
RETURNING id, title, description, target_date, progress, created_at, updated_at;

-- name: UpdateGoalFields :execrows
UPDATE goals
SET title = $3, description = $4, target_date = $5, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: UpdateGoalProgress :execrows
UPDATE goals
SET progress = $3, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: DeleteGoal :execrows
DELETE FROM goals WHERE account_id = $1 AND id = $2;

-- name: ListGoals :many
SELECT id, title, description, target_date, progress, created_at, updated_at
FROM goals
WHERE account_id = $1
ORDER BY created_at DESC, id;
