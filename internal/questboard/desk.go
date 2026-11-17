package questboard

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// The desk (slice S4): the API-shaped face of the Exec for the watch routes (internal/watch/questsapi.go). Every
// action answers a Reply that carries `ok`, a refusal code and a reason in words (the Pause bug lesson: a reply
// without them was shown as "could not do that"); the page reads the result back from View, never from the reply.
// A hand-in is asynchronous (the dialogue runs on the bot loop), so its reply is 202 and the progress is in View.

const (
	SchemaQuests = "watch.quests/1"

	// Refusal codes of Reply.Code.
	CodeOff       = "off"
	CodeNoHero    = "no_hero"
	CodePaused    = "paused"
	CodeBusy      = "busy"
	CodeNotReady  = "not_ready"
	CodeNotAwait  = "not_awaiting"
	CodeUnknown   = "unknown"
	CodeDone      = "done"
	CodeBadMode   = "bad_mode"
	CodeInternal  = "internal"
	viewCacheTTL  = time.Second
	handInTimeout = 5 * time.Minute
	historyMax    = 20
)

// DeskConfig is what the desk needs from the bot side. Every field but Exec may be nil.
type DeskConfig struct {
	// Exec returns the executor and the plan store of the loaded hero; a nil Exec means no hero yet.
	Exec func() (*Exec, *Store)
	// Paused reports the player's Pause.
	Paused func() bool
	// Explain is the full readiness (reasons, legacy floor, sentence) of a chunk; without it the offers carry the
	// word and the legacy floor only.
	Explain func(q Quest, c Chunk) (Readiness, bool)
	// Class is the hero's class, for the legacy floor.
	Class func() string
	Now   func() time.Time
	// Ctx bounds the asynchronous hand-in goroutines (the process); default Background.
	Ctx context.Context
}

// Desk serves the API of one process. It is safe for concurrent use.
type Desk struct {
	b   *Board
	cfg DeskConfig

	mu     sync.Mutex
	cache  *View
	cached time.Time
	hand   map[string]*handRun // by chunk key
	last   *HandInView
}

type handRun struct{ started time.Time }

// NewDesk builds the desk over a board.
func NewDesk(b *Board, cfg DeskConfig) *Desk {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Ctx == nil {
		cfg.Ctx = context.Background()
	}
	return &Desk{b: b, cfg: cfg, hand: map[string]*handRun{}}
}

// Reply is the answer of every action.
type Reply struct {
	OK      bool   `json:"ok"`
	Pending bool   `json:"pending,omitempty"` // the action runs on; read the result from GET /api/quests
	Already bool   `json:"already,omitempty"` // a repeated press: nothing changed
	Code    string `json:"code,omitempty"`    // refusal code
	Reason  string `json:"reason,omitempty"`  // the refusal (or the news) in words
	Quest   string `json:"quest,omitempty"`
	Chunk   string `json:"chunk,omitempty"`
	State   string `json:"state,omitempty"` // the chunk's state after the call
	Mode    string `json:"mode,omitempty"`
	Status  int    `json:"-"` // HTTP status the route uses
}

func refuse(status int, code, format string, a ...any) Reply {
	return Reply{Code: code, Reason: fmt.Sprintf(format, a...), Status: status}
}

// ---- views -----------------------------------------------------------------------------------------------

// View is GET /api/quests.
type View struct {
	Schema     string       `json:"schema"`
	Enabled    bool         `json:"enabled"`
	Reason     string       `json:"reason,omitempty"` // why the board shows nothing (off, no hero)
	Difficulty string       `json:"difficulty,omitempty"`
	Paused     bool         `json:"paused,omitempty"`
	Mode       string       `json:"mode"`
	ModeActs   bool         `json:"modeActs"` // false until slice S6 implements the modes; the choice is stored
	AutoStart  AutoView     `json:"autoStart"`
	Offers     []OfferView  `json:"offers"`
	Quests     []QuestRow   `json:"quests"`
	Running    *RunView     `json:"running,omitempty"`
	Awaiting   *RunView     `json:"awaiting,omitempty"`
	Blocked    *RunView     `json:"blocked,omitempty"`
	HandIn     *HandInView  `json:"handIn,omitempty"`
	History    []HistoryRow `json:"history"`
	Needs      []NeedRow    `json:"needs"`
	Strip      *Strip       `json:"strip,omitempty"`
	Extra      []string     `json:"extra,omitempty"` // further accepted chunks in the plan file (never run)
	Generated  time.Time    `json:"generated"`
}

// AutoView is the auto-start block: a stub until slice S8 (the flags per quest are real, nothing acts on them).
type AutoView struct {
	Available bool   `json:"available"`
	Note      string `json:"note"`
}

// Pays is the meta payout of an offer; omitted until the vault wallet exists (slice S10).
type Pays struct {
	Meta int `json:"meta"`
}

// OfferView is one chunk that raises an offer.
type OfferView struct {
	Quest       string   `json:"quest"`
	QuestTitle  string   `json:"questTitle"`
	Chunk       string   `json:"chunk"`
	Title       string   `json:"title"`
	Act         int      `json:"act"`
	Risk        string   `json:"risk,omitempty"`
	Readiness   Word     `json:"readiness"`
	Reasons     []Reason `json:"reasons"`
	LegacyFloor int      `json:"legacyFloor,omitempty"`
	Sentence    string   `json:"sentence,omitempty"`
	Minutes     [2]int   `json:"minutes"`
	Startable   bool     `json:"startable"`
	Why         string   `json:"why,omitempty"` // why Start is refused now
	Auto        bool     `json:"auto"`
	Pays        *Pays    `json:"pays,omitempty"`
}

// QuestRow is one quest of the difficulty with its chunks (the board sheet).
type QuestRow struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	Act    int        `json:"act"`
	State  State      `json:"state"`
	Auto   bool       `json:"auto"`
	Chunks []ChunkRow `json:"chunks"`
}

// ChunkRow is one part of a quest.
type ChunkRow struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	State     State  `json:"state"`
	Readiness Word   `json:"readiness,omitempty"`
	Why       string `json:"why,omitempty"`
	Confirmed bool   `json:"confirmed,omitempty"`
}

// RunView is the accepted chunk: running, awaiting the tap or parked.
type RunView struct {
	Quest      string `json:"quest"`
	QuestTitle string `json:"questTitle"`
	Chunk      string `json:"chunk"`
	Title      string `json:"title"`
	State      State  `json:"state"`
	Step       string `json:"step,omitempty"`
	NPC        string `json:"npc,omitempty"` // who the hero walks to or waits at
	Why        string `json:"why,omitempty"`
	Failures   int    `json:"failures,omitempty"`
	Deaths     int    `json:"deaths,omitempty"`
}

// HandInView is the hand-in in flight or the last one that finished.
type HandInView struct {
	Quest   string `json:"quest"`
	Chunk   string `json:"chunk"`
	State   string `json:"state"` // running, done, confirming, blocked, failed
	Reason  string `json:"reason,omitempty"`
	Attempt int    `json:"attempts,omitempty"`
	At      string `json:"at,omitempty"`
}

// HistoryRow is a chunk the board ran and that is done now.
type HistoryRow struct {
	Quest     string    `json:"quest"`
	Chunk     string    `json:"chunk"`
	Title     string    `json:"title"`
	Accepted  time.Time `json:"accepted"`
	Confirmed bool      `json:"confirmed"`
}

// Verb is a button on a need row.
type Verb struct {
	Label string         `json:"label"`
	Post  string         `json:"post"`
	Body  map[string]any `json:"body,omitempty"`
}

// NeedRow is a Needs-you row (design 4.6), glance-ready.
type NeedRow struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Sev     string `json:"sev"`
	Text    string `json:"text"`
	Actions []Verb `json:"actions"`
}

// Strip is the video-card strip: the current chunk, step and NPC.
type Strip struct {
	Quest      string `json:"quest"`
	QuestTitle string `json:"questTitle"`
	Chunk      string `json:"chunk"`
	Title      string `json:"title"`
	State      State  `json:"state"`
	Step       string `json:"step,omitempty"`
	NPC        string `json:"npc,omitempty"`
}

// QuestRoute is the path of one quest's action.
func QuestRoute(quest, action string) string { return "/api/quests/" + quest + "/" + action }

func (d *Desk) exec() (*Exec, *Store) {
	if d.cfg.Exec == nil {
		return nil, nil
	}
	return d.cfg.Exec()
}

func (d *Desk) paused() bool { return d.cfg.Paused != nil && d.cfg.Paused() }

func (d *Desk) invalidate() { d.mu.Lock(); d.cache = nil; d.mu.Unlock() }

// View derives the board (cached for one second: the page polls, derive reads the game evidence).
func (d *Desk) View() View {
	d.mu.Lock()
	if d.cache != nil && d.cfg.Now().Sub(d.cached) < viewCacheTTL {
		v := *d.cache
		d.mu.Unlock()
		return v
	}
	d.mu.Unlock()
	v := d.build()
	d.mu.Lock()
	d.cache, d.cached = &v, d.cfg.Now()
	d.mu.Unlock()
	return v
}

func (d *Desk) build() View {
	v := View{Schema: SchemaQuests, Mode: DefaultMode, Generated: d.cfg.Now().UTC(),
		AutoStart: AutoView{Note: "Opt-in auto-start arrives with slice S8; the per-quest flag is stored but nothing acts on it."},
		Offers:    []OfferView{}, Quests: []QuestRow{}, History: []HistoryRow{}, Needs: []NeedRow{}}
	x, st := d.exec()
	if x == nil || st == nil {
		v.Reason = "No hero is loaded yet."
		return v
	}
	if m, err := st.Mode(); err == nil {
		v.Mode = m
	}
	v.Paused = d.paused()
	status, err := x.State()
	if err == ErrOff {
		v.Reason = "The quest board is off (knob quests.board is 0)."
		return v
	}
	if err != nil {
		v.Reason = "The quest board could not be read: " + err.Error()
		return v
	}
	v.Enabled = true
	v.Difficulty = Difficulties[status.Diff]
	plan, _, _ := st.Load()
	busy := status.Active != nil
	class := ""
	if d.cfg.Class != nil {
		class = d.cfg.Class()
	}
	for _, qv := range status.Views {
		q, qerr := d.b.Quest(qv.ID)
		if qerr != nil {
			continue
		}
		row := QuestRow{ID: qv.ID, Title: qv.Title, Act: q.Act, State: qv.State, Auto: plan.AutoStart(status.Diff, qv.ID), Chunks: []ChunkRow{}}
		for _, cv := range qv.Chunks {
			c, cerr := d.b.Chunk(qv.ID, cv.ID)
			if cerr != nil {
				continue
			}
			row.Chunks = append(row.Chunks, ChunkRow{ID: cv.ID, Title: c.Title, State: cv.State, Readiness: cv.Reason, Why: cv.Why, Confirmed: cv.Confirmed})
			if cv.State == Ready {
				v.Offers = append(v.Offers, d.offer(q, c, cv, plan.AutoStart(status.Diff, qv.ID), !busy, class, v.Paused))
			}
			if cv.State == Done && plan.accepted(status.Diff, qv.ID, cv.ID) {
				v.History = append(v.History, HistoryRow{Quest: qv.ID, Chunk: cv.ID, Title: c.Title, Accepted: acceptedAt(plan, status.Diff, qv.ID, cv.ID), Confirmed: cv.Confirmed})
			}
		}
		v.Quests = append(v.Quests, row)
	}
	// newest first, capped
	for i, j := 0, len(v.History)-1; i < j; i, j = i+1, j-1 {
		v.History[i], v.History[j] = v.History[j], v.History[i]
	}
	if len(v.History) > historyMax {
		v.History = v.History[:historyMax]
	}
	if a := status.Active; a != nil {
		rv := d.run(*a)
		switch {
		case a.State == Awaiting:
			v.Awaiting = &rv
		case a.State == Blocked:
			v.Blocked = &rv
		default:
			v.Running = &rv
		}
		v.Strip = &Strip{Quest: rv.Quest, QuestTitle: rv.QuestTitle, Chunk: rv.Chunk, Title: rv.Title, State: rv.State, Step: rv.Step, NPC: rv.NPC}
	}
	for _, r := range status.Extra {
		v.Extra = append(v.Extra, r.String())
	}
	d.mu.Lock()
	v.HandIn = d.last
	d.mu.Unlock()
	v.Needs = d.needs(v)
	return v
}

func acceptedAt(p Plan, diff int, quest, chunk string) time.Time {
	for _, r := range p.Accepted {
		if r.Quest == quest && r.Chunk == chunk && r.Diff == Difficulties[diff] {
			return r.At
		}
	}
	return time.Time{}
}

func (d *Desk) offer(q Quest, c Chunk, cv ChunkView, auto, startable bool, class string, paused bool) OfferView {
	risk := c.Risk
	if risk == "" {
		risk = q.Risk
	}
	o := OfferView{Quest: q.ID, QuestTitle: q.Title, Chunk: c.ID, Title: c.Title, Act: q.Act, Risk: risk,
		Readiness: cv.Reason, Reasons: []Reason{}, Minutes: c.Minutes, Startable: startable && !paused, Auto: auto}
	if class != "" {
		if f, ok, err := d.b.Floor(q, class); err == nil && ok {
			o.LegacyFloor = f
		}
	}
	if d.cfg.Explain != nil {
		if r, ok := d.cfg.Explain(q, c); ok {
			if r.Reasons != nil {
				o.Reasons = r.Reasons
			}
			o.Sentence = r.Sentence
			if r.LegacyFloor > 0 {
				o.LegacyFloor = r.LegacyFloor
			}
		}
	}
	switch {
	case paused:
		o.Why = "The bot is paused."
	case !startable:
		o.Why = "Another chunk is running; finish or stop it first."
	}
	return o
}

func (d *Desk) run(a Active) RunView {
	rv := RunView{Quest: a.Quest, Chunk: a.Chunk, State: a.State, Step: a.Step, Why: a.Why, Failures: a.Failures, Deaths: a.Deaths}
	if q, err := d.b.Quest(a.Quest); err == nil {
		rv.QuestTitle = q.Title
		if c, err := d.b.Chunk(a.Quest, a.Chunk); err == nil {
			rv.Title = c.Title
			switch {
			case c.Giver != nil:
				rv.NPC = c.Giver.NPC
			case q.Accept != nil:
				rv.NPC = q.Accept.NPC
			}
		}
	}
	return rv
}

// needs are the Needs-you rows (design 4.6): the hand-in the hero waits at, a parked chunk. Offers are raised by
// the page once per offer (S5); the "paused: you were away" row belongs to S8.
func (d *Desk) needs(v View) []NeedRow {
	var out []NeedRow
	if a := v.Awaiting; a != nil {
		who := "The quest giver"
		if a.NPC != "" {
			who = a.NPC
		}
		out = append(out, NeedRow{ID: "quest:handin:" + a.Quest + "/" + a.Chunk, Kind: "quest_handin", Sev: "event",
			Text:    fmt.Sprintf("%s is waiting: hand in %s", who, a.QuestTitle),
			Actions: []Verb{{Label: "Hand in", Post: QuestRoute(a.Quest, "handin"), Body: map[string]any{"chunk": a.Chunk}}}})
	}
	if b := v.Blocked; b != nil {
		why := b.Why
		if why == "" {
			why = "the step failed"
		}
		out = append(out, NeedRow{ID: "quest:blocked:" + b.Quest + "/" + b.Chunk, Kind: "quest_blocked", Sev: "warn",
			Text: fmt.Sprintf("%s is stuck: %s", b.QuestTitle, why),
			Actions: []Verb{
				{Label: "Try again", Post: QuestRoute(b.Quest, "retry"), Body: map[string]any{"chunk": b.Chunk}},
				{Label: "Stop", Post: QuestRoute(b.Quest, "cancel"), Body: map[string]any{"chunk": b.Chunk}}}})
	}
	if h := v.HandIn; h != nil && (h.State == "blocked" || h.State == "failed") && v.Blocked == nil {
		out = append(out, NeedRow{ID: "quest:handin-failed:" + h.Quest + "/" + h.Chunk, Kind: "quest_handin_failed", Sev: "warn",
			Text:    "The hand-in did not go through: " + h.Reason,
			Actions: []Verb{{Label: "Hand in", Post: QuestRoute(h.Quest, "handin"), Body: map[string]any{"chunk": h.Chunk}}}})
	}
	return out
}

// ---- actions ---------------------------------------------------------------------------------------------

// gate is the common refusal of every action: no hero, board off.
func (d *Desk) gate() (*Exec, *Store, *Reply) {
	x, st := d.exec()
	if x == nil || st == nil {
		r := refuse(http.StatusServiceUnavailable, CodeNoHero, "No hero is loaded yet; try again in a moment.")
		return nil, nil, &r
	}
	if !x.Enabled() {
		r := refuse(http.StatusConflict, CodeOff, "The quest board is off: set the knob quests.board to 1 to use it.")
		return nil, nil, &r
	}
	return x, st, nil
}

// pick finds the chunk an action means: the named one, else the quest's current chunk (the Ready one for Start,
// the accepted one otherwise).
func (d *Desk) pick(x *Exec, quest, chunk string, start bool) (string, State, *Reply) {
	if _, err := d.b.Quest(quest); err != nil {
		r := refuse(http.StatusNotFound, CodeUnknown, "There is no quest %q.", quest)
		return "", "", &r
	}
	if chunk != "" {
		if _, err := d.b.Chunk(quest, chunk); err != nil {
			r := refuse(http.StatusNotFound, CodeUnknown, "Quest %s has no part %q.", quest, chunk)
			return "", "", &r
		}
	}
	st, err := x.State()
	if err != nil {
		r := refuse(http.StatusInternalServerError, CodeInternal, "The board could not be read: %v", err)
		return "", "", &r
	}
	qv, ok := QuestView{}, false
	for _, v := range st.Views {
		if v.ID == quest {
			qv, ok = v, true
		}
	}
	if !ok {
		r := refuse(http.StatusConflict, CodeNotReady, "This quest does not exist in %s.", Difficulties[st.Diff])
		return "", "", &r
	}
	if chunk != "" {
		for _, c := range qv.Chunks {
			if c.ID == chunk {
				return chunk, c.State, nil
			}
		}
	}
	// the current chunk: the offered one for Start; else one that runs, waits or is parked
	order := func(want func(State) bool) (string, State, bool) {
		for _, c := range qv.Chunks {
			if want(c.State) {
				return c.ID, c.State, true
			}
		}
		return "", "", false
	}
	if start {
		if len(qv.Offered) > 0 {
			for _, c := range qv.Chunks {
				if c.ID == qv.Offered[0] {
					return c.ID, c.State, nil
				}
			}
		}
		if id, s, ok := order(func(s State) bool { return s == Accepted || s == InProgress || s == Awaiting }); ok {
			return id, s, nil
		}
	} else if id, s, ok := order(func(s State) bool { return s == Accepted || s == InProgress || s == Awaiting || s == Blocked }); ok {
		return id, s, nil
	}
	if len(qv.Chunks) > 0 {
		c := qv.Chunks[0]
		return c.ID, c.State, nil
	}
	r := refuse(http.StatusConflict, CodeNotReady, "%s has no part to start.", qv.Title)
	return "", "", &r
}

func (d *Desk) title(quest string) string {
	if q, err := d.b.Quest(quest); err == nil {
		return q.Title
	}
	return quest
}

// Start accepts the current chunk of a quest (the player's Start tap; it never covers the next part).
func (d *Desk) Start(quest, chunk string) Reply {
	x, st, bad := d.gate()
	if bad != nil {
		return *bad
	}
	chunk, state, bad := d.pick(x, quest, chunk, true)
	if bad != nil {
		return *bad
	}
	base := Reply{Quest: quest, Chunk: chunk}
	switch state {
	case Accepted, InProgress, Awaiting:
		base.OK, base.Already, base.State, base.Status = true, true, string(state), http.StatusOK
		base.Reason = d.title(quest) + " is already started."
		return base
	case Done:
		return withBase(base, refuse(http.StatusConflict, CodeDone, "%s is already done.", d.title(quest)))
	}
	if d.paused() {
		return withBase(base, refuse(http.StatusConflict, CodePaused, "The bot is paused; resume it before starting %s.", d.title(quest)))
	}
	if m, merr := st.Mode(); merr == nil {
		x.SetMode(m) // the record carries the mode the player picked, also after a restart
	}
	err := x.Accept(quest, chunk)
	if err != nil {
		code, status := CodeNotReady, http.StatusConflict
		if s := err.Error(); strings.Contains(s, "already accepted") {
			code = CodeBusy
		}
		return withBase(base, refuse(status, code, "%s", cleanErr(err)))
	}
	d.invalidate()
	base.OK, base.State, base.Status = true, string(Accepted), http.StatusOK
	base.Reason = d.title(quest) + " started."
	return base
}

// Cancel stops an accepted chunk and puts it back to ready. Nothing to stop is an `already` no-op.
func (d *Desk) Cancel(quest, chunk string) Reply {
	x, _, bad := d.gate()
	if bad != nil {
		return *bad
	}
	chunk, state, bad := d.pick(x, quest, chunk, false)
	if bad != nil {
		return *bad
	}
	base := Reply{Quest: quest, Chunk: chunk}
	switch state {
	case Done:
		return withBase(base, refuse(http.StatusConflict, CodeDone, "%s is already done; nothing to stop.", d.title(quest)))
	case Accepted, InProgress, Awaiting, Blocked:
	default:
		base.OK, base.Already, base.State, base.Status = true, true, string(state), http.StatusOK
		base.Reason = d.title(quest) + " is not running."
		return base
	}
	if err := x.Cancel(quest, chunk); err != nil {
		return withBase(base, refuse(http.StatusInternalServerError, CodeInternal, "Could not stop it: %v", err))
	}
	d.invalidate()
	base.OK, base.State, base.Status, base.Reason = true, string(Ready), http.StatusOK, d.title(quest)+" stopped."
	return base
}

// Retry clears the park of a blocked chunk (the "Try again" button of a Needs-you row).
func (d *Desk) Retry(quest, chunk string) Reply {
	x, _, bad := d.gate()
	if bad != nil {
		return *bad
	}
	chunk, state, bad := d.pick(x, quest, chunk, false)
	if bad != nil {
		return *bad
	}
	base := Reply{Quest: quest, Chunk: chunk, State: string(state)}
	if state != Blocked {
		base.OK, base.Already, base.Status, base.Reason = true, true, http.StatusOK, d.title(quest)+" is not stuck."
		return base
	}
	x.Unblock(quest, chunk)
	d.invalidate()
	base.OK, base.Status, base.Reason = true, http.StatusOK, d.title(quest)+": trying again."
	return base
}

// HandIn is the player's hand-in tap. The dialogue runs on the bot loop and can take a while (a town trip, a
// death), so the reply is 202 and the outcome is in View.HandIn.
func (d *Desk) HandIn(quest, chunk string) Reply {
	x, _, bad := d.gate()
	if bad != nil {
		return *bad
	}
	chunk, state, bad := d.pick(x, quest, chunk, false)
	if bad != nil {
		return *bad
	}
	base := Reply{Quest: quest, Chunk: chunk, State: string(state)}
	k := ChunkKey(quest, chunk)
	d.mu.Lock()
	_, running := d.hand[k]
	d.mu.Unlock()
	switch {
	case running:
		base.OK, base.Already, base.Pending, base.Status, base.Reason = true, true, true, http.StatusAccepted, "The hand-in is already running."
		return base
	case state == Done:
		base.OK, base.Already, base.Status, base.Reason = true, true, http.StatusOK, d.title(quest)+" is already handed in."
		return base
	case state != Awaiting:
		return withBase(base, refuse(http.StatusConflict, CodeNotAwait, "%s is not waiting for the hand-in%s.", d.title(quest), whyState(state)))
	case d.paused():
		return withBase(base, refuse(http.StatusConflict, CodePaused, "The bot is paused; resume it to hand in."))
	}
	d.mu.Lock()
	d.hand[k] = &handRun{started: d.cfg.Now()}
	d.last = &HandInView{Quest: quest, Chunk: chunk, State: "running"}
	d.cache = nil
	d.mu.Unlock()
	go d.runHandIn(x, quest, chunk, k)
	base.OK, base.Pending, base.Status, base.Reason = true, true, http.StatusAccepted, "Handing in "+d.title(quest)+"."
	return base
}

func (d *Desk) runHandIn(x *Exec, quest, chunk, k string) {
	ctx, cancel := context.WithTimeout(d.cfg.Ctx, handInTimeout)
	defer cancel()
	res, err := x.HandIn(ctx, quest, chunk)
	hv := HandInView{Quest: quest, Chunk: chunk, Attempt: res.Attempts, At: d.cfg.Now().UTC().Format(time.RFC3339)}
	switch {
	case err != nil:
		hv.State, hv.Reason = "failed", cleanErr(err)
	case res.Outcome == HandInDone:
		hv.State = "done"
	case res.Outcome == HandInConfirming:
		hv.State, hv.Reason = "confirming", "The reward was seen; waiting for the game to confirm it."
	default:
		hv.State, hv.Reason = "blocked", res.Why
	}
	d.mu.Lock()
	delete(d.hand, k)
	d.last, d.cache = &hv, nil
	d.mu.Unlock()
}

// SetMode stores the between-chunks mode (train, idle, farm). Slice S6 gives it behaviour; until then it is
// stored and echoed (View.ModeActs is false).
func (d *Desk) SetMode(mode string) Reply {
	x, st := d.exec()
	if x == nil || st == nil {
		return refuse(http.StatusServiceUnavailable, CodeNoHero, "No hero is loaded yet; try again in a moment.")
	}
	if !ValidMode(mode) {
		return refuse(http.StatusBadRequest, CodeBadMode, "Mode must be train, idle or farm, not %q.", mode)
	}
	cur, err := st.Mode()
	if err != nil {
		return refuse(http.StatusInternalServerError, CodeInternal, "Could not read the mode: %v", err)
	}
	if cur == mode {
		return Reply{OK: true, Already: true, Mode: mode, Status: http.StatusOK, Reason: "The mode is already " + mode + "."}
	}
	if err := st.SetMode(mode); err != nil {
		return refuse(http.StatusInternalServerError, CodeInternal, "Could not save the mode: %v", err)
	}
	x.SetMode(mode)
	d.invalidate()
	return Reply{OK: true, Mode: mode, Status: http.StatusOK, Reason: "Mode set to " + mode + "."}
}

func withBase(b, r Reply) Reply {
	r.Quest, r.Chunk = b.Quest, b.Chunk
	if b.State != "" {
		r.State = b.State
	}
	return r
}

func whyState(s State) string {
	if s == "" {
		return ""
	}
	return " (it is " + string(s) + ")"
}

func cleanErr(err error) string {
	s := err.Error()
	return strings.TrimPrefix(s, "questboard: ")
}
