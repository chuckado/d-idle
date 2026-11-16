package main

import (
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

type fakeResetRunner struct {
	soft         bool
	state        string
	softN, hardN int
	lastForce    bool
}

func (f *fakeResetRunner) ResetView() bot.ResetView {
	return bot.ResetView{Schema: bot.ResetSchema, State: f.state}
}
func (f *fakeResetRunner) CanSoftReset() bool { return f.soft }
func (f *fakeResetRunner) PlayerReset(force bool) bot.ResetView {
	f.softN, f.lastForce, f.state = f.softN+1, force, bot.ResetResetting
	return f.ResetView()
}
func (f *fakeResetRunner) BeginHardReset() bot.ResetView {
	f.hardN, f.state = f.hardN+1, bot.ResetResetting
	return bot.ResetView{Schema: bot.ResetSchema, State: bot.ResetResetting, Hard: true}
}

func TestResetAdapterLevels(t *testing.T) {
	var code atomic.Int32
	exit := func(c int) { code.Store(int32(c)) }
	mk := func(soft bool) (*resetAdapter, *fakeResetRunner) {
		f := &fakeResetRunner{soft: soft, state: bot.ResetIdle}
		return &resetAdapter{r: f, exit: exit}, f
	}
	// soft: no process exit
	a, f := mk(true)
	if _, no := a.Set(false, true); no != nil || f.softN != 1 || !f.lastForce || f.hardN != 0 {
		t.Fatalf("soft: %+v %v", f, no)
	}
	// a second press: the running view, nothing started
	v, no := a.Set(true, false)
	if no != nil || f.softN != 1 || f.hardN != 0 || !v.(bot.ResetView).Already {
		t.Fatalf("double press: %+v %v %+v", f, no, v)
	}
	// hard: the process exits like the watchdog (exitStorm) after the reply
	a, f = mk(true)
	if _, no := a.Set(true, false); no != nil || f.hardN != 1 || f.softN != 0 {
		t.Fatalf("hard: %+v %v", f, no)
	}
	deadline := time.Now().Add(3 * time.Second)
	for code.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if code.Load() != exitStorm {
		t.Fatalf("exit code %d, want exitStorm %d", code.Load(), exitStorm)
	}
	// a backend with no soft reset falls back to the hard path
	code.Store(0)
	a, f = mk(false)
	if _, no := a.Set(false, false); no != nil || f.hardN != 1 || f.softN != 0 {
		t.Fatalf("fallback: %+v %v", f, no)
	}
}

func TestResetAdapterRefusals(t *testing.T) {
	f := &fakeResetRunner{soft: true, state: bot.ResetIdle}
	a := &resetAdapter{r: f, exit: func(int) { t.Error("exit") }, measured: true}
	if _, no := a.Set(false, false); no == nil || no.Code != "measured_run" || f.softN+f.hardN != 0 {
		t.Fatalf("measured: %+v %v", f, no)
	}
	a = &resetAdapter{r: f, exit: func(int) { t.Error("exit") }, hardOff: true}
	if _, no := a.Set(true, false); no == nil || no.Code != "hard_off" || f.hardN != 0 {
		t.Fatalf("hard off: %+v %v", f, no)
	}
	if _, no := a.Set(false, false); no != nil || f.softN != 1 {
		t.Fatalf("soft with hard off: %+v %v", f, no)
	}
}

func TestStuckNeedAndBand(t *testing.T) {
	if _, ok := stuckNeed(bot.ResetIdle, time.Minute, time.Second); ok {
		t.Error("a moving bot gets no row")
	}
	n, ok := stuckNeed(bot.ResetIdle, 6*time.Minute, time.Second)
	if !ok || n.Kind != "stuck" || len(n.Actions) != 1 || n.Actions[0].Label != "Reset" || n.Actions[0].Post != watch.ResetRoute {
		t.Fatalf("stalled: %+v", n)
	}
	if n, ok := stuckNeed(bot.ResetIdle, 0, 3*time.Minute); !ok || n.Text != "The bot has not moved for 3 min" {
		t.Fatalf("hung loop: %+v", n)
	}
	if _, ok := stuckNeed(bot.ResetResetting, 9*time.Minute, 0); ok {
		t.Error("no chip while a reset runs")
	}
	var src watch.GlanceSources
	deps := glanceDeps{Reset: func() bot.ResetView {
		return bot.ResetView{State: bot.ResetRejoining, Reason: "leaving the game and rejoining"}
	}}
	for _, fn := range glanceExtra {
		fn(&src, deps)
	}
	if on, text := src.Resetting(); !on || len(text) > 60 || text == "" {
		t.Errorf("band: %v %q", on, text)
	}
}
