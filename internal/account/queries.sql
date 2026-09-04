-- name: CreateAccount :one
INSERT INTO accounts (email, password_hash, timezone)
VALUES ($1, $2, $3)
RETURNING id, email, timezone, created_at;

-- name: GetAccountByEmail :one
SELECT id, email, password_hash, timezone, created_at
FROM accounts
WHERE email = $1;

-- name: GetAccountByID :one
SELECT id, email, timezone, created_at
FROM accounts
WHERE id = $1;

-- name: GetAccountPasswordHash :one
SELECT password_hash
FROM accounts
WHERE id = $1;

-- name: UpdateAccountTimezone :exec
UPDATE accounts SET timezone = $2 WHERE id = $1;

-- name: UpdateAccountPassword :exec
UPDATE accounts SET password_hash = $2 WHERE id = $1;

-- name: CreateSession :one
INSERT INTO sessions (token, account_id, expires_at)
VALUES ($1, $2, $3)
RETURNING token, account_id, created_at, expires_at, last_seen_at;

-- name: GetSession :one
SELECT token, account_id, created_at, expires_at, last_seen_at
FROM sessions
WHERE token = $1;

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = now() WHERE token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;

-- name: DeleteAccountSessions :exec
DELETE FROM sessions WHERE account_id = $1;
