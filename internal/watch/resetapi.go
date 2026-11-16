package watch

import (
	"net/http"
	"sync"
)

// The Reset route (docs/reset.md): GET /api/reset on the viewer shows the player's reset (idle, resetting, rejoining,
// failed, with the last result), POST /api/reset {hard, wait} on the answer listener asks for one. The desk is
// late-bound like the pause desk's; without one (no runner yet) both answer 503 "off". A refusal the desk names (a
// measured run, a disabled hard reset) is a 409 with its code.

// ResetRoute serves both methods; the viewer mux only the GET, the answer mux only the POST.
const ResetRoute = "/api/reset"

// ResetRefusal is why the desk will not reset now.
type ResetRefusal struct {
	Code  string
	Error string
}

// ResetDesk is what the route needs from the bot side (bot.Runner through resetAdapter in cmd/d2bot).
type ResetDesk interface {
	View() any
	// Set asks for a reset: hard restarts the process, wait (default false) holds for a safe moment; the press itself
	// never waits. A press while one runs returns the running view (already set) and no refusal. Never blocks.
	Set(hard, wait bool) (any, *ResetRefusal)
}

// ResetAPI holds the desk once it exists.
type ResetAPI struct {
	mu sync.Mutex
	d  ResetDesk
}

func NewResetAPI() *ResetAPI { return &ResetAPI{} }

func (a *ResetAPI) Bind(d ResetDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *ResetAPI) desk() ResetDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET.
func (a *ResetAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "reset is not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	}))
}

// PostHandler serves POST /api/reset {hard?, wait?}; an empty object is an instant soft reset. The reply is 202.
func (a *ResetAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Hard bool `json:"hard"`
			Wait bool `json:"wait"`
		}
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "reset is not available"})
			return
		}
		v, no := d.Set(b.Hard, b.Wait)
		if no != nil {
			writeJSON(w, http.StatusConflict, policyReply{Code: no.Code, Error: no.Error})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusAccepted, v)
	})
}
