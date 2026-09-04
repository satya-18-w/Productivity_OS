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
	mux.Handle("GET /api/categories", read(h.listCategories))
	mux.Handle("POST /api/categories", write(h.createCategory))
	mux.Handle("PATCH /api/categories/{id}", write(h.renameCategory))
	mux.Handle("POST /api/categories/{id}/archive", write(h.archiveCategory))

	mux.Handle("POST /api/blocks", write(h.createBlock))
	mux.Handle("PUT /api/blocks/{id}", write(h.updateBlock))
	mux.Handle("DELETE /api/blocks/{id}", write(h.deleteBlock))

	mux.Handle("GET /api/timeline", read(h.getTimeline))
	mux.Handle("GET /api/comparison", read(h.getComparison))
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

type comparisonRow struct {
	CategoryID        *string `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	PlannedSeconds    int64   `json:"planned_seconds"`
	ActualSeconds     int64   `json:"actual_seconds"`
	DifferenceSeconds int64   `json:"difference_seconds"`
}

type comparisonResponse struct {
	Date       string          `json:"date"`
	Categories []comparisonRow `json:"categories"`
}

func (h *Handler) getComparison(w http.ResponseWriter, r *http.Request) {
	date, ok := queryDate(w, r)
	if !ok {
		return
	}
	cmp, err := h.svc.Comparison(r.Context(), accountID(r), date)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	resp := comparisonResponse{Date: date.String(), Categories: []comparisonRow{}}
	for _, c := range cmp.Categories {
		row := comparisonRow{
			CategoryName:      c.CategoryName,
			PlannedSeconds:    c.PlannedSeconds,
			ActualSeconds:     c.ActualSeconds,
			DifferenceSeconds: c.DifferenceSeconds,
		}
		if c.CategoryID != nil {
			s := c.CategoryID.String()
			row.CategoryID = &s
		}
		resp.Categories = append(resp.Categories, row)
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

type categoryBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func toCategoryBody(c Category) categoryBody {
	return categoryBody{ID: c.ID.String(), Name: c.Name}
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListActiveCategories(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := make([]categoryBody, len(cats))
	for i, c := range cats {
		out[i] = toCategoryBody(c)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"categories": out})
}

type nameRequest struct {
	Name string `json:"name"`
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req nameRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	cat, err := h.svc.CreateCategory(r.Context(), accountID(r), req.Name)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toCategoryBody(cat))
}

func (h *Handler) renameCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req nameRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.RenameCategory(r.Context(), accountID(r), id, req.Name); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) archiveCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.ArchiveCategory(r.Context(), accountID(r), id); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
}

type blockBody struct {
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	StartsAt     string  `json:"starts_at"`
	EndsAt       string  `json:"ends_at"`
	CategoryID   *string `json:"category_id"`
	CategoryName *string `json:"category_name,omitempty"`
}

func categoryIDString(id *uuid.UUID) *string {
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
		CategoryID:   categoryIDString(b.CategoryID),
		CategoryName: b.CategoryName,
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
	if err := h.svc.EditBlock(r.Context(), accountID(r), id, in.StartsAt, in.EndsAt, in.CategoryID); err != nil {
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
	case errors.Is(err, ErrCategoryNameTaken):
		httpx.WriteError(w, r, httpx.NewError(http.StatusConflict, httpx.CodeConflict,
			"A category with this name already exists"))
	case errors.Is(err, ErrCategoryNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Category not found"))
	case errors.Is(err, ErrBlockNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Time block not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
