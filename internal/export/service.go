package export

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

type service struct {
	categories CategoriesReader
	timeline   TimelineReader
	tasks      TasksReader
	habits     HabitsReader
	goals      GoalsReader
	reviews    ReviewsReader
	notes      NotesReader
}

// NewService builds the export service over the other modules' published
// interfaces. It owns no connection pool — it never queries a table directly.
func NewService(
	categoriesReader CategoriesReader,
	timelineReader TimelineReader,
	tasksReader TasksReader,
	habitsReader HabitsReader,
	goalsReader GoalsReader,
	reviewsReader ReviewsReader,
	notesReader NotesReader,
) Service {
	return &service{
		categories: categoriesReader, timeline: timelineReader, tasks: tasksReader,
		habits: habitsReader, goals: goalsReader, reviews: reviewsReader, notes: notesReader,
	}
}

func (s *service) Export(ctx context.Context, accountID uuid.UUID) (Export, error) {
	cats, err := s.categories.ListAll(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list categories: %w", err)
	}

	blocks, err := s.timeline.ListAllBlocks(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list blocks: %w", err)
	}
	planned, actual := splitBlocks(blocks)

	board, err := s.tasks.Board(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list tasks: %w", err)
	}

	allHabits, err := s.habits.ListAll(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list habits: %w", err)
	}
	completions, err := s.habits.AllCompletions(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list habit completions: %w", err)
	}

	allGoals, err := s.goals.ListGoals(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list goals: %w", err)
	}
	goalDone, goalTotal, err := s.tasks.ProgressByGoal(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("goal progress: %w", err)
	}

	dailyReviews, err := s.reviews.ListDaily(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list daily reviews: %w", err)
	}
	weeklyReviews, err := s.reviews.ListWeekly(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list weekly reviews: %w", err)
	}

	allNotes, err := s.notes.ListNotes(ctx, accountID)
	if err != nil {
		return Export{}, fmt.Errorf("list notes: %w", err)
	}

	return Export{
		ExportedAt:       time.Now().UTC(),
		Categories:       cats,
		PlannedBlocks:    planned,
		ActualBlocks:     actual,
		Tasks:            flattenBoard(board),
		Habits:           allHabits,
		HabitCompletions: completions,
		Goals:            allGoals,
		GoalDoneTasks:    goalDone,
		GoalTotalTasks:   goalTotal,
		DailyReviews:     dailyReviews,
		WeeklyReviews:    weeklyReviews,
		Notes:            allNotes,
	}, nil
}

// splitBlocks separates the unbounded block list by kind — planned and actual are
// named as two distinct export entities (v1.md §14).
func splitBlocks(blocks []timeline.Block) (planned, actual []timeline.Block) {
	planned = []timeline.Block{}
	actual = []timeline.Block{}
	for _, b := range blocks {
		if b.Kind == timeline.Planned {
			planned = append(planned, b)
		} else {
			actual = append(actual, b)
		}
	}
	return planned, actual
}

func flattenBoard(board tasks.Board) []tasks.Task {
	out := []tasks.Task{}
	for _, col := range board.Columns {
		out = append(out, col.Tasks...)
	}
	return out
}
