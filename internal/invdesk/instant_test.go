package invdesk

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/status"
)

var allCaps = []string{CapItemEquip, CapItemUnequip, CapItemMove, CapItemIdentify, CapItemDrop, CapGoldStash}

// flight counts verbs and clicks in flight on the shared channel: it must never be above one.
type flight struct {
	now, max atomic.Int32
}

func (f *flight) enter() {
	n := f.now.Add(1)
	for {
		m := f.max.Load()
		if n <= m || f.max.CompareAndSwap(m, n) {
			break
		}
	}
	time.Sleep(50 * time.Microsecond)
}
func (f *flight) leave() { f.now.Add(-1) }

// fakeFast is the backend: it records the verbs, refuses by verb name and counts concurrent use.
type fakeFast struct {
	mu    sync.Mutex
	caps  []string
	calls []string
	err   map[string]error // by verb: equip, unequip, drop, identify, move, gold
	boom  string           // a verb that panics
	fl    *flight
}

func (f *fakeFast) rec(verb, call string) error {
	if f.fl != nil {
		f.fl.enter()
		defer f.fl.leave()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
	if f.boom == verb {
		panic("verb exploded")
	}
	return f.err[verb]
}
func (f *fakeFast) Caps(context.Context) ([]string, error) { return f.caps, nil }
func (f *fakeFast) ItemEquip(_ context.Context, id uint32, slot int) (bool, error) {
	return false, f.rec("equip", "equip:"+itoa(id)+":"+itoa(uint32(slot)))
}
func (f *fakeFast) ItemUnequip(_ context.Context, id uint32) (bool, error) {
	return false, f.rec("unequip", "unequip:"+itoa(id))
}
func (f *fakeFast) ItemMove(_ context.Context, id uint32, panel string, _, _ int) (bool, error) {
	return false, f.rec("move", "move:"+itoa(id)+":"+panel)
}
func (f *fakeFast) ItemIdentify(_ context.Context, id uint32) (bool, error) {
	return false, f.rec("identify", "identify:"+itoa(id))
}
func (f *fakeFast) ItemDrop(_ context.Context, id uint32) error {
	return f.rec("drop", "drop:"+itoa(id))
}
func (f *fakeFast) GoldStash(_ context.Context, amount int, deposit bool) (int, int, error) {
	d := "withdraw"
	if deposit {
		d = "deposit"
	}
	return 0, 0, f.rec("gold", "gold:"+d+":"+itoa(uint32(amount)))
}
func (f *fakeFast) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

func itoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

// unEq is the click unequipper of the rig.
type unEq struct{ done []uint32 }

func (u *unEq) PlayerUnequip(_ context.Context, id uint32) error {
	u.done = append(u.done, id)
	return nil
}

type fastEnv struct {
	d    *Desk
	fv   *fakeFast
	fl   *FastLane
	eq   *fakeEq
	tr   *fakeTr
	fs   *fakeStash
	un   *unEq
	snap *d2.Snapshot
	good uint32 // a pack item the hero class may wear
}

// fastRig is the stash rig with every click executor, the lane and a backend with all verbs. Town or field by area.
func fastRig(t *testing.T, class d2.Class, area int) *fastEnv {
	t.Helper()
	d, fs, snap := stashRig(t, class, area)
	e := &fastEnv{d: d, fs: fs, snap: snap, fv: &fakeFast{caps: allCaps, err: map[string]error{}}, un: &unEq{},
		good: map[d2.Class]uint32{d2.Paladin: 1, d2.Sorceress: 2}[class]}
	e.eq, _ = d.eq.(*fakeEq)
	e.tr = &fakeTr{snap: snap}
	d.SetTrasher(e.tr)
	d.SetUnequipper(e.un)
	d.c.Inventory = func() status.InventoryData {
		return status.InventoryData{Gold: 5000, StashGold: 20000,
			Equipment: map[string]*status.Item{"torso": {ID: 5, Code: "13", W: 2, H: 3}},
			Inventory: []status.Item{{ID: 1, Code: "10", W: 1, H: 3}, {ID: 2, Code: "11", W: 1, H: 2}, {ID: 4}},
			Stash:     []status.Item{{ID: 3, Code: "12"}}}
	}
	e.fl = &FastLane{Verbs: Serial{Ch: NewChannel(), V: e.fv}, On: on, Instant: on, Pause: func(context.Context, time.Duration) error { return nil }, Settles: 1}
	d.SetFast(e.fl)
	return e
}

var classes = []d2.Class{d2.Paladin, d2.Sorceress}

// Each verb-capable kind completes inside the POST: the request comes back done with how=fast, the backend got exactly
// one verb, the queue is empty and the click paths were not used.
func TestFastLaneEveryKindCompletes(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, town1)
		d := e.d
		for _, c := range []struct {
			name string
			do   func() (Request, error)
			call string
		}{
			{"equip", func() (Request, error) { return d.RequestEquip(e.good, "rhand") }, "equip:" + itoa(e.good) + ":" + itoa(uint32(d2.BodyRHand))},
			{"unequip", func() (Request, error) { return d.RequestUnequip("torso") }, "unequip:5"},
			{"trash", func() (Request, error) { return d.RequestTrash(e.good, true) }, "drop:" + itoa(e.good)},
			{"identify", func() (Request, error) { return d.RequestIdentify(1) }, "identify:1"},
			{"stash", func() (Request, error) { return d.RequestStash(1) }, "move:1:stash"},
			{"take", func() (Request, error) { return d.RequestTake(3) }, "move:3:inv"},
			{"gold", func() (Request, error) { return d.RequestGold(true, 1000) }, "gold:deposit:1000"},
		} {
			r, err := c.do()
			if err != nil || r.State != StateDone || r.How != HowFast {
				t.Fatalf("%v %s: %+v %v", class, c.name, r, err)
			}
			if got := e.fv.log(); got[len(got)-1] != c.call {
				t.Errorf("%v %s: verbs %v, want %s last", class, c.name, got, c.call)
			}
		}
		if d.Pending() != 0 || len(e.eq.worn) != 0 || len(e.tr.dropped) != 0 || len(e.un.done) != 0 || len(e.fs.visits) != 0 || len(e.fs.identified) != 0 {
			t.Errorf("%v: something took the queue or click path: pending %d", class, d.Pending())
		}
		if !d.Ignored(d2.Unit{ID: e.good, Class: 10, Item: &d2.Item{Quality: 4, ILvl: 5}}) && class == d2.Paladin {
			t.Errorf("%v: the dropped item is not on the looter's ignore list", class)
		}
		if v := d.View(); v.Instant == nil || v.Instant.Done[HowFast] != 7 || len(v.Recent) != 7 {
			t.Errorf("%v: telemetry %+v recent %d", class, v.Instant, len(v.Recent))
		}
	}
}

// A rule refusal of the verb is final: the verb's code and text come back, the request is failed in Recent, never queued.
func TestFastLaneRefusalCodesAreFinal(t *testing.T) {
	for _, class := range classes {
		for _, c := range []string{"item", "place", "slot", "identified", "class", "requirements", "twohand", "nospace", "dead"} {
			e := fastRig(t, class, town1)
			e.fv.err["equip"] = &VerbRefusal{Code: c, Why: "because " + c}
			_, err := e.d.RequestEquip(e.good, "rhand")
			if code(err) != c || err.Error() != "because "+c {
				t.Errorf("%v %s: %v", class, c, err)
			}
			v := e.d.View()
			if len(v.Queue) != 0 || len(v.Recent) != 1 || v.Recent[0].State != StateFailed || v.Recent[0].Code != c || v.Recent[0].Reason != "because "+c {
				t.Errorf("%v %s: queue %v recent %v", class, c, v.Queue, v.Recent)
			}
			if len(e.eq.worn) != 0 || e.d.Pinned(d2.BodyRHand) {
				t.Errorf("%v %s: a refused equip must neither click nor pin", class, c)
			}
		}
	}
}

// A verb that answers ok but did not move the item is reported failed (code verify), not retried and not done.
func TestFastLaneReadBackFailure(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, town1)
		e.fl.Fresh = func() (*d2.Snapshot, error) { return e.snap, nil } // the snapshot never shows the item worn
		_, err := e.d.RequestEquip(e.good, "rhand")
		if code(err) != "verify" {
			t.Fatalf("%v: %v", class, err)
		}
		if e.d.Pinned(d2.BodyRHand) || e.d.Pending() != 0 {
			t.Errorf("%v: pinned or still queued after a failed read-back", class)
		}
	}
}

// Everything that closes the lane leaves the request queued as before: nothing is sent, the bot's queue path does it.
func TestFastLaneClosedQueuesAsBefore(t *testing.T) {
	for _, class := range classes {
		for _, c := range []struct {
			name  string
			setup func(e *fastEnv)
		}{
			{"knob off", func(e *fastEnv) { e.fl.On = off }},
			{"no caps", func(e *fastEnv) { e.fv.caps = nil }},
			{"no verb for the kind", func(e *fastEnv) { e.fv.caps = []string{CapItemMove} }},
			{"named run", func(e *fastEnv) { e.fl.Named = on }},
			{"instant knob off", func(e *fastEnv) { e.fl.Instant = off }},
			{"vendor, stash or menu panel open", func(e *fastEnv) {
				for i := range e.snap.UI { // any index the table calls a blocking panel
					var u d2.UIState
					u[i] = 1
					if u.BlockingPanelOpen() {
						e.snap.UI[i] = 1
						break
					}
				}
				e.fl.Hold = func(s *d2.Snapshot) string {
					if s.UI.BlockingPanelOpen() {
						return "a panel is open"
					}
					return ""
				}
			}},
			{"busy cursor", func(e *fastEnv) {
				e.fl.Hold = func(s *d2.Snapshot) string { return "the cursor holds an item" }
			}},
			{"safety action", func(e *fastEnv) { e.fl.Hold = func(*d2.Snapshot) string { return "a safety action holds the input" } }},
			{"dead", func(e *fastEnv) { e.snap.Player.MaxLife, e.snap.Player.Life = 100, 0 }},
			{"another request is running", func(e *fastEnv) {
				e.d.st.Queue = append(e.d.st.Queue, Request{ID: 77, Kind: KindTrash, State: StateDoing})
			}},
			{"the verb says cursor", func(e *fastEnv) { e.fv.err["equip"] = &VerbRefusal{Code: "cursor", Why: "the hand is not empty"} }},
		} {
			e := fastRig(t, class, town1)
			c.setup(e)
			r, err := e.d.RequestEquip(e.good, "rhand")
			if err != nil || r.State != StateQueued || r.How != HowQueue || r.Wait != WaitSafe {
				t.Fatalf("%v %s: %+v %v", class, c.name, r, err)
			}
			if c.name != "the verb says cursor" && len(e.fv.log()) != 0 {
				t.Errorf("%v %s: a verb was sent: %v", class, c.name, e.fv.log())
			}
			if c.name == "another request is running" {
				continue
			}
			if e.d.Pending() != 1 {
				t.Fatalf("%v %s: pending %d", class, c.name, e.d.Pending())
			}
			e.d.RunPending(context.Background()) // the queue path (click) does it
			if len(e.eq.worn) != 1 || !e.d.Pinned(d2.BodyRHand) {
				t.Errorf("%v %s: queue path did not run: %v", class, c.name, e.eq.worn)
			}
		}
	}
}

// Without the lane (no SetFast) the request and the view are the old ones: no how, no wait, no instant block.
func TestNoFastLaneChangesNothing(t *testing.T) {
	d, eq, _ := rig(t, d2.Paladin, town1)
	r, err := d.RequestEquip(1, "rhand")
	if err != nil || r.How != "" || r.Wait != "" || d.View().Instant != nil {
		t.Fatalf("%+v %v", r, err)
	}
	d.RunPending(context.Background())
	if len(eq.worn) != 1 {
		t.Fatal("click path")
	}
}

// Stash, take and gold: at the chest (in town) instant; in the field they wait for the next town visit and then run
// without a new click; a request waiting can be cancelled.
func TestFastLaneStashTakeGoldNeedTheChest(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, field)
		d := e.d
		e.fl.capSet(context.Background()) // the view reads the cached list only
		for _, do := range []func() (Request, error){
			func() (Request, error) { return d.RequestStash(1) },
			func() (Request, error) { return d.RequestTake(3) },
			func() (Request, error) { return d.RequestGold(true, 500) },
		} {
			r, err := do()
			if err != nil || r.State != StateQueued || r.Wait != WaitTown || r.How != HowQueue {
				t.Fatalf("%v: %+v %v", class, r, err)
			}
		}
		if len(e.fv.log()) != 0 {
			t.Fatalf("%v: a stash verb was sent from the field: %v", class, e.fv.log())
		}
		v := d.View()
		if len(v.Queue) != 3 || v.Queue[0].Wait != WaitTown || v.Instant.Stash || v.Instant.Take || v.Instant.Gold || !v.Instant.Equip {
			t.Errorf("%v: view %+v", class, v.Instant)
		}
		if err := d.Cancel(0, KindGold); err != nil { // cancel while waiting
			t.Fatal(err)
		}
		if got := len(d.View().Queue); got != 2 {
			t.Errorf("%v: cancel left %d", class, got)
		}
		e.snap.Player.Area = town1 // arrival: the next tick runs it, no new click
		if d.Pending() != 2 {
			t.Fatalf("%v: pending on arrival %d", class, d.Pending())
		}
		d.RunPending(context.Background())
		if d.Pending() != 0 || len(e.fs.visits) != 1 {
			t.Errorf("%v: visit %v", class, e.fs.visits)
		}
		if v := d.View(); !v.Instant.Stash || !v.Instant.Take || !v.Instant.Gold {
			t.Errorf("%v: at the chest the view says %+v", class, v.Instant)
		}
		// at the chest the verb is sent at once
		if r, err := d.RequestStash(1); err != nil || r.State != StateDone || r.How != HowFast {
			t.Errorf("%v: %+v %v", class, r, err)
		}
		// the chest is too far for the verb: the request waits for the town visit
		e.fv.err["move"] = &VerbRefusal{Code: "far", Why: "not at the chest"}
		if r, err := d.RequestTake(3); err != nil || r.State != StateQueued || r.Wait != WaitSafe {
			t.Errorf("%v: far: %+v %v", class, r, err)
		}
	}
}

// The bot's own verbs and the player's instant requests never run together; neither does a click of the queue path
// run while a lane verb is on the channel (-race).
func TestFastLaneAndBotNeverInterleave(t *testing.T) {
	e := fastRig(t, d2.Paladin, town1)
	fl := &flight{}
	e.fv.fl = fl
	ch := NewChannel()
	serial := Serial{Ch: ch, V: e.fv}
	e.fl.Verbs = serial
	e.d.SetEquipper(flightEq{e.eq, fl})
	e.fv.err = map[string]error{}

	botVerbs := NewVerbEquipper(flightEq{e.eq, fl}, serial, on) // the bot's executor over the same Serial
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(2)
	go func() { // the bot: one goroutine, its own verbs and its queue loop (click path of the desk)
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				_ = botVerbs.PlayerEquip(context.Background(), 1, d2.BodyRHand)
				e.d.RunPending(context.Background())
			}
		}
	}()
	go func() { // the player: instant requests
		defer wg.Done()
		defer close(stop)
		for i := 0; i < 60; i++ {
			if _, err := e.d.RequestEquip(1, "rhand"); err != nil {
				t.Error(err)
			}
		}
	}()
	wg.Wait()
	if m := fl.max.Load(); m != 1 {
		t.Errorf("%d actions were on the channel at once", m)
	}
	if len(e.fv.log()) == 0 {
		t.Error("no verbs ran")
	}
}

// flightEq counts a click equip like a verb: a click while a verb is in flight would be seen as two.
type flightEq struct {
	*fakeEq
	fl *flight
}

func (f flightEq) PlayerEquip(ctx context.Context, id uint32, body int) error {
	f.fl.enter()
	defer f.fl.leave()
	return f.fakeEq.PlayerEquip(ctx, id, body)
}

// The refusals the requirement check already knows come back before anything is sent (C1's reasons plug in there).
func TestFastLaneChecksBeforeSending(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, town1)
		bad := map[d2.Class]uint32{d2.Paladin: 2, d2.Sorceress: 1}[class]
		if _, err := e.d.RequestEquip(bad, "rhand"); code(err) != "requirements" {
			t.Errorf("%v: %v", class, err)
		}
		if len(e.fv.log()) != 0 {
			t.Errorf("%v: sent a verb for a refused request", class)
		}
	}
}

// Schema 9 golden: the instant block, how=fast in Recent and a stash job waiting for town.
func TestViewSchema9InstantGolden(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, field)
		now := time.Unix(1_800_000_000, 0)
		e.d.c.Now = func() time.Time { return now }
		if _, err := e.d.RequestEquip(e.good, "rhand"); err != nil {
			t.Fatal(err)
		}
		if _, err := e.d.RequestTake(3); err != nil {
			t.Fatal(err)
		}
		v := e.d.View()
		v.Schema = 9 // schema 10 only added the sell kind: everything else in the view is byte for byte as before
		b, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		b = append(b, '\n')
		path := filepath.Join("testdata", "view_schema9_instant_"+class.String()+".json")
		if os.Getenv("UPDATE") != "" {
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(want) != string(b) {
			t.Errorf("%v: view differs from %s\n got: %s\nwant: %s", class, path, b, want)
		}
	}
}

var _ = errors.New

// With the instant knob off the view says so and every request queues; the verbs may still be on for the bot.
func TestFastLaneOwnKnob(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, town1)
		e.fl.Instant = off
		v := e.d.View()
		if v.Instant == nil || v.Instant.Enabled || v.Instant.Equip || v.Instant.Why == "" {
			t.Errorf("%v: %+v", class, v.Instant)
		}
		e.fl.Instant = on
		e.fl.capSet(context.Background()) // the view reads the cached list only
		if v := e.d.View(); !v.Instant.Enabled || !v.Instant.Equip || v.Instant.Why != "" {
			t.Errorf("%v: %+v", class, v.Instant)
		}
	}
}

// A verb that was sent and not answered in time may have been applied: gold and drop are never repeated through the queue
// (a second deposit would move the gold twice); the idempotent kinds go back to the queue. A read-back that runs out of
// time after an ok reply is final for every kind.
func TestFastLaneLostAnswerIsNeverRepeated(t *testing.T) {
	for _, class := range classes {
		for _, c := range []struct {
			name string
			do   func(e *fastEnv) (Request, error)
			verb string
		}{
			{"gold", func(e *fastEnv) (Request, error) { return e.d.RequestGold(true, 100) }, "gold"},
			{"trash", func(e *fastEnv) (Request, error) { return e.d.RequestTrash(e.good, true) }, "drop"},
		} {
			e := fastRig(t, class, town1)
			e.fv.err[c.verb] = context.DeadlineExceeded
			_, err := c.do(e)
			v := e.d.View()
			if code(err) != "timeout" || len(v.Queue) != 0 || len(v.Recent) != 1 || v.Recent[0].Code != "timeout" || e.d.Pending() != 0 {
				t.Errorf("%v %s: %v queue %v recent %v", class, c.name, err, v.Queue, v.Recent)
			}
			if n := len(e.fv.log()); n != 1 {
				t.Errorf("%v %s: %d verbs", class, c.name, n)
			}
		}
		// an idempotent kind is safe to repeat: back to the queue
		e := fastRig(t, class, town1)
		e.fv.err["equip"] = context.DeadlineExceeded
		if r, err := e.d.RequestEquip(e.good, "rhand"); err != nil || r.State != StateQueued || r.How != HowQueue {
			t.Errorf("%v: equip %+v %v", class, r, err)
		}
		// the channel was never acquired: nothing was sent, the request queues for any kind
		e = fastRig(t, class, town1)
		e.fl.Wait = 20 * time.Millisecond
		ch := NewChannel()
		e.fl.Verbs = Serial{Ch: ch, V: e.fv}
		e.fl.capSet(context.Background())
		_ = ch.Acquire(context.Background())
		if r, err := e.d.RequestGold(true, 100); err != nil || r.State != StateQueued || len(e.fv.log()) != 0 {
			t.Errorf("%v: busy channel %+v %v", class, r, err)
		}
		// a read-back that times out after an ok reply is final, whatever the kind
		e = fastRig(t, class, town1)
		e.fl.Fresh = func() (*d2.Snapshot, error) { return e.snap, nil }
		e.fl.Pause = func(context.Context, time.Duration) error { return context.DeadlineExceeded }
		if _, err := e.d.RequestEquip(e.good, "rhand"); code(err) != "timeout" || e.d.Pending() != 0 {
			t.Errorf("%v: read-back timeout %v", class, err)
		}
	}
}

// A panic in the verb path must not leave the request "doing" (that would stop the whole queue).
func TestFastLanePanicLeavesNothingDoing(t *testing.T) {
	for _, class := range classes {
		e := fastRig(t, class, town1)
		e.fv.boom = "equip"
		func() {
			defer func() { _ = recover() }()
			_, _ = e.d.RequestEquip(e.good, "rhand")
		}()
		e.fv.boom = ""
		v := e.d.View()
		if len(v.Queue) != 0 || len(v.Recent) != 1 || v.Recent[0].State != StateFailed || v.Recent[0].Code != "internal" {
			t.Fatalf("%v: queue %v recent %v", class, v.Queue, v.Recent)
		}
		// the queue still works: the next request is done by the lane, and the bot's own path is not blocked
		if r, err := e.d.RequestEquip(e.good, "rhand"); err != nil || r.State != StateDone {
			t.Errorf("%v: next request %+v %v", class, r, err)
		}
	}
}
