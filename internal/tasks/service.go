package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/tasks/tasksdb"
)

const (
	maxTitleLen = 200
	maxDescLen  = 5000
)

type service struct {
	pool  *pgxpool.Pool
	q     *tasksdb.Queries
	cats  CategoryChecker
	goals GoalChecker
}

// NewService builds the tasks service over a connection pool. cats validates a
// category_id a caller assigns (wired to categories.Service by cmd/server —
// ADR-0009); goals validates a goal_id a caller links (wired to goals.Service —
// MX3).
func NewService(pool *pgxpool.Pool, cats CategoryChecker, goals GoalChecker) Service {
	return &service{pool: pool, q: tasksdb.New(pool), cats: cats, goals: goals}
}

func (s *service) assertAssignableCategory(ctx context.Context, accountID uuid.UUID, categoryID *uuid.UUID) error {
	if categoryID == nil {
		return nil
	}
	ok, err := s.cats.AssignableToAccount(ctx, accountID, *categoryID)
	if err != nil {
		return fmt.Errorf("check category: %w", err)
	}
	if !ok {
		return &ValidationError{Fields: map[string]string{"category_id": "category not found or archived"}}
	}
	return nil
}

func (s *service) assertAssignableGoal(ctx context.Context, accountID uuid.UUID, goalID *uuid.UUID) error {
	if goalID == nil {
		return nil
	}
	ok, err := s.goals.AssignableToAccount(ctx, accountID, *goalID)
	if err != nil {
		return fmt.Errorf("check goal: %w", err)
	}
	if !ok {
		return &ValidationError{Fields: map[string]string{"goal_id": "goal not found"}}
	}
	return nil
}

func validateInput(in TaskInput) (TaskInput, *ValidationError) {
	fields := map[string]string{}

	title := strings.TrimSpace(in.Title)
	switch {
	case title == "":
		fields["title"] = "title is required"
	case len(title) > maxTitleLen:
		fields["title"] = "title must be at most 200 characters"
	}
	if len(in.Description) > maxDescLen {
		fields["description"] = "description must be at most 5000 characters"
	}

	if in.Priority != nil && !validPriority(*in.Priority) {
		fields["priority"] = "must be HIGH, MEDIUM, or LOW"
	}

	if len(fields) > 0 {
		return TaskInput{}, &ValidationError{Fields: fields}
	}
	return TaskInput{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		DueDate:     in.DueDate,
		CategoryID:  in.CategoryID,
		GoalID:      in.GoalID,
		Priority:    in.Priority,
	}, nil
}

func (s *service) CreateTask(ctx context.Context, accountID uuid.UUID, raw TaskInput) (Task, error) {
	in, verr := validateInput(raw)
	if verr != nil {
		return Task{}, verr
	}
	if err := s.assertAssignableCategory(ctx, accountID, in.CategoryID); err != nil {
		return Task{}, err
	}
	if err := s.assertAssignableGoal(ctx, accountID, in.GoalID); err != nil {
		return Task{}, err
	}

	var task Task
	err := s.inTx(ctx, func(q *tasksdb.Queries) error {
		row, err := q.CreateTask(ctx, tasksdb.CreateTaskParams{
			AccountID:   accountID,
			Title:       in.Title,
			Description: pgText(in.Description),
			DueDate:     pgDate(in.DueDate),
			State:       string(Backlog),
			CategoryID:  toPgUUID(in.CategoryID),
			GoalID:      toPgUUID(in.GoalID),
			Priority:    toPgPriority(in.Priority),
		})
		if err != nil {
			return fmt.Errorf("insert task: %w", err)
		}
		task = toTask(row.ID, row.Title, row.Description, row.DueDate, row.State, row.CategoryID, row.GoalID, row.Priority, row.CreatedAt, row.UpdatedAt)

		return q.RecordTransition(ctx, tasksdb.RecordTransitionParams{
			TaskID:    row.ID,
			AccountID: accountID,
			FromState: pgtype.Text{},
			ToState:   string(Backlog),
		})
	})
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (s *service) UpdateTask(ctx context.Context, accountID, taskID uuid.UUID, raw TaskInput) error {
	in, verr := validateInput(raw)
	if verr != nil {
		return verr
	}
	if err := s.assertAssignableCategory(ctx, accountID, in.CategoryID); err != nil {
		return err
	}
	if err := s.assertAssignableGoal(ctx, accountID, in.GoalID); err != nil {
		return err
	}
	rows, err := s.q.UpdateTaskFields(ctx, tasksdb.UpdateTaskFieldsParams{
		AccountID:   accountID,
		ID:          taskID,
		Title:       in.Title,
		Description: pgText(in.Description),
		DueDate:     pgDate(in.DueDate),
		CategoryID:  toPgUUID(in.CategoryID),
		GoalID:      toPgUUID(in.GoalID),
		Priority:    toPgPriority(in.Priority),
	})
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *service) MoveTask(ctx context.Context, accountID, taskID uuid.UUID, to State) error {
	if !validState(to) {
		return &ValidationError{Fields: map[string]string{"state": "must be BACKLOG, TODO, IN_PROGRESS, or DONE"}}
	}

	return s.inTx(ctx, func(q *tasksdb.Queries) error {
		current, err := q.GetTaskState(ctx, tasksdb.GetTaskStateParams{AccountID: accountID, ID: taskID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTaskNotFound
			}
			return fmt.Errorf("get task state: %w", err)
		}
		if State(current) == to {
			return nil // no-op, records nothing
		}

		if _, err := q.UpdateTaskState(ctx, tasksdb.UpdateTaskStateParams{
			AccountID: accountID, ID: taskID, State: string(to),
		}); err != nil {
			return fmt.Errorf("update state: %w", err)
		}
		return q.RecordTransition(ctx, tasksdb.RecordTransitionParams{
			TaskID:    taskID,
			AccountID: accountID,
			FromState: pgtype.Text{String: current, Valid: true},
			ToState:   string(to),
		})
	})
}

func (s *service) DeleteTask(ctx context.Context, accountID, taskID uuid.UUID) error {
	rows, err := s.q.DeleteTask(ctx, tasksdb.DeleteTaskParams{AccountID: accountID, ID: taskID})
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *service) Board(ctx context.Context, accountID uuid.UUID) (Board, error) {
	rows, err := s.q.ListTasks(ctx, accountID)
	if err != nil {
		return Board{}, fmt.Errorf("list tasks: %w", err)
	}

	byState := map[State][]Task{}
	for _, r := range rows {
		t := toTask(r.ID, r.Title, r.Description, r.DueDate, r.State, r.CategoryID, r.GoalID, r.Priority, r.CreatedAt, r.UpdatedAt)
		byState[t.State] = append(byState[t.State], t)
	}

	board := Board{Columns: make([]Column, len(Columns))}
	for i, st := range Columns {
		col := Column{State: st, Tasks: byState[st]}
		if col.Tasks == nil {
			col.Tasks = []Task{}
		}
		board.Columns[i] = col
	}
	return board, nil
}

func (s *service) CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := s.q.CountTasksByCategory(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count tasks by category: %w", err)
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		if id := fromPgUUID(r.CategoryID); id != nil {
			out[*id] = int(r.Total)
		}
	}
	return out, nil
}

func (s *service) DoneCountInRange(ctx context.Context, accountID uuid.UUID, from, to time.Time) (int, error) {
	n, err := s.q.DoneTaskCountInRange(ctx, tasksdb.DoneTaskCountInRangeParams{
		AccountID:   accountID,
		FromInstant: pgtype.Timestamptz{Time: from, Valid: true},
		ToInstant:   pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("done count in range: %w", err)
	}
	return int(n), nil
}

func (s *service) ProgressByGoal(ctx context.Context, accountID uuid.UUID) (done, total map[uuid.UUID]int, err error) {
	rows, err := s.q.ProgressByGoal(ctx, accountID)
	if err != nil {
		return nil, nil, fmt.Errorf("progress by goal: %w", err)
	}
	done = make(map[uuid.UUID]int, len(rows))
	total = make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		id := fromPgUUID(r.GoalID)
		if id == nil {
			continue
		}
		done[*id] = int(r.Done)
		total[*id] = int(r.Total)
	}
	return done, total, nil
}

func (s *service) AssignableToAccount(ctx context.Context, accountID, taskID uuid.UUID) (bool, error) {
	n, err := s.q.CountAssignableTask(ctx, tasksdb.CountAssignableTaskParams{
		AccountID: accountID, ID: taskID,
	})
	if err != nil {
		return false, fmt.Errorf("check task: %w", err)
	}
	return n > 0, nil
}

func (s *service) CategoriesForTasks(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]*uuid.UUID, error) {
	rows, err := s.q.TaskCategories(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("task categories: %w", err)
	}
	out := make(map[uuid.UUID]*uuid.UUID, len(rows))
	for _, r := range rows {
		out[r.ID] = fromPgUUID(r.CategoryID)
	}
	return out, nil
}

func (s *service) inTx(ctx context.Context, fn func(*tasksdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func toTask(id uuid.UUID, title string, desc pgtype.Text, due pgtype.Date, state string, category, goal pgtype.UUID, priority pgtype.Text, created, updated pgtype.Timestamptz) Task {
	t := Task{
		ID:         id,
		Title:      title,
		State:      State(state),
		DueDate:    fromPgDate(due),
		CategoryID: fromPgUUID(category),
		GoalID:     fromPgUUID(goal),
		Priority:   fromPgPriority(priority),
		CreatedAt:  created.Time.UTC().Format(time.RFC3339),
		UpdatedAt:  updated.Time.UTC().Format(time.RFC3339),
	}
	if desc.Valid {
		t.Description = desc.String
	}
	return t
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromPgUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := uuid.UUID(v.Bytes)
	return &id
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toPgPriority(p *Priority) pgtype.Text {
	if p == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*p), Valid: true}
}

func fromPgPriority(v pgtype.Text) *Priority {
	if !v.Valid {
		return nil
	}
	p := Priority(v.String)
	return &p
}

func pgDate(d *timezone.Date) pgtype.Date {
	if d == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func fromPgDate(v pgtype.Date) *timezone.Date {
	if !v.Valid {
		return nil
	}
	return &timezone.Date{Year: v.Time.Year(), Month: v.Time.Month(), Day: v.Time.Day()}
}
