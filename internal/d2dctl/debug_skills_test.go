package d2dctl_test

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
)

// echoServer answers every line with "ok <line>" and records them.
func echoServer(t *testing.T, refuse bool) (path string, lines func() []string) {
	t.Helper()
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	var mu sync.Mutex
	var got []string
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				sc := bufio.NewScanner(c)
				for sc.Scan() {
					mu.Lock()
					got = append(got, sc.Text())
					mu.Unlock()
					if refuse {
						c.Write([]byte("err refused: not an offline --no-save game\n"))
					} else {
						c.Write([]byte("ok " + sc.Text() + "\n"))
					}
				}
			}()
		}
	}()
	return path, func() []string { mu.Lock(); defer mu.Unlock(); return append([]string(nil), got...) }
}

func newSD(t *testing.T, refuse bool) (*d2dctl.SkillDebug, func() []string, *cheat.Ledger) {
	path, lines := echoServer(t, refuse)
	c := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	led := cheat.NewLedger(filepath.Join(t.TempDir(), "cheats.jsonl"))
	return &d2dctl.SkillDebug{C: c, Ledger: led, Actor: "drills:M1", Instance: "i2", Character: "Combi"}, lines, led
}

func ip(v int) *int { return &v }

func TestSkillDebugLinesAndLedger(t *testing.T) {
	sd, lines, led := newSD(t, false)
	ctx := context.Background()
	calls := []func() (string, error){
		func() (string, error) { return sd.ManaFull(ctx, "M1 refill") },
		func() (string, error) { return sd.StatesSet(ctx, 40, 0, "M1 armor") },
		func() (string, error) { return sd.StatesSet(ctx, 40, 5, "M1 armor") },
		func() (string, error) { return sd.StatesClear(ctx, 40, "M1 armor off") },
		func() (string, error) { return sd.StatesClear(ctx, -1, "M1 all off") },
		func() (string, error) {
			return sd.MonSet(ctx, 17, d2dctl.MonEdit{Life: ip(6), Dmg: &[2]int{50, 60}, Res: map[string]int{"light": 0, "fire": -20}}, "M4 roll")
		},
		func() (string, error) { return sd.CastAt(ctx, 36, 12.5, 30, "M1 cast") },
		func() (string, error) { return sd.CastAtUnit(ctx, 36, 17, "M1 cast") },
		func() (string, error) { return sd.CastNear(ctx, 36, "M1 cast") },
	}
	for i, f := range calls {
		if _, err := f(); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	want := []string{"debug mana full", "debug states set 40", "debug states set 40 5", "debug states clear 40", "debug states clear all",
		"debug mon set 17 life 6 dmg 50 60 res fire -20 res light 0", "debug cast 36 12.5 30", "debug cast 36 unit 17", "debug cast 36 near"}
	got := lines()
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("lines\n got %q\nwant %q", got, want)
	}
	es, err := led.List(cheat.Filter{Kind: d2dctl.KindD2DDebug})
	if err != nil || len(es) != 2*len(want) {
		t.Fatalf("ledger %d entries, err %v", len(es), err)
	}
	if es[0].Success || !es[1].Success || es[0].Reason != "M1 refill" || es[0].Actor != "drills:M1" || es[0].Params["verb"] != "debug mana full" {
		t.Fatalf("entries %+v %+v", es[0], es[1])
	}
	if es[1].After["reply"] != "ok debug mana full" && !strings.Contains(es[1].After["reply"].(string), "mana full") {
		t.Fatalf("reply %v", es[1].After)
	}
	if es[12].Params["test_action"] != true || es[0].Params["test_action"] != nil {
		t.Fatalf("test_action: cast %v, mana %v", es[12].Params, es[0].Params)
	}
}

func TestSkillDebugRefusals(t *testing.T) {
	sd, lines, led := newSD(t, false)
	ctx := context.Background()
	bad := []func() (string, error){
		func() (string, error) { return sd.ManaFull(ctx, "  ") },
		func() (string, error) { return sd.StatesSet(ctx, 1001, 0, "r") },
		func() (string, error) { return sd.StatesSet(ctx, 40, 100, "r") },
		func() (string, error) { return sd.StatesClear(ctx, -2, "r") },
		func() (string, error) { return sd.MonSet(ctx, 1, d2dctl.MonEdit{}, "r") },
		func() (string, error) { return sd.MonSet(ctx, 1, d2dctl.MonEdit{Life: ip(0)}, "r") },
		func() (string, error) { return sd.MonSet(ctx, 1, d2dctl.MonEdit{LifeFrac: ip(256)}, "r") },
		func() (string, error) { return sd.MonSet(ctx, 1, d2dctl.MonEdit{Dmg: &[2]int{9, 3}}, "r") },
		func() (string, error) {
			return sd.MonSet(ctx, 1, d2dctl.MonEdit{Res: map[string]int{"fire": 101}}, "r")
		},
		func() (string, error) { return sd.MonSet(ctx, 1, d2dctl.MonEdit{Res: map[string]int{"water": 1}}, "r") },
		func() (string, error) { return sd.CastAt(ctx, 36, 5000, 0, "r") },
		func() (string, error) { return sd.CastAtUnit(ctx, 36, -1, "r") },
		func() (string, error) { return sd.CastNear(ctx, 2000, "r") },
	}
	for i, f := range bad {
		if _, err := f(); err == nil {
			t.Fatalf("call %d should be refused", i)
		}
	}
	if got := lines(); len(got) != 0 {
		t.Fatalf("nothing may be sent, sent %q", got)
	}
	if es, _ := led.List(cheat.Filter{}); len(es) != 0 {
		t.Fatalf("refused calls write no entry, got %d", len(es))
	}
	var none d2dctl.SkillDebug
	if _, err := none.ManaFull(ctx, "r"); err == nil {
		t.Fatal("a zero SkillDebug must refuse")
	}
}

func TestSkillDebugLedgerFirst(t *testing.T) {
	sd, lines, _ := newSD(t, false)
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	sd.Ledger = cheat.NewLedger(filepath.Join(blocker, "sub", "cheats.jsonl")) // a path that cannot be created
	if _, err := sd.ManaFull(context.Background(), "r"); err == nil {
		t.Fatal("an unwritable ledger must stop the verb")
	}
	if got := lines(); len(got) != 0 {
		t.Fatalf("sent %q with no ledger", got)
	}
}

func TestSkillDebugServerRefusalIsLedgered(t *testing.T) {
	sd, _, led := newSD(t, true)
	if _, err := sd.CastNear(context.Background(), 36, "M1"); err == nil {
		t.Fatal("the server's refusal must be an error")
	}
	es, _ := led.List(cheat.Filter{})
	if len(es) != 2 || es[1].Success || !strings.Contains(es[1].Error, "refused") {
		t.Fatalf("entries %+v", es)
	}
}
