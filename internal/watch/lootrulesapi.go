package watch

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"

	"d2lod-bot/internal/lootrules"
)

// The Loot Rules routes (docs/lootrules.md section 6, wave 2a): GET /api/lootrules (the active rule set, plain-language
// lines, presets, the recent decisions; schema watch.lootrules/1) on the viewer, PUT /api/lootrules on the answer
// listener with the same gates as the other control POSTs (origin list, token, JSON only) and a larger body limit,
// because a rule set is bigger than an answer. Late-bound, 503 "off" until the bot side exists.

var (
	LootRulesRoutes    = []string{"/api/lootrules"}
	LootRulesPutRoutes = []string{"/api/lootrules"}
	lootRulesMaxBody   = int64(256 << 10)
)

// LootRulesDesk is what the routes need from the bot side (cmd/d2bot: the lootrules.Store plus the shadow's recent list).
type LootRulesDesk interface {
	View() any
	Put(body []byte) (any, error) // lootrules.Errors, lootrules.ErrConflict or another error
}

// LootRulesAPI holds the desk once it exists.
type LootRulesAPI struct {
	mu sync.Mutex
	d  LootRulesDesk
}

func NewLootRulesAPI() *LootRulesAPI { return &LootRulesAPI{} }

func (a *LootRulesAPI) Bind(d LootRulesDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *LootRulesAPI) desk() LootRulesDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET route.
func (a *LootRulesAPI) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/lootrules", func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the loot rules are not available yet"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	})
	return readOnly(m)
}

type lootRulesReply struct {
	OK     bool              `json:"ok"`
	Code   string            `json:"code,omitempty"`
	Error  string            `json:"error,omitempty"`
	Seq    int               `json:"seq,omitempty"`
	Errors []lootrules.Error `json:"errors,omitempty"`
}

// PutHandler serves PUT /api/lootrules on the answer listener: 200 {ok, seq}, 400 {errors:[{path,msg,line,col}]},
// 409 {seq} when the stored set changed since the client read it.
func (a *LootRulesAPI) PutHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := origin != "" && containsStr(o.Origins, origin)
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		fail := func(code int, c, msg string) { writeJSON(w, code, lootRulesReply{Code: c, Error: msg}) }
		switch {
		case r.Method == http.MethodOptions && allowed:
			w.Header().Set("Access-Control-Allow-Methods", "PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		case r.Method != http.MethodPut:
			w.Header().Set("Allow", "PUT, OPTIONS")
			fail(http.StatusMethodNotAllowed, "method", "PUT only")
			return
		case origin != "" && !allowed:
			fail(http.StatusForbidden, "origin", "origin not allowed")
			return
		case !o.tokenOK(r):
			fail(http.StatusUnauthorized, "token", "unauthorized")
			return
		}
		if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
			fail(http.StatusUnsupportedMediaType, "content-type", "send application/json")
			return
		}
		b, err := io.ReadAll(io.LimitReader(r.Body, lootRulesMaxBody+1))
		if err != nil || int64(len(b)) > lootRulesMaxBody {
			fail(http.StatusRequestEntityTooLarge, "body", "the rule set is too large")
			return
		}
		d := a.desk()
		if d == nil {
			fail(http.StatusServiceUnavailable, "off", "the loot rules are not available yet")
			return
		}
		res, err := d.Put(b)
		var (
			es lootrules.Errors
			cf lootrules.ErrConflict
		)
		switch {
		case err == nil:
			seq := 0
			if rs, ok := res.(*lootrules.RuleSet); ok {
				seq = rs.Seq
			}
			writeJSON(w, 200, lootRulesReply{OK: true, Seq: seq})
		case errors.As(err, &es):
			writeJSON(w, http.StatusBadRequest, lootRulesReply{Code: "invalid", Error: "the rule set has errors", Errors: es})
		case errors.As(err, &cf):
			writeJSON(w, http.StatusConflict, lootRulesReply{Code: "conflict", Error: cf.Error(), Seq: cf.Seq})
		default:
			fail(http.StatusInternalServerError, "save", "could not save the rule set")
		}
	})
}
