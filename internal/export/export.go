// Package export is a read-only composition over categories, timeline, tasks,
// habits, goals, reviews, and notes — the account-data export of v1.md §14. It owns
// no table and performs no writes; every record is gathered from the other modules'
// own "list everything" read methods. Service is its entire public surface
// (ADR-0002).
package export

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/notes"
	"github.com/satya-18-w/productivity-os/internal/reviews"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// CategoriesReader is the slice of the categories module export needs.
// categories.Service already satisfies this structurally.
type CategoriesReader interface {
	ListAll(ctx context.Context, accountID uuid.UUID) ([]categories.Category, error)
}

// TimelineReader is the slice of the timeline module export needs. timeline.Service
// already satisfies this structurally.
type TimelineReader interface {
	ListAllBlocks(ctx context.Context, accountID uuid.UUID) ([]timeline.Block, error)
}

// TasksReader is the slice of the tasks module export needs. Board already
// returns every task, unbounded — no new method needed. ProgressByGoal supplies
// each goal's derived task-completion progress, matching what GET /api/goals
// already returns live (closes a completeness gap flagged during MX3-follow-up,
// docs/export-format.md). tasks.Service already satisfies this structurally.
type TasksReader interface {
	Board(ctx context.Context, accountID uuid.UUID) (tasks.Board, error)
	ProgressByGoal(ctx context.Context, accountID uuid.UUID) (done, total map[uuid.UUID]int, err error)
}

// HabitsReader is the slice of the habits module export needs. habits.Service
// already satisfies this structurally.
type HabitsReader interface {
	ListAll(ctx context.Context, accountID uuid.UUID) ([]habits.Habit, error)
	AllCompletions(ctx context.Context, accountID uuid.UUID) ([]habits.HabitCompletion, error)
}

// GoalsReader is the slice of the goals module export needs. ListGoals already
// returns every goal, unbounded — no new method needed. goals.Service already
// satisfies this structurally.
type GoalsReader interface {
	ListGoals(ctx context.Context, accountID uuid.UUID) ([]goals.Goal, error)
}

// ReviewsReader is the slice of the reviews module export needs. reviews.Service
// already satisfies this structurally.
type ReviewsReader interface {
	ListDaily(ctx context.Context, accountID uuid.UUID) ([]reviews.DailyReview, error)
	ListWeekly(ctx context.Context, accountID uuid.UUID) ([]reviews.WeeklyReview, error)
}

// NotesReader is the slice of the notes module export needs (MX4). ListNotes
// already returns every note, unbounded — no new method needed. notes.Service
// already satisfies this structurally.
type NotesReader interface {
	ListNotes(ctx context.Context, accountID uuid.UUID) ([]notes.Note, error)
}

// Export is the account's complete data snapshot (v1.md §14). Every slice is
// exactly what the account owns for that entity — never truncated, never
// filtered beyond account ownership.
type Export struct {
	ExportedAt       time.Time
	Categories       []categories.Category
	PlannedBlocks    []timeline.Block
	ActualBlocks     []timeline.Block
	Tasks            []tasks.Task
	Habits           []habits.Habit
	HabitCompletions []habits.HabitCompletion
	Goals            []goals.Goal
	GoalDoneTasks    map[uuid.UUID]int // keyed by goal id; absent = 0 linked-done tasks
	GoalTotalTasks   map[uuid.UUID]int // keyed by goal id; absent = 0 linked tasks
	DailyReviews     []reviews.DailyReview
	WeeklyReviews    []reviews.WeeklyReview
	Notes            []notes.Note
}

// Service is the export module's published interface.
type Service interface {
	// Export gathers all of the account's data into one snapshot — categories,
	// planned blocks, actual blocks, tasks, habits and their completions, goals,
	// daily reviews, weekly reviews, and notes (v1.md §14). Scoped entirely to
	// accountID, which the caller takes only from the request context (ADR-0004).
	Export(ctx context.Context, accountID uuid.UUID) (Export, error)
}
