package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/timeline"
)

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestAddBlock_PlannedAndActual(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	p, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, timeline.Planned, p.Kind)
	require.Nil(t, p.CategoryID)

	a, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:30:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, timeline.Actual, a.Kind)
	require.NotEqual(t, p.ID, a.ID)
}

func TestAddBlock_MidnightSpanning(t *testing.T) {
	svc, _, acc := setup(t)
	b, err := svc.AddBlock(context.Background(), acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T23:00:00Z"), EndsAt: ts("2025-06-16T02:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 3*time.Hour, b.EndsAt.Sub(b.StartsAt))
}

func TestAddBlock_Validation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: "sideways", StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	requireField(t, err, "kind")

	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T10:00:00Z"), EndsAt: ts("2025-06-15T09:00:00Z"),
	})
	requireField(t, err, "end")

	_, err = svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T10:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	requireField(t, err, "end")
}

func TestAddBlock_CategoryMustBeOwnedAndActive(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-cat@test")

	mine, _ := svc.CreateCategory(ctx, acc, "Deep Work")
	theirs, _ := svc.CreateCategory(ctx, other, "Theirs")
	archived, _ := svc.CreateCategory(ctx, acc, "Old")
	require.NoError(t, svc.ArchiveCategory(ctx, acc, archived.ID))

	in := func(cat uuid.UUID) timeline.BlockInput {
		return timeline.BlockInput{
			Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
			CategoryID: &cat,
		}
	}

	b, err := svc.AddBlock(ctx, acc, in(mine.ID))
	require.NoError(t, err)
	require.Equal(t, mine.ID, *b.CategoryID)

	_, err = svc.AddBlock(ctx, acc, in(theirs.ID))
	requireField(t, err, "category_id")

	_, err = svc.AddBlock(ctx, acc, in(archived.ID))
	requireField(t, err, "category_id")

	_, err = svc.AddBlock(ctx, acc, in(uuid.New()))
	requireField(t, err, "category_id")
}

func TestEditBlock(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	b, _ := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})
	cat, _ := svc.CreateCategory(ctx, acc, "Reading")

	require.NoError(t, svc.EditBlock(ctx, acc, b.ID, ts("2025-06-15T09:00:00Z"), ts("2025-06-15T11:30:00Z"), &cat.ID))

	requireField(t,
		svc.EditBlock(ctx, acc, b.ID, ts("2025-06-15T11:00:00Z"), ts("2025-06-15T10:00:00Z"), nil), "end")

	require.ErrorIs(t, svc.EditBlock(ctx, acc, uuid.New(),
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil), timeline.ErrBlockNotFound)

	// isolation
	other := newAccount(t, pool, "edit-iso@test")
	require.ErrorIs(t, svc.EditBlock(ctx, other, b.ID,
		ts("2025-06-15T09:00:00Z"), ts("2025-06-15T10:00:00Z"), nil), timeline.ErrBlockNotFound)
}

func TestDeleteBlock(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()

	b, _ := svc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"),
	})

	other := newAccount(t, pool, "del-iso@test")
	require.ErrorIs(t, svc.DeleteBlock(ctx, other, b.ID), timeline.ErrBlockNotFound)

	require.NoError(t, svc.DeleteBlock(ctx, acc, b.ID))
	require.ErrorIs(t, svc.DeleteBlock(ctx, acc, b.ID), timeline.ErrBlockNotFound)
}

func requireField(t *testing.T, err error, field string) {
	t.Helper()
	var verr *timeline.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, field)
}
