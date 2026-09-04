-- name: CreateCategory :one
INSERT INTO categories (account_id, name)
VALUES ($1, $2)
RETURNING id, name, archived_at, created_at;

-- name: RenameCategory :execrows
UPDATE categories
SET name = $3
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: ArchiveCategory :execrows
UPDATE categories
SET archived_at = now()
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: ListActiveCategories :many
SELECT id, name, archived_at, created_at
FROM categories
WHERE account_id = $1 AND archived_at IS NULL
ORDER BY lower(name);

-- name: GetCategory :one
SELECT id, name, archived_at, created_at
FROM categories
WHERE account_id = $1 AND id = $2;

-- name: CountAssignableCategory :one
SELECT count(*)
FROM categories
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: CreateTimeBlock :one
INSERT INTO time_blocks (account_id, kind, starts_at, ends_at, category_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, kind, starts_at, ends_at, category_id, created_at;

-- name: UpdateTimeBlock :execrows
UPDATE time_blocks
SET starts_at = $3, ends_at = $4, category_id = $5
WHERE account_id = $1 AND id = $2;

-- name: DeleteTimeBlock :execrows
DELETE FROM time_blocks
WHERE account_id = $1 AND id = $2;

-- name: GetTimeBlock :one
SELECT id, kind, starts_at, ends_at, category_id, created_at
FROM time_blocks
WHERE account_id = $1 AND id = $2;

-- name: ListBlocksOverlapping :many
SELECT b.id, b.kind, b.starts_at, b.ends_at, b.category_id, c.name AS category_name
FROM time_blocks b
LEFT JOIN categories c ON c.id = b.category_id
WHERE b.account_id = $1
  AND b.starts_at < sqlc.arg(window_end)
  AND b.ends_at > sqlc.arg(window_start)
ORDER BY b.starts_at, b.ends_at;
