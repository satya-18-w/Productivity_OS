package goals_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

func setup(t *testing.T) (goals.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return goals.NewService(pool), pool, newAccount(t, pool, "owner@test")
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func TestGoalLifecycle(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	due := timezone.Date{Year: 2025, Month: 12, Day: 31}
	g, err := svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "  Learn Go  ", Description: "deeply", TargetDate: &due})
	require.NoError(t, err)
	require.Equal(t, "Learn Go", g.Title)
	require.Equal(t, goals.NotStarted, g.Progress)
	require.Equal(t, "2025-12-31", g.TargetDate.String())

	for _, p := range []goals.Progress{goals.InProgress, goals.Achieved, goals.Abandoned, goals.NotStarted} {
		require.NoError(t, svc.SetProgress(ctx, acc, g.ID, p))
	}
	require.Error(t, svc.SetProgress(ctx, acc, g.ID, goals.Progress("SOMEDAY")))

	require.NoError(t, svc.UpdateGoal(ctx, acc, g.ID, goals.GoalInput{Title: "Master Go"}))
	list, _ := svc.ListGoals(ctx, acc)
	require.Len(t, list, 1)
	require.Equal(t, "Master Go", list[0].Title)
	require.Nil(t, list[0].TargetDate, "cleared on update")

	require.NoError(t, svc.DeleteGoal(ctx, acc, g.ID))
	require.ErrorIs(t, svc.DeleteGoal(ctx, acc, g.ID), goals.ErrGoalNotFound)
}

func TestGoalValidation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "  "})
	var verr *goals.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "title")

	_, err = svc.CreateGoal(ctx, acc, goals.GoalInput{Title: strings.Repeat("x", 201)})
	require.ErrorAs(t, err, &verr)
}

func TestGoalIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	ga, _ := svc.CreateGoal(ctx, a, goals.GoalInput{Title: "A's goal"})
	_, _ = svc.CreateGoal(ctx, b, goals.GoalInput{Title: "B's goal"})

	bl, _ := svc.ListGoals(ctx, b)
	require.Len(t, bl, 1)
	require.Equal(t, "B's goal", bl[0].Title)

	require.ErrorIs(t, svc.UpdateGoal(ctx, b, ga.ID, goals.GoalInput{Title: "x"}), goals.ErrGoalNotFound)
	require.ErrorIs(t, svc.SetProgress(ctx, b, ga.ID, goals.Achieved), goals.ErrGoalNotFound)
	require.ErrorIs(t, svc.DeleteGoal(ctx, b, ga.ID), goals.ErrGoalNotFound)
}

// --- HTTP ---

func stubProtector(accountID uuid.UUID) goals.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func TestGoalEndpoints(t *testing.T) {
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	goals.NewHandler(goals.NewService(pool)).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	do := func(method, path, body string) (*http.Response, map[string]any) {
		req, _ := http.NewRequestWithContext(context.Background(), method, srv.URL+path, strings.NewReader(body))
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var m map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&m)
		return resp, m
	}

	resp, body := do(http.MethodPost, "/api/goals", `{"title":"Ship V1","target_date":"2026-01-01"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "NOT_STARTED", body["progress"])
	id := body["id"].(string)

	resp, _ = do(http.MethodPost, "/api/goals", `{"title":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, _ = do(http.MethodPut, "/api/goals/"+id+"/progress", `{"progress":"ACHIEVED"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, _ = do(http.MethodPut, "/api/goals/"+id+"/progress", `{"progress":"NOPE"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	_, list := do(http.MethodGet, "/api/goals", "")
	gs := list["goals"].([]any)
	require.Len(t, gs, 1)
	require.Equal(t, "ACHIEVED", gs[0].(map[string]any)["progress"])

	resp, _ = do(http.MethodDelete, "/api/goals/"+id, "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, _ = do(http.MethodDelete, "/api/goals/not-a-uuid", "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
