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

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/tasks"
)

// fakeProgress is a no-op goals.ProgressReader for HTTP tests that don't exercise
// task-linkage (MX3).
type fakeProgress struct{}

func (fakeProgress) ProgressByGoal(context.Context, uuid.UUID) (map[uuid.UUID]int, map[uuid.UUID]int, error) {
	return map[uuid.UUID]int{}, map[uuid.UUID]int{}, nil
}

func setup(t *testing.T) (goals.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return goals.NewService(pool, categories.NewService(pool)), pool, newAccount(t, pool, "owner@test")
}

// mkCategory inserts an active category directly (goals tests only need one to
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

func TestGoalCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-cat@test")

	mine := mkCategory(t, pool, acc, "Career")
	theirs := mkCategory(t, pool, other, "Theirs")
	archived := mkCategory(t, pool, acc, "Old")
	archiveCategory(t, pool, archived)

	g, err := svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "Learn Go", CategoryID: &mine})
	require.NoError(t, err)
	require.Equal(t, mine, *g.CategoryID)

	for _, cat := range []uuid.UUID{theirs, archived, uuid.New()} {
		_, err := svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "x", CategoryID: &cat})
		var verr *goals.ValidationError
		require.ErrorAs(t, err, &verr)
		require.Contains(t, verr.Fields, "category_id")
	}

	other2 := mkCategory(t, pool, acc, "Health")
	require.NoError(t, svc.UpdateGoal(ctx, acc, g.ID, goals.GoalInput{Title: "Learn Go", CategoryID: &other2}))
	list, _ := svc.ListGoals(ctx, acc)
	require.Equal(t, other2, *list[0].CategoryID)

	require.NoError(t, svc.UpdateGoal(ctx, acc, g.ID, goals.GoalInput{Title: "Learn Go"}))
	list, _ = svc.ListGoals(ctx, acc)
	require.Nil(t, list[0].CategoryID, "cleared on update")
}

func TestCountByCategory(t *testing.T) {
	svc, pool, acc := setup(t)
	ctx := context.Background()
	other := newAccount(t, pool, "other-count@test")

	catA := mkCategory(t, pool, acc, "A")
	catB := mkCategory(t, pool, acc, "B")
	otherCat := mkCategory(t, pool, other, "Other")

	_, _ = svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "1", CategoryID: &catA})
	_, _ = svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "2", CategoryID: &catA})
	_, _ = svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "3", CategoryID: &catB})
	_, _ = svc.CreateGoal(ctx, acc, goals.GoalInput{Title: "4"}) // uncategorized
	_, _ = svc.CreateGoal(ctx, other, goals.GoalInput{Title: "5", CategoryID: &otherCat})

	counts, err := svc.CountByCategory(ctx, acc)
	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]int{catA: 2, catB: 1}, counts)
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
	goals.NewHandler(goals.NewService(pool, categories.NewService(pool)), fakeProgress{}).Mount(mux, stubProtector(acc), stubProtector(acc))
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

	// category_id: unknown -> 400
	resp, _ = do(http.MethodPost, "/api/goals",
		`{"title":"x","category_id":"`+uuid.NewString()+`"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var catID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, 'Career') RETURNING id`, acc).Scan(&catID))
	resp, body = do(http.MethodPost, "/api/goals", `{"title":"Grow","category_id":"`+catID.String()+`"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, catID.String(), body["category_id"])
}

// TestGoalEndpoints_TaskProgress exercises MX3's derived progress end to end: link
// tasks to a goal via the real tasks module, mark some DONE, confirm GET /api/goals
// reads back done_tasks/total_tasks correctly; then delete the goal and confirm its
// tasks survive with goal_id cleared (v1.md §10 amendment — CP 2).
func TestGoalEndpoints_TaskProgress(t *testing.T) {
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "progress@test")
	catSvc := categories.NewService(pool)
	goalsSvc := goals.NewService(pool, catSvc)
	tasksSvc := tasks.NewService(pool, catSvc, goalsSvc)

	mux := http.NewServeMux()
	goals.NewHandler(goalsSvc, tasksSvc).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ctx := context.Background()
	goal, err := goalsSvc.CreateGoal(ctx, acc, goals.GoalInput{Title: "Ship MX3"})
	require.NoError(t, err)

	var taskIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		task, err := tasksSvc.CreateTask(ctx, acc, tasks.TaskInput{Title: "t", GoalID: &goal.ID})
		require.NoError(t, err)
		taskIDs = append(taskIDs, task.ID)
	}
	require.NoError(t, tasksSvc.MoveTask(ctx, acc, taskIDs[0], tasks.Done))
	require.NoError(t, tasksSvc.MoveTask(ctx, acc, taskIDs[1], tasks.Done))

	do := func(method, path string) map[string]any {
		req, _ := http.NewRequestWithContext(ctx, method, srv.URL+path, nil)
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var m map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&m)
		return m
	}

	list := do(http.MethodGet, "/api/goals")
	gs := list["goals"].([]any)
	require.Len(t, gs, 1)
	g := gs[0].(map[string]any)
	require.InEpsilon(t, 2, g["done_tasks"], 0)
	require.InEpsilon(t, 3, g["total_tasks"], 0)

	// deleting the goal clears goal_id on its tasks without deleting them
	require.NoError(t, goalsSvc.DeleteGoal(ctx, acc, goal.ID))
	board, err := tasksSvc.Board(ctx, acc)
	require.NoError(t, err)
	var found int
	for _, col := range board.Columns {
		for _, tk := range col.Tasks {
			found++
			require.Nil(t, tk.GoalID)
		}
	}
	require.Equal(t, 3, found)
}
