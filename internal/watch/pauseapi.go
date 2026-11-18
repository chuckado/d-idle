package watch

import (
	"net/http"
	"sync"
)

// The Pause route (docs/pause.md): GET /api/pause on the viewer shows the player's pause (running, going_home, paused),
// POST /api/pause {on} on the answer listener sets it. The desk is late-bound like the repair desk's; without one (no
// runner yet, or a measured run) both answer 503 "off".

// PauseRoute serves both methods; the viewer mux only the GET, the answer mux only the POST.
const PauseRoute = "/api/pause"

// PauseDesk is what the route needs from the bot side (bot.Runner through pauseAdapter in cmd/d2bot).
type PauseDesk interface {
	View() any
	Set(on bool) any
}

// PauseModeDesk is the optional second face of a desk: SetMode picks between the frozen pause and the older walk-home
// pause (freeze nil: the desk decides, freeze when it can). A desk without it only takes Set.
type PauseModeDesk interface {
	SetMode(on bool, freeze *bool) any
}

// PauseAPI holds the desk once it exists.
type PauseAPI struct {
	mu sync.Mutex
	d  PauseDesk
}

func NewPauseAPI() *PauseAPI { return &PauseAPI{} }

func (a *PauseAPI) Bind(d PauseDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *PauseAPI) desk() PauseDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET.
func (a *PauseAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "pause is not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	}))
}

// PostHandler serves POST /api/pause {on: true|false}; the field is required.
func (a *PauseAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			On     *bool `json:"on"`
			Freeze *bool `json:"freeze"` // optional: true freeze (refused where the game cannot freeze), false walk home first
		}
		if !policyGate(o, w, r, &b) {
			return
		}
		if b.On == nil {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: "on is required"})
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "pause is not available"})
			return
		}
		if md, ok := d.(PauseModeDesk); ok && b.Freeze != nil {
			writeJSON(w, 200, pauseAnswer(md.SetMode(*b.On, b.Freeze)))
			return
		}
		writeJSON(w, 200, pauseAnswer(d.Set(*b.On)))
	})
}

// pauseAnswer adds the "ok" flag every command route's reply carries: the Orders panel (POST /api/commands relays this
// reply) reads "ok" to tell an accepted command from a refusal, and a view without it was shown as "could not do that"
// although the game had frozen (live bug 2026-11-16). ok is false only when the view carries an error; the view's own
// fields stay as they were.
func pauseAnswer(v any) any {
	var m map[string]any
	if !decode(marshal(v), &m) || m == nil {
		return v
	}
	e, _ := m["error"].(string)
	m["ok"] = e == ""
	if e != "" {
		m["code"] = "pause_refused"
	}
	return m
}
