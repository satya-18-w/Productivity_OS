-- name: CreateTask :one
INSERT INTO tasks (account_id, title, description, due_date, state)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, due_date, state, created_at, updated_at;

-- name: RecordTransition :exec
INSERT INTO task_transitions (task_id, account_id, from_state, to_state)
VALUES ($1, $2, $3, $4);

-- name: GetTaskState :one
SELECT state
FROM tasks
WHERE account_id = $1 AND id = $2;

-- name: UpdateTaskFields :execrows
UPDATE tasks
SET title = $3, description = $4, due_date = $5, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: UpdateTaskState :execrows
UPDATE tasks
SET state = $3, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE account_id = $1 AND id = $2;

-- name: ListTasks :many
SELECT id, title, description, due_date, state, created_at, updated_at
FROM tasks
WHERE account_id = $1
ORDER BY created_at DESC, id;
