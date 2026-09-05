package timeline

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the timeline HTTP endpoints.
type Handler struct {
	svc  Service
	zone AccountZone
}

// NewHandler builds the timeline handler. zone converts the wall-clock times the
// client submits into instants using the account's timezone (ADR-0005: no
// client-side tz math).
func NewHandler(svc Service, zone AccountZone) *Handler { return &Handler{svc: svc, zone: zone} }

// Protector wraps a handler with auth (write also adds CSRF). cmd/server supplies
// the account module's middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the timeline routes. write must enforce auth + CSRF; read only
// auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("POST /api/blocks", write(h.createBlock))
	mux.Handle("PUT /api/blocks/{id}", write(h.updateBlock))
	mux.Handle("DELETE /api/blocks/{id}", write(h.deleteBlock))

	mux.Handle("GET /api/timeline", read(h.getTimeline))
	mux.Handle("GET /api/timeline/range", read(h.getTimelineRange))
	mux.Handle("GET /api/comparison", read(h.getComparison))
	mux.Handle("GET /api/tasks/{id}/blocks", read(h.getBlocksForTask))
}

func queryDate(w http.ResponseWriter, r *http.Request) (timezone.Date, bool) {
	raw := r.URL.Query().Get("date")
	if raw == "" {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{
			"date": "date query parameter is required (YYYY-MM-DD)",
		}))
		return timezone.Date{}, false
	}
	d, err := timezone.ParseDate(raw)
	if err != nil {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "must be YYYY-MM-DD"}))
		return timezone.Date{}, false
	}
	return d, true
}

type positionedBlockBody struct {
	blockBody
	StartMinute int    `json:"start_minute"`
	EndMinute   int    `json:"end_minute"`
	FromPrevDay bool   `json:"from_prev_day"`
	ToNextDay   bool   `json:"to_next_day"`
	LocalDate   string `json:"local_date"`
	LocalStart  string `json:"local_start"`
	LocalEnd    string `json:"local_end"`
	EndsNextDay bool   `json:"ends_next_day"`
}

type timelineResponse struct {
	Date    string                `json:"date"`
	Planned []positionedBlockBody `json:"planned"`
	Actual  []positionedBlockBody `json:"actual"`
}

func toPositionedBody(b PositionedBlock) positionedBlockBody {
	return positionedBlockBody{
		blockBody:   toBlockBody(b.Block),
		StartMinute: b.StartMinute,
		EndMinute:   b.EndMinute,
		FromPrevDay: b.FromPrevDay,
		ToNextDay:   b.ToNextDay,
		LocalDate:   b.LocalDate,
		LocalStart:  b.LocalStart,
		LocalEnd:    b.LocalEnd,
		EndsNextDay: b.EndsNextDay,
	}
}

func (h *Handler) getTimeline(w http.ResponseWriter, r *http.Request) {
	date, ok := queryDate(w, r)
	if !ok {
		return
	}
	tl, err := h.svc.Timeline(r.Context(), accountID(r), date)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	resp := timelineResponse{
		Date:    date.String(),
		Planned: []positionedBlockBody{},
		Actual:  []positionedBlockBody{},
	}
	for _, b := range tl.Planned {
		resp.Planned = append(resp.Planned, toPositionedBody(b))
	}
	for _, b := range tl.Actual {
		resp.Actual = append(resp.Actual, toPositionedBody(b))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// rangeTimelineResponse is the batched shape the frontend's Week/Month views
// expect (docs/left.md) — one entry per day, each shaped exactly like
// timelineResponse's planned/actual arrays.
type rangeTimelineResponse struct {
	From string           `json:"from"`
	To   string           `json:"to"`
	Days []dayTimelineDay `json:"days"`
}

type dayTimelineDay struct {
	Date    string                `json:"date"`
	Planned []positionedBlockBody `json:"planned"`
	Actual  []positionedBlockBody `json:"actual"`
}

func (h *Handler) getTimelineRange(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fields := map[string]string{}

	from, err := timezone.ParseDate(q.Get("from"))
	if err != nil {
		fields["from"] = "must be YYYY-MM-DD"
	}
	to, err := timezone.ParseDate(q.Get("to"))
	if err != nil {
		fields["to"] = "must be YYYY-MM-DD"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return
	}

	rt, err := h.svc.TimelineRange(r.Context(), accountID(r), from, to)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	resp := rangeTimelineResponse{From: rt.From.String(), To: rt.To.String(), Days: make([]dayTimelineDay, len(rt.Days))}
	for i, d := range rt.Days {
		day := dayTimelineDay{Date: d.Date.String(), Planned: []positionedBlockBody{}, Actual: []positionedBlockBody{}}
		for _, b := range d.Planned {
			day.Planned = append(day.Planned, toPositionedBody(b))
		}
		for _, b := range d.Actual {
			day.Actual = append(day.Actual, toPositionedBody(b))
		}
		resp.Days[i] = day
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type comparisonRow struct {
	CategoryID        *string `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	PlannedSeconds    int64   `json:"planned_seconds"`
	ActualSeconds     int64   `json:"actual_seconds"`
	DifferenceSeconds int64   `json:"difference_seconds"`
}

func toComparisonRows(totals []CategoryTotals) []comparisonRow {
	rows := make([]comparisonRow, len(totals))
	for i, c := range totals {
		rows[i] = comparisonRow{
			CategoryName:      c.CategoryName,
			PlannedSeconds:    c.PlannedSeconds,
			ActualSeconds:     c.ActualSeconds,
			DifferenceSeconds: c.DifferenceSeconds,
		}
		if c.CategoryID != nil {
			s := c.CategoryID.String()
			rows[i].CategoryID = &s
		}
	}
	return rows
}

// comparisonResponse serves both the single-date and the range shape: Date is set
// for `?date=`, From/To for `?from=&to=` (M6/M7 foundation).
type comparisonResponse struct {
	Date       string          `json:"date,omitempty"`
	From       string          `json:"from,omitempty"`
	To         string          `json:"to,omitempty"`
	Categories []comparisonRow `json:"categories"`
}

// getComparison serves GET /api/comparison. `?date=` returns one day; `?from=&to=`
// (an alternative to `?date=`) returns the same per-category shape summed over the
// inclusive range — the M6/M7 foundation.
func (h *Handler) getComparison(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("from") != "" || q.Get("to") != "" {
		h.getComparisonRange(w, r)
		return
	}

	date, ok := queryDate(w, r)
	if !ok {
		return
	}
	cmp, err := h.svc.Comparison(r.Context(), accountID(r), date)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, comparisonResponse{
		Date: date.String(), Categories: toComparisonRows(cmp.Categories),
	})
}

func (h *Handler) getComparisonRange(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fields := map[string]string{}

	from, err := timezone.ParseDate(q.Get("from"))
	if err != nil {
		fields["from"] = "must be YYYY-MM-DD (required together with to)"
	}
	to, err := timezone.ParseDate(q.Get("to"))
	if err != nil {
		fields["to"] = "must be YYYY-MM-DD (required together with from)"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return
	}

	cmp, err := h.svc.ComparisonRange(r.Context(), accountID(r), from, to)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, comparisonResponse{
		From: from.String(), To: to.String(), Categories: toComparisonRows(cmp.Categories),
	})
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

// blockRequest is wall-clock: a date, HH:MM start/end, and whether the end falls
// on the next day (v1.md §3). The server converts to instants in the account's
// timezone — the client never does tz math (ADR-0005).
type blockRequest struct {
	Kind        string  `json:"kind"`
	Date        string  `json:"date"`
	Start       string  `json:"start"`
	End         string  `json:"end"`
	EndsNextDay bool    `json:"ends_next_day"`
	CategoryID  *string `json:"category_id"`
	TaskID      *string `json:"task_id"`
}

type blockBody struct {
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	StartsAt     string  `json:"starts_at"`
	EndsAt       string  `json:"ends_at"`
	CategoryID   *string `json:"category_id"`
	CategoryName *string `json:"category_name,omitempty"`
	TaskID       *string `json:"task_id"`
}

func uuidPtrString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func toBlockBody(b Block) blockBody {
	return blockBody{
		ID:           b.ID.String(),
		Kind:         string(b.Kind),
		StartsAt:     b.StartsAt.UTC().Format(time.RFC3339),
		EndsAt:       b.EndsAt.UTC().Format(time.RFC3339),
		CategoryID:   uuidPtrString(b.CategoryID),
		CategoryName: b.CategoryName,
		TaskID:       uuidPtrString(b.TaskID),
	}
}

func parseHHMM(s string) (hour, minute int, ok bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, 0, false
	}
	return t.Hour(), t.Minute(), true
}

// blockInstants turns a wall-clock request into a BlockInput using loc.
func blockInstants(loc *time.Location, req blockRequest) (BlockInput, *ValidationError) {
	fields := map[string]string{}

	d, dErr := timezone.ParseDate(req.Date)
	if dErr != nil {
		fields["date"] = "must be YYYY-MM-DD"
	}
	sh, sm, sok := parseHHMM(req.Start)
	if !sok {
		fields["start"] = "must be HH:MM"
	}
	eh, em, eok := parseHHMM(req.End)
	if !eok {
		fields["end"] = "must be HH:MM"
	}

	var catID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			fields["category_id"] = "must be a UUID"
		} else {
			catID = &id
		}
	}
	var taskID *uuid.UUID
	if req.TaskID != nil && *req.TaskID != "" {
		id, err := uuid.Parse(*req.TaskID)
		if err != nil {
			fields["task_id"] = "must be a UUID"
		} else {
			taskID = &id
		}
	}

	if len(fields) > 0 {
		return BlockInput{}, &ValidationError{Fields: fields}
	}

	endDayOffset := 0
	if req.EndsNextDay {
		endDayOffset = 1
	}
	return BlockInput{
		Kind:       BlockKind(req.Kind),
		StartsAt:   time.Date(d.Year, d.Month, d.Day, sh, sm, 0, 0, loc),
		EndsAt:     time.Date(d.Year, d.Month, d.Day+endDayOffset, eh, em, 0, 0, loc),
		CategoryID: catID,
		TaskID:     taskID,
	}, nil
}

func (h *Handler) blockInputFromRequest(w http.ResponseWriter, r *http.Request) (BlockInput, bool) {
	var req blockRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return BlockInput{}, false
	}
	loc, err := h.zone.Zone(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return BlockInput{}, false
	}
	in, verr := blockInstants(loc, req)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return BlockInput{}, false
	}
	return in, true
}

func (h *Handler) createBlock(w http.ResponseWriter, r *http.Request) {
	in, ok := h.blockInputFromRequest(w, r)
	if !ok {
		return
	}
	block, err := h.svc.AddBlock(r.Context(), accountID(r), in)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toBlockBody(block))
}

func (h *Handler) updateBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := h.blockInputFromRequest(w, r)
	if !ok {
		return
	}
	if err := h.svc.EditBlock(r.Context(), accountID(r), id, in.StartsAt, in.EndsAt, in.CategoryID, in.TaskID); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteBlock(r.Context(), accountID(r), id); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getBlocksForTask serves GET /api/tasks/{id}/blocks — v1.md §7's reverse-lookup
// view: every block linked to a task, across any date.
func (h *Handler) getBlocksForTask(w http.ResponseWriter, r *http.Request) {
	taskID, ok := pathID(w, r)
	if !ok {
		return
	}
	blocks, err := h.svc.BlocksForTask(r.Context(), accountID(r), taskID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	out := make([]blockBody, len(blocks))
	for i, b := range blocks {
		out[i] = toBlockBody(b)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"blocks": out})
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Not found"))
		return uuid.Nil, false
	}
	return id, true
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
	case errors.Is(err, ErrBlockNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Time block not found"))
	case errors.Is(err, ErrTaskNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Task not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
