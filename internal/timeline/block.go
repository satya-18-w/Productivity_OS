package timeline

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/satya-18-w/productivity-os/internal/timeline/timelinedb"
)

func (s *service) AddBlock(ctx context.Context, accountID uuid.UUID, in BlockInput) (Block, error) {
	if verr := validateBlockTimes(in.StartsAt, in.EndsAt); verr != nil {
		return Block{}, verr
	}
	if in.Kind != Planned && in.Kind != Actual {
		return Block{}, &ValidationError{Fields: map[string]string{"kind": `kind must be "planned" or "actual"`}}
	}
	if verr := validateCategoryTaskExclusive(in.CategoryID, in.TaskID); verr != nil {
		return Block{}, verr
	}
	if err := s.assertAssignableCategory(ctx, accountID, in.CategoryID); err != nil {
		return Block{}, err
	}
	if err := s.assertAssignableTask(ctx, accountID, in.TaskID); err != nil {
		return Block{}, err
	}

	row, err := s.q.CreateTimeBlock(ctx, timelinedb.CreateTimeBlockParams{
		AccountID:  accountID,
		Kind:       string(in.Kind),
		StartsAt:   pgtype.Timestamptz{Time: in.StartsAt, Valid: true},
		EndsAt:     pgtype.Timestamptz{Time: in.EndsAt, Valid: true},
		CategoryID: toPgUUID(in.CategoryID),
		TaskID:     toPgUUID(in.TaskID),
	})
	if err != nil {
		return Block{}, fmt.Errorf("create time block: %w", err)
	}
	return Block{
		ID:         row.ID,
		Kind:       BlockKind(row.Kind),
		StartsAt:   row.StartsAt.Time,
		EndsAt:     row.EndsAt.Time,
		CategoryID: fromPgUUID(row.CategoryID),
		TaskID:     fromPgUUID(row.TaskID),
	}, nil
}

func (s *service) EditBlock(ctx context.Context, accountID, blockID uuid.UUID, starts, ends time.Time, categoryID, taskID *uuid.UUID) error {
	if verr := validateBlockTimes(starts, ends); verr != nil {
		return verr
	}
	if verr := validateCategoryTaskExclusive(categoryID, taskID); verr != nil {
		return verr
	}
	if err := s.assertAssignableCategory(ctx, accountID, categoryID); err != nil {
		return err
	}
	if err := s.assertAssignableTask(ctx, accountID, taskID); err != nil {
		return err
	}

	rows, err := s.q.UpdateTimeBlock(ctx, timelinedb.UpdateTimeBlockParams{
		AccountID:  accountID,
		ID:         blockID,
		StartsAt:   pgtype.Timestamptz{Time: starts, Valid: true},
		EndsAt:     pgtype.Timestamptz{Time: ends, Valid: true},
		CategoryID: toPgUUID(categoryID),
		TaskID:     toPgUUID(taskID),
	})
	if err != nil {
		return fmt.Errorf("update time block: %w", err)
	}
	if rows == 0 {
		return ErrBlockNotFound
	}
	return nil
}

func (s *service) DeleteBlock(ctx context.Context, accountID, blockID uuid.UUID) error {
	rows, err := s.q.DeleteTimeBlock(ctx, timelinedb.DeleteTimeBlockParams{AccountID: accountID, ID: blockID})
	if err != nil {
		return fmt.Errorf("delete time block: %w", err)
	}
	if rows == 0 {
		return ErrBlockNotFound
	}
	return nil
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

func (s *service) assertAssignableTask(ctx context.Context, accountID uuid.UUID, taskID *uuid.UUID) error {
	if taskID == nil {
		return nil
	}
	ok, err := s.tasks.AssignableToAccount(ctx, accountID, *taskID)
	if err != nil {
		return fmt.Errorf("check task: %w", err)
	}
	if !ok {
		return &ValidationError{Fields: map[string]string{"task_id": "task not found"}}
	}
	return nil
}

// validateCategoryTaskExclusive enforces the MX-TL inheritance rule: a task-linked
// block never carries its own category (mirrors the DB CHECK constraint at the
// application boundary, so the caller gets a field-level 400 instead of a raw
// constraint-violation error).
func validateCategoryTaskExclusive(categoryID, taskID *uuid.UUID) *ValidationError {
	if categoryID != nil && taskID != nil {
		return &ValidationError{Fields: map[string]string{
			"category_id": "cannot set a category on a task-linked block; it inherits the task's category",
		}}
	}
	return nil
}

func validateBlockTimes(starts, ends time.Time) *ValidationError {
	if starts.IsZero() || ends.IsZero() {
		return &ValidationError{Fields: map[string]string{"start": "start and end are required"}}
	}
	if !ends.After(starts) {
		return &ValidationError{Fields: map[string]string{"end": "end must be after start"}}
	}
	return nil
}
