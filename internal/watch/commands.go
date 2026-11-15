package watch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// GET /api/commands (watch.commands/1, docs/ui/redesign-2026-11-14/C-orders.md section 6): the command sources the player
// can see as one list in one state vocabulary. Read-only: every row names the existing route that cancels it; nothing
// here changes a desk. The join takes a snapshot of each source on its own (no desk lock is held while another is read).

const (
	// CommandsRoute is GET (viewer).
	CommandsRoute  = "/api/commands"
	commandsSchema = "watch.commands/1"
)

// Row states, the one vocabulary of the Orders panel (C.3).
const (
	CmdQueued    = "queued"    // accepted, not started
	CmdWaiting   = "waiting"   // wants to start; the reason says what for
	CmdActive    = "active"    // running (Phase says where)
	CmdDone      = "done"      // finished
	CmdFailed    = "failed"    // ended with a reason
	CmdCancelled = "cancelled" // withdrawn
)

// failedShowSec is how long a failed row stays on the list (C.5), counted in the source's own clock.
const failedShowSec = 20 * 60

// Row sources, also the id prefixes.
const (
	srcOrder    = "order"    // the watch orders box (POST /api/orders)
	srcQueue    = "queue"    // the order queue (POST /api/orderqueue)
	srcLoop     = "loop"     // the Loop order (POST /api/orderloop)
	srcRepair   = "repair"   // the repair desk (POST /api/repair/*)
	srcItem     = "inv"      // the inventory desk's queue (POST /api/inventory/cancel)
	srcGoldFarm = "farmgold" // the gold-farm desk (POST /api/farmgold)
)

// CommandWhy is the reason line of a row.
type CommandWhy struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

// CommandProgress is a real counter (a farm leg), never a guess.
type CommandProgress struct {
	Pct   float64 `json:"pct"`
	Label string  `json:"label"`
}

// CommandCancel is the existing route that cancels a row.
type CommandCancel struct {
	Path string         `json:"path"`
	Body map[string]any `json:"body"`
}

// CommandCan says what the row's buttons may do.
type CommandCan struct {
	Cancel bool `json:"cancel"`
	Retry  bool `json:"retry"`
}

// CommandRow is one player command.
type CommandRow struct {
	ID         string           `json:"id"`
	Intent     string           `json:"intent"`
	Text       string           `json:"text"`
	Source     string           `json:"source"`
	State      string           `json:"state"`
	Phase      string           `json:"phase,omitempty"`
	Next       bool             `json:"next,omitempty"`
	Why        *CommandWhy      `json:"why,omitempty"`
	Progress   *CommandProgress `json:"progress,omitempty"`
	ETASeconds *int             `json:"etaSeconds,omitempty"`
	Can        CommandCan       `json:"can"`
	Cancel     *CommandCancel   `json:"cancel,omitempty"` // present exactly when Can.Cancel
}

// CommandBlocked is a row that is not a command: the plan cannot go on and says why.
type CommandBlocked struct {
	ID      string   `json:"id"`
	Code    string   `json:"code"`
	Text    string   `json:"text"`
	Actions []string `json:"actions"`
}

// CommandPlan is the grey last row: what the hero does when the list is empty.
type CommandPlan struct {
	Text       string `json:"text"`
	ETASeconds *int   `json:"etaSeconds,omitempty"`
}

// CommandSafety is the one-line banner: safety comes before every order.
type CommandSafety struct {
	Active bool   `json:"active"`
	Text   string `json:"text"`
}

// CommandVerb is a button of the verb row; only verbs whose desk is bound are listed.
type CommandVerb struct {
	Intent  string `json:"intent"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

// CommandLimits are the caps the panel names.
type CommandLimits struct {
	ZoneOrders  int `json:"zoneOrders"`
	ItemActions int `json:"itemActions"`
}

// CommandsView is GET /api/commands.
type CommandsView struct {
	Schema  string           `json:"schema"`
	Seq     uint64           `json:"seq"`
	Safety  CommandSafety    `json:"safety"`
	Paused  bool             `json:"paused"`
	Rows    []CommandRow     `json:"rows"`
	Blocked []CommandBlocked `json:"blocked"`
	Plan    *CommandPlan     `json:"plan"`
	Verbs   []CommandVerb    `json:"verbs"`
	Limits  CommandLimits    `json:"limits"`
}

// CommandHooks are facts the watch package does not own. Every field is optional; Paused nil means there is no pause
// desk (no Pause verb), Blocked nil means the journey is never reported blocked here.
type CommandHooks struct {
	Paused  func() bool
	Blocked func() string
}

// CommandSources is what BuildCommands reads. A nil function means that desk is not bound: its rows and verbs are absent.
// The JSON sources are the desks' own views, decoded into the few fields used here, so the join does not import them.
type CommandSources struct {
	Orders   func() OrdersView
	Queue    func() *OrderQueueView
	Loop     func() *LoopView
	Repair   func() []byte // internal/repair.View
	Items    func() []byte // internal/invdesk.View (the cached one: never builds on this goroutine)
	GoldFarm func() []byte // internal/goldfarm.View
	Plan     func() []byte // internal/campaign.View
	Safety   func() CommandSafety
	Hooks    CommandHooks
	Now      func() time.Time // wall clock for sources that stamp unix seconds
}

// sourceStates is the table every source's own states map through (the table test walks it).
var sourceStates = map[string]map[string]string{
	srcOrder: {OrderQueued: CmdQueued, OrderActive: CmdActive, OrderDone: CmdDone, OrderCancelled: CmdCancelled, OrderFailed: CmdFailed},
	srcQueue: {QueueWaiting: CmdQueued, QueueRunning: CmdActive, QueueDone: CmdDone, QueueFailed: CmdFailed, QueueCancelled: CmdCancelled},
	srcLoop:  {LoopActive: CmdActive, LoopDone: CmdDone, LoopCancelled: CmdCancelled, LoopFailed: CmdFailed},
	// item actions: queued (with a Wait: waiting), doing shows as active
	srcItem: {"queued": CmdQueued, "doing": CmdActive, "done": CmdDone, "failed": CmdFailed, "cancelled": CmdCancelled},
	// repair: the request's own state plus its phase (repair.Phase*): the five middle phases are waiting or active
	srcRepair: {"queued": CmdQueued, "waiting": CmdWaiting, "going": CmdActive, "repairing": CmdActive, "returning": CmdActive,
		"farming": CmdActive, "done": CmdDone, "failed": CmdFailed, "cancelled": CmdCancelled},
	srcGoldFarm: {"on": CmdActive},
}

// commandState maps a source's state to the row vocabulary; ok is false for a word the table does not know.
func commandState(source, state string) (string, bool) {
	s, ok := sourceStates[source][state]
	return s, ok
}

// terminal says a row state is finished.
func terminal(s string) bool { return s == CmdDone || s == CmdFailed || s == CmdCancelled }

func itemIntent(kind string) string {
	switch kind {
	case "", "equip":
		return "equip"
	case "buyid", "cain":
		return "identify"
	case "sell":
		return "sell_now"
	}
	return kind
}

func itemText(kind, name string) string {
	switch kind {
	case "", "equip":
		return "Equip " + name
	case "trash":
		return "Trash " + name
	case "identify":
		return "Identify " + name
	case "stash":
		return "Stash " + name
	case "take":
		return "Take " + name
	case "unequip":
		return "Take off " + name
	}
	return name
}

func minsOf(sec int) int { return (sec + 30) / 60 }

// BuildCommands joins the sources into the view. Pure: no desk is changed.
func BuildCommands(s CommandSources) CommandsView {
	v := CommandsView{Schema: commandsSchema, Rows: []CommandRow{}, Blocked: []CommandBlocked{}, Verbs: []CommandVerb{},
		Limits: CommandLimits{ZoneOrders: MaxOpenOrders, ItemActions: 8}}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	var ov OrdersView
	if s.Orders != nil {
		ov = s.Orders()
		v.Seq = ov.Seq
	}
	var qv *OrderQueueView
	if s.Queue != nil {
		qv = s.Queue()
	}
	var lv *LoopView
	if s.Loop != nil {
		lv = s.Loop()
	}
	if s.Safety != nil {
		v.Safety = s.Safety()
	}
	if s.Hooks.Paused != nil {
		v.Paused = s.Hooks.Paused()
	}

	var items, repair, farm, loop, zone, failed []CommandRow
	add := func(dst *[]CommandRow, r CommandRow) {
		if r.State == CmdFailed {
			failed = append(failed, r)
			return
		}
		if terminal(r.State) {
			return
		}
		*dst = append(*dst, r)
	}

	// item actions, identify first (the ladder of docs/inventory-actions.md)
	var iv struct {
		Queue  []itemReq `json:"queue"`
		Recent []itemReq `json:"recent"`
	}
	if s.Items != nil && decode(s.Items(), &iv) {
		var first, rest []CommandRow
		seen := map[string]bool{}
		for _, q := range append(iv.Queue, iv.Recent...) {
			r, ok := q.row(now().Unix())
			if !ok || seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			if r.Intent == "identify" {
				add(&first, r)
			} else {
				add(&rest, r)
			}
		}
		items = append(first, rest...)
	}

	var rv struct {
		Request *struct {
			State, Kind, Phase, Note, Area string
			Done                           int64 `json:"doneAt"`
			Have, Target                   int
		} `json:"request"`
		Trip *struct {
			Line   string
			Button struct {
				Action, Label, Reason string
				Enabled               bool
			}
		} `json:"trip"`
	}
	haveRepair := s.Repair != nil && decode(s.Repair(), &rv)
	if haveRepair && rv.Request != nil {
		q := rv.Request
		st := ""
		switch q.State {
		case "requested":
			st = q.Phase
			if st == "" {
				st = "queued"
			}
		default:
			st = q.State
		}
		text := "Repair now"
		if q.Kind == "farm" {
			text = "Farm until repair"
		}
		if m, ok := commandState(srcRepair, st); ok {
			r := CommandRow{ID: "repair:1", Intent: "repair", Text: text, Source: "gear", State: m}
			if m == CmdActive {
				r.Phase = st
			}
			switch {
			case m == CmdFailed && now().Unix()-q.Done > failedShowSec:
				r.State = CmdDone // an old failure is not listed
			case m == CmdFailed:
				r.Why = &CommandWhy{"repair_failed", safeText(q.Note, 120)}
			case rv.Trip != nil && rv.Trip.Line != "":
				r.Why = &CommandWhy{"repair_" + st, safeText(rv.Trip.Line, 120)}
			default:
				r.Why = &CommandWhy{"repair_" + st, repairWord[st]}
			}
			if q.Kind == "farm" && q.Target > 0 && st == "farming" {
				r.Progress = &CommandProgress{Pct: cmdRound(float64(q.Have) / float64(q.Target)), Label: fmt.Sprintf("%d of %d gold", q.Have, q.Target)}
				r.Why = &CommandWhy{"repair_farm", "farming for repair: " + r.Progress.Label}
			}
			if m == CmdQueued || m == CmdWaiting || m == CmdActive {
				r.Can.Cancel = true
				r.Cancel = &CommandCancel{Path: "/api/repair/cancel", Body: map[string]any{}}
			}
			add(&repair, r)
		}
	}

	var gv struct {
		On   bool
		Text string
		For  string
	}
	if s.GoldFarm != nil && decode(s.GoldFarm(), &gv) && gv.On && gv.For == "" {
		add(&farm, CommandRow{ID: "farmgold:1", Intent: "farm_gold", Text: "Farm gold", Source: "gold", State: CmdActive, Phase: "farming",
			Why: &CommandWhy{"farm_gold", safeText(gv.Text, 120)}, Can: CommandCan{Cancel: true},
			Cancel: &CommandCancel{Path: "/api/farmgold", Body: map[string]any{"on": false}}})
	}

	loopHolds := false
	if lv != nil {
		switch lv.State {
		case LoopActive:
			loopHolds = true
			r := CommandRow{ID: "loop:1", Intent: "loop", Text: safeText(lv.Text, 120), Source: "loop", State: CmdActive, Phase: lv.Phase,
				Can: CommandCan{Cancel: true}, Cancel: &CommandCancel{Path: OrderLoopRoute, Body: map[string]any{"action": "stop"}}}
			if lv.Waiting != "" {
				r.Why = &CommandWhy{"loop_wait", safeText(lv.Waiting, 120)}
			}
			add(&loop, r)
			if lv.JourneyBlocked != "" {
				v.Blocked = append(v.Blocked, blockedRow(lv.JourneyBlocked))
			}
		case LoopFailed, LoopCancelled:
			if lv.StoppedWhy != "" && lv.State == LoopFailed {
				add(&loop, CommandRow{ID: "loop:1", Intent: "loop", Text: safeText(lv.Text, 120), Source: "loop", State: CmdFailed,
					Why: &CommandWhy{"loop_stopped", safeText(lv.StoppedWhy, 120)}})
			}
		}
	}
	if s.Hooks.Blocked != nil {
		if b := s.Hooks.Blocked(); b != "" {
			v.Blocked = append(v.Blocked, blockedRow(b))
		}
	}

	// zone orders: the running one first, then the orders waiting, then the queue entries behind them
	runningOrder := map[int]bool{}
	var queued []CommandRow
	if qv != nil {
		for _, e := range qv.Entries {
			if e.State == QueueRunning && e.OrderID > 0 {
				runningOrder[e.OrderID] = true
			}
		}
	}
	for _, o := range ov.Orders {
		if runningOrder[o.ID] || (loopHolds && (o.Kind == OrderGo || o.Kind == OrderHome)) {
			continue // the queue entry or the loop row stands for it
		}
		m, ok := commandState(srcOrder, o.State)
		if !ok {
			continue
		}
		r := CommandRow{ID: fmt.Sprintf("order:%d", o.ID), Intent: orderIntent(o.Kind), Text: safeText(o.Text, 120), Source: "map", State: m}
		switch m {
		case CmdActive:
			r.Phase = "farming"
			if o.Kind == OrderHome {
				r.Phase = "going home"
			}
			r.Why = &CommandWhy{"order_active", activeWhy(o, ov.Now)}
			if o.Minutes > 0 && o.StartedAt > 0 {
				left := max(0, o.Minutes-int((ov.Now-o.StartedAt)/60))
				secs := left * 60
				r.ETASeconds = &secs
			}
		case CmdQueued:
			r.Why = &CommandWhy{"order_queued", "after the running order"}
		case CmdFailed:
			if ov.Now-o.GivenAt > failedShowSec+float64(o.Minutes*60) {
				continue
			}
			r.Why = &CommandWhy{"order_failed", safeText(o.Note, 120)}
		}
		if m == CmdQueued || m == CmdActive {
			r.Can.Cancel = true
			r.Cancel = &CommandCancel{Path: "/api/orders/cancel", Body: map[string]any{"id": o.ID}}
		}
		if m == CmdActive {
			zone = append(zone, r)
		} else {
			add(&queued, r)
		}
	}
	if qv != nil {
		n := 0
		for _, e := range qv.Entries {
			m, ok := commandState(srcQueue, e.State)
			if !ok {
				continue
			}
			r := CommandRow{ID: fmt.Sprintf("queue:%d", e.ID), Intent: orderIntent(e.Kind), Text: safeText(e.Text, 120), Source: "queue", State: m}
			switch m {
			case CmdQueued:
				n++
				r.Why = &CommandWhy{"queue_wait", fmt.Sprintf("behind the running order (place %d)", n)}
			case CmdActive:
				r.Phase = "farming"
				r.Why = &CommandWhy{"queue_active", safeText(e.Note, 120)}
			case CmdFailed:
				r.Why = &CommandWhy{"queue_failed", safeText(e.Note, 120)}
			}
			if m == CmdQueued || m == CmdActive {
				r.Can.Cancel = true
				r.Cancel = &CommandCancel{Path: OrderQueueRoute, Body: map[string]any{"action": "cancel", "id": e.ID}}
			}
			if m == CmdActive {
				zone = append(zone, r)
			} else {
				add(&queued, r)
			}
		}
	}
	zone = append(zone, queued...)

	for _, g := range [][]CommandRow{items, repair, farm, loop, zone} {
		for _, r := range g {
			if terminal(r.State) {
				continue
			}
			v.Rows = append(v.Rows, r)
		}
	}
	sort.SliceStable(failed, func(i, j int) bool { return failed[i].ID < failed[j].ID })
	v.Rows = append(v.Rows, failed...)
	for i := range v.Rows {
		if v.Rows[i].State == CmdQueued || v.Rows[i].State == CmdWaiting {
			v.Rows[i].Next = true // the first one runs at the next safe moment
			break
		}
	}

	var pv struct {
		Decision *struct{ Why string } `json:"decision"`
		Meter    *struct {
			ETA int `json:"etaSeconds"`
		} `json:"meter"`
	}
	if s.Plan != nil && decode(s.Plan(), &pv) && pv.Decision != nil && pv.Decision.Why != "" {
		v.Plan = &CommandPlan{Text: safeText(pv.Decision.Why, 160)}
		if pv.Meter != nil && pv.Meter.ETA > 0 {
			e := pv.Meter.ETA
			v.Plan.ETASeconds = &e
		}
	}

	v.Verbs = buildVerbs(s, ov, haveRepair, rv.Trip != nil, func() (string, bool, string) {
		if rv.Trip == nil {
			return "Repair", true, ""
		}
		b := rv.Trip.Button
		l := b.Label
		if l == "" {
			l = "Repair"
		}
		return l, b.Enabled, b.Reason
	}, v.Paused)
	return v
}

var repairWord = map[string]string{"queued": "asked; waiting to be seen", "waiting": "safe moment first", "going": "going to town",
	"repairing": "repairing in town", "returning": "going back", "farming": "farming gold first"}

func cmdRound(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }

func orderIntent(kind string) string {
	switch kind {
	case OrderHere:
		return "fight_here"
	case OrderHome:
		return "town"
	}
	return kind
}

func activeWhy(o Order, now float64) string {
	var p []string
	if o.Minutes > 0 && o.StartedAt > 0 {
		p = append(p, fmt.Sprintf("%d min left", max(0, o.Minutes-int((now-o.StartedAt)/60))))
	}
	if len(p) == 0 && o.Kills > 0 { // the row is one line: minutes first, kills only when there are no minutes
		p = append(p, fmt.Sprintf("%d kills", o.Kills))
	}
	return strings.Join(p, ", ")
}

func blockedRow(text string) CommandBlocked {
	return CommandBlocked{ID: "journey:blocked", Code: "journey_blocked", Text: safeText(text, 160), Actions: []string{"go_now", "hold"}}
}

// itemReq is the part of an inventory desk request this view reads.
type itemReq struct {
	ID     uint32 `json:"id"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Reason string `json:"reason"`
	At     int64  `json:"at"`
	Wait   string `json:"wait"`
	Group  string `json:"group"`
}

func (q itemReq) row(now int64) (CommandRow, bool) {
	m, ok := commandState(srcItem, q.State)
	if !ok {
		return CommandRow{}, false
	}
	if m == CmdFailed && now-q.At > failedShowSec {
		return CommandRow{}, false
	}
	kind := q.Kind
	if kind == "" {
		kind = "equip"
	}
	r := CommandRow{ID: fmt.Sprintf("inv:%d:%s", q.ID, kind), Intent: itemIntent(kind), Text: safeText(itemText(kind, q.Name), 120), Source: "inventory", State: m}
	switch m {
	case CmdQueued:
		switch q.Wait {
		case "safe":
			r.State, r.Why = CmdWaiting, &CommandWhy{"player_items", "safe moment first"}
		case "town":
			r.State, r.Why = CmdWaiting, &CommandWhy{"player_items", "waiting for the next town visit"}
		default:
			r.Why = &CommandWhy{"player_items", "queued"}
		}
	case CmdActive:
		r.Why = &CommandWhy{"player_items", "working on it"}
	case CmdFailed:
		r.Why = &CommandWhy{"player_items", safeText(q.Reason, 120)}
	}
	if m == CmdQueued { // never from "doing" (docs/inventory-actions.md)
		r.Can.Cancel = true
		if q.Group != "" {
			r.Cancel = &CommandCancel{Path: "/api/inventory/sets/cancel", Body: map[string]any{"group": q.Group}}
		} else {
			r.Cancel = &CommandCancel{Path: "/api/inventory/cancel", Body: map[string]any{"id": q.ID, "kind": kind}}
		}
	}
	return r, true
}

func decode(b []byte, v any) bool {
	return len(b) > 0 && json.Unmarshal(b, v) == nil
}

// buildVerbs lists the buttons whose desk is bound (C.4: at most five).
func buildVerbs(s CommandSources, ov OrdersView, haveRepair, haveTrip bool, repairBtn func() (string, bool, string), paused bool) []CommandVerb {
	var vs []CommandVerb
	if s.Orders != nil {
		reason := ""
		if len(ov.Areas) == 0 {
			reason = "no orderable area yet"
		}
		vs = append(vs, CommandVerb{"go", "Go to", reason == "", reason}, CommandVerb{"town", "Town", true, ""})
	}
	if haveRepair {
		l, en, why := repairBtn()
		vs = append(vs, CommandVerb{"repair", l, en, why})
	}
	if s.Hooks.Paused != nil {
		l := "Pause"
		if paused {
			l = "Resume"
		}
		vs = append(vs, CommandVerb{"pause", l, true, ""})
	}
	if s.Items != nil || s.GoldFarm != nil {
		vs = append(vs, CommandVerb{"more", "More", true, ""})
	}
	return vs
}

// CommandsHandler serves GET /api/commands.
func CommandsHandler(src func() CommandSources) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, 200, BuildCommands(src()))
	})
}

// CommandSources builds the sources from what the Fun holds, bound late: a desk bound after the first request shows up
// at the next one.
func (f *Fun) CommandSources() CommandSources {
	s := CommandSources{Orders: f.Hub.OrdersView, Hooks: f.Cmd}
	s.Safety = func() CommandSafety {
		switch f.Hub.Uptime().State {
		case KindDeath, KindRecovery:
			return CommandSafety{Active: true, Text: "Safety first: the hero is recovering. Your orders wait."}
		}
		return CommandSafety{}
	}
	if f.Queue != nil {
		s.Queue = func() *OrderQueueView { v := f.Queue.View(); return &v }
	}
	if f.Loop != nil {
		s.Loop = func() *LoopView { v := f.Loop.View(); return &v }
	}
	if f.Repair != nil {
		s.Repair = func() []byte {
			if d := f.Repair.desk(); d != nil {
				return marshal(d.View())
			}
			return nil
		}
	}
	if f.Inventory != nil {
		s.Items = f.Inventory.viewNoWait
	}
	if f.GoldFarm != nil {
		s.GoldFarm = func() []byte {
			if d := f.GoldFarm.desk(); d != nil {
				return marshal(d.View())
			}
			return nil
		}
	}
	if f.Campaign != nil {
		s.Plan = func() []byte {
			if d := f.Campaign.desk(); d != nil {
				return marshal(d.View())
			}
			return nil
		}
	}
	return s
}

func marshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// CommandsHandler serves GET /api/commands from the Fun's desks.
func (f *Fun) CommandsHandler() http.Handler { return CommandsHandler(f.CommandSources) }
