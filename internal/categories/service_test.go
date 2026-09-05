package categories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
)

func setup(t *testing.T) (categories.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return categories.NewService(pool), pool, newAccount(t, pool, "owner@test")
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func names(cats []categories.Category) []string {
	out := make([]string, len(cats))
	for i, c := range cats {
		out[i] = c.Name
	}
	return out
}

func in(name string) categories.Input { return categories.Input{Name: name} }

func updateName(name string) categories.UpdateInput { return categories.UpdateInput{Name: &name} }

func strp(s string) *string { return &s }

func TestCreateAndList(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, acc, categories.Input{Name: "  Gym  ", Colour: "green", Icon: "dumbbell"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, acc, in("DSA"))
	require.NoError(t, err)

	cats, err := svc.List(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, []string{"DSA", "Gym"}, names(cats), "name-ordered, trimmed")
	require.Equal(t, "green", cats[1].Colour)
	require.Equal(t, "dumbbell", cats[1].Icon)
	require.Empty(t, cats[0].Colour, "unset colour is empty, not null-ish")
}

func TestCreate_ColourAndIconRoundTrip(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, err := svc.Create(ctx, acc, categories.Input{Name: "Work", Colour: "  blue  ", Icon: "  briefcase  "})
	require.NoError(t, err)
	require.Equal(t, "blue", c.Colour, "trimmed")
	require.Equal(t, "briefcase", c.Icon)

	updated, err := svc.Update(ctx, acc, c.ID, categories.UpdateInput{Name: strp("Work"), Colour: strp("red"), Icon: strp("")})
	require.NoError(t, err)
	require.Equal(t, "red", updated.Colour)
	require.Empty(t, updated.Icon, "an explicit empty icon clears it")

	cats, _ := svc.List(ctx, acc)
	require.Equal(t, "red", cats[0].Colour)
	require.Empty(t, cats[0].Icon)
}

func TestUpdate_PartialLeavesOmittedFieldsUnchanged(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, err := svc.Create(ctx, acc, categories.Input{Name: "Gym", Colour: "blue", Icon: "dumbbell"})
	require.NoError(t, err)

	// rename only — colour/icon must survive (R3: the bug docs/left.md flagged)
	updated, err := svc.Update(ctx, acc, c.ID, updateName("Gym & Cardio"))
	require.NoError(t, err)
	require.Equal(t, "Gym & Cardio", updated.Name)
	require.Equal(t, "blue", updated.Colour, "omitted colour is left untouched")
	require.Equal(t, "dumbbell", updated.Icon, "omitted icon is left untouched")

	cats, _ := svc.List(ctx, acc)
	require.Equal(t, "blue", cats[0].Colour)
	require.Equal(t, "dumbbell", cats[0].Icon)

	// colour only — name/icon untouched
	updated, err = svc.Update(ctx, acc, c.ID, categories.UpdateInput{Colour: strp("green")})
	require.NoError(t, err)
	require.Equal(t, "Gym & Cardio", updated.Name)
	require.Equal(t, "green", updated.Colour)
	require.Equal(t, "dumbbell", updated.Icon)

	// explicit empty icon clears it; name/colour untouched
	updated, err = svc.Update(ctx, acc, c.ID, categories.UpdateInput{Icon: strp("")})
	require.NoError(t, err)
	require.Equal(t, "Gym & Cardio", updated.Name)
	require.Equal(t, "green", updated.Colour)
	require.Empty(t, updated.Icon)

	// an entirely empty UpdateInput{} changes nothing
	updated, err = svc.Update(ctx, acc, c.ID, categories.UpdateInput{})
	require.NoError(t, err)
	require.Equal(t, "Gym & Cardio", updated.Name)
	require.Equal(t, "green", updated.Colour)
	require.Empty(t, updated.Icon)
}

func TestCreate_DuplicateActiveNameCaseInsensitive(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, acc, in("Gym"))
	require.NoError(t, err)
	_, err = svc.Create(ctx, acc, in("  gym "))
	require.ErrorIs(t, err, categories.ErrNameTaken)
}

func TestCreate_ArchivedNameDoesNotBlock(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, err := svc.Create(ctx, acc, in("Gym"))
	require.NoError(t, err)
	require.NoError(t, svc.Archive(ctx, acc, c.ID))

	_, err = svc.Create(ctx, acc, in("Gym"))
	require.NoError(t, err, "an archived name is free to reuse")
}

func TestCreate_Validation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	cases := []struct {
		in    categories.Input
		field string
	}{
		{in(""), "name"},
		{in("   "), "name"},
		{in(strings.Repeat("x", 61)), "name"},
		{categories.Input{Name: "ok", Colour: strings.Repeat("c", 41)}, "colour"},
		{categories.Input{Name: "ok", Icon: strings.Repeat("i", 41)}, "icon"},
	}
	for _, tc := range cases {
		_, err := svc.Create(ctx, acc, tc.in)
		var verr *categories.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, tc.field)
	}
}

func TestUpdate(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, _ := svc.Create(ctx, acc, in("Gm"))
	_, err := svc.Update(ctx, acc, c.ID, updateName("Gym"))
	require.NoError(t, err)

	cats, _ := svc.List(ctx, acc)
	require.Equal(t, []string{"Gym"}, names(cats))

	// collision
	_, _ = svc.Create(ctx, acc, in("DSA"))
	_, err = svc.Update(ctx, acc, c.ID, updateName("dsa"))
	require.ErrorIs(t, err, categories.ErrNameTaken)

	// unknown id
	_, err = svc.Update(ctx, acc, uuid.New(), updateName("Whatever"))
	require.ErrorIs(t, err, categories.ErrNotFound)
}

func TestArchive(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	c, _ := svc.Create(ctx, acc, in("Gym"))
	require.NoError(t, svc.Archive(ctx, acc, c.ID))

	cats, _ := svc.List(ctx, acc)
	require.Empty(t, cats)

	// archiving again -> not found (already archived)
	require.ErrorIs(t, svc.Archive(ctx, acc, c.ID), categories.ErrNotFound)
	// updating an archived one -> not found
	_, err := svc.Update(ctx, acc, c.ID, updateName("X"))
	require.ErrorIs(t, err, categories.ErrNotFound)
}

func TestAssignableToAccount(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other@test")

	active, _ := svc.Create(ctx, acc, in("Active"))
	archived, _ := svc.Create(ctx, acc, in("Archived"))
	require.NoError(t, svc.Archive(ctx, acc, archived.ID))
	foreign, _ := svc.Create(ctx, other, in("Foreign"))

	ok, err := svc.AssignableToAccount(ctx, acc, active.ID)
	require.NoError(t, err)
	require.True(t, ok, "owned + active")

	for name, id := range map[string]uuid.UUID{
		"archived": archived.ID,
		"foreign":  foreign.ID,
		"unknown":  uuid.New(),
	} {
		ok, err := svc.AssignableToAccount(ctx, acc, id)
		require.NoError(t, err)
		require.False(t, ok, name)
	}
}

func TestNamesForAccount_IncludesArchived(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	a, _ := svc.Create(ctx, acc, in("Active"))
	b, _ := svc.Create(ctx, acc, in("Gone"))
	require.NoError(t, svc.Archive(ctx, acc, b.ID))

	m, err := svc.NamesForAccount(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, "Active", m[a.ID])
	require.Equal(t, "Gone", m[b.ID], "read models still label items on an archived category")
}

func TestListAll_IncludesArchived(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-listall@test")

	a, _ := svc.Create(ctx, acc, categories.Input{Name: "Active", Colour: "blue"})
	b, _ := svc.Create(ctx, acc, in("Gone"))
	require.NoError(t, svc.Archive(ctx, acc, b.ID))
	_, _ = svc.Create(ctx, other, in("Not mine"))

	all, err := svc.ListAll(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, []string{"Active", "Gone"}, names(all), "active and archived, nothing foreign")

	byID := map[uuid.UUID]categories.Category{}
	for _, c := range all {
		byID[c.ID] = c
	}
	require.Equal(t, "blue", byID[a.ID].Colour)
	require.Nil(t, byID[a.ID].ArchivedAt)
	require.NotNil(t, byID[b.ID].ArchivedAt)
}

func TestIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	ca, _ := svc.Create(ctx, a, in("A-only"))
	_, _ = svc.Create(ctx, b, in("B-only"))

	// B cannot see, update, or archive A's category.
	bCats, _ := svc.List(ctx, b)
	require.Equal(t, []string{"B-only"}, names(bCats))

	_, err := svc.Update(ctx, b, ca.ID, updateName("hijacked"))
	require.ErrorIs(t, err, categories.ErrNotFound)
	require.ErrorIs(t, svc.Archive(ctx, b, ca.ID), categories.ErrNotFound)

	bNames, _ := svc.NamesForAccount(ctx, b)
	require.NotContains(t, bNames, ca.ID, "B's name map never contains A's category")

	// A's category is untouched.
	aCats, _ := svc.List(ctx, a)
	require.Equal(t, []string{"A-only"}, names(aCats))
}
