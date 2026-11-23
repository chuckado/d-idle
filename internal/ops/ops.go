// Package ops is the operator control of the quest board: a local unix socket through which someone who can already
// reach the container (docker exec) presses the same board verbs the viewer buttons press, without the viewer's
// passcode. Off by default, quest verbs only, every mutating verb audited and rate limited. docs/operator.md.
package ops

import (
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/questboard"
)

// SockName is the socket file inside the instance state directory.
const SockName = "operator.sock"

const (
	MaxWhy       = 200
	PerQuestHour = 6
	PerTotalHour = 30
	RetryCool    = 60 * time.Second
)

// The allowlist. Nothing else is dispatched, whatever the request says. mode and dismiss are board verbs too, but the
// operator does not get them: they change what the player sees or how the bot plays, not a quest.
var mutating = map[string]bool{"start": true, "retry": true, "cancel": true, "handin": true, "auto": true}

func allowed(v string) bool { return mutating[v] || v == "status" }

// Desk is what the operator verbs need from the board (questboard.Desk; the same calls the HTTP handlers make).
type Desk interface {
	View() questboard.View
	Start(quest, chunk string) questboard.Reply
	HandIn(quest, chunk string) questboard.Reply
	Cancel(quest, chunk string) questboard.Reply
	Retry(quest, chunk string) questboard.Reply
	SetAuto(quest string, on bool) questboard.Reply
	Record(questboard.AuditEvent)
}

// Request is one line of JSON on the socket.
type Request struct {
	Verb  string `json:"verb"`
	Quest string `json:"quest,omitempty"`
	Chunk string `json:"chunk,omitempty"`
	On    *bool  `json:"on,omitempty"`
	Why   string `json:"why,omitempty"`
}

// Response is the one line answered. Reply is the desk's own answer of a mutating verb; Board answers status.
type Response struct {
	OK     bool              `json:"ok"`
	Code   string            `json:"code,omitempty"`
	Reason string            `json:"reason,omitempty"`
	Reply  *questboard.Reply `json:"reply,omitempty"`
	Board  *Board            `json:"board,omitempty"`
}

// Board is the status view: one row per chunk of the difficulty.
type Board struct {
	Enabled    bool   `json:"enabled"`
	Difficulty string `json:"difficulty,omitempty"`
	Why        string `json:"why,omitempty"`
	Rows       []Row  `json:"rows"`
}

// Row is a chunk with its state, its park reason (Code and Text) and its step while it runs.
type Row struct {
	Quest string `json:"quest"`
	Chunk string `json:"chunk"`
	State string `json:"state"`
	Step  string `json:"step,omitempty"`
	// Progress counts the changes of state or step the server has seen for this chunk since it started; a wait that
	// times out with a moved counter is slow progress, not a stall.
	Progress int    `json:"progress,omitempty"`
	Code     string `json:"code,omitempty"`
	Text     string `json:"text,omitempty"`
}

func boardOf(v questboard.View) *Board {
	b := &Board{Enabled: v.Enabled, Difficulty: v.Difficulty, Why: v.Reason, Rows: []Row{}}
	runs := []*questboard.RunView{v.Running, v.Awaiting, v.Blocked}
	for _, q := range v.Quests {
		for _, c := range q.Chunks {
			r := Row{Quest: q.ID, Chunk: c.ID, State: string(c.State)}
			if c.Reason != nil {
				r.Code, r.Text = c.Reason.Code, c.Reason.Text
			}
			for _, rv := range runs {
				if rv == nil || rv.Quest != q.ID || rv.Chunk != c.ID {
					continue
				}
				if rv.Step != "" {
					r.Step = rv.Step
				}
				if r.Code == "" && rv.Reason != nil {
					r.Code, r.Text = rv.Reason.Code, rv.Reason.Text
				}
			}
			b.Rows = append(b.Rows, r)
		}
	}
	return b
}

// Enabled says whether the operator socket may exist: only with D2BOT_OPERATOR=1 and never in a measured run.
func Enabled(getenv func(string) string) (bool, string) {
	if getenv("D2BOT_OPERATOR") != "1" {
		return false, "operator control is off (D2BOT_OPERATOR is not 1)"
	}
	if getenv("D2BOT_RUN") != "" {
		return false, "operator control is off in a measured run (D2BOT_RUN is set)"
	}
	return true, ""
}

// Config wires a Server. Getenv and Now default to the process; Logf may be nil.
type Config struct {
	Desk   func() Desk
	Getenv func(string) string
	Now    func() time.Time
	Logf   func(msg string, kv ...any)
}

// Server dispatches requests. Safe for concurrent use.
type Server struct {
	cfg  Config
	mu   sync.Mutex
	all  []time.Time
	per  map[string][]time.Time
	last map[string]time.Time // last retry per quest
	prog map[string]*progress
}

func New(c Config) *Server {
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.Getenv == nil {
		c.Getenv = func(string) string { return "" }
	}
	if c.Logf == nil {
		c.Logf = func(string, ...any) {}
	}
	return &Server{cfg: c, per: map[string][]time.Time{}, last: map[string]time.Time{}, prog: map[string]*progress{}}
}

func no(code, reason string) Response { return Response{Code: code, Reason: reason} }

func safeID(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func cleanWhy(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > MaxWhy {
		return "", false
	}
	for _, c := range s {
		if c < ' ' || c == 0x7f {
			return "", false
		}
	}
	return s, true
}

// admit applies the rate limits and books the verb on success.
func (s *Server) admit(verb, quest string) *Response {
	now := s.cfg.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	prune := func(l []time.Time) []time.Time {
		i := 0
		for i < len(l) && now.Sub(l[i]) >= time.Hour {
			i++
		}
		return l[i:]
	}
	s.all = prune(s.all)
	s.per[quest] = prune(s.per[quest])
	if t, ok := s.last[quest]; ok && (verb == "retry" || verb == "start") && now.Sub(t) < RetryCool {
		r := no("cooldown", "a retry of this quest was pressed less than a minute ago; wait "+(RetryCool-now.Sub(t)).Round(time.Second).String())
		return &r
	}
	if len(s.per[quest]) >= PerQuestHour {
		r := no("rate_limited", "this quest already had 6 operator verbs in the last hour")
		return &r
	}
	if len(s.all) >= PerTotalHour {
		r := no("rate_limited", "30 operator verbs in the last hour is the limit for the whole board")
		return &r
	}
	s.all = append(s.all, now)
	s.per[quest] = append(s.per[quest], now)
	if verb == "retry" {
		s.last[quest] = now
	}
	return nil
}

// Handle runs one request. Every refusal before the desk is logged, not written to the audit ring (a runaway loop
// must not push real events out of it); everything the desk ran is audited with remote "operator".
func (s *Server) Handle(rq Request) Response {
	rsp := s.handle(rq)
	if !rsp.OK && rsp.Reply == nil {
		s.cfg.Logf("operator verb refused", "verb", rq.Verb, "quest", rq.Quest, "code", rsp.Code)
	}
	return rsp
}

func (s *Server) handle(rq Request) Response {
	if !allowed(rq.Verb) {
		return no("verb", "only start, retry, cancel, handin, auto and status are served here")
	}
	var d Desk
	if s.cfg.Desk != nil {
		d = s.cfg.Desk()
	}
	if d == nil {
		return no("off", "the quest board is not available yet")
	}
	if rq.Verb == "status" {
		b := boardOf(d.View())
		s.track(b)
		return Response{OK: true, Board: b}
	}
	if s.cfg.Getenv("D2BOT_RUN") != "" {
		return no("measured_run", "a measured run never takes operator verbs (D2BOT_RUN is set)")
	}
	if !safeID(rq.Quest) {
		return no("quest", "send a quest id (a-z, 0-9, _ and -, at most 64)")
	}
	if rq.Chunk != "" && !safeID(rq.Chunk) {
		return no("chunk", "the chunk id has characters a chunk never has")
	}
	why, ok := cleanWhy(rq.Why)
	if !ok {
		return no("why", "every mutating verb needs a reason: one line, 1 to 200 characters")
	}
	if rq.Verb == "auto" && rq.On == nil {
		return no("body", "auto needs on or off")
	}
	if rq.Verb == "retry" {
		if r := notParked(d.View(), rq.Quest, rq.Chunk); r != nil {
			return *r
		}
	}
	if r := s.admit(rq.Verb, rq.Quest); r != nil {
		return *r
	}
	var rep questboard.Reply
	switch rq.Verb {
	case "start":
		rep = d.Start(rq.Quest, rq.Chunk)
	case "handin":
		rep = d.HandIn(rq.Quest, rq.Chunk)
	case "cancel":
		rep = d.Cancel(rq.Quest, rq.Chunk)
	case "retry":
		rep = d.Retry(rq.Quest, rq.Chunk)
	case "auto":
		rep = d.SetAuto(rq.Quest, *rq.On)
	}
	st := rep.Status
	if st == 0 {
		st = 200
	}
	d.Record(questboard.AuditEvent{Action: rq.Verb, Quest: rq.Quest, Chunk: rq.Chunk, Path: "operator/" + rq.Verb,
		Remote: "operator", Agent: "ops-cli", Why: why, OK: rep.OK, Already: rep.Already, Status: st, Code: rep.Code, Reason: rep.Reason})
	s.cfg.Logf("operator verb", "verb", rq.Verb, "quest", rq.Quest, "chunk", rq.Chunk, "ok", rep.OK, "code", rep.Code)
	return Response{OK: rep.OK, Code: rep.Code, Reason: rep.Reason, Reply: &rep}
}

type progress struct {
	sig string
	n   int
}

// track numbers the state and step changes of each chunk across status calls.
func (s *Server) track(b *Board) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range b.Rows {
		r := &b.Rows[i]
		k := r.Quest + "/" + r.Chunk
		p := s.prog[k]
		if p == nil {
			p = &progress{sig: r.State + "|" + r.Step}
			s.prog[k] = p
		}
		if sig := r.State + "|" + r.Step; sig != p.sig {
			p.sig, p.n = sig, p.n+1
		}
		r.Progress = p.n
	}
}

// notParked refuses a retry unless a matching chunk of the quest is blocked: retry clears a park, and on a chunk that
// runs it would disturb the run (it put a running chunk into "trying again" live).
func notParked(v questboard.View, quest, chunk string) *Response {
	state := ""
	for _, q := range v.Quests {
		if q.ID != quest {
			continue
		}
		for _, c := range q.Chunks {
			if chunk != "" && c.ID != chunk {
				continue
			}
			if c.State == questboard.Blocked {
				return nil
			}
			if state == "" || c.State == questboard.InProgress || c.State == questboard.Accepted {
				state = string(c.State)
			}
		}
	}
	if state == "" {
		state = "not on the board"
	}
	r := no("not_parked", "retry only clears a park; "+quest+" is "+state+", not blocked")
	return &r
}
