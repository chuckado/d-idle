package watch

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/task"
	"d2lod-bot/internal/unlock"
)

// loopDesk is the repair side of the loop: the tests move its request state by hand.
type loopDesk struct {
	worn     bool
	err      error // what Request answers with
	state    string
	auto     bool
	note     string
	requests int
	cancels  int
}

func (d *loopDesk) Worn() bool { return d.worn }
func (d *loopDesk) Request() error {
	if d.err != nil {
		return d.err
	}
	d.requests++
	d.state, d.auto, d.note = LoopReqOpen, false, ""
	return nil
}
func (d *loopDesk) Status() (string, bool, string) { return d.state, d.auto, d.note }
func (d *loopDesk) Cancel() error                  { d.cancels++; d.state = LoopReqCancelled; return nil }

type loopRefusal string

func (r loopRefusal) Error() string { return "refused: " + string(r) }
func (r loopRefusal) Kind() string  { return string(r) }

type loopRig struct {
	*queueRig
	desk *loopDesk
	path string
}

func newLoopRig(t *testing.T, class string) *loopRig {
	t.Helper()
	l, err := unlock.Shipped()
	if err != nil {
		t.Fatal(err)
	}
	eng, err := unlock.Open(t.TempDir(), strings.ToLower(class), l, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := &loopRig{desk: &loopDesk{worn: true}, path: filepath.Join(t.TempDir(), "orderloop.json")}
	g.queueRig = &queueRig{eng: eng}
	g.queueRig.funRig = newFunRigWith(t, class, 5, func(r *funRig, fc *FunConfig) {
		fc.Unlock = eng
		fc.OrderQueue = &OrderQueueConfig{Quote: r.strip.Quote, Halt: func() string { return g.halt }}
		fc.OrderLoop = &OrderLoopConfig{Path: g.path, Quote: r.strip.Quote, Halt: func() string { return g.halt }}
	})
	g.fun.Loop.Bind(g.desk)
	return g
}

func (g *loopRig) openLoop() { g.fun.Unlocks.Seen(LoopManualEvent) }

func (g *loopRig) start(spec string) (int, string) {
	resp, b := g.do("POST", g.answer, OrderLoopRoute, funToken, `{"action":"start",`+spec+`}`)
	return resp.StatusCode, string(b)
}

func (g *loopRig) lview() LoopView {
	var v LoopView
	g.getJSON(OrderLoopRoute, &v)
	return v
}

// legDone ends the running order after a full farm leg.
func (g *loopRig) legDone(state string) {
	g.clock.Advance(21 * time.Minute)
	g.finishRunning(state)
	g.fun.Settle()
}

func (g *loopRig) only(t *testing.T) Order {
	t.Helper()
	open := g.openOrders()
	if len(open) != 1 {
		t.Fatalf("want one open order, got %+v", open)
	}
	return open[0]
}

var classes = []string{"Sorceress", "Paladin"}

func TestLoopThreeRoundsGoRepairGoAndOpensTheQueue(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			g.openLoop()
			if code, b := g.start(`"area":"Blood Moor","minutes":20,"rounds":3`); code != 200 {
				t.Fatalf("start: %d %s", code, b)
			}
			if g.eng.Allowed(OrderQueueID) {
				t.Fatal("the queue was open before any round")
			}
			for r := 1; r <= 3; r++ {
				g.fun.Settle()
				if o := g.only(t); o.Kind != OrderGo || o.Area != "Blood Moor" || o.Minutes != 20 {
					t.Fatalf("round %d farm order: %+v", r, o)
				}
				if v := g.lview(); v.Round != r || v.Phase != LoopPhaseFarming {
					t.Fatalf("round %d view: %+v", r, v)
				}
				g.legDone(OrderDone)
				if g.desk.requests != r || len(g.openOrders()) != 0 {
					t.Fatalf("round %d: requests %d, open %v", r, g.desk.requests, g.openOrders())
				}
				if v := g.lview(); v.Phase != LoopPhaseRepair {
					t.Fatalf("round %d repair phase: %+v", r, v)
				}
				g.fun.Settle() // the request is still open: nothing moves
				g.desk.state = LoopReqDone
				g.fun.Settle()
			}
			v := g.lview()
			if v.State != LoopDone || v.Finished != 3 || !strings.Contains(v.StoppedWhy, "3 rounds") {
				t.Fatalf("end view: %+v", v)
			}
			if !g.eng.Allowed(OrderQueueID) {
				t.Fatal("three finished rounds did not open the first queue slot")
			}
			if _, ok := g.fun.Loop.Command(); ok {
				t.Fatal("a finished loop still has a command row")
			}
		})
	}
}

func TestLoopDamagedSkipsTheSmithAndAlwaysAsks(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			g.openLoop()
			g.desk.worn = false
			if code, b := g.start(`"area":"Blood Moor","minutes":10,"rounds":2`); code != 200 {
				t.Fatalf("start: %d %s", code, b)
			}
			g.fun.Settle()
			g.legDone(OrderDone)
			if g.desk.requests != 0 {
				t.Fatal("damaged mode asked the smith with nothing worn")
			}
			if o := g.only(t); o.Kind != OrderHome {
				t.Fatalf("expected the home order, got %+v", o)
			}
			g.finishRunning(OrderDone)
			g.fun.Settle()
			if o := g.only(t); o.Kind != OrderGo {
				t.Fatalf("round two: %+v", o)
			}
			if g.lview().Round != 2 {
				t.Fatalf("round: %+v", g.lview())
			}
			// always: the desk answers undamaged, which counts as done and goes home
			if err := g.fun.Loop.Stop(); err != nil {
				t.Fatal(err)
			}
			g.desk.err = loopRefusal("undamaged")
			if code, b := g.start(`"area":"Blood Moor","minutes":10,"rounds":1,"repair":"always"`); code != 200 {
				t.Fatalf("start always: %d %s", code, b)
			}
			g.fun.Settle()
			g.legDone(OrderDone)
			if o := g.only(t); o.Kind != OrderHome {
				t.Fatalf("undamaged answer should go home: %+v", o)
			}
			g.finishRunning(OrderDone)
			g.fun.Settle()
			if v := g.lview(); v.State != LoopDone {
				t.Fatalf("undamaged round did not finish: %+v", v)
			}
		})
	}
}

func TestLoopStopConditions(t *testing.T) {
	type tc struct {
		name  string
		state string
		why   string
		do    func(g *loopRig)
	}
	cases := []tc{
		{"player cancel", LoopCancelled, "cancelled by you", func(g *loopRig) {
			if err := g.fun.Loop.Stop(); err != nil {
				t.Fatal(err)
			}
			if len(g.openOrders()) != 0 {
				t.Fatal("the farm order stayed open after the cancel")
			}
		}},
		{"zone order with replace", LoopCancelled, "another order", func(g *loopRig) {
			if _, err := g.hub.Give(OrderRequest{Kind: OrderGo, Area: "Cold Plains", Minutes: 10, Replace: true}); err != nil {
				t.Fatal(err)
			}
			g.fun.Settle()
		}},
		{"death", LoopFailed, "died", func(g *loopRig) { g.feed(goals.Event{Type: goals.EvDeath, Seq: 1}) }},
		{"guard refuses", LoopFailed, "farm leg failed", func(g *loopRig) {
			g.finishRunning(OrderFailed)
			g.fun.Settle()
		}},
		{"halt", LoopFailed, "low life", func(g *loopRig) { g.halt = "low life pause"; g.fun.Settle() }},
		{"repair failed", LoopFailed, "repair failed", func(g *loopRig) {
			g.legDone(OrderDone)
			g.desk.state, g.desk.note = LoopReqFailed, "smith unreachable"
			g.fun.Settle()
		}},
		{"repair dropped by the bot", LoopFailed, "dropped by the bot", func(g *loopRig) {
			g.legDone(OrderDone)
			g.desk.state, g.desk.auto = LoopReqCancelled, true
			g.fun.Settle()
		}},
		{"repair cancelled by the player", LoopCancelled, "repair was cancelled", func(g *loopRig) {
			g.legDone(OrderDone)
			g.desk.state = LoopReqCancelled
			g.fun.Settle()
		}},
		{"pact forbids repair", LoopFailed, "Pact", func(g *loopRig) {
			g.desk.err = loopRefusal("pact")
			g.legDone(OrderDone)
		}},
		{"farm leg ends at once, three times", LoopFailed, "ended at once", func(g *loopRig) {
			for i := 0; i < 3; i++ {
				g.finishRunning(OrderDone) // no time passes: the clock stays
				g.fun.Settle()
				g.desk.state = LoopReqDone
				g.fun.Settle()
			}
		}},
	}
	for _, class := range classes {
		for _, c := range cases {
			t.Run(class+"/"+c.name, func(t *testing.T) {
				g := newLoopRig(t, class)
				g.openLoop()
				if code, b := g.start(`"area":"Blood Moor","minutes":20,"rounds":5`); code != 200 {
					t.Fatalf("start: %d %s", code, b)
				}
				g.fun.Settle()
				g.only(t)
				c.do(g)
				v := g.lview()
				if v.State != c.state || !strings.Contains(v.StoppedWhy, c.why) {
					t.Fatalf("view after %s: %+v", c.name, v)
				}
				if g.fun.Loop.Holding() {
					t.Fatal("a stopped loop still holds the place")
				}
				before := len(g.hub.OrdersView().Orders)
				g.clock.Advance(time.Hour)
				g.fun.Settle()
				if len(g.hub.OrdersView().Orders) != before {
					t.Fatal("a stopped loop placed another order")
				}
			})
		}
	}
}

func TestLoopBacksOffThenGivesUpWhenTheHeroIsBusy(t *testing.T) {
	g := newLoopRig(t, "Sorceress")
	g.openLoop()
	if _, err := g.hub.Give(OrderRequest{Kind: OrderHere, Minutes: 600 / 10}); err != nil {
		t.Fatal(err)
	}
	if code, b := g.start(`"area":"Blood Moor","minutes":20,"rounds":2`); code != 200 {
		t.Fatalf("start: %d %s", code, b)
	}
	g.fun.Settle()
	if v := g.lview(); v.Phase != LoopPhaseWaiting || v.Waiting == "" || v.State != LoopActive {
		t.Fatalf("first backoff: %+v", v)
	}
	g.fun.Settle() // no time passed: no second try
	for i, wait := range []time.Duration{30 * time.Second, 60 * time.Second, 120 * time.Second} {
		g.clock.Advance(wait)
		g.fun.Settle()
		if i < 2 {
			if v := g.lview(); v.State != LoopActive {
				t.Fatalf("gave up too early at try %d: %+v", i+2, v)
			}
		}
	}
	if v := g.lview(); v.State != LoopFailed || !strings.Contains(v.StoppedWhy, "gave up") {
		t.Fatalf("end: %+v", v)
	}
}

func TestLoopRestartResumesAtTheStartOfTheRound(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			g.openLoop()
			g.start(`"area":"Blood Moor","minutes":20,"rounds":4`)
			g.fun.Settle()
			g.legDone(OrderDone)
			g.desk.state = LoopReqDone
			g.fun.Settle()
			g.fun.Settle() // round 2 farm leg is placed
			if g.lview().Round != 2 {
				t.Fatalf("setup: %+v", g.lview())
			}
			g.finishRunning(OrderCancelled) // the old process died with its order
			l2, err := NewOrderLoop(g.fun.Unlocks, g.hub, OrderLoopConfig{Path: g.path})
			if err != nil {
				t.Fatal(err)
			}
			l2.Bind(g.desk)
			v := l2.View()
			if v.State != LoopActive || v.Round != 2 || v.Rounds != 4 || v.Phase != "" || v.Finished != 1 {
				t.Fatalf("resumed view: %+v", v)
			}
			l2.Tick()
			if o := g.only(t); o.Kind != OrderGo || o.Area != "Blood Moor" {
				t.Fatalf("the resumed round did not place its farm leg: %+v", o)
			}
		})
	}
}

func TestLoopLadderRowsStrictOrder(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			if g.eng.Allowed(OrderLoopID) {
				t.Fatal("the loop row is open at the start")
			}
			g.fun.Settle()
			// a repair the player asked for, before any zone order: not a cycle
			g.desk.state = LoopReqDone
			g.fun.Settle()
			if g.eng.Allowed(OrderLoopID) {
				t.Fatal("a repair without a zone order opened the loop")
			}
			if _, err := g.hub.Give(OrderRequest{Kind: OrderGo, Area: "Blood Moor", Minutes: 10}); err != nil {
				t.Fatal(err)
			}
			g.fun.Settle()
			if !g.eng.Allowed(OrderLoopID) {
				t.Fatal("zone order then a done repair did not open the loop")
			}
			if g.eng.Allowed(OrderQueueID) {
				t.Fatal("the loop row opened the queue")
			}
			// the Den of Evil no longer opens the first slot
			g.feed(questEv(task.QDen, 2, 0, 1))
			if g.eng.Allowed(OrderQueueID) {
				t.Fatal("the Den of Evil opened the queue")
			}
			g.fun.Unlocks.Seen(LoopRoundsEvent)
			if !g.eng.Allowed(OrderQueueID) {
				t.Fatal("loop_rounds_3 did not open the queue")
			}
		})
	}
}

func TestLoopManualCycleIgnoresTheLoopsOwnRepair(t *testing.T) {
	g := newLoopRig(t, "Sorceress")
	g.openLoop() // the row is open: observeManual is off, the loop's own repairs never emit the event
	g.desk.state = LoopReqDone
	g.fun.Settle()
	if g.fun.Loop.st.ManualSeen {
		t.Fatal("observed a manual cycle with the row open")
	}
}

// A hero whose queue row was opened by the old trigger (the Den) keeps it after the ladder change.
func TestOrderQueueOpenedByTheDenStaysOpen(t *testing.T) {
	raw, err := os.ReadFile("../../data/unlocks/ladder.json")
	if err != nil {
		t.Fatal(err)
	}
	newTrig := `"trigger": {"kind": "seen", "args": {"what": "loop_rounds_3"}}`
	if !strings.Contains(string(raw), newTrig) {
		t.Fatal("the shipped ladder lost the loop_rounds_3 trigger")
	}
	oldTrig := `"trigger": {"kind": "quest_stage", "args": {"quest": "den_of_evil", "min": 2}}`
	oldLadder, err := unlock.Parse([]byte(strings.Replace(string(raw), newTrig, oldTrig, 1)))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	old, err := unlock.Open(dir, "sorceress", oldLadder, nil)
	if err != nil {
		t.Fatal(err)
	}
	old.Observe(unlock.Event{Type: unlock.EvQuestStage, Name: "den_of_evil", N: 2})
	if !old.Allowed(OrderQueueID) {
		t.Fatal("setup: the old trigger did not open the queue")
	}
	shipped, err := unlock.Shipped()
	if err != nil {
		t.Fatal(err)
	}
	now, err := unlock.Open(dir, "sorceress", shipped, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !now.Allowed(OrderQueueID) {
		t.Fatal("a hero with the queue open lost it under the new ladder")
	}
	if now.Allowed(OrderLoopID) {
		t.Fatal("the loop row opened without its event")
	}
}

func TestLoopRoutes(t *testing.T) {
	for _, class := range classes {
		lc := strings.ToLower(class)
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			gold := func(name string) {
				b, _ := json.MarshalIndent(g.lview(), "", "  ")
				golden(t, "orderloop/view-"+name+"-"+lc, string(b)+"\n")
			}
			if resp, _ := g.do("GET", g.view, OrderLoopRoute, "", ""); resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("GET without a token: %d", resp.StatusCode)
			}
			if resp, _ := g.do("POST", g.answer, OrderLoopRoute, "", `{"action":"stop"}`); resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("POST without a token: %d", resp.StatusCode)
			}
			gold("locked")
			if code, b := g.start(`"area":"Blood Moor","minutes":10`); code != http.StatusConflict || !strings.Contains(b, `"locked"`) {
				t.Fatalf("start while locked: %d %s", code, b)
			}
			g.openLoop()
			gold("idle")
			for _, bad := range []string{`"area":"Moo Moo Farm","minutes":10`, `"area":"Blood Moor","minutes":4`, `"area":"Blood Moor","minutes":10,"rounds":21`,
				`"area":"Blood Moor","minutes":10,"rounds":2,"forever":true`, `"area":"Blood Moor","minutes":10,"repair":"never"`} {
				if code, b := g.start(bad); code != http.StatusBadRequest || !strings.Contains(b, "bad-order") {
					t.Errorf("%s: %d %s", bad, code, b)
				}
			}
			if code, b := g.post2(`{"action":"fly"}`); code != http.StatusBadRequest {
				t.Errorf("unknown action: %d %s", code, b)
			}
			if code, b := g.post2(`{"action":"stop"}`); code != http.StatusConflict || !strings.Contains(b, `"none"`) {
				t.Errorf("stop with no loop: %d %s", code, b)
			}
			if code, b := g.start(`"area":"Blood Moor","minutes":20,"rounds":3`); code != 200 {
				t.Fatalf("start: %d %s", code, b)
			}
			if code, b := g.start(`"area":"Blood Moor","minutes":20`); code != http.StatusConflict || !strings.Contains(b, `"active"`) {
				t.Errorf("second start: %d %s", code, b)
			}
			g.fun.Settle()
			gold("active")
			cmd, ok := g.fun.Loop.Command()
			if !ok || cmd.Text != "Loop: Blood Moor 20 min, repair, repeat (round 1 of 3)" {
				t.Fatalf("command row: %+v %v", cmd, ok)
			}
			if code, _ := g.post2(`{"action":"stop"}`); code != 200 {
				t.Fatalf("stop: %d", code)
			}
			gold("cancelled")
		})
	}
}

func TestLoopForeverAndTheQueueRefuseEachOther(t *testing.T) {
	g := newLoopRig(t, "Sorceress")
	g.openLoop()
	g.denCleared() // slot one
	g.add(t, "Blood Moor", `,"minutes":10`)
	if code, b := g.start(`"area":"Blood Moor","minutes":20,"forever":true`); code != http.StatusConflict || !strings.Contains(b, "queue-busy") {
		t.Fatalf("forever with a waiting entry: %d %s", code, b)
	}
	if code, b := g.start(`"area":"Blood Moor","minutes":20,"rounds":2`); code != 200 {
		t.Fatalf("a finite loop with a waiting entry: %d %s", code, b)
	}
	g.fun.Settle()
	// the loop holds the place: the queue entry waits while the loop is between its steps
	g.legDone(OrderDone)
	for _, o := range g.openOrders() {
		t.Fatalf("an order was placed while the loop repairs: %+v", o)
	}
}

func (g *loopRig) post2(body string) (int, string) {
	resp, b := g.do("POST", g.answer, OrderLoopRoute, funToken, body)
	return resp.StatusCode, string(b)
}
