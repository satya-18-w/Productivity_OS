-- name: CreateNote :one
INSERT INTO notes (account_id, title, body)
VALUES ($1, $2, $3)
RETURNING id, title, body, created_at, updated_at;

-- name: UpdateNoteFields :execrows
UPDATE notes
SET title = $3, body = $4, updated_at = now()
WHERE account_id = $1 AND id = $2;

-- name: DeleteNote :execrows
DELETE FROM notes WHERE account_id = $1 AND id = $2;

-- name: ListNotes :many
SELECT id, title, body, created_at, updated_at
FROM notes
WHERE account_id = $1
ORDER BY created_at DESC, id;
