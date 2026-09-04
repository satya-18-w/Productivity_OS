package goals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/goals/goalsdb"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

const (
	maxTitleLen = 200
	maxDescLen  = 5000
)

type service struct {
	q *goalsdb.Queries
}

// NewService builds the goals service over a connection pool.
func NewService(pool *pgxpool.Pool) Service {
	return &service{q: goalsdb.New(pool)}
}

func validateInput(in GoalInput) (GoalInput, *ValidationError) {
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
	if len(fields) > 0 {
		return GoalInput{}, &ValidationError{Fields: fields}
	}
	return GoalInput{Title: title, Description: strings.TrimSpace(in.Description), TargetDate: in.TargetDate}, nil
}

func (s *service) CreateGoal(ctx context.Context, accountID uuid.UUID, raw GoalInput) (Goal, error) {
	in, verr := validateInput(raw)
	if verr != nil {
		return Goal{}, verr
	}
	row, err := s.q.CreateGoal(ctx, goalsdb.CreateGoalParams{
		AccountID:   accountID,
		Title:       in.Title,
		Description: pgText(in.Description),
		TargetDate:  pgDate(in.TargetDate),
	})
	if err != nil {
		return Goal{}, fmt.Errorf("create goal: %w", err)
	}
	return toGoal(row.ID, row.Title, row.Description, row.TargetDate, row.Progress, row.CreatedAt, row.UpdatedAt), nil
}

func (s *service) UpdateGoal(ctx context.Context, accountID, goalID uuid.UUID, raw GoalInput) error {
	in, verr := validateInput(raw)
	if verr != nil {
		return verr
	}
	rows, err := s.q.UpdateGoalFields(ctx, goalsdb.UpdateGoalFieldsParams{
		AccountID:   accountID,
		ID:          goalID,
		Title:       in.Title,
		Description: pgText(in.Description),
		TargetDate:  pgDate(in.TargetDate),
	})
	if err != nil {
		return fmt.Errorf("update goal: %w", err)
	}
	if rows == 0 {
		return ErrGoalNotFound
	}
	return nil
}

func (s *service) SetProgress(ctx context.Context, accountID, goalID uuid.UUID, p Progress) error {
	if !validProgress(p) {
		return &ValidationError{Fields: map[string]string{
			"progress": "must be NOT_STARTED, IN_PROGRESS, ACHIEVED, or ABANDONED",
		}}
	}
	rows, err := s.q.UpdateGoalProgress(ctx, goalsdb.UpdateGoalProgressParams{
		AccountID: accountID, ID: goalID, Progress: string(p),
	})
	if err != nil {
		return fmt.Errorf("update progress: %w", err)
	}
	if rows == 0 {
		return ErrGoalNotFound
	}
	return nil
}

func (s *service) DeleteGoal(ctx context.Context, accountID, goalID uuid.UUID) error {
	rows, err := s.q.DeleteGoal(ctx, goalsdb.DeleteGoalParams{AccountID: accountID, ID: goalID})
	if err != nil {
		return fmt.Errorf("delete goal: %w", err)
	}
	if rows == 0 {
		return ErrGoalNotFound
	}
	return nil
}

func (s *service) ListGoals(ctx context.Context, accountID uuid.UUID) ([]Goal, error) {
	rows, err := s.q.ListGoals(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	out := make([]Goal, len(rows))
	for i, r := range rows {
		out[i] = toGoal(r.ID, r.Title, r.Description, r.TargetDate, r.Progress, r.CreatedAt, r.UpdatedAt)
	}
	return out, nil
}

func toGoal(id uuid.UUID, title string, desc pgtype.Text, target pgtype.Date, progress string, created, updated pgtype.Timestamptz) Goal {
	g := Goal{
		ID:         id,
		Title:      title,
		Progress:   Progress(progress),
		TargetDate: fromPgDate(target),
		CreatedAt:  created.Time.UTC().Format(time.RFC3339),
		UpdatedAt:  updated.Time.UTC().Format(time.RFC3339),
	}
	if desc.Valid {
		g.Description = desc.String
	}
	return g
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgDate(d *timezone.Date) pgtype.Date {
	if d == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func fromPgDate(v pgtype.Date) *timezone.Date {
	if !v.Valid {
		return nil
	}
	return &timezone.Date{Year: v.Time.Year(), Month: v.Time.Month(), Day: v.Time.Day()}
}
