package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
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

func setup(t *testing.T) (timeline.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return timeline.NewService(pool, fakeZone{}), pool, newAccount(t, pool, "owner@test")
}

func setupZone(t *testing.T, loc *time.Location) (timeline.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return timeline.NewService(pool, fakeZone{loc: loc}), pool, newAccount(t, pool, "owner@test")
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func names(cats []timeline.Category) []string {
	out := make([]string, len(cats))
	for i, c := range cats {
		out[i] = c.Name
	}
	return out
}

func TestCreateAndList(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.CreateCategory(ctx, acc, "  Gym  ")
	require.NoError(t, err)
	_, err = svc.CreateCategory(ctx, acc, "DSA")
	require.NoError(t, err)

	cats, err := svc.ListActiveCategories(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, []string{"DSA", "Gym"}, names(cats), "name-ordered, trimmed")
}

func TestCreateCategory_DuplicateActiveNameCaseInsensitive(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.CreateCategory(ctx, acc, "Gym")
	require.NoError(t, err)
	_, err = svc.CreateCategory(ctx, acc, "  gym ")
	require.ErrorIs(t, err, timeline.ErrCategoryNameTaken)
}

func TestCreateCategory_ArchivedNameDoesNotBlock(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, err := svc.CreateCategory(ctx, acc, "Gym")
	require.NoError(t, err)
	require.NoError(t, svc.ArchiveCategory(ctx, acc, c.ID))

	_, err = svc.CreateCategory(ctx, acc, "Gym")
	require.NoError(t, err, "an archived name is free to reuse")
}

func TestCreateCategory_Validation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	for _, name := range []string{"", "   ", string(make([]byte, 61))} {
		_, err := svc.CreateCategory(ctx, acc, name)
		var verr *timeline.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, "name")
	}
}

func TestRenameCategory(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, _ := svc.CreateCategory(ctx, acc, "Gm")
	require.NoError(t, svc.RenameCategory(ctx, acc, c.ID, "Gym"))

	cats, _ := svc.ListActiveCategories(ctx, acc)
	require.Equal(t, []string{"Gym"}, names(cats))

	// collision
	_, _ = svc.CreateCategory(ctx, acc, "DSA")
	require.ErrorIs(t, svc.RenameCategory(ctx, acc, c.ID, "dsa"), timeline.ErrCategoryNameTaken)

	// unknown id
	require.ErrorIs(t, svc.RenameCategory(ctx, acc, uuid.New(), "Whatever"), timeline.ErrCategoryNotFound)
}

func TestArchiveCategory(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, _ := svc.CreateCategory(ctx, acc, "Gym")
	require.NoError(t, svc.ArchiveCategory(ctx, acc, c.ID))

	cats, _ := svc.ListActiveCategories(ctx, acc)
	require.Empty(t, cats)

	// archiving again -> not found (already archived)
	require.ErrorIs(t, svc.ArchiveCategory(ctx, acc, c.ID), timeline.ErrCategoryNotFound)
	// renaming an archived one -> not found
	require.ErrorIs(t, svc.RenameCategory(ctx, acc, c.ID, "X"), timeline.ErrCategoryNotFound)
}

func TestCategoryIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	ca, _ := svc.CreateCategory(ctx, a, "A-only")
	_, _ = svc.CreateCategory(ctx, b, "B-only")

	// B cannot see, rename, or archive A's category.
	bCats, _ := svc.ListActiveCategories(ctx, b)
	require.Equal(t, []string{"B-only"}, names(bCats))

	require.ErrorIs(t, svc.RenameCategory(ctx, b, ca.ID, "hijacked"), timeline.ErrCategoryNotFound)
	require.ErrorIs(t, svc.ArchiveCategory(ctx, b, ca.ID), timeline.ErrCategoryNotFound)

	// A's category is untouched.
	aCats, _ := svc.ListActiveCategories(ctx, a)
	require.Equal(t, []string{"A-only"}, names(aCats))
}
