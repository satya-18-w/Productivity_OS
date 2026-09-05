package reviews_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/reviews"
)

func setup(t *testing.T) (reviews.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return reviews.NewService(pool), pool, newAccount(t, pool, "owner@test")
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func TestGetDaily_NeverSavedReturnsBlank(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	rev, err := svc.GetDaily(ctx, acc, timezone.Date{Year: 2025, Month: 6, Day: 15})
	require.NoError(t, err)
	require.Empty(t, rev.Answers)
	require.Empty(t, rev.UpdatedAt, "never saved -> blank, not an error")
}

func TestSaveAndReloadDaily(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	date := timezone.Date{Year: 2025, Month: 6, Day: 15}

	saved, err := svc.SaveDaily(ctx, acc, date, reviews.Answers{"went_well": "Shipped M6"})
	require.NoError(t, err)
	require.Equal(t, "Shipped M6", saved.Answers["went_well"])
	require.NotEmpty(t, saved.UpdatedAt)

	reloaded, err := svc.GetDaily(ctx, acc, date)
	require.NoError(t, err)
	require.Equal(t, "Shipped M6", reloaded.Answers["went_well"])
}

func TestSaveDaily_UpsertOverwrites(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	date := timezone.Date{Year: 2025, Month: 6, Day: 15}

	_, err := svc.SaveDaily(ctx, acc, date, reviews.Answers{"went_well": "first draft"})
	require.NoError(t, err)
	_, err = svc.SaveDaily(ctx, acc, date, reviews.Answers{"went_well": "final answer"})
	require.NoError(t, err)

	rev, err := svc.GetDaily(ctx, acc, date)
	require.NoError(t, err)
	require.Equal(t, "final answer", rev.Answers["went_well"], "edit replaces, not appends")
	require.Len(t, rev.Answers, 1)
}

func TestSaveDaily_UnknownKeysDropped(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	date := timezone.Date{Year: 2025, Month: 6, Day: 15}

	saved, err := svc.SaveDaily(ctx, acc, date, reviews.Answers{
		"went_well": "ok", "not_a_real_prompt": "should vanish",
	})
	require.NoError(t, err)
	require.Contains(t, saved.Answers, "went_well")
	require.NotContains(t, saved.Answers, "not_a_real_prompt")

	reloaded, err := svc.GetDaily(ctx, acc, date)
	require.NoError(t, err)
	require.NotContains(t, reloaded.Answers, "not_a_real_prompt")
}

func TestSaveDaily_EmptyAnswersAllowed(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()
	date := timezone.Date{Year: 2025, Month: 6, Day: 15}

	saved, err := svc.SaveDaily(ctx, acc, date, reviews.Answers{})
	require.NoError(t, err)
	require.Empty(t, saved.Answers)
	require.NotEmpty(t, saved.UpdatedAt, "a partially-filled (empty) review still saves")
}

func TestDailyReview_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")
	date := timezone.Date{Year: 2025, Month: 6, Day: 15}

	_, err := svc.SaveDaily(ctx, a, date, reviews.Answers{"went_well": "A's answer"})
	require.NoError(t, err)

	rev, err := svc.GetDaily(ctx, b, date)
	require.NoError(t, err)
	require.Empty(t, rev.Answers, "B sees none of A's review")
}

func TestGetWeekly_NeverSavedReturnsBlank(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	rev, err := svc.GetWeekly(ctx, acc, 2025, 24)
	require.NoError(t, err)
	require.Empty(t, rev.Answers)
	require.Empty(t, rev.UpdatedAt)
}

func TestSaveAndReloadWeekly(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	saved, err := svc.SaveWeekly(ctx, acc, 2025, 24, reviews.Answers{"highlights": "Shipped M6"})
	require.NoError(t, err)
	require.Equal(t, "Shipped M6", saved.Answers["highlights"])

	reloaded, err := svc.GetWeekly(ctx, acc, 2025, 24)
	require.NoError(t, err)
	require.Equal(t, "Shipped M6", reloaded.Answers["highlights"])
}

func TestSaveWeekly_UpsertOverwritesAndUnknownKeysDropped(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.SaveWeekly(ctx, acc, 2025, 24, reviews.Answers{"highlights": "draft"})
	require.NoError(t, err)
	saved, err := svc.SaveWeekly(ctx, acc, 2025, 24, reviews.Answers{
		"highlights": "final", "not_a_real_prompt": "gone",
	})
	require.NoError(t, err)
	require.Equal(t, "final", saved.Answers["highlights"])
	require.NotContains(t, saved.Answers, "not_a_real_prompt")
	require.Len(t, saved.Answers, 1)
}

func TestWeekly_DifferentWeeksAreIndependent(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.SaveWeekly(ctx, acc, 2025, 1, reviews.Answers{"highlights": "week 1"})
	require.NoError(t, err)
	_, err = svc.SaveWeekly(ctx, acc, 2025, 2, reviews.Answers{"highlights": "week 2"})
	require.NoError(t, err)
	// same week number, different year, must not collide
	_, err = svc.SaveWeekly(ctx, acc, 2026, 1, reviews.Answers{"highlights": "next year week 1"})
	require.NoError(t, err)

	w1, _ := svc.GetWeekly(ctx, acc, 2025, 1)
	w2, _ := svc.GetWeekly(ctx, acc, 2025, 2)
	w1n, _ := svc.GetWeekly(ctx, acc, 2026, 1)
	require.Equal(t, "week 1", w1.Answers["highlights"])
	require.Equal(t, "week 2", w2.Answers["highlights"])
	require.Equal(t, "next year week 1", w1n.Answers["highlights"])
}

func TestWeekly_InvalidWeekNumber(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	for _, week := range []int{0, 54, -1} {
		_, err := svc.GetWeekly(ctx, acc, 2025, week)
		var verr *reviews.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, "iso_week")

		_, err = svc.SaveWeekly(ctx, acc, 2025, week, reviews.Answers{})
		require.ErrorAs(t, err, &verr)
	}
}

func TestListDailyAndListWeekly(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-listall@test")

	_, err := svc.SaveDaily(ctx, acc, timezone.Date{Year: 2025, Month: 6, Day: 15}, reviews.Answers{"went_well": "a"})
	require.NoError(t, err)
	_, err = svc.SaveDaily(ctx, acc, timezone.Date{Year: 2025, Month: 6, Day: 16}, reviews.Answers{"went_well": "b"})
	require.NoError(t, err)
	_, err = svc.SaveWeekly(ctx, acc, 2025, 24, reviews.Answers{"highlights": "week 24"})
	require.NoError(t, err)
	_, err = svc.SaveWeekly(ctx, acc, 2025, 25, reviews.Answers{"highlights": "week 25"})
	require.NoError(t, err)

	_, err = svc.SaveDaily(ctx, other, timezone.Date{Year: 2025, Month: 6, Day: 15}, reviews.Answers{"went_well": "not mine"})
	require.NoError(t, err)

	dailies, err := svc.ListDaily(ctx, acc)
	require.NoError(t, err)
	require.Len(t, dailies, 2, "both of the caller's, none of other's")

	weeklies, err := svc.ListWeekly(ctx, acc)
	require.NoError(t, err)
	require.Len(t, weeklies, 2)
	require.Equal(t, "week 24", weeklies[0].Answers["highlights"])
	require.Equal(t, "week 25", weeklies[1].Answers["highlights"])
}

func TestWeeklyReview_Isolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other-weekly@test")

	_, err := svc.SaveWeekly(ctx, a, 2025, 24, reviews.Answers{"highlights": "A's week"})
	require.NoError(t, err)

	rev, err := svc.GetWeekly(ctx, b, 2025, 24)
	require.NoError(t, err)
	require.Empty(t, rev.Answers, "B sees none of A's review")
}
