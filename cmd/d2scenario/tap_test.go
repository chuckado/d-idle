package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/scenario"
)

// board is a fake watch server: the view JSON is swapped by the test, the POSTs are recorded.
type board struct {
	mu   sync.Mutex
	view string
	got  []string
}

func (b *board) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		b.mu.Lock()
		defer b.mu.Unlock()
		if r.Method == http.MethodPost {
			b.got = append(b.got, r.URL.Path)
			w.WriteHeader(202)
			return
		}
		io.WriteString(w, b.view)
	})
}

func (b *board) set(v string) { b.mu.Lock(); b.view = v; b.mu.Unlock() }

func TestTapperTapsOnlyWhatIsAskedFor(t *testing.T) {
	b := &board{}
	srv := httptest.NewServer(b.handler(t))
	defer srv.Close()
	var log strings.Builder
	tp := newTapper(srv.URL, srv.URL, "tok", "search_for_cain", []string{"read"}, &log)
	ctx := context.Background()

	b.set(`{"quests":[{"id":"search_for_cain","chunks":[{"id":"scroll","state":"x","action":{"verb":"start"}}]}]}`)
	if tap, _, err := tp.Step(ctx); err != nil || tap != "" {
		t.Fatalf("scroll is not on the list: %q %v", tap, err)
	}
	b.set(`{"quests":[{"id":"search_for_cain","chunks":[{"id":"read","action":null}]},{"id":"den_of_evil","chunks":[{"id":"read","action":{"verb":"start"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "" {
		t.Fatalf("not startable or another quest: %q", tap)
	}
	b.set(`{"quests":[{"id":"search_for_cain","chunks":[{"id":"read","state":"x","action":{"verb":"start"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "start read" {
		t.Fatalf("tap %q", tap)
	}
	b.set(`{"running":{"quest":"search_for_cain","chunk":"read","state":"running"},"quests":[{"id":"search_for_cain","chunks":[{"id":"read","state":"x","action":{"verb":"cancel"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "" {
		t.Fatal("a running chunk is not tapped again")
	}
	b.set(`{"awaiting":{"quest":"search_for_cain","chunk":"read","state":"awaiting"},"quests":[{"id":"search_for_cain","chunks":[{"id":"read","state":"x","action":{"verb":"handin"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "handin read" {
		t.Fatalf("tap %q", tap)
	}
	b.set(`{"awaiting":{"quest":"search_for_cain","chunk":"read"},"handIn":{"state":"confirming"},"quests":[{"id":"search_for_cain","chunks":[{"id":"read","state":"x","action":{"verb":"handin""disabled":true}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "" {
		t.Fatal("hand-in already in progress")
	}
	b.set(`{"blocked":{"quest":"search_for_cain","chunk":"read","why":"x"}}`)
	if _, blocked, _ := tp.Step(ctx); !blocked {
		t.Fatal("a parked quest stops the tapper")
	}
	b.mu.Lock()
	got := strings.Join(b.got, ",")
	b.mu.Unlock()
	if got != "/api/quests/search_for_cain/start,/api/quests/search_for_cain/handin" || strings.Join(tp.Taps, ",") != "start read,handin read" {
		t.Fatalf("posts %s taps %v", got, tp.Taps)
	}
}

func TestTapperRunStopsWithTheContext(t *testing.T) {
	b := &board{view: `{"quests":[{"id":"q","chunks":[{"id":"a","state":"x","action":{"verb":"start"}}]}]}`}
	srv := httptest.NewServer(b.handler(t))
	defer srv.Close()
	tp := newTapper(srv.URL, srv.URL, "tok", "q", nil, io.Discard)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() { tp.Run(ctx, 20*time.Millisecond); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run ignored the context")
	}
	if len(tp.Taps) == 0 {
		t.Fatal("no tap made")
	}
}

func TestShippedBuildFieldStartIsBlockedBeforeRunning(t *testing.T) {
	scs, err := loadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	s, err := findScenario(scs, "cain.stones.shipped")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("D2D_BIN_SANDBOX", "/x")
	why := blockedBefore(s, scenario.Run{ID: s.ID, Class: "Paladin", Seed: 11})
	if !strings.Contains(why, "field start") {
		t.Fatalf("reason %q", why)
	}
}

func TestEveryRunnableScenarioIsBuildable(t *testing.T) {
	scs, err := loadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range scs {
		if s.Pending() || s.Build == "sandbox" {
			continue
		}
		for _, c := range s.Matrix.Classes {
			su, err := scenario.BuildSetup(s, c, s.Matrix.Seeds[0])
			if err != nil || len(su.Unsupported) > 0 {
				t.Errorf("%s/%s: not buildable: %v %v", s.ID, c, err, su.Unsupported)
			}
		}
	}
}

func TestAddTrailReadsTheBotLog(t *testing.T) {
	f := t.TempDir() + "/x.bot.log"
	if err := os.WriteFile(f, []byte(`{"msg":"zone entry","area":1}`+"\n"+`{"msg":"zone entry","area":5}`+"\n"+`{"msg":"zone entry","area":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var st scenario.State
	addTrail(&st, f)
	if st.AreaChanges == nil || *st.AreaChanges != 2 || len(st.AreaTrail) != 3 {
		t.Fatalf("%+v", st)
	}
	var none scenario.State
	addTrail(&none, f+".missing")
	if none.AreaChanges != nil {
		t.Fatal("no log: not observed")
	}
}

// vboard serves /api/thoughts with a virtual clock the test advances and a board that offers the chunk from a given
// virtual second on; it records the virtual second of each POST.
type vboard struct {
	mu      sync.Mutex
	now     float64
	offerAt float64
	posts   []float64
}

func (b *vboard) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.mu.Lock()
		defer b.mu.Unlock()
		switch {
		case r.Method == http.MethodPost:
			b.posts = append(b.posts, b.now)
			w.WriteHeader(202)
		case r.URL.Path == "/api/thoughts":
			fmt.Fprintf(w, `{"now":%v}`, b.now)
		case len(b.posts) > 0:
			io.WriteString(w, `{"running":{"quest":"q","chunk":"a","state":"running"}}`)
		case b.now >= b.offerAt:
			io.WriteString(w, `{"quests":[{"id":"q","chunks":[{"id":"a","state":"x","action":{"verb":"start"}}]}]}`)
		default:
			io.WriteString(w, `{"offers":[]}`)
		}
	})
}

func (b *vboard) advance(to float64) { b.mu.Lock(); b.now = to; b.mu.Unlock() }

// The tap lands on the first grid point of VIRTUAL time at or after the offer, however fast the wall poll runs and
// however far the clock jumped between two polls.
func TestRunVirtualTapsOnTheVirtualGrid(t *testing.T) {
	for _, poll := range []time.Duration{2 * time.Millisecond, 15 * time.Millisecond} {
		b := &vboard{offerAt: 47}
		srv := httptest.NewServer(b.handler())
		tp := newTapper(srv.URL, srv.URL, "tok", "q", nil, io.Discard)
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { tp.RunVirtual(ctx, 30, poll, time.Minute); close(done) }()
		for _, v := range []float64{5, 29, 46, 47, 55, 61, 62, 100} {
			b.advance(v)
			time.Sleep(60 * time.Millisecond)
		}
		cancel()
		<-done
		srv.Close()
		if len(tp.Taps) != 1 || tp.Taps[0] != "start a" {
			t.Fatalf("poll %v: taps %v", poll, tp.Taps)
		}
		if len(tp.TapsAt) != 1 || tp.TapsAt[0] < 60 || tp.TapsAt[0] >= 90 {
			t.Fatalf("poll %v: the tap must land on the grid point 60 (first read at or after it), got %v", poll, tp.TapsAt)
		}
	}
}

func TestRunVirtualFallsBackToTheWallClock(t *testing.T) {
	b := &board{view: `{"quests":[{"id":"q","chunks":[{"id":"a","state":"x","action":{"verb":"start"}}]}]}`}
	srv := httptest.NewServer(b.handler(t)) // its GET /api/thoughts reply has no "now"
	defer srv.Close()
	var log strings.Builder
	tp := newTapper(srv.URL, srv.URL, "tok", "q", nil, &log)
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	tp.RunVirtual(ctx, 30, 5*time.Millisecond, 50*time.Millisecond)
	if !strings.Contains(log.String(), "falling back") {
		t.Fatalf("log %q", log.String())
	}
}

// A quest that is on the board but never startable says why in the taps log, once per change (card andariel1).
func TestTapperLogsWhyAChunkIsNotStartable(t *testing.T) {
	b := &board{}
	srv := httptest.NewServer(b.handler(t))
	defer srv.Close()
	var log strings.Builder
	tp := newTapper(srv.URL, srv.URL, "tok", "sisters_to_the_slaughter", []string{"andariel"}, &log)
	b.set(`{"quests":[{"id":"sisters_to_the_slaughter","state":"available","chunks":[{"id":"andariel","state":"available","readiness":"too_hard","why":"You have no merc: too hard"}]}]}`)
	for i := 0; i < 3; i++ {
		if tap, _, err := tp.Step(context.Background()); err != nil || tap != "" {
			t.Fatalf("tap %q %v", tap, err)
		}
	}
	got := log.String()
	if strings.Count(got, "not startable") != 1 || !strings.Contains(got, `readiness "too_hard"`) || !strings.Contains(got, "no merc") {
		t.Fatalf("log %q", got)
	}
}

// The Kashya case (2026-11-21): the part reads awaiting_handin with no run, so the view has no `awaiting` object; the row's
// button is Start and that is the tap, then Hand in once the run waits for it.
func TestTapperPressesTheRowsButton(t *testing.T) {
	b := &board{}
	srv := httptest.NewServer(b.handler(t))
	defer srv.Close()
	tp := newTapper(srv.URL, srv.URL, "tok", "sisters_burial_grounds", nil, io.Discard)
	ctx := context.Background()
	b.set(`{"quests":[{"id":"sisters_burial_grounds","chunks":[{"id":"blood_raven","state":"awaiting_handin","action":{"verb":"start","label":"Start"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "start blood_raven" {
		t.Fatalf("tap %q", tap)
	}
	b.set(`{"awaiting":{"quest":"sisters_burial_grounds","chunk":"blood_raven"},"quests":[{"id":"sisters_burial_grounds","chunks":[{"id":"blood_raven","state":"awaiting_handin","action":{"verb":"handin","label":"Hand in"}}]}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "handin blood_raven" {
		t.Fatalf("tap %q", tap)
	}
}
