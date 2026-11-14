package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// the content of a scenario differs per class so a class-shaped value cannot hide behind the default hero
type cmdClass struct{ area, item, area2 string }

var cmdClasses = map[string]cmdClass{
	"sorceress": {"Cold Plains", "Frost Rune Staff", "Blood Moor"},
	"paladin":   {"Stony Field", "Tower Shield", "Dark Wood"},
}

func ordersFixture(cl cmdClass, orders ...Order) func() OrdersView {
	return func() OrdersView {
		return OrdersView{Schema: ordersSchema, Seq: 41, Now: 3600, Orders: orders, Areas: []string{cl.area, cl.area2}, Kinds: []string{OrderGo, OrderHere, OrderHome}, MaxMinutes: MaxOrderMinutes}
	}
}

func jsrc(s string) func() []byte { return func() []byte { return []byte(s) } }

const repairIdle = `{"schema":2,"trip":{"state":"ready","line":"","button":{"action":"request","label":"Repair now","enabled":true}}}`

func scenario(name string, cl cmdClass) CommandSources {
	base := CommandSources{Orders: ordersFixture(cl), Repair: jsrc(repairIdle), Items: jsrc(`{"queue":[],"recent":[]}`),
		GoldFarm: jsrc(`{"on":false,"text":"Farm gold"}`), Hooks: CommandHooks{Paused: func() bool { return false }},
		Plan: jsrc(`{"decision":{"why":"Farming ` + cl.area + ` for level 10 to unlock Blood Raven"},"meter":{"etaSeconds":1500}}`),
		Now:  func() time.Time { return time.Unix(1_800_000_000, 0) }}
	active := Order{ID: 12, Kind: OrderGo, Area: cl.area, Minutes: 30, State: OrderActive, Text: "Farm " + cl.area + " for 30 min", StartedAt: 3000, Kills: 41}
	switch name {
	case "three-one-failed":
		base.Orders = ordersFixture(cl, active)
		base.Items = jsrc(`{"queue":[{"id":44,"kind":"identify","name":"` + cl.item + `","state":"queued","wait":"safe","at":1799999000}],
			"recent":[{"id":45,"kind":"trash","name":"Cracked Helm","state":"failed","reason":"no room on the ground here","at":1799999500}]}`)
	case "blocked":
		base.Orders = ordersFixture(cl, active)
		base.Hooks.Blocked = func() string { return "Journey blocked: Blood Raven needs level 9 (you are 7)" }
	case "paused":
		base.Orders = ordersFixture(cl, active)
		base.Hooks.Paused = func() bool { return true }
	case "safety":
		base.Orders = ordersFixture(cl, active)
		base.Safety = func() CommandSafety {
			return CommandSafety{Active: true, Text: "Safety first: retreating to town. Your orders wait."}
		}
	}
	return base
}

func TestCommandsGoldens(t *testing.T) {
	if err := os.MkdirAll("testdata/commands", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"empty", "three-one-failed", "blocked", "paused", "safety"} {
		for class, cl := range cmdClasses {
			rec := httptest.NewRecorder()
			s := scenario(name, cl)
			CommandsHandler(func() CommandSources { return s }).ServeHTTP(rec, httptest.NewRequest("GET", CommandsRoute, nil))
			if rec.Code != 200 {
				t.Fatalf("%s-%s: %d", name, class, rec.Code)
			}
			golden(t, "commands/view-"+name+"-"+class, rec.Body.String())
		}
	}
}

// Every state a source can report maps to exactly one row state, and the table has no word the source lacks.
func TestCommandStateTable(t *testing.T) {
	known := map[string][]string{
		srcOrder:    {OrderQueued, OrderActive, OrderDone, OrderCancelled, OrderFailed},
		srcQueue:    {QueueWaiting, QueueRunning, QueueDone, QueueFailed, QueueCancelled},
		srcLoop:     {LoopActive, LoopDone, LoopCancelled, LoopFailed},
		srcItem:     {"queued", "doing", "done", "failed", "cancelled"},
		srcRepair:   {"queued", "waiting", "going", "repairing", "returning", "farming", "done", "failed", "cancelled"},
		srcGoldFarm: {"on"},
	}
	rows := map[string]bool{CmdQueued: true, CmdWaiting: true, CmdActive: true, CmdDone: true, CmdFailed: true, CmdCancelled: true}
	for src, states := range known {
		tab := sourceStates[src]
		if len(tab) != len(states) {
			t.Errorf("%s: table has %d states, the source has %d", src, len(tab), len(states))
		}
		for _, s := range states {
			m, ok := commandState(src, s)
			if !ok || !rows[m] {
				t.Errorf("%s %q maps to %q, %v", src, s, m, ok)
			}
		}
		if _, ok := commandState(src, "no-such-state"); ok {
			t.Errorf("%s: an unknown state must not map", src)
		}
	}
	if _, ok := commandState("nope", "queued"); ok {
		t.Error("an unknown source must not map")
	}
}

// A row's Cancel names a route that really exists as a POST route, and only a row that can be cancelled has one.
func TestCommandCancelRoutesExist(t *testing.T) {
	post := map[string]bool{"/api/orders/cancel": true, OrderQueueRoute: true, OrderLoopRoute: true, GoldFarmRoute: true}
	for _, p := range append(append([]string{}, RepairPostRoutes...), InvPostRoutes...) {
		post[p] = true
	}
	cl := cmdClasses["sorceress"]
	s := scenario("three-one-failed", cl)
	s.Orders = ordersFixture(cl, Order{ID: 12, Kind: OrderHere, Minutes: 30, State: OrderActive, Text: "Fight here for 30 min", StartedAt: 3000}) // a loop holds go and home orders only
	s.Queue = func() *OrderQueueView {
		return &OrderQueueView{Entries: []QueueEntry{{ID: 3, Kind: OrderGo, Area: cl.area2, Minutes: 20, Text: "Farm " + cl.area2 + " for 20 min", State: QueueWaiting}}}
	}
	s.Loop = func() *LoopView { return &LoopView{State: LoopActive, Text: "Loop: x", Phase: LoopPhaseFarming} }
	s.Repair = jsrc(`{"request":{"state":"requested","kind":"farm","phase":"farming","have":140,"target":300},"trip":{"line":"Farming gold","button":{"action":"cancel","label":"Cancel","enabled":true}}}`)
	s.GoldFarm = jsrc(`{"on":true,"text":"Farming gold in Blood Moor"}`)
	s.Items = jsrc(`{"queue":[{"id":7,"kind":"equip","name":"Helm","state":"queued"},{"id":8,"kind":"trash","name":"Gem","state":"doing"},{"id":9,"kind":"equip","name":"Ring","state":"queued","group":"g1"}]}`)
	v := BuildCommands(s)
	seen := map[string]bool{}
	for _, r := range v.Rows {
		seen[strings.SplitN(r.ID, ":", 2)[0]] = true
		if r.Can.Cancel != (r.Cancel != nil) {
			t.Errorf("%s: can.cancel %v but cancel %v", r.ID, r.Can.Cancel, r.Cancel)
		}
		if r.Cancel != nil && !post[r.Cancel.Path] {
			t.Errorf("%s: cancel route %s is not a POST route", r.ID, r.Cancel.Path)
		}
		if r.ID == "inv:8:trash" && r.Can.Cancel {
			t.Error("an item action that is being done cannot be cancelled")
		}
	}
	for _, src := range []string{"order", "queue", "loop", "repair", "inv", "farmgold"} {
		if !seen[src] {
			t.Errorf("no %s row in %v", src, v.Rows)
		}
	}
}

// The ladder: item actions (identify first), repair, gold, the loop, then zone orders; only the first waiting row is next.
func TestCommandLadderAndNext(t *testing.T) {
	cl := cmdClasses["paladin"]
	s := scenario("empty", cl)
	s.Orders = ordersFixture(cl,
		Order{ID: 1, Kind: OrderGo, Area: cl.area, Minutes: 30, State: OrderActive, Text: "Farm", StartedAt: 3500},
		Order{ID: 2, Kind: OrderHome, State: OrderQueued, Text: "Rest in town"})
	s.Items = jsrc(`{"queue":[{"id":1,"kind":"equip","name":"A","state":"queued"},{"id":2,"kind":"identify","name":"B","state":"queued","wait":"town"}]}`)
	v := BuildCommands(s)
	var ids []string
	for _, r := range v.Rows {
		ids = append(ids, r.ID)
	}
	want := []string{"inv:2:identify", "inv:1:equip", "order:1", "order:2"}
	if strings.Join(ids, " ") != strings.Join(want, " ") {
		t.Fatalf("ladder %v, want %v", ids, want)
	}
	n := 0
	for _, r := range v.Rows {
		if r.Next {
			n++
			if r.ID != "inv:2:identify" || r.State != CmdWaiting {
				t.Errorf("next is %s %s", r.ID, r.State)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d rows marked next", n)
	}
}

// A queue entry that runs and the loop stand for the watch order they placed: no row twice.
func TestCommandNoDuplicateRows(t *testing.T) {
	cl := cmdClasses["sorceress"]
	s := scenario("empty", cl)
	s.Orders = ordersFixture(cl, Order{ID: 5, Kind: OrderGo, Area: cl.area, Minutes: 20, State: OrderActive, Text: "Farm", StartedAt: 3000})
	s.Queue = func() *OrderQueueView {
		return &OrderQueueView{Entries: []QueueEntry{{ID: 1, Kind: OrderGo, Text: "Farm", State: QueueRunning, OrderID: 5}}}
	}
	if v := BuildCommands(s); len(v.Rows) != 1 || v.Rows[0].ID != "queue:1" {
		t.Errorf("queue: %+v", v.Rows)
	}
	s.Queue = nil
	s.Loop = func() *LoopView { return &LoopView{State: LoopActive, Text: "Loop: a", Phase: "farming"} }
	if v := BuildCommands(s); len(v.Rows) != 1 || v.Rows[0].ID != "loop:1" {
		t.Errorf("loop: %+v", v.Rows)
	}
}

// Unbound desks have no rows and no verbs; a repair trip carries its phase, a farm leg its real counter.
func TestCommandVerbsFollowDesks(t *testing.T) {
	v := BuildCommands(CommandSources{Orders: ordersFixture(cmdClasses["sorceress"])})
	var got []string
	for _, x := range v.Verbs {
		got = append(got, x.Intent)
	}
	if strings.Join(got, ",") != "go,town" {
		t.Errorf("verbs %v", got)
	}
	if v.Rows == nil || v.Blocked == nil || v.Verbs == nil {
		t.Error("lists must be [] not null")
	}
	s := scenario("empty", cmdClasses["sorceress"])
	s.Repair = jsrc(`{"request":{"state":"requested","kind":"farm","phase":"farming","have":140,"target":300},"trip":{"button":{"label":"Cancel","enabled":true}}}`)
	r := BuildCommands(s).Rows[0]
	if r.State != CmdActive || r.Phase != "farming" || r.Progress == nil || r.Progress.Label != "140 of 300 gold" {
		t.Errorf("%+v", r)
	}
}

type raceDesk struct {
	mu   sync.Mutex
	view map[string]any
}

func (d *raceDesk) View() any {
	d.mu.Lock()
	defer d.mu.Unlock()
	c := map[string]any{}
	for k, v := range d.view {
		c[k] = v
	}
	return c
}
func (d *raceDesk) Request() (any, error) { return nil, nil }
func (d *raceDesk) Cancel() error         { return nil }
func (d *raceDesk) set(v map[string]any)  { d.mu.Lock(); d.view = v; d.mu.Unlock() }

// Concurrent desk updates and reads: run under -race. The route also answers through the real viewer.
func TestCommandsRace(t *testing.T) {
	g := newLoopRig(t, "Sorceress")
	g.openLoop()
	rd := &raceDesk{}
	g.fun.Repair = NewRepairAPI()
	g.fun.Repair.Bind(rd)
	h := g.fun.CommandsHandler()
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest("GET", CommandsRoute, nil))
				var v CommandsView
				if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &v) != nil || v.Schema != commandsSchema {
					t.Errorf("bad reply %d %s", rec.Code, rec.Body.String())
					return
				}
			}
		}()
	}
	for i := 0; i < 60; i++ {
		id, err := g.hub.GiveOrder(OrderGo, "Cold Plains", 5)
		if err == nil {
			_ = g.hub.CancelOrder(id.ID)
		}
		rd.set(map[string]any{"request": map[string]any{"state": "requested", "phase": []string{"queued", "going", "farming"}[i%3], "kind": "repair"}})
		g.fun.Settle()
	}
	close(stop)
	wg.Wait()
	resp, b := g.do("GET", g.view, CommandsRoute, funToken, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(b), commandsSchema) {
		t.Fatalf("GET through the viewer: %d %s", resp.StatusCode, b)
	}
	var ids []string
	var v CommandsView
	_ = json.Unmarshal(b, &v)
	for _, r := range v.Rows {
		ids = append(ids, r.ID)
	}
	sort.Strings(ids)
	if !hasID(ids, "repair:1") {
		t.Errorf("rows %v miss the repair", ids)
	}
}

func hasID(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
