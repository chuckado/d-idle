package main

import (
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
	t := dangerTable(db)
	if t == nil {
		return nil
	}
	return danger.Handler(t)
}
