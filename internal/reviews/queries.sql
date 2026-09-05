-- name: UpsertDailyReview :one
INSERT INTO daily_reviews (account_id, on_date, answers)
VALUES ($1, $2, $3)
ON CONFLICT (account_id, on_date) DO UPDATE
SET answers = EXCLUDED.answers, updated_at = now()
RETURNING updated_at;

-- name: GetDailyReview :one
SELECT answers, updated_at
FROM daily_reviews
WHERE account_id = $1 AND on_date = $2;

-- name: UpsertWeeklyReview :one
INSERT INTO weekly_reviews (account_id, iso_year, iso_week, answers)
VALUES ($1, $2, $3, $4)
ON CONFLICT (account_id, iso_year, iso_week) DO UPDATE
SET answers = EXCLUDED.answers, updated_at = now()
RETURNING updated_at;

-- name: GetWeeklyReview :one
SELECT answers, updated_at
FROM weekly_reviews
WHERE account_id = $1 AND iso_year = $2 AND iso_week = $3;

-- name: ListAllDailyReviews :many
-- Every saved daily review, for M8 export completeness.
SELECT on_date, answers, updated_at
FROM daily_reviews
WHERE account_id = $1
ORDER BY on_date;

-- name: ListAllWeeklyReviews :many
-- Every saved weekly review, for M8 export completeness.
SELECT iso_year, iso_week, answers, updated_at
FROM weekly_reviews
WHERE account_id = $1
ORDER BY iso_year, iso_week;
