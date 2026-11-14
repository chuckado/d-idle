package main

import (
	"log/slog"
	"net/http"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/data"
)

// wireDanger builds the danger profile from the game tables (docs/survival/design-2026-11-14.md section 3) and serves it
// read-only at /api/danger. nil without game tables or on an error (logged): the endpoint is then simply absent.
func wireDanger(db *data.DB) http.Handler {
	if db == nil {
		return nil
	}
	ov, err := danger.ShippedOverrides()
	if err != nil {
		slog.Warn("danger: bad overrides", "err", err)
		return nil
	}
	t, err := danger.Build(db, danger.DefaultParams(), ov)
	if err != nil {
		slog.Warn("danger: no profile", "err", err)
		return nil
	}
	return danger.Handler(t)
}
