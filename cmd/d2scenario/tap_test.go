package main

import (
	"context"
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

	b.set(`{"offers":[{"quest":"search_for_cain","chunk":"scroll","startable":true}]}`)
	if tap, _, err := tp.Step(ctx); err != nil || tap != "" {
		t.Fatalf("scroll is not on the list: %q %v", tap, err)
	}
	b.set(`{"offers":[{"quest":"search_for_cain","chunk":"read","startable":false},{"quest":"den_of_evil","chunk":"read","startable":true}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "" {
		t.Fatalf("not startable or another quest: %q", tap)
	}
	b.set(`{"offers":[{"quest":"search_for_cain","chunk":"read","startable":true}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "start read" {
		t.Fatalf("tap %q", tap)
	}
	b.set(`{"running":{"quest":"search_for_cain","chunk":"read","state":"running"},"offers":[{"quest":"search_for_cain","chunk":"read","startable":true}]}`)
	if tap, _, _ := tp.Step(ctx); tap != "" {
		t.Fatal("a running chunk is not tapped again")
	}
	b.set(`{"awaiting":{"quest":"search_for_cain","chunk":"read","state":"awaiting"}}`)
	if tap, _, _ := tp.Step(ctx); tap != "handin read" {
		t.Fatalf("tap %q", tap)
	}
	b.set(`{"awaiting":{"quest":"search_for_cain","chunk":"read"},"handIn":{"state":"confirming"}}`)
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
	b := &board{view: `{"offers":[{"quest":"q","chunk":"a","startable":true}]}`}
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
