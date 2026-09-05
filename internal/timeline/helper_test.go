package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// fakeZone returns a fixed location for every account.
type fakeZone struct{ loc *time.Location }

func (f fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) {
	if f.loc == nil {
		return time.UTC, nil
	}
	return f.loc, nil
}

func newTasksSvc(pool *pgxpool.Pool) tasks.Service {
	catSvc := categories.NewService(pool)
	goalsSvc := goals.NewService(pool, catSvc)
	return tasks.NewService(pool, catSvc, goalsSvc)
}

func setup(t *testing.T) (timeline.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	svc := timeline.NewService(pool, fakeZone{}, categories.NewService(pool), newTasksSvc(pool))
	return svc, pool, newAccount(t, pool, "owner@test")
}

func setupZone(t *testing.T, loc *time.Location) (timeline.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	svc := timeline.NewService(pool, fakeZone{loc: loc}, categories.NewService(pool), newTasksSvc(pool))
	return svc, pool, newAccount(t, pool, "owner@test")
}

// mkTask inserts a task directly (timeline tests only need one to exist and be
// assignable — task CRUD is exercised in the tasks package).
func mkTask(t *testing.T, pool *pgxpool.Pool, acc uuid.UUID, title string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO tasks (account_id, title, state) VALUES ($1, $2, 'BACKLOG') RETURNING id`, acc, title).Scan(&id))
	return id
}

// mkTaskWithCategory inserts a task with a category directly, for testing
// category inheritance on a task-linked block.
func mkTaskWithCategory(t *testing.T, pool *pgxpool.Pool, acc, category uuid.UUID, title string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO tasks (account_id, title, state, category_id) VALUES ($1, $2, 'BACKLOG', $3) RETURNING id`,
		acc, title, category).Scan(&id))
	return id
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

// mkCategory inserts an active category directly (timeline tests only need one to
// exist and be assignable — category CRUD is exercised in the categories package).
func mkCategory(t *testing.T, pool *pgxpool.Pool, acc uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, $2) RETURNING id`, acc, name).Scan(&id))
	return id
}

func archiveCategory(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE categories SET archived_at = now() WHERE id = $1`, id)
	require.NoError(t, err)
}
