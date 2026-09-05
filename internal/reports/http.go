package reports

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the report HTTP endpoint. It is read-only — reports performs no
// writes (v1.md §13: "The user can view").
type Handler struct {
	svc Service
}

// NewHandler builds the reports handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth. cmd/server supplies the account module's
// middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the report route behind read (auth only — no writes exist to
// protect with CSRF).
func (h *Handler) Mount(mux *http.ServeMux, read Protector) {
	mux.Handle("GET /api/reports", read(h.report))
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

func queryRange(w http.ResponseWriter, r *http.Request) (from, to timezone.Date, ok bool) {
	q := r.URL.Query()
	fields := map[string]string{}

	from, fromErr := timezone.ParseDate(q.Get("from"))
	if fromErr != nil {
		fields["from"] = "from query parameter is required (YYYY-MM-DD)"
	}
	to, toErr := timezone.ParseDate(q.Get("to"))
	if toErr != nil {
		fields["to"] = "to query parameter is required (YYYY-MM-DD)"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return timezone.Date{}, timezone.Date{}, false
	}
	return from, to, true
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	httpx.WriteError(w, r, err)
}

func categoryIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// timeByCategoryBody / plannedVsActualBody / habitCompletionBody / dailyActualBody
// and the top-level reportBody match docs/left.md Phase 9's response shape exactly
// — the frontend's reportsData.ts swap point consumes these field names verbatim.

type timeByCategoryBody struct {
	CategoryID   *string `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Seconds      int64   `json:"seconds"`
}

type plannedVsActualBody struct {
	CategoryID     *string `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	PlannedSeconds int64   `json:"planned_seconds"`
	ActualSeconds  int64   `json:"actual_seconds"`
}

type habitCompletionBody struct {
	HabitID       string `json:"habit_id"`
	HabitName     string `json:"habit_name"`
	CompletedDays int    `json:"completed_days"`
	RangeDays     int    `json:"range_days"`
}

type dailyActualTotalBody struct {
	Date    string `json:"date"`
	Seconds int64  `json:"seconds"`
}

type reportBody struct {
	From              string                 `json:"from"`
	To                string                 `json:"to"`
	TimeByCategory    []timeByCategoryBody   `json:"time_by_category"`
	PlannedVsActual   []plannedVsActualBody  `json:"planned_vs_actual"`
	HabitCompletion   []habitCompletionBody  `json:"habit_completion"`
	TaskThroughput    int                    `json:"task_throughput"`
	DailyActualTotals []dailyActualTotalBody `json:"daily_actual_totals"`
}

func toReportBody(rep Report) reportBody {
	out := reportBody{
		From: rep.From.String(), To: rep.To.String(),
		TimeByCategory:    make([]timeByCategoryBody, len(rep.TimeByCategory)),
		PlannedVsActual:   make([]plannedVsActualBody, len(rep.PlannedVsActual)),
		HabitCompletion:   make([]habitCompletionBody, len(rep.HabitCompletion)),
		TaskThroughput:    rep.TaskThroughput,
		DailyActualTotals: make([]dailyActualTotalBody, len(rep.DailyActualTotals)),
	}
	for i, c := range rep.TimeByCategory {
		out.TimeByCategory[i] = timeByCategoryBody{
			CategoryID: categoryIDString(c.CategoryID), CategoryName: c.CategoryName, Seconds: c.ActualSeconds,
		}
	}
	for i, c := range rep.PlannedVsActual {
		out.PlannedVsActual[i] = plannedVsActualBody{
			CategoryID: categoryIDString(c.CategoryID), CategoryName: c.CategoryName,
			PlannedSeconds: c.PlannedSeconds, ActualSeconds: c.ActualSeconds,
		}
	}
	for i, hc := range rep.HabitCompletion {
		out.HabitCompletion[i] = habitCompletionBody{
			HabitID: hc.HabitID.String(), HabitName: hc.Name,
			CompletedDays: hc.CompletedDays, RangeDays: hc.RangeDays,
		}
	}
	for i, d := range rep.DailyActualTotals {
		out.DailyActualTotals[i] = dailyActualTotalBody{Date: d.Date.String(), Seconds: d.ActualSeconds}
	}
	return out
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	from, to, ok := queryRange(w, r)
	if !ok {
		return
	}
	rep, err := h.svc.Report(r.Context(), accountID(r), from, to)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toReportBody(rep))
}
