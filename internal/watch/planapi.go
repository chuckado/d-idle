package watch

import (
	"net/http"
	"sync"
)

// PlanRoute is GET /api/hero/plan (docs/agency.md, roadmap R7 agency P1): the rotation the planner derives from the
// skills the hero actually has, its warnings and the viability line. Read-only and shadow: the hero does not play by it.
const PlanRoute = "/api/hero/plan"

// PlanAPI holds the view source once it exists (late-bound like the points desk). The source returns a value that is
// marshalled as JSON (agency.View); the watch package does not import the planner.
type PlanAPI struct {
	mu  sync.Mutex
	get func() any
}

func NewPlanAPI() *PlanAPI { return &PlanAPI{} }

// Bind sets the source; a nil source answers the "off" reply.
func (a *PlanAPI) Bind(get func() any) { a.mu.Lock(); a.get = get; a.mu.Unlock() }

func (a *PlanAPI) source() func() any { a.mu.Lock(); defer a.mu.Unlock(); return a.get }

// Handler serves the viewer GET.
func (a *PlanAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PlanRoute {
			http.NotFound(w, r)
			return
		}
		get := a.source()
		if get == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the planner is not available yet"})
			return
		}
		writeJSON(w, http.StatusOK, get())
	}))
}
