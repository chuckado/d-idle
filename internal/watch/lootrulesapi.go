package watch

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"sync"

	"d2lod-bot/internal/lootrules"
)

// The Loot Rules routes (docs/lootrules.md section 6, wave 2a): GET /api/lootrules (the active rule set, plain-language
// lines, presets, the recent decisions; schema watch.lootrules/1) on the viewer, PUT /api/lootrules on the answer
// listener with the same gates as the other control POSTs (origin list, token, JSON only) and a larger body limit,
// because a rule set is bigger than an answer. Late-bound, 503 "off" until the bot side exists.

var (
	LootRulesRoutes     = []string{"/api/lootrules", "/api/lootrules/export"}
	LootRulesPutRoutes  = []string{"/api/lootrules"}
	LootRulesPostRoutes = []string{"/api/lootrules/preview", "/api/lootrules/undo", "/api/lootrules/import"}
	lootRulesMaxBody    = int64(256 << 10)
)

// LootRulesDesk is what the routes need from the bot side (cmd/d2bot: the lootrules.Store plus the shadow's recent list).
type LootRulesDesk interface {
	View() any
	Put(body []byte) (any, error) // lootrules.Errors, lootrules.ErrConflict or another error
}

// LootRulesTools are the wave 2b routes; a desk that implements them gets them (the others answer 404).
// Preview decides the pack and stash with the body's rule set (a PUT body, nothing is saved) and acts on nothing; Undo goes
// back to the previous version; Import replaces the set with an exported file; Export is the active set as a file.
type LootRulesTools interface {
	Preview(body []byte) (any, error)
	Undo(seq int) (any, error)
	Import(body []byte, seq int) (any, error)
	Export() ([]byte, error)
}

// LootRulesAPI holds the desk once it exists.
type LootRulesAPI struct {
	mu sync.Mutex
	d  LootRulesDesk
}

func NewLootRulesAPI() *LootRulesAPI { return &LootRulesAPI{} }

func (a *LootRulesAPI) Bind(d LootRulesDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *LootRulesAPI) tools() (LootRulesTools, bool) {
	t, ok := a.desk().(LootRulesTools)
	return t, ok
}

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
	m.HandleFunc("/api/lootrules/export", func(w http.ResponseWriter, r *http.Request) {
		t, ok := a.tools()
		if !ok {
			writeJSON(w, http.StatusNotFound, policyReply{Code: "off", Error: "export is not available"})
			return
		}
		b, err := t.Export()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, policyReply{Code: "export", Error: "could not export the rule set"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="lootrules.json"`)
		_, _ = w.Write(append(b, '\n'))
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

// gate runs the gates every control write shares (CORS preflight, method, origin, token, JSON, body size) and returns the
// body, or ok=false after answering.
func (a *LootRulesAPI) gate(o AnswerOptions, method string, w http.ResponseWriter, r *http.Request) (b []byte, ok bool) {
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
		w.Header().Set("Access-Control-Allow-Methods", method)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return nil, false
	case r.Method != method:
		w.Header().Set("Allow", method+", OPTIONS")
		fail(http.StatusMethodNotAllowed, "method", method+" only")
		return nil, false
	case origin != "" && !allowed:
		fail(http.StatusForbidden, "origin", "origin not allowed")
		return nil, false
	case !o.tokenOK(r):
		fail(http.StatusUnauthorized, "token", "unauthorized")
		return nil, false
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		fail(http.StatusUnsupportedMediaType, "content-type", "send application/json")
		return nil, false
	}
	b, err := io.ReadAll(io.LimitReader(r.Body, lootRulesMaxBody+1))
	if err != nil || int64(len(b)) > lootRulesMaxBody {
		fail(http.StatusRequestEntityTooLarge, "body", "the rule set is too large")
		return nil, false
	}
	return b, true
}

// reply turns a desk result into the answer: 200, 400 with the errors, 409 with the stored seq.
func (a *LootRulesAPI) reply(w http.ResponseWriter, res any, err error, ok func(any) any) {
	var (
		es lootrules.Errors
		cf lootrules.ErrConflict
	)
	switch {
	case err == nil:
		writeJSON(w, 200, ok(res))
	case errors.As(err, &es):
		writeJSON(w, http.StatusBadRequest, lootRulesReply{Code: "invalid", Error: "the rule set has errors", Errors: es})
	case errors.As(err, &cf):
		writeJSON(w, http.StatusConflict, lootRulesReply{Code: "conflict", Error: cf.Error(), Seq: cf.Seq})
	case errors.Is(err, lootrules.ErrNoUndo):
		writeJSON(w, http.StatusConflict, lootRulesReply{Code: "no-undo", Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, lootRulesReply{Code: "save", Error: "could not complete that"})
	}
}

func seqOf(res any) any {
	seq := 0
	if rs, ok := res.(*lootrules.RuleSet); ok {
		seq = rs.Seq
	}
	return lootRulesReply{OK: true, Seq: seq}
}

// PostHandler serves the wave 2b writes on the answer listener, same gates as PUT: POST /api/lootrules/preview (body: a
// PUT body; answers the dry run, saves nothing), /undo (body {"seq":N}) and /import (the exported file as the body, the
// version you were looking at in the query, ?seq=N).
func (a *LootRulesAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := a.gate(o, http.MethodPost, w, r)
		if !ok {
			return
		}
		t, ok := a.tools()
		if !ok {
			writeJSON(w, http.StatusServiceUnavailable, lootRulesReply{Code: "off", Error: "the loot rules are not available yet"})
			return
		}
		switch r.URL.Path {
		case "/api/lootrules/preview":
			res, err := t.Preview(b)
			a.reply(w, res, err, func(v any) any { return v })
		case "/api/lootrules/undo":
			var in struct {
				Seq int `json:"seq"`
			}
			if err := json.Unmarshal(b, &in); err != nil {
				writeJSON(w, http.StatusBadRequest, lootRulesReply{Code: "invalid", Error: "send {\"seq\": N}"})
				return
			}
			res, err := t.Undo(in.Seq)
			a.reply(w, res, err, seqOf)
		case "/api/lootrules/import":
			seq, err := strconv.Atoi(r.URL.Query().Get("seq"))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, lootRulesReply{Code: "invalid", Error: "say which version you are replacing: ?seq=N"})
				return
			}
			res, err := t.Import(b, seq)
			a.reply(w, res, err, seqOf)
		default:
			http.NotFound(w, r)
		}
	})
}

// PutHandler serves PUT /api/lootrules on the answer listener: 200 {ok, seq}, 400 {errors:[{path,msg,line,col}]},
// 409 {seq} when the stored set changed since the client read it.
func (a *LootRulesAPI) PutHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := a.gate(o, http.MethodPut, w, r)
		if !ok {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, lootRulesReply{Code: "off", Error: "the loot rules are not available yet"})
			return
		}
		res, err := d.Put(b)
		a.reply(w, res, err, seqOf)
	})
}
