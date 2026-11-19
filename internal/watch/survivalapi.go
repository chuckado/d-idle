package watch

import (
	"net/http"
	"sync/atomic"
)

// GET /api/survival: the survival controller's shadow choice in WORDS ONLY (docs/survival/ledger.md; user decision
// 2026-11-14: the watch screen never shows a margin number). It is the block that /api/glance (UI slice S2) will embed
// as `survival` once that endpoint exists; until then the UI reads it from here behind its feature flag. Read-only; the
// reply is {"enabled": false} until the bot reports a choice (SetSurvivalSource), so nothing changes without the hook.

// SurvivalRoute is the path of the shadow choice.
const SurvivalRoute = "/api/survival"

// SurvivalSource returns the latest words-only block (a survival.Glance) or nil. It must not block.
type SurvivalSource func() any

type survivalHolder struct {
	f atomic.Pointer[SurvivalSource]
}

// SetSurvivalSource sets where /api/survival reads from (nil: the endpoint answers enabled false).
func (h *Hub) SetSurvivalSource(f SurvivalSource) {
	if f == nil {
		h.survival.f.Store(nil)
		return
	}
	h.survival.f.Store(&f)
}

func (h *Hub) survivalHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if f := h.survival.f.Load(); f != nil {
			if v := (*f)(); v != nil {
				writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "survival": v})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
	}
}
