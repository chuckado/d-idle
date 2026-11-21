package watch

import (
	"net/http"
	"sync"
)

// The Recall now route (docs/process/management-fast-path.md): GET /api/recall on the viewer shows the command's state,
// POST /api/recall {} on the answer listener takes the hero home at the next safe moment, {"cancel": true} withdraws it.
// Without a desk (no runner, a measured run) both answer 503 "off".

// RecallRoute serves both methods; the viewer mux only the GET, the answer mux only the POST.
const RecallRoute = "/api/recall"

// RecallDesk is what the route needs from the bot side (bot.Runner through recallAdapter in cmd/d2bot).
type RecallDesk interface {
	View() any
	Recall() any
	Cancel() any
}

// RecallAPI holds the desk once it exists.
type RecallAPI struct {
	mu sync.Mutex
	d  RecallDesk
}

func NewRecallAPI() *RecallAPI { return &RecallAPI{} }

func (a *RecallAPI) Bind(d RecallDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *RecallAPI) desk() RecallDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET.
func (a *RecallAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "recall is not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	}))
}

// PostHandler serves POST /api/recall. The reply is the view plus "ok" (false when the view carries an error).
func (a *RecallAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Cancel bool `json:"cancel"`
		}
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "recall is not available"})
			return
		}
		var v any
		if b.Cancel {
			v = d.Cancel()
		} else {
			v = d.Recall()
		}
		var m map[string]any
		if !decode(marshal(v), &m) || m == nil {
			writeJSON(w, 200, v)
			return
		}
		e, _ := m["error"].(string)
		m["ok"] = e == ""
		if e != "" {
			m["code"] = "recall_refused"
		}
		writeJSON(w, 200, m)
	})
}
