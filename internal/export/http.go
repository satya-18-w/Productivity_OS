package export

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/notes"
	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/reviews"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// Handler serves the export HTTP endpoint. It is read-only — export performs no
// writes (v1.md §14: "a single user-initiated download").
type Handler struct {
	svc Service
}

// NewHandler builds the export handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth. cmd/server supplies the account module's
// middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the export route behind read (auth only — no writes exist to
// protect with CSRF).
func (h *Handler) Mount(mux *http.ServeMux, read Protector) {
	mux.Handle("GET /api/export", read(h.export))
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

func categoryIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func timePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

type categoryBody struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Colour     string  `json:"colour"`
	Icon       string  `json:"icon"`
	ArchivedAt *string `json:"archived_at"`
}

func toCategoryBody(c categories.Category) categoryBody {
	return categoryBody{ID: c.ID.String(), Name: c.Name, Colour: c.Colour, Icon: c.Icon, ArchivedAt: timePtr(c.ArchivedAt)}
}

type blockBody struct {
	ID         string  `json:"id"`
	StartsAt   string  `json:"starts_at"`
	EndsAt     string  `json:"ends_at"`
	CategoryID *string `json:"category_id"`
	TaskID     *string `json:"task_id"`
}

func toBlockBody(b timeline.Block) blockBody {
	return blockBody{
		ID: b.ID.String(), StartsAt: b.StartsAt.UTC().Format(time.RFC3339), EndsAt: b.EndsAt.UTC().Format(time.RFC3339),
		CategoryID: categoryIDString(b.CategoryID), TaskID: categoryIDString(b.TaskID),
	}
}

type taskBody struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
	State       string  `json:"state"`
	CategoryID  *string `json:"category_id"`
	GoalID      *string `json:"goal_id"`
	Priority    *string `json:"priority"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toTaskBody(t tasks.Task) taskBody {
	b := taskBody{
		ID: t.ID.String(), Title: t.Title, Description: t.Description, State: string(t.State),
		CategoryID: categoryIDString(t.CategoryID), CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	if t.DueDate != nil {
		s := t.DueDate.String()
		b.DueDate = &s
	}
	if t.GoalID != nil {
		s := t.GoalID.String()
		b.GoalID = &s
	}
	if t.Priority != nil {
		s := string(*t.Priority)
		b.Priority = &s
	}
	return b
}

type habitBody struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CategoryID *string `json:"category_id"`
	Target     *string `json:"target"`
	ArchivedAt *string `json:"archived_at"`
}

func toHabitBody(h habits.Habit) habitBody {
	return habitBody{
		ID: h.ID.String(), Name: h.Name, CategoryID: categoryIDString(h.CategoryID),
		Target: h.Target, ArchivedAt: timePtr(h.ArchivedAt),
	}
}

type habitCompletionBody struct {
	HabitID string `json:"habit_id"`
	Date    string `json:"date"`
}

func toHabitCompletionBody(c habits.HabitCompletion) habitCompletionBody {
	return habitCompletionBody{HabitID: c.HabitID.String(), Date: c.Date.String()}
}

type goalBody struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TargetDate  *string `json:"target_date"`
	Progress    string  `json:"progress"`
	CategoryID  *string `json:"category_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DoneTasks   int     `json:"done_tasks"`
	TotalTasks  int     `json:"total_tasks"`
}

func toGoalBody(g goals.Goal) goalBody {
	b := goalBody{
		ID: g.ID.String(), Title: g.Title, Description: g.Description, Progress: string(g.Progress),
		CategoryID: categoryIDString(g.CategoryID), CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
	if g.TargetDate != nil {
		s := g.TargetDate.String()
		b.TargetDate = &s
	}
	return b
}

type dailyReviewBody struct {
	Date      string            `json:"date"`
	Answers   map[string]string `json:"answers"`
	UpdatedAt string            `json:"updated_at"`
}

func toDailyReviewBody(r reviews.DailyReview) dailyReviewBody {
	return dailyReviewBody{Date: r.Date.String(), Answers: r.Answers, UpdatedAt: r.UpdatedAt}
}

type weeklyReviewBody struct {
	ISOYear   int               `json:"iso_year"`
	ISOWeek   int               `json:"iso_week"`
	Answers   map[string]string `json:"answers"`
	UpdatedAt string            `json:"updated_at"`
}

func toWeeklyReviewBody(r reviews.WeeklyReview) weeklyReviewBody {
	return weeklyReviewBody{ISOYear: r.ISOYear, ISOWeek: r.ISOWeek, Answers: r.Answers, UpdatedAt: r.UpdatedAt}
}

type noteBody struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toNoteBody(n notes.Note) noteBody {
	return noteBody{ID: n.ID.String(), Title: n.Title, Body: n.Body, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt}
}

// exportBody is the documented export schema (docs/export-format.md).
type exportBody struct {
	ExportedAt       string                `json:"exported_at"`
	Categories       []categoryBody        `json:"categories"`
	PlannedBlocks    []blockBody           `json:"planned_blocks"`
	ActualBlocks     []blockBody           `json:"actual_blocks"`
	Tasks            []taskBody            `json:"tasks"`
	Habits           []habitBody           `json:"habits"`
	HabitCompletions []habitCompletionBody `json:"habit_completions"`
	Goals            []goalBody            `json:"goals"`
	DailyReviews     []dailyReviewBody     `json:"daily_reviews"`
	WeeklyReviews    []weeklyReviewBody    `json:"weekly_reviews"`
	Notes            []noteBody            `json:"notes"`
}

func toExportBody(e Export) exportBody {
	out := exportBody{
		ExportedAt:       e.ExportedAt.Format(time.RFC3339),
		Categories:       make([]categoryBody, len(e.Categories)),
		PlannedBlocks:    make([]blockBody, len(e.PlannedBlocks)),
		ActualBlocks:     make([]blockBody, len(e.ActualBlocks)),
		Tasks:            make([]taskBody, len(e.Tasks)),
		Habits:           make([]habitBody, len(e.Habits)),
		HabitCompletions: make([]habitCompletionBody, len(e.HabitCompletions)),
		Goals:            make([]goalBody, len(e.Goals)),
		DailyReviews:     make([]dailyReviewBody, len(e.DailyReviews)),
		WeeklyReviews:    make([]weeklyReviewBody, len(e.WeeklyReviews)),
		Notes:            make([]noteBody, len(e.Notes)),
	}
	for i, c := range e.Categories {
		out.Categories[i] = toCategoryBody(c)
	}
	for i, b := range e.PlannedBlocks {
		out.PlannedBlocks[i] = toBlockBody(b)
	}
	for i, b := range e.ActualBlocks {
		out.ActualBlocks[i] = toBlockBody(b)
	}
	for i, t := range e.Tasks {
		out.Tasks[i] = toTaskBody(t)
	}
	for i, h := range e.Habits {
		out.Habits[i] = toHabitBody(h)
	}
	for i, c := range e.HabitCompletions {
		out.HabitCompletions[i] = toHabitCompletionBody(c)
	}
	for i, g := range e.Goals {
		b := toGoalBody(g)
		b.DoneTasks = e.GoalDoneTasks[g.ID]
		b.TotalTasks = e.GoalTotalTasks[g.ID]
		out.Goals[i] = b
	}
	for i, r := range e.DailyReviews {
		out.DailyReviews[i] = toDailyReviewBody(r)
	}
	for i, r := range e.WeeklyReviews {
		out.WeeklyReviews[i] = toWeeklyReviewBody(r)
	}
	for i, n := range e.Notes {
		out.Notes[i] = toNoteBody(n)
	}
	return out
}

func (h *Handler) export(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Export(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	filename := fmt.Sprintf("productivity-os-export-%s.json", out.ExportedAt.Format("2006-01-02"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	httpx.WriteJSON(w, http.StatusOK, toExportBody(out))
}
