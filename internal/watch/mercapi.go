package watch

import (
	"errors"
	"net/http"
	"sync"
)

// The mercenary routes (docs/merc-revive.md): GET /api/merc (state, price, gold, the Revive mercenary button and the
// open request, schema watch.merc/1) on the viewer, POST /api/merc/revive and /api/merc/cancel on the answer listener
// with the same gates as the repair and inventory POSTs. The desk is late-bound, 503 off until then.

var (
	MercRoutes     = []string{"/api/merc"}
	MercPostRoutes = []string{"/api/merc/revive", "/api/merc/cancel"}
)

// MercDesk is what the routes need from the bot side (merc.PlayerRevive through mercAPIAdapter in cmd/d2bot).
type MercDesk interface {
	View() any
	Request() (any, error)
	Cancel() error
}

// MercAPI holds the desk once it exists.
type MercAPI struct {
	mu sync.Mutex
	d  MercDesk
}

func NewMercAPI() *MercAPI { return &MercAPI{} }

func (a *MercAPI) Bind(d MercDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *MercAPI) desk() MercDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET route.
func (a *MercAPI) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/merc", func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the mercenary is not available yet"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	})
	return readOnly(m)
}

// PostHandler serves the two POST routes.
func (a *MercAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct{}
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the mercenary is not available yet"})
			return
		}
		var (
			err error
			res any
		)
		switch r.URL.Path {
		case "/api/merc/revive":
			res, err = d.Request()
		case "/api/merc/cancel":
			err = d.Cancel()
		default:
			http.NotFound(w, r)
			return
		}
		var ref invRefusal
		switch {
		case err == nil:
			writeJSON(w, 200, map[string]any{"ok": true, "request": res})
		case errors.As(err, &ref):
			writeJSON(w, http.StatusConflict, policyReply{Code: ref.Kind(), Error: ref.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, policyReply{Code: "internal", Error: "could not save the request"})
		}
	})
}
