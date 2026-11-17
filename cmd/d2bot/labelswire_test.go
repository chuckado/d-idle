package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"d2lod-bot/internal/d2dctl"
)

// fakeGame holds the flag like the fork does: in the process, not per hero.
type fakeGame struct {
	on   bool
	err  error
	sent []string
}

func (f *fakeGame) Labels(_ context.Context, mode string) (bool, error) {
	f.sent = append(f.sent, mode)
	if f.err != nil {
		return false, f.err
	}
	switch mode {
	case "on":
		f.on = true
	case "off":
		f.on = false
	}
	return f.on, nil
}

func TestLabelsDeskPerHeroAndReapply(t *testing.T) {
	dir := t.TempDir()
	hero := "Ana"
	g := &fakeGame{}
	d := newLabelsDesk(g, dir, func() string { return hero })
	if st := d.Get(); !st.OK || st.On || !st.Available {
		t.Fatalf("default off: %+v", st)
	}
	if st := d.Set(true); !st.OK || !st.On {
		t.Fatalf("set on: %+v", st)
	}
	// A restarted game forgets it: Apply (entering a game) puts it back, once.
	g.on, g.sent = false, nil
	_ = d.Apply(context.Background())
	if !g.on || len(g.sent) != 2 || g.sent[1] != "on" {
		t.Fatalf("reapply: on=%v sent=%v", g.on, g.sent)
	}
	g.sent = nil
	_ = d.Apply(context.Background())
	if len(g.sent) != 1 {
		t.Fatalf("idempotent: a matching game is only read, sent %v", g.sent)
	}
	// Another hero has no wish: the game is put off. The wish of the first hero survives a new desk (a restart of the bot).
	hero = "Bo"
	if st := d.Get(); !st.OK || st.On || g.on {
		t.Fatalf("other hero: %+v on=%v", st, g.on)
	}
	hero = "Ana"
	d2 := newLabelsDesk(g, dir, func() string { return hero })
	if st := d2.Get(); !st.On {
		t.Fatalf("persisted wish: %+v", st)
	}
	if _, err := os.Stat(filepath.Join(dir, "labels.json")); err != nil {
		t.Fatal(err)
	}
}

func TestLabelsDeskFailures(t *testing.T) {
	g := &fakeGame{err: &d2dctl.CmdError{Msg: "lab verbs off: labels needs d2d --item-verbs"}}
	d := newLabelsDesk(g, "", func() string { return "Ana" })
	if st := d.Set(true); st.OK || st.Available || st.Code != "unsupported" {
		t.Fatalf("stock d2d: %+v", st)
	}
	g.err = context.DeadlineExceeded
	if st := d.Get(); st.OK || st.Code != "game" || !st.Available {
		t.Fatalf("unreachable: %+v", st)
	}
	if d.pref["Ana"] {
		t.Fatal("a refused set must not be remembered")
	}
}
