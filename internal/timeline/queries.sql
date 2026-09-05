-- name: CreateTimeBlock :one
INSERT INTO time_blocks (account_id, kind, starts_at, ends_at, category_id, task_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, kind, starts_at, ends_at, category_id, task_id, created_at;

-- name: UpdateTimeBlock :execrows
UPDATE time_blocks
SET starts_at = $3, ends_at = $4, category_id = $5, task_id = $6
WHERE account_id = $1 AND id = $2;

-- name: DeleteTimeBlock :execrows
DELETE FROM time_blocks
WHERE account_id = $1 AND id = $2;

-- name: GetTimeBlock :one
SELECT id, kind, starts_at, ends_at, category_id, task_id, created_at
FROM time_blocks
WHERE account_id = $1 AND id = $2;

-- name: ListBlocksOverlapping :many
SELECT b.id, b.kind, b.starts_at, b.ends_at, b.category_id, b.task_id
FROM time_blocks b
WHERE b.account_id = $1
  AND b.starts_at < sqlc.arg(window_end)
  AND b.ends_at > sqlc.arg(window_start)
ORDER BY b.starts_at, b.ends_at;

-- name: CountBlocksByCategory :many
SELECT category_id, count(*) AS total
FROM time_blocks
WHERE account_id = $1 AND category_id IS NOT NULL
GROUP BY category_id;

-- name: CountBlocksByTask :many
-- A task-linked block's own category_id is always null (CHECK constraint); this
-- powers the inherited-category count so a task-linked block still contributes to
-- its (inherited) category's total in the categories overview (MX-TL).
SELECT task_id, count(*) AS total
FROM time_blocks
WHERE account_id = $1 AND task_id IS NOT NULL
GROUP BY task_id;

-- name: ListAllBlocks :many
-- Every planned and actual block, unbounded, for M8 export completeness.
SELECT id, kind, starts_at, ends_at, category_id, task_id, created_at
FROM time_blocks
WHERE account_id = $1
ORDER BY starts_at;

-- name: ListBlocksByTask :many
-- Every block (planned and actual, across any date) linked to one task —
-- v1.md §7's "see all of a task's linked time blocks" (reverse of task_id).
SELECT id, kind, starts_at, ends_at, category_id, task_id
FROM time_blocks
WHERE account_id = $1 AND task_id = $2
ORDER BY starts_at;
