package watch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/currency"
	"d2lod-bot/internal/goals"
)

// The Loop order (docs/loop-order.md, design C-orders section 8): go to an area for X minutes, home, repair, repeat.
// A sequencer on the pattern of OrderQueue: it places ONE existing order at a time through Hub.place (charter, 5-open
// limit and the bot's route guard apply unchanged) and asks the repair desk for the repair, then reacts to their end.
// It never loops on failure: a refusal stops it with a reason, a busy hero is retried with a growing wait and then
// stops, rounds that end at once stop it too.

const (
	// OrderLoopID is the ladder row that opens the Loop.
	OrderLoopID = "orderloop"
	// OrderLoopRoute is GET (viewer) and POST (answer listener): {action: start | stop, ...}.
	OrderLoopRoute = "/api/orderloop"

	orderLoopSchema = "watch.orderloop/1"
	loopStateSchema = "watch.orderloop.state/1"
	loopStateFile   = "orderloop.json"

	// LoopManualEvent opens the Loop row: a repair the player asked for ended done after a zone order was given.
	LoopManualEvent = "manual_repair_cycle"
	// LoopRoundsEvent opens the first queue slot: LoopRoundsNeeded finished rounds in all.
	LoopRoundsEvent  = "loop_rounds_3"
	LoopRoundsNeeded = 3

	LoopMinMinutes = 5
	LoopMaxMinutes = 240
	LoopMaxRounds  = 20

	loopQuickSecs   = 60 // a farm leg that ends done sooner than this did nothing
	loopQuickLimit  = 3  // rounds in a row that did nothing before the loop gives up
	loopBackoffBase = 30 * time.Second
	loopMaxAttempts = 4 // a busy hero is retried 30 s, 60 s, 120 s, then the loop stops
)

// Repair choices.
const (
	LoopRepairDamaged = "damaged" // the smith only when something worn is under the warn line or broken
	LoopRepairAlways  = "always"  // ask for a repair every round (the desk still answers undamaged when nothing is hurt)
)

// States of the loop (the view's `state`).
const (
	LoopLocked    = "locked"
	LoopIdle      = "idle"
	LoopActive    = "active"
	LoopDone      = "done"
	LoopCancelled = "cancelled"
	LoopFailed    = "failed"
)

// Phases of an active loop.
const (
	loopPhaseStart   = ""           // the start of a round
	LoopPhaseFarming = "farming"    // the go order runs
	LoopPhaseHome    = "going home" // the plain home order runs
	LoopPhaseRepair  = "repairing"  // the repair request is open
	LoopPhaseWaiting = "waiting"    // backing off before the next try
)

// Request states the loop reads from the repair desk.
const (
	LoopReqNone      = ""
	LoopReqOpen      = "requested"
	LoopReqDone      = "done"
	LoopReqFailed    = "failed"
	LoopReqCancelled = "cancelled"
)

// LoopRepairDesk is what the loop needs from the repair side (internal/repair through cmd/d2bot/loopwire.go).
type LoopRepairDesk interface {
	// Worn says something worn is under the warn line or broken.
	Worn() bool
	// Request asks for a repair; the error carries Kind() "undamaged", "cooldown", "pact" or "active".
	Request() error
	// Status is the state of the last request ("" none) and whether the bot, not the player, ended it.
	Status() (state string, auto bool, note string)
	Cancel() error
}

// OrderLoopConfig builds an OrderLoop. Every field is optional.
type OrderLoopConfig struct {
	Path  string                                    // "" keeps the state in memory
	Quote func(area string) (currency.Quote, error) // charter check at start
	Halt  func() string                             // the Pact or low-life pause reason ("" = running)
	// Blocked says why the quest journey is refused now ("" = not); shown on the row, never a reason to stop (the loop
	// is a player order, it goes before the plan).
	Blocked func() string
	// Waiting is how many queue entries wait (the loop and a forever loop refuse each other).
	Waiting func() int
}

// LoopSpec is one loop as the player gave it.
type LoopSpec struct {
	Area    string `json:"area"`
	Minutes int    `json:"minutes"`
	Repair  string `json:"repair"`
	Rounds  int    `json:"rounds,omitempty"`
	Forever bool   `json:"forever,omitempty"`
	Approve bool   `json:"approve,omitempty"` // the player accepted the charter of an above-range area
}

type loopState struct {
	Schema     string   `json:"schema"`
	Spec       LoopSpec `json:"spec"`
	State      string   `json:"state"`
	Round      int      `json:"round"` // 1-based, the round that runs or ran last
	Phase      string   `json:"phase,omitempty"`
	Finished   int      `json:"finished"` // rounds finished by every loop of this hero (opens the queue)
	StoppedWhy string   `json:"stoppedWhy,omitempty"`
	ZoneOrder  bool     `json:"zoneOrder,omitempty"` // a zone order was given before the ladder row opened
	ManualSeen bool     `json:"manualSeen,omitempty"`

	// Not persisted in a way that matters: a restart resumes at the start of the round.
	OrderID  int    `json:"-"`
	StartAt  int64  `json:"-"` // virtual unix seconds the farm leg was placed
	Quick    int    `json:"-"`
	Attempts int    `json:"-"`
	RetryAt  int64  `json:"-"`
	Wait     string `json:"-"`
	repReq   bool   // the loop asked for the repair request that is open
}

// OrderLoop is the sequencer. Safe for concurrent use.
type OrderLoop struct {
	u   *Unlocks
	hub *Hub
	cfg OrderLoopConfig

	mu   sync.Mutex
	desk LoopRepairDesk
	st   loopState
}

// Errors from the loop routes.
var (
	ErrLoopLocked    = errors.New("watch: the Loop order is not unlocked yet")
	ErrLoopActive    = errors.New("watch: a loop is already running")
	ErrLoopNone      = errors.New("watch: no loop is running")
	ErrLoopQueueBusy = errors.New("watch: orders wait in the queue; a loop until you stop it would starve them")
	ErrLoopOff       = errors.New("watch: the repair desk is not available yet")
)

// NewOrderLoop checks the ladder row and loads the saved state. An active loop resumes at the start of its round.
func NewOrderLoop(u *Unlocks, h *Hub, c OrderLoopConfig) (*OrderLoop, error) {
	if h == nil || u == nil {
		return nil, errors.New("watch: the loop order needs a hub and the unlock ladder")
	}
	if _, err := u.Engine.Check(OrderLoopID); err != nil {
		return nil, err
	}
	l := &OrderLoop{u: u, hub: h, cfg: c}
	l.st.Schema = loopStateSchema
	if c.Path != "" {
		b, err := os.ReadFile(c.Path)
		switch {
		case err == nil:
			var s loopState
			if err := json.Unmarshal(b, &s); err != nil || s.Schema != loopStateSchema {
				return nil, fmt.Errorf("order loop: %s is not a loop state", c.Path)
			}
			l.st = s
			l.st.Phase = loopPhaseStart
		case !errors.Is(err, os.ErrNotExist):
			return nil, err
		}
	}
	return l, nil
}

// Bind gives the loop its repair desk (late-bound like the repair routes).
func (l *OrderLoop) Bind(d LoopRepairDesk) { l.mu.Lock(); l.desk = d; l.mu.Unlock() }

func (l *OrderLoop) open() bool { return l.u.Engine.Allowed(OrderLoopID) }

func (l *OrderLoop) now() int64 { return int64(l.hub.cfg.Clock.Now() / time.Second) }

func (l *OrderLoop) saveLocked() {
	if l.cfg.Path == "" {
		return
	}
	b, err := json.MarshalIndent(l.st, "", "  ")
	if err == nil {
		if err = os.MkdirAll(filepath.Dir(l.cfg.Path), 0o755); err == nil {
			tmp := l.cfg.Path + ".tmp"
			if err = os.WriteFile(tmp, b, 0o600); err == nil {
				err = os.Rename(tmp, l.cfg.Path)
			}
		}
	}
	if err != nil {
		l.u.fail(fmt.Errorf("order loop: saving the state: %w", err))
	}
}

func (l *OrderLoop) journal(text string) {
	h := l.hub
	h.mu.Lock()
	h.addExtraLocked(extra{kind: "order", text: safeText(text, 400)})
	h.notifyLocked()
	h.mu.Unlock()
}

func (l *OrderLoop) activeLocked() bool { return l.st.State == LoopActive }

// Holding says a loop is active: the queue must not place anything while it is (a finite loop holds the running place,
// and the gaps between its steps belong to it).
func (l *OrderLoop) Holding() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.activeLocked()
}

func (l *OrderLoop) textLocked() string {
	s := l.st.Spec
	rep := ", repair"
	r := fmt.Sprintf(" (round %d of %d)", l.st.Round, s.Rounds)
	if s.Forever {
		r = fmt.Sprintf(" (round %d)", l.st.Round)
	}
	return fmt.Sprintf("Loop: %s %d min%s, repeat%s", s.Area, s.Minutes, rep, r)
}

// Start begins a loop. A forever loop is refused while queue entries wait.
func (l *OrderLoop) Start(s LoopSpec) error {
	if !l.open() {
		return ErrLoopLocked
	}
	if s.Repair == "" {
		s.Repair = LoopRepairDamaged
	}
	if s.Repair != LoopRepairDamaged && s.Repair != LoopRepairAlways {
		return fmt.Errorf("%w: repair must be damaged or always", ErrBadOrder)
	}
	if s.Minutes == 0 {
		s.Minutes = 20
	}
	if s.Minutes < LoopMinMinutes || s.Minutes > LoopMaxMinutes {
		return fmt.Errorf("%w: minutes must be %d to %d", ErrBadOrder, LoopMinMinutes, LoopMaxMinutes)
	}
	switch {
	case s.Forever && s.Rounds != 0:
		return fmt.Errorf("%w: forever takes no rounds", ErrBadOrder)
	case s.Forever:
	case s.Rounds == 0:
		s.Rounds = 5
	case s.Rounds < 1 || s.Rounds > LoopMaxRounds:
		return fmt.Errorf("%w: rounds must be 1 to %d", ErrBadOrder, LoopMaxRounds)
	}
	o, err := validateOrder(OrderGo, s.Area, s.Minutes, false, l.hub.cfg.Areas)
	if err != nil {
		return err
	}
	s.Area = o.Area
	if l.cfg.Quote != nil {
		qt, err := l.cfg.Quote(s.Area)
		if err != nil {
			return err
		}
		if qt.Price > 0 && !s.Approve {
			return fmt.Errorf("%w: %s", currency.ErrCharterNeeded, qt.Text)
		}
	}
	if s.Forever && l.cfg.Waiting != nil && l.cfg.Waiting() > 0 {
		return ErrLoopQueueBusy
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.desk == nil {
		return ErrLoopOff
	}
	if l.activeLocked() {
		return ErrLoopActive
	}
	l.st.Spec, l.st.State, l.st.Round, l.st.Phase = s, LoopActive, 1, loopPhaseStart
	l.st.StoppedWhy, l.st.OrderID, l.st.Quick, l.st.Attempts, l.st.RetryAt, l.st.Wait = "", 0, 0, 0, 0, ""
	l.saveLocked()
	l.journal("Loop started: " + l.textLocked())
	return nil
}

// Stop is the player's cancel: the running order and an open repair request of the loop are cancelled too.
func (l *OrderLoop) Stop() error {
	l.mu.Lock()
	if !l.activeLocked() {
		l.mu.Unlock()
		return ErrLoopNone
	}
	oid, rep := l.st.OrderID, l.st.repReq
	desk := l.desk
	l.endLocked(LoopCancelled, "cancelled by you")
	l.mu.Unlock()
	if oid != 0 {
		_ = l.hub.CancelOrder(oid)
	}
	if rep && desk != nil {
		_ = desk.Cancel()
	}
	return nil
}

func (l *OrderLoop) endLocked(state, why string) {
	text := l.textLocked()
	l.st.State, l.st.StoppedWhy, l.st.Phase, l.st.OrderID, l.st.Wait = state, why, loopPhaseStart, 0, ""
	l.st.repReq = false
	l.saveLocked()
	l.journal(fmt.Sprintf("Loop ended (%s): %s: %s", state, text, why))
}

// OnGoalEvent stops the loop when the hero dies (the next one must be started by a person).
func (l *OrderLoop) OnGoalEvent(ev goals.Event) {
	if ev.Type != goals.EvDeath {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.activeLocked() {
		l.endLocked(LoopFailed, "the hero died")
	}
}

// finishRoundLocked counts a finished round, opens the queue after the third one in all, and starts the next round or
// ends the loop.
func (l *OrderLoop) finishRoundLocked() {
	l.st.Finished++
	if l.st.Finished == LoopRoundsNeeded {
		l.u.Seen(LoopRoundsEvent)
	}
	s := l.st.Spec
	if !s.Forever && l.st.Round >= s.Rounds {
		l.endLocked(LoopDone, fmt.Sprintf("%d round%s done", s.Rounds, riskS(s.Rounds)))
		return
	}
	l.st.Round++
	l.st.Phase, l.st.OrderID, l.st.repReq = loopPhaseStart, 0, false
	l.st.Attempts, l.st.RetryAt, l.st.Wait = 0, 0, ""
	l.saveLocked()
}

func (l *OrderLoop) backoffLocked(why string) {
	l.st.Attempts++
	if l.st.Attempts >= loopMaxAttempts {
		l.endLocked(LoopFailed, why+"; gave up after "+fmt.Sprint(loopMaxAttempts)+" tries")
		return
	}
	d := loopBackoffBase << (l.st.Attempts - 1)
	l.st.RetryAt = l.now() + int64(d/time.Second)
	l.st.Wait, l.st.Phase = why, LoopPhaseWaiting
}

// Tick advances the loop (call it from the loop that polls the status, Fun.Settle).
func (l *OrderLoop) Tick() {
	if !l.open() {
		l.observeManual()
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.activeLocked() {
		return
	}
	if l.cfg.Halt != nil {
		if why := l.cfg.Halt(); why != "" {
			oid := l.st.OrderID
			l.endLocked(LoopFailed, why)
			l.mu.Unlock()
			if oid != 0 {
				_ = l.hub.CancelOrder(oid)
			}
			l.mu.Lock()
			return
		}
	}
	if l.hub.heroNow().Level <= 0 {
		return // the first status poll has not told the hero yet
	}
	for i := 0; i < 4 && l.activeLocked(); i++ { // a step that ends at once (a skipped repair) flows into the next
		if !l.stepLocked() {
			return
		}
	}
}

// stepLocked runs one step of the state machine; true when the state changed and another step may follow.
func (l *OrderLoop) stepLocked() bool {
	switch l.st.Phase {
	case loopPhaseStart, LoopPhaseWaiting:
		if l.st.Phase == LoopPhaseWaiting && l.now() < l.st.RetryAt {
			return false
		}
		return l.placeFarmLocked()
	case LoopPhaseFarming:
		return l.farmingLocked()
	case LoopPhaseHome:
		return l.homeLocked()
	case LoopPhaseRepair:
		return l.repairingLocked()
	}
	l.endLocked(LoopFailed, "unknown phase "+l.st.Phase)
	return false
}

func (l *OrderLoop) placeFarmLocked() bool {
	if len(l.hub.openIDs()) != 0 {
		l.backoffLocked("another order is running")
		return false
	}
	s := l.st.Spec
	o, _, err := l.hub.place(OrderRequest{Kind: OrderGo, Area: s.Area, Minutes: s.Minutes}, s.Approve)
	if err != nil {
		if errors.Is(err, ErrOrdersFull) {
			l.backoffLocked("too many open orders")
			return false
		}
		l.endLocked(LoopFailed, "could not send the hero to "+s.Area+": "+safeText(err.Error(), 160))
		return false
	}
	l.st.OrderID, l.st.StartAt, l.st.Phase, l.st.Wait = o.ID, l.now(), LoopPhaseFarming, ""
	l.st.Attempts = 0
	l.saveLocked()
	return false
}

func (l *OrderLoop) farmingLocked() bool {
	st, note := l.hub.orderStateOf(l.st.OrderID)
	switch st {
	case OrderQueued, OrderActive:
		return false
	case OrderDone:
		if l.now()-l.st.StartAt < loopQuickSecs {
			l.st.Quick++
			if l.st.Quick >= loopQuickLimit {
				l.endLocked(LoopFailed, fmt.Sprintf("%s ended at once %d rounds in a row", l.st.Spec.Area, l.st.Quick))
				return false
			}
		} else {
			l.st.Quick = 0
		}
		l.st.OrderID = 0
		return l.afterFarmLocked()
	case OrderFailed:
		l.endLocked(LoopFailed, "the farm leg failed"+noteSuffix(note))
	default: // cancelled outside the loop (a player zone order with replace, "Go now", a proposal)
		l.endLocked(LoopCancelled, "the farm leg was cancelled by another order")
	}
	return false
}

// afterFarmLocked is the way home: a repair request, or the plain home order.
func (l *OrderLoop) afterFarmLocked() bool {
	if l.desk == nil {
		l.endLocked(LoopFailed, "the repair desk is gone")
		return false
	}
	if l.st.Spec.Repair == LoopRepairDamaged && !l.desk.Worn() {
		return l.placeHomeLocked("nothing worn needs the smith")
	}
	err := l.desk.Request()
	var ref invRefusal
	switch {
	case err == nil:
		l.st.Phase, l.st.repReq = LoopPhaseRepair, true
		l.saveLocked()
		return true
	case errors.As(err, &ref):
		switch ref.Kind() {
		case "undamaged", "cooldown": // nothing to do or repaired minutes ago: the round counts as done
			return l.placeHomeLocked("repair skipped: " + ref.Error())
		case "active": // a request (the player's) is already open: wait for it
			l.st.Phase = LoopPhaseRepair
			l.saveLocked()
			return true
		case "pact":
			l.endLocked(LoopFailed, "a Pact forbids repairs")
			return false
		}
	}
	l.endLocked(LoopFailed, "the repair could not be asked for: "+safeText(err.Error(), 160))
	return false
}

func (l *OrderLoop) placeHomeLocked(why string) bool {
	o, _, err := l.hub.place(OrderRequest{Kind: OrderHome}, false)
	if err != nil {
		if errors.Is(err, ErrOrdersFull) {
			l.backoffLocked("too many open orders")
			if l.activeLocked() {
				l.st.Phase = LoopPhaseHome // the retry places home again, not a farm leg
			}
			return false
		}
		l.endLocked(LoopFailed, "could not send the hero home: "+safeText(err.Error(), 160))
		return false
	}
	l.st.OrderID, l.st.Phase = o.ID, LoopPhaseHome
	l.saveLocked()
	l.journal("Loop: " + why + ", going home")
	return false
}

func (l *OrderLoop) homeLocked() bool {
	if l.st.OrderID == 0 { // a retry after a full orders box
		if l.now() < l.st.RetryAt {
			return false
		}
		return l.placeHomeLocked("retry")
	}
	st, note := l.hub.orderStateOf(l.st.OrderID)
	switch st {
	case OrderQueued, OrderActive:
		return false
	case OrderDone:
		l.finishRoundLocked()
		return true
	case OrderFailed:
		l.endLocked(LoopFailed, "going home failed"+noteSuffix(note))
	default:
		l.endLocked(LoopCancelled, "the way home was cancelled by another order")
	}
	return false
}

func (l *OrderLoop) repairingLocked() bool {
	if l.desk == nil {
		l.endLocked(LoopFailed, "the repair desk is gone")
		return false
	}
	state, auto, note := l.desk.Status()
	switch state {
	case LoopReqOpen:
		if ids := l.hub.openIDs(); len(ids) > 0 { // a person gave a zone order while the hero repairs
			l.endLocked(LoopCancelled, "a zone order took over")
		}
		return false
	case LoopReqDone:
		l.finishRoundLocked()
		return true
	case LoopReqFailed:
		l.endLocked(LoopFailed, "the repair failed"+noteSuffix(note))
	case LoopReqCancelled:
		if auto {
			l.endLocked(LoopFailed, "the repair was dropped by the bot"+noteSuffix(note))
		} else {
			l.endLocked(LoopCancelled, "the repair was cancelled")
		}
	default:
		l.endLocked(LoopFailed, "the repair request vanished")
	}
	return false
}

// observeManual runs while the Loop row is shut: a repair the player asked for that ended done, after a zone order was
// given, opens the row (the first full manual cycle, user answer 5).
func (l *OrderLoop) observeManual() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.st.ManualSeen {
		return
	}
	if !l.st.ZoneOrder {
		for _, o := range l.hub.OrdersView().Orders {
			if o.Kind == OrderGo || o.Kind == OrderHere {
				l.st.ZoneOrder = true
				l.saveLocked()
				break
			}
		}
	}
	if !l.st.ZoneOrder || l.desk == nil {
		return
	}
	if state, auto, _ := l.desk.Status(); state == LoopReqDone && !auto {
		l.st.ManualSeen = true
		l.saveLocked()
		l.u.Seen(LoopManualEvent)
	}
}

// LoopView is GET /api/orderloop.
type LoopView struct {
	Schema         string `json:"schema"`
	Open           bool   `json:"open"`
	Title          string `json:"title,omitempty"`
	Teaser         string `json:"teaser,omitempty"`
	State          string `json:"state"`
	Phase          string `json:"phase,omitempty"`
	Round          int    `json:"round,omitempty"`
	Rounds         int    `json:"rounds,omitempty"`
	Forever        bool   `json:"forever,omitempty"`
	Area           string `json:"area,omitempty"`
	Minutes        int    `json:"minutes,omitempty"`
	Repair         string `json:"repair,omitempty"`
	Text           string `json:"text,omitempty"`
	Finished       int    `json:"finished"`
	StoppedWhy     string `json:"stoppedWhy,omitempty"`
	Waiting        string `json:"waitingFor,omitempty"`
	JourneyBlocked string `json:"journeyBlocked,omitempty"`
}

// View is the payload. Locked: the teaser of the row only.
func (l *OrderLoop) View() LoopView {
	row, _ := l.u.state(OrderLoopID)
	v := LoopView{Schema: orderLoopSchema, Open: row.Open, Title: row.Title, State: LoopLocked}
	if !row.Open {
		v.Teaser = row.Teaser
		return v
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	v.State, v.Finished = LoopIdle, l.st.Finished
	if l.st.State != "" {
		s := l.st.Spec
		v.State, v.Phase, v.Round, v.Rounds, v.Forever = l.st.State, l.st.Phase, l.st.Round, s.Rounds, s.Forever
		v.Area, v.Minutes, v.Repair, v.Text = s.Area, s.Minutes, s.Repair, l.textLocked()
		v.StoppedWhy, v.Waiting = l.st.StoppedWhy, l.st.Wait
		if l.activeLocked() && l.cfg.Blocked != nil {
			v.JourneyBlocked = safeText(l.cfg.Blocked(), 160)
		}
	}
	return v
}

// LoopCommand is the one row of /api/commands for an active loop (S4 reads it).
type LoopCommand struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	State string `json:"state"` // active
	Phase string `json:"phase,omitempty"`
}

// Command returns the loop's command row, false when no loop is active.
func (l *OrderLoop) Command() (LoopCommand, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.activeLocked() {
		return LoopCommand{}, false
	}
	ph := l.st.Phase
	if ph == loopPhaseStart {
		ph = LoopPhaseFarming
	}
	return LoopCommand{ID: "loop:1", Text: l.textLocked(), State: "active", Phase: ph}, true
}

// Handler serves GET /api/orderloop.
func (l *OrderLoop) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, l.View())
	})
}

type loopBody struct {
	Action  string `json:"action"`
	Area    string `json:"area"`
	Minutes int    `json:"minutes"`
	Repair  string `json:"repair"`
	Rounds  int    `json:"rounds"`
	Forever bool   `json:"forever"`
	Approve bool   `json:"approve"`
}

type loopReply struct {
	OK      bool      `json:"ok"`
	Code    string    `json:"code,omitempty"`
	Error   string    `json:"error,omitempty"`
	Charter *Charter  `json:"charter,omitempty"`
	Loop    *LoopView `json:"loop,omitempty"`
}

// PostHandler serves POST /api/orderloop on the answer listener: same origin list, token, JSON-only and small-body rules
// as the other answer routes.
func (l *OrderLoop) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !unlockPostGate(o, w, r) {
			return
		}
		fail := func(code int, c, msg string) { writeJSON(w, code, loopReply{Code: c, Error: msg}) }
		var b loopBody
		if err := decodeSmall(r, &b); err != nil {
			fail(http.StatusBadRequest, "body", "body must be one small JSON object with known fields")
			return
		}
		var err error
		switch b.Action {
		case "start":
			err = l.Start(LoopSpec{Area: b.Area, Minutes: b.Minutes, Repair: strings.ToLower(b.Repair), Rounds: b.Rounds, Forever: b.Forever, Approve: b.Approve})
		case "stop":
			err = l.Stop()
		default:
			fail(http.StatusBadRequest, "bad-request", "action must be start or stop")
			return
		}
		switch {
		case err == nil:
			v := l.View()
			writeJSON(w, 200, loopReply{OK: true, Loop: &v})
		case errors.Is(err, ErrLoopLocked):
			fail(http.StatusConflict, "locked", err.Error())
		case errors.Is(err, ErrLoopActive):
			fail(http.StatusConflict, "active", err.Error())
		case errors.Is(err, ErrLoopNone):
			fail(http.StatusConflict, "none", err.Error())
		case errors.Is(err, ErrLoopQueueBusy):
			fail(http.StatusConflict, "queue-busy", err.Error())
		case errors.Is(err, ErrLoopOff):
			fail(http.StatusServiceUnavailable, "off", err.Error())
		case errors.Is(err, currency.ErrCharterNeeded):
			rep := loopReply{Code: "charter-needed", Error: err.Error()}
			if l.cfg.Quote != nil {
				if qt, qerr := l.cfg.Quote(b.Area); qerr == nil && qt.Price > 0 {
					rep.Charter = &Charter{Area: qt.Area, Price: qt.Price, Text: qt.Text}
				}
			}
			writeJSON(w, http.StatusPaymentRequired, rep)
		case errors.Is(err, ErrBadOrder):
			fail(http.StatusBadRequest, "bad-order", err.Error())
		default:
			fail(http.StatusInternalServerError, "internal", err.Error())
		}
	})
}

// waitingCount is how many queue entries wait (the loop reads it for queue-busy).
func (q *OrderQueue) waitingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.waitingLocked()
}
