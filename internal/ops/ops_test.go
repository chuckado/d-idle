package ops

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/questboard"
)

// fakeDesk records the calls and answers from canned data. The quest board is class-agnostic (it reads the hero's
// class only through the board data), so no second class is needed here.
type fakeDesk struct {
	mu    sync.Mutex
	calls []string
	view  questboard.View
	rep   questboard.Reply
	rec   []questboard.AuditEvent
}

func (f *fakeDesk) call(s string) questboard.Reply {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, s)
	r := f.rep
	if r.Status == 0 {
		r.Status = 200
	}
	return r
}
func (f *fakeDesk) View() questboard.View               { f.mu.Lock(); defer f.mu.Unlock(); return f.view }
func (f *fakeDesk) Start(q, c string) questboard.Reply  { return f.call("start " + q + " " + c) }
func (f *fakeDesk) HandIn(q, c string) questboard.Reply { return f.call("handin " + q + " " + c) }
func (f *fakeDesk) Cancel(q, c string) questboard.Reply { return f.call("cancel " + q + " " + c) }
func (f *fakeDesk) Retry(q, c string) questboard.Reply  { return f.call("retry " + q + " " + c) }
func (f *fakeDesk) SetAuto(q string, on bool) questboard.Reply {
	if on {
		return f.call("auto " + q + " on")
	}
	return f.call("auto " + q + " off")
}
func (f *fakeDesk) Record(e questboard.AuditEvent) {
	f.mu.Lock()
	f.rec = append(f.rec, e)
	f.mu.Unlock()
}

func newSrv(d *fakeDesk, env map[string]string) (*Server, *time.Time) {
	now := time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)
	s := New(Config{Desk: func() Desk { return d }, Getenv: func(k string) string { return env[k] }, Now: func() time.Time { return now }})
	return s, &now
}

func TestEnabled(t *testing.T) {
	cases := []struct {
		env map[string]string
		on  bool
	}{
		{nil, false},
		{map[string]string{"D2BOT_OPERATOR": "0"}, false},
		{map[string]string{"D2BOT_OPERATOR": "1"}, true},
		{map[string]string{"D2BOT_OPERATOR": "1", "D2BOT_RUN": "g2"}, false},
	}
	for i, c := range cases {
		on, why := Enabled(func(k string) string { return c.env[k] })
		if on != c.on || (!on && why == "") {
			t.Errorf("case %d: on=%v why=%q", i, on, why)
		}
	}
}

func TestOffByDefaultNoSocket(t *testing.T) {
	dir := t.TempDir()
	stop, err := Start(dir, func() Desk { return &fakeDesk{} }, func(string) string { return "" }, nil)
	if err != nil || stop == nil {
		t.Fatalf("Start: %v", err)
	}
	stop()
	if _, err := os.Stat(filepath.Join(dir, SockName)); !os.IsNotExist(err) {
		t.Fatalf("socket exists without the env: %v", err)
	}
	// and refused during a measured run even with the env
	env := map[string]string{"D2BOT_OPERATOR": "1", "D2BOT_RUN": "x"}
	stop, _ = Start(dir, func() Desk { return &fakeDesk{} }, func(k string) string { return env[k] }, nil)
	stop()
	if _, err := os.Stat(filepath.Join(dir, SockName)); !os.IsNotExist(err) {
		t.Fatalf("socket exists in a measured run: %v", err)
	}
}

func TestSocketMode0600AndRoundTrip(t *testing.T) {
	dir := t.TempDir()
	d := &fakeDesk{rep: questboard.Reply{OK: true, Reason: "started"}}
	env := map[string]string{"D2BOT_OPERATOR": "1"}
	stop, err := Start(dir, func() Desk { return d }, func(k string) string { return env[k] }, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	p := filepath.Join(dir, SockName)
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 || fi.Mode()&os.ModeSocket == 0 {
		t.Fatalf("mode %v", fi.Mode())
	}
	resp, err := Call(p, Request{Verb: "start", Quest: "den_of_evil", Why: "overnight test"})
	if err != nil || !resp.OK {
		t.Fatalf("%+v %v", resp, err)
	}
	if len(d.calls) != 1 || d.calls[0] != "start den_of_evil " {
		t.Fatalf("calls %v", d.calls)
	}
	if _, err := net.Dial("unix", p); err != nil {
		t.Fatal(err)
	}
}

func TestAllowlistRefusesEverythingElse(t *testing.T) {
	d := &fakeDesk{}
	s, _ := newSrv(d, nil)
	for _, v := range []string{"sell", "vault", "hero", "character", "tuning", "respec", "order", "mode", "dismiss", "wait", "", "START", "start;sell", "../start"} {
		r := s.Handle(Request{Verb: v, Quest: "a", Why: "x"})
		if r.OK || r.Code != "verb" {
			t.Errorf("verb %q: %+v", v, r)
		}
	}
	if len(d.calls) != 0 || len(d.rec) != 0 {
		t.Fatalf("desk touched: %v %v", d.calls, d.rec)
	}
}

func TestWhyRequiredAndBounded(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: true}}
	s, _ := newSrv(d, nil)
	for _, why := range []string{"", "   ", strings.Repeat("x", 201), "a\nb"} {
		if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: why}); r.OK || r.Code != "why" {
			t.Errorf("why %q: %+v", why, r)
		}
	}
	if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: strings.Repeat("x", 200)}); !r.OK {
		t.Errorf("200 chars refused: %+v", r)
	}
	// status needs no reason
	if r := s.Handle(Request{Verb: "status"}); !r.OK {
		t.Errorf("status: %+v", r)
	}
}

func TestBadIDs(t *testing.T) {
	d := &fakeDesk{}
	s, _ := newSrv(d, nil)
	for _, q := range []string{"", "A b", "a/b", strings.Repeat("a", 65)} {
		if r := s.Handle(Request{Verb: "start", Quest: q, Why: "x"}); r.OK || r.Code != "quest" {
			t.Errorf("quest %q: %+v", q, r)
		}
	}
	if r := s.Handle(Request{Verb: "start", Quest: "a", Chunk: "x y", Why: "x"}); r.OK || r.Code != "chunk" {
		t.Errorf("chunk: %+v", r)
	}
	if r := s.Handle(Request{Verb: "auto", Quest: "a", Why: "x"}); r.OK || r.Code != "body" {
		t.Errorf("auto without on: %+v", r)
	}
}

func TestMeasuredRunRefusesMutatingAtRequestTime(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: true}}
	env := map[string]string{}
	s, _ := newSrv(d, env)
	env["D2BOT_RUN"] = "gate"
	r := s.Handle(Request{Verb: "start", Quest: "a", Why: "x"})
	if r.OK || r.Code != "measured_run" || len(d.calls) != 0 {
		t.Fatalf("%+v %v", r, d.calls)
	}
	if r := s.Handle(Request{Verb: "status"}); !r.OK {
		t.Fatalf("status must stay readable: %+v", r)
	}
}

func TestRateLimitPerQuestHour(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: true}}
	s, now := newSrv(d, nil)
	for i := 0; i < 6; i++ {
		*now = now.Add(2 * time.Minute) // past the retry cooldown
		if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: "x"}); !r.OK {
			t.Fatalf("retry %d: %+v", i, r)
		}
	}
	*now = now.Add(2 * time.Minute)
	r := s.Handle(Request{Verb: "retry", Quest: "a", Why: "x"})
	if r.OK || r.Code != "rate_limited" || r.Reason == "" {
		t.Fatalf("7th: %+v", r)
	}
	if len(d.calls) != 6 {
		t.Fatalf("calls %d", len(d.calls))
	}
	// another quest is unaffected; after an hour the first is free again
	if r := s.Handle(Request{Verb: "retry", Quest: "b", Why: "x"}); !r.OK {
		t.Fatalf("other quest: %+v", r)
	}
	*now = now.Add(time.Hour)
	if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: "x"}); !r.OK {
		t.Fatalf("after an hour: %+v", r)
	}
}

func TestRateLimitTotalHour(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: true}}
	s, now := newSrv(d, nil)
	for i := 0; i < 30; i++ {
		*now = now.Add(time.Second)
		q := string(rune('a'+i/5)) + "q"
		if r := s.Handle(Request{Verb: "cancel", Quest: q, Why: "x"}); !r.OK {
			t.Fatalf("%d: %+v", i, r)
		}
	}
	if r := s.Handle(Request{Verb: "cancel", Quest: "zq", Why: "x"}); r.OK || r.Code != "rate_limited" {
		t.Fatalf("31st: %+v", r)
	}
}

func TestRetryCooldown(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: true}}
	s, now := newSrv(d, nil)
	if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: "x"}); !r.OK {
		t.Fatal(r)
	}
	*now = now.Add(30 * time.Second)
	for _, v := range []string{"retry", "start"} {
		if r := s.Handle(Request{Verb: v, Quest: "a", Why: "x"}); r.OK || r.Code != "cooldown" {
			t.Errorf("%s in cooldown: %+v", v, r)
		}
	}
	if r := s.Handle(Request{Verb: "cancel", Quest: "a", Why: "x"}); !r.OK {
		t.Errorf("cancel must pass: %+v", r)
	}
	*now = now.Add(31 * time.Second)
	if r := s.Handle(Request{Verb: "retry", Quest: "a", Why: "x"}); !r.OK {
		t.Errorf("after cooldown: %+v", r)
	}
}

func TestAuditLine(t *testing.T) {
	d := &fakeDesk{rep: questboard.Reply{OK: false, Code: "busy", Reason: "another chunk runs", Status: 409}}
	s, _ := newSrv(d, nil)
	r := s.Handle(Request{Verb: "start", Quest: "den_of_evil", Chunk: "c1", Why: "overnight retry 2"})
	if r.OK || r.Code != "busy" {
		t.Fatalf("the desk refusal must pass through: %+v", r)
	}
	if len(d.rec) != 1 {
		t.Fatalf("audit lines: %d", len(d.rec))
	}
	e := d.rec[0]
	if e.Remote != "operator" || e.Agent != "ops-cli" || e.Why != "overnight retry 2" || e.Action != "start" ||
		e.Quest != "den_of_evil" || e.Chunk != "c1" || e.Code != "busy" || e.Status != 409 || e.OK {
		t.Fatalf("%+v", e)
	}
	// a refusal before the desk is not an audit line of the desk (it is logged); no ring entry
	s.Handle(Request{Verb: "sell", Quest: "a", Why: "x"})
	if len(d.rec) != 1 {
		t.Fatal("allowlist refusal wrote to the ring")
	}
}

func TestStatusRows(t *testing.T) {
	d := &fakeDesk{view: questboard.View{Enabled: true, Difficulty: "normal",
		Quests: []questboard.QuestRow{{ID: "den", Chunks: []questboard.ChunkRow{
			{ID: "c1", State: questboard.Done},
			{ID: "c2", State: questboard.Blocked, Reason: &questboard.ReasonView{Code: "gave_up", Text: "could not find it"}},
			{ID: "c3", State: questboard.InProgress},
		}}},
		Running: &questboard.RunView{Quest: "den", Chunk: "c3", Step: "walk_to_cave"},
	}}
	s, _ := newSrv(d, nil)
	r := s.Handle(Request{Verb: "status"})
	if !r.OK || r.Board == nil || len(r.Board.Rows) != 3 {
		t.Fatalf("%+v", r)
	}
	rows := r.Board.Rows
	if rows[1].Code != "gave_up" || rows[1].Text != "could not find it" || rows[2].Step != "walk_to_cave" || rows[2].State != "in_progress" {
		t.Fatalf("%+v", rows)
	}
}

func TestWait(t *testing.T) {
	board := func(states ...string) func() (*Board, error) {
		i := 0
		return func() (*Board, error) {
			st := states[min(i, len(states)-1)]
			i++
			row := Row{Quest: "den", Chunk: "c1", State: st, Step: "s" + st}
			if st == "blocked" {
				row.Code, row.Text = "gave_up", "stuck"
			}
			return &Board{Rows: []Row{row}}, nil
		}
	}
	nosleep := func(time.Duration) {}
	clock := func() func() time.Time {
		t0 := time.Unix(0, 0)
		return func() time.Time { t0 = t0.Add(time.Second); return t0 }
	}
	cases := []struct {
		name    string
		states  []string
		timeout time.Duration
		want    string
	}{
		{"finished", []string{"accepted", "in_progress", "done"}, time.Hour, OutFinished},
		{"parked", []string{"in_progress", "blocked"}, time.Hour, OutBlocked},
		{"stopped", []string{"in_progress", "ready"}, time.Hour, OutStopped},
		{"timeout", []string{"in_progress"}, 5 * time.Second, OutTimeout},
		{"idle", []string{"ready"}, time.Hour, OutIdle},
	}
	for _, c := range cases {
		res := Wait(board(c.states...), "den", "", c.timeout, time.Second, nosleep, clock())
		if res.Outcome != c.want {
			t.Errorf("%s: %+v", c.name, res)
		}
		if c.want == OutBlocked && (res.Code != "gave_up" || res.Step == "") {
			t.Errorf("blocked evidence: %+v", res)
		}
	}
}

func TestProgressCounter(t *testing.T) {
	d := &fakeDesk{view: questboard.View{Enabled: true, Quests: []questboard.QuestRow{{ID: "q", Chunks: []questboard.ChunkRow{{ID: "c", State: questboard.InProgress}}}},
		Running: &questboard.RunView{Quest: "q", Chunk: "c", Step: "a"}}}
	s, _ := newSrv(d, nil)
	get := func() int { return s.Handle(Request{Verb: "status"}).Board.Rows[0].Progress }
	if get() != 0 || get() != 0 {
		t.Fatal("unchanged board must not count")
	}
	d.mu.Lock()
	d.view.Running = &questboard.RunView{Quest: "q", Chunk: "c", Step: "b"}
	d.mu.Unlock()
	if get() != 1 {
		t.Fatal("step change must count")
	}
}
