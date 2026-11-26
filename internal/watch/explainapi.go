package watch

import (
	"net/http"
	"sync/atomic"
)

// GET /api/explain: what the hero decided lately and what is refusing it now (Fable logic review 2026-11-25, P7). The
// bot supplies the answer through SetExplainSource; read-only, GET and HEAD only, mounted with the other read routes
// (same access as /api/why). Until a source is set it answers {"enabled": false}.

// ExplainRoute is the path of the decision ring.
const ExplainRoute = "/api/explain"

// ExplainSource returns the current explain view (bot.ExplainView) or nil. It must not block.
type ExplainSource func() any

type explainHolder struct {
	f atomic.Pointer[ExplainSource]
}

// SetExplainSource sets where /api/explain reads from (nil: the endpoint answers enabled false).
func (h *Hub) SetExplainSource(f ExplainSource) {
	if f == nil {
		h.explain.f.Store(nil)
		return
	}
	h.explain.f.Store(&f)
}

func (h *Hub) explainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if f := h.explain.f.Load(); f != nil {
			if v := (*f)(); v != nil {
				writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "explain": v})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
	}
}
