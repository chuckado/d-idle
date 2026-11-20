package watch

import (
	"net/http"
	"sync"
)

// The item labels route (docs/video-feed.md, "Item labels"): GET /api/labels on the viewer shows whether the game draws
// the ground item names (D2's Alt, "Show Items"), POST /api/labels {on} on the answer listener sets it. The desk is
// late-bound like the pause desk; without one both answer 503 "off".

// LabelsRoute serves both methods; the viewer mux only the GET, the answer mux only the POST.
const LabelsRoute = "/api/labels"

// LabelsState is the reply of both methods. On is what the game reported back (not what was asked); OK is false when it
// could not be read or set (Code and Reason say why: "unsupported" for a d2d without the verb, "busy", "game").
// Available false hides the button's effect: the fork cannot do it.
type LabelsState struct {
	OK        bool   `json:"ok"`
	On        bool   `json:"on"`
	Available bool   `json:"available"`
	Code      string `json:"code,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// LabelsDesk is what the route needs from the bot side (cmd/d2bot/labelswire.go).
type LabelsDesk interface {
	Get() LabelsState
	Set(on bool) LabelsState
}

// LabelsAPI holds the desk once it exists.
type LabelsAPI struct {
	mu sync.Mutex
	d  LabelsDesk
}

func NewLabelsAPI() *LabelsAPI { return &LabelsAPI{} }

func (a *LabelsAPI) Bind(d LabelsDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *LabelsAPI) desk() LabelsDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET.
func (a *LabelsAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, LabelsState{Code: "off", Reason: "item labels are not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.Get())
	}))
}

// PostHandler serves POST /api/labels {on: true|false}; the field is required.
func (a *LabelsAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			On *bool `json:"on"`
		}
		if !policyGate(o, w, r, &b) {
			return
		}
		if b.On == nil {
			writeJSON(w, http.StatusBadRequest, LabelsState{Code: "body", Reason: "on is required"})
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, LabelsState{Code: "off", Reason: "item labels are not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.Set(*b.On))
	})
}
