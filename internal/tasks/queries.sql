-- name: CreateTask :one
INSERT INTO tasks (account_id, title, description, due_date, state, category_id, goal_id, priority)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, title, description, due_date, state, category_id, goal_id, priority, created_at, updated_at;

-- name: RecordTransition :exec
INSERT INTO task_transitions (task_id, account_id, from_state, to_state)
VALUES ($1, $2, $3, $4);

-- name: GetTaskState :one
SELECT state
FROM tasks
WHERE account_id = $1 AND id = $2;

-- name: UpdateTaskFields :execrows
UPDATE tasks
SET title = $3, description = $4, due_date = $5, category_id = $6, goal_id = $7, priority = $8, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: UpdateTaskState :execrows
UPDATE tasks
SET state = $3, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE account_id = $1 AND id = $2;

-- name: ListTasks :many
SELECT id, title, description, due_date, state, category_id, goal_id, priority, created_at, updated_at
FROM tasks
WHERE account_id = $1
ORDER BY created_at DESC, id;

-- name: CountTasksByCategory :many
SELECT category_id, count(*) AS total
FROM tasks
WHERE account_id = $1 AND category_id IS NOT NULL
GROUP BY category_id;

-- name: DoneTaskCountInRange :one
SELECT count(DISTINCT task_id)
FROM task_transitions
WHERE account_id = $1
  AND to_state = 'DONE'
  AND at >= sqlc.arg(from_instant)
  AND at < sqlc.arg(to_instant);

-- name: ProgressByGoal :many
SELECT goal_id, count(*) AS total, count(*) FILTER (WHERE state = 'DONE') AS done
FROM tasks
WHERE account_id = $1 AND goal_id IS NOT NULL
GROUP BY goal_id;

-- name: CountAssignableTask :one
SELECT count(*)
FROM tasks
WHERE account_id = $1 AND id = $2;

-- name: TaskCategories :many
-- Every task's own category, for time_blocks that inherit a category from a linked
-- task (MX-TL). A task with no category is still listed, with a null category_id.
SELECT id, category_id
FROM tasks
WHERE account_id = $1;
