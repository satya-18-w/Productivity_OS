-- name: CreateCategory :one
INSERT INTO categories (account_id, name, colour, icon)
VALUES ($1, $2, $3, $4)
RETURNING id, name, COALESCE(colour, '')::text AS colour, COALESCE(icon, '')::text AS icon, archived_at, created_at;

-- name: UpdateCategory :execrows
UPDATE categories
SET name = $3, colour = $4, icon = $5
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: ArchiveCategory :execrows
UPDATE categories
SET archived_at = now()
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: ListActiveCategories :many
SELECT id, name, COALESCE(colour, '')::text AS colour, COALESCE(icon, '')::text AS icon, archived_at, created_at
FROM categories
WHERE account_id = $1 AND archived_at IS NULL
ORDER BY lower(name);

-- name: GetActiveCategory :one
-- Read-before-write for the partial Update (R3) — merges a caller's provided
-- fields onto the current row before the full-replace UpdateCategory write below.
SELECT id, name, COALESCE(colour, '')::text AS colour, COALESCE(icon, '')::text AS icon, archived_at, created_at
FROM categories
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: CountAssignableCategory :one
SELECT count(*)
FROM categories
WHERE account_id = $1 AND id = $2 AND archived_at IS NULL;

-- name: ListCategoryNames :many
SELECT id, name
FROM categories
WHERE account_id = $1;

-- name: ListAllCategories :many
-- Active and archived, for M8 export completeness.
SELECT id, name, COALESCE(colour, '')::text AS colour, COALESCE(icon, '')::text AS icon, archived_at, created_at
FROM categories
WHERE account_id = $1
ORDER BY lower(name);
