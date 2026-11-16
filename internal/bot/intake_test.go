package bot

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// The intake sums the drops between polls, ignores gains (potions, regen), and never counts a max-life change as damage
// (level-up, gear swap). A series is the same for every class: the code reads no class (two builds below).
func TestIntakeSeries(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		g := newRig(t)
		o := &g.r.ob
		p := &d2.Player{Class: class, Life: 100, MaxLife: 100}
		g.r.obMu.Lock()
		for _, st := range []struct{ life, max int }{
			{100, 100}, {80, 100}, // hit: 20
			{100, 100},           // potion heal: ignored
			{90, 100}, {85, 100}, // two hits: 10 + 5
			{120, 130}, // level-up: max rises, life rises
			{60, 130},  // hit: 60
			{50, 110},  // max life drops (unequip) and life with it: rebased, not damage
			{45, 110},  // hit: 5
		} {
			p.Life, p.MaxLife = st.life, st.max
			o.noteHP(p, 0)
		}
		lost, hits, polls := o.takeIntake()
		g.r.obMu.Unlock()
		if lost != 100 || hits != 5 || polls != 9 {
			t.Errorf("class %v: lost %d hits %d polls %d, want 100 5 9", class, lost, hits, polls)
		}
		if l, h, _ := func() (int, int, int) { g.r.obMu.Lock(); defer g.r.obMu.Unlock(); return o.takeIntake() }(); l != 0 || h != 0 {
			t.Errorf("minute did not reset: %d %d", l, h)
		}
	}
}

func TestIntakeWindowAndLine(t *testing.T) {
	g := newRig(t)
	buf := &lockedBuf{}
	g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
	if _, ok := g.r.Intake(); ok {
		t.Fatal("intake before the first minute line")
	}
	p := &d2.Player{Life: 100, MaxLife: 100}
	s := beltSnap(0, 2)
	for i := 0; i < intakeWindow+2; i++ { // the window keeps the last intakeWindow minutes
		g.r.obMu.Lock()
		p.Life = 100
		g.r.ob.noteHP(p, 0)
		p.Life = 100 - 10*(i+1)
		g.r.ob.noteHP(p, 0)
		g.r.minuteLine(s, s.Player, g.r.e.Clock.Now().Add(time.Duration(i)*time.Minute))
		g.r.obMu.Unlock()
	}
	in, ok := g.r.Intake()
	if !ok || in.WindowMinutes != intakeWindow || in.LifeLostPerMin != 50 { // minutes 3..7: 30..70
		t.Errorf("intake %+v ok %v", in, ok)
	}
	if !strings.Contains(buf.String(), " lifeLost=70 hitsTaken=1 hpPolls=2") {
		t.Errorf("minute line lacks the intake fields: %s", buf.String())
	}
}
