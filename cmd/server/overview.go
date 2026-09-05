package main

import (
	"net/http"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// categoryCountsBody is one category's cross-module item counts.
type categoryCountsBody struct {
	Tasks  int `json:"tasks"`
	Habits int `json:"habits"`
	Goals  int `json:"goals"`
	Blocks int `json:"blocks"`
}

type categoryOverviewRow struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Colour string             `json:"colour"`
	Icon   string             `json:"icon"`
	Counts categoryCountsBody `json:"counts"`
}

// categoriesOverviewHandler composes categories.List with one categories.Counter
// per domain module into GET /api/categories/overview. It lives here, not inside
// internal/categories, because categories must never know another module's schema
// (ADR-0009) — this is the "thin composition handler in cmd/server" the ADR calls
// for.
func categoriesOverviewHandler(cats categories.Service, tasksC, habitsC, goalsC, blocksC categories.Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, _ := reqctx.IdentityFrom(ctx)
		accountID := id.AccountID

		list, err := cats.List(ctx, accountID)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		taskCounts, err := tasksC.CountByCategory(ctx, accountID)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		habitCounts, err := habitsC.CountByCategory(ctx, accountID)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		goalCounts, err := goalsC.CountByCategory(ctx, accountID)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		blockCounts, err := blocksC.CountByCategory(ctx, accountID)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}

		out := make([]categoryOverviewRow, len(list))
		for i, c := range list {
			out[i] = categoryOverviewRow{
				ID: c.ID.String(), Name: c.Name, Colour: c.Colour, Icon: c.Icon,
				Counts: categoryCountsBody{
					Tasks:  taskCounts[c.ID],
					Habits: habitCounts[c.ID],
					Goals:  goalCounts[c.ID],
					Blocks: blockCounts[c.ID],
				},
			}
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"categories": out})
	}
}
