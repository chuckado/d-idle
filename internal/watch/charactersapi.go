package watch

import (
	"net/http"
	"sync"

	"d2lod-bot/internal/charsel"
)

// The Characters routes (docs/charselect.md, watch.characters/1): GET /api/characters on the viewer lists the save
// folder; POST on the answer listener creates, switches, deletes (archive first) and restores. The desk is late-bound;
// without one all answer 503 "off". A refusal the desk names is its own status with {"code","error"}.

const (
	CharactersRoute        = "/api/characters"
	CharactersSwitchRoute  = "/api/characters/switch"
	CharactersDeleteRoute  = "/api/characters/delete"
	CharactersRestoreRoute = "/api/characters/restore"
	CharactersRulesRoute   = "/api/characters/rules" // GET: the class table and the name rule (no desk needed)
)

// CharactersDesk is what the routes need from the backend (charsel.Manager implements it; fakes in tests).
type CharactersDesk interface {
	List() (any, *charsel.Refusal)
	Create(name, class string, hardcore bool) (any, *charsel.Refusal)
	Switch(name string) (any, *charsel.Refusal)
	Delete(name, confirm string) (any, *charsel.Refusal)
	Restore(archive string) (any, *charsel.Refusal)
}

type CharactersAPI struct {
	mu sync.Mutex
	d  CharactersDesk
}

func NewCharactersAPI() *CharactersAPI { return &CharactersAPI{} }

func (a *CharactersAPI) Bind(d CharactersDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *CharactersAPI) desk() CharactersDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

func charReply(w http.ResponseWriter, v any, no *charsel.Refusal) {
	if no != nil {
		writeJSON(w, no.Status, policyReply{Code: no.Code, Error: no.Error})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, v)
}

// Handler serves the viewer GET.
func (a *CharactersAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "characters are not available"})
			return
		}
		v, no := d.List()
		charReply(w, v, no)
	}))
}

// RulesHandler serves the page's rules; they are static, so it answers even before a desk is bound.
func (a *CharactersAPI) RulesHandler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, charsel.Rules())
	}))
}

// PostHandler serves one POST route of the four; each has its own known fields.
func (a *CharactersAPI) PostHandler(o AnswerOptions, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Name     string `json:"name"`
			Class    string `json:"class"`
			Hardcore bool   `json:"hardcore"`
			Confirm  string `json:"confirm"`
			Archive  string `json:"archive"`
		}
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "characters are not available"})
			return
		}
		// A field that does not belong to the verb is a mistake, not a hint.
		bad := func(fields ...string) bool {
			for _, f := range fields {
				if f != "" {
					writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: "unexpected field for this route"})
					return true
				}
			}
			return false
		}
		switch route {
		case CharactersRoute:
			if bad(b.Confirm, b.Archive) {
				return
			}
			v, no := d.Create(b.Name, b.Class, b.Hardcore)
			charReply(w, v, no)
		case CharactersSwitchRoute:
			if bad(b.Class, b.Confirm, b.Archive) || b.Hardcore {
				return
			}
			v, no := d.Switch(b.Name)
			charReply(w, v, no)
		case CharactersDeleteRoute:
			if bad(b.Class, b.Archive) || b.Hardcore {
				return
			}
			v, no := d.Delete(b.Name, b.Confirm)
			charReply(w, v, no)
		case CharactersRestoreRoute:
			if bad(b.Name, b.Class, b.Confirm) || b.Hardcore {
				return
			}
			v, no := d.Restore(b.Archive)
			charReply(w, v, no)
		}
	})
}

// CharactersPostRoutes are the four POST paths.
var CharactersPostRoutes = []string{CharactersRoute, CharactersSwitchRoute, CharactersDeleteRoute, CharactersRestoreRoute}
