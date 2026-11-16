package watch

import (
	"net/http"
	"sync"
)

// GET /api/lootrules/shadow (watch.lootshadow/1): the read-only report of the loot rules shadow stage
// (docs/lootrules.md, wave 1): totals, the top disagreements by rule and the last 50 lines. It only reports; the
// engine acts nowhere. Late-bound like the other desks.

var LootShadowRoutes = []string{"/api/lootrules/shadow"}

// LootShadowAPI holds the report source once the bot side exists.
type LootShadowAPI struct {
	mu  sync.Mutex
	rep func() any
}

func NewLootShadowAPI() *LootShadowAPI { return &LootShadowAPI{} }

// Bind sets the report function (cmd/d2bot: loot.DefaultShadow.Report with the knob).
func (a *LootShadowAPI) Bind(rep func() any) { a.mu.Lock(); a.rep = rep; a.mu.Unlock() }

func (a *LootShadowAPI) report() func() any { a.mu.Lock(); defer a.mu.Unlock(); return a.rep }

// Handler serves the GET route.
func (a *LootShadowAPI) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/lootrules/shadow", func(w http.ResponseWriter, r *http.Request) {
		rep := a.report()
		if rep == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the loot rules shadow is not available yet"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, rep())
	})
	return readOnly(m)
}
