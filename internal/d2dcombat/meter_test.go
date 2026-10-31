package d2dcombat

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
	"d2lod-bot/internal/d2dinput"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/nav"
)

func TestMeterSamples(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
	defer c.Close()
	m := &Meter{P: &d2dinput.Pacer{C: c}, C: c, Ctx: context.Background()}
	m.Sleep(80 * time.Millisecond)
	if tot := m.Totals(); tot.Samples != 1 || tot.MinLife != 45 {
		t.Fatalf("%+v", tot)
	}
}

func mon(id uint32, mode, life int) d2.Unit {
	return d2.Unit{ID: id, Type: d2.TypeMonster, Mode: mode, Monster: &d2.MonsterInfo{Life: life, MaxLife: 100}}
}

type seqClock struct{ now time.Time }

func (c *seqClock) Now() time.Time        { return c.now }
func (c *seqClock) Sleep(d time.Duration) { c.now = c.now.Add(d) }

func fakeMeter(t *testing.T, frames *[]*d2.Snapshot) (*Meter, *seqClock) {
	t.Helper()
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	c := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	clk := &seqClock{}
	i := 0
	m := &Meter{C: c, Ctx: context.Background(), Clk: clk, Mons: func() (*d2.Snapshot, error) {
		f := (*frames)[min(i, len(*frames)-1)]
		i++
		return f, nil
	}}
	return m, clk
}

func TestMeterFakeFight(t *testing.T) {
	alive := d2.MonModeStand
	frames := []*d2.Snapshot{
		{Monsters: []d2.Unit{mon(1, alive, 100), mon(2, alive, 80)}},
		{Monsters: []d2.Unit{mon(1, alive, 60), mon(2, alive, 80)}},          // 40 dealt
		{Monsters: []d2.Unit{mon(1, d2.MonModeDeath, 0), mon(2, alive, 50)}}, // 1 dies (60 left), 30 dealt
		{Monsters: []d2.Unit{mon(2, alive, 50), mon(3, alive, 100)}},         // corpse 1 gone: nothing; 3 appears
		{Monsters: []d2.Unit{mon(2, alive, 70), mon(3, alive, 90)}},          // regen is not damage; 10 dealt
		{Monsters: []d2.Unit{mon(2, d2.MonModeDead, 0)}},                     // 2 dies (70 left), 3 vanished: no kill
	}
	m, clk := fakeMeter(t, &frames)
	for range frames {
		m.Sleep(40 * time.Millisecond)
	}
	tot := m.Totals()
	if !tot.DealtOK {
		t.Fatalf("%+v", tot)
	}
	// 100->60 (40), 1 dies at 60 (60), 80->50 (30), 70 vs 50 up (0), 3: 100->90 (10), 2 dies at 70 (70)
	if tot.DamageDealt != 40+60+30+10+70 || tot.Kills != 2 || tot.Samples != 6 {
		t.Fatalf("%+v", tot)
	}
	if !clk.now.Equal(time.Time{}.Add(240 * time.Millisecond)) {
		t.Fatalf("clock not forwarded: %v", clk.now)
	}
	if s := tot.String(); !strings.Contains(s, "dealt=210 kills=2 hits=n/a misses=n/a") {
		t.Fatal(s)
	}
}

func TestMeterUnavailableIsNA(t *testing.T) {
	m := &Meter{P: &d2dinput.Pacer{}} // no Mons: monster metrics unavailable
	tot := m.Totals()
	if tot.DealtOK {
		t.Fatal("dealt must be unavailable")
	}
	if s := tot.String(); !strings.Contains(s, "dealt=n/a kills=n/a hits=n/a misses=n/a") {
		t.Fatal(s)
	}
}

func TestMeterOffForwardsOnly(t *testing.T) {
	frames := []*d2.Snapshot{{}}
	m, clk := fakeMeter(t, &frames)
	m.Off = true
	m.Sleep(time.Second)
	if m.Totals().Samples != 0 || clk.now.IsZero() {
		t.Fatalf("off meter sampled or did not forward: %+v", m.Totals())
	}
	m.Enable()
	m.Sleep(time.Second)
	if m.Totals().Samples != 1 {
		t.Fatalf("%+v", m.Totals())
	}
}

func TestAttachBuildsOffMeterOnTheBackendClock(t *testing.T) {
	clk := &seqClock{}
	e := &env.Env{ID: "t", Read: stubReader{s: &d2.Snapshot{Player: &d2.Player{}}}, Clock: clk, Log: slog.New(slog.DiscardHandler)}
	ctl := d2dinput.New(nil)
	rig := Attach(context.Background(), e, nav.New(e), ctl)
	if rig.Meter == nil || !rig.Meter.Off || rig.Env.Clock != env.Clock(rig.Meter) {
		t.Fatalf("meter %+v", rig.Meter)
	}
	rig.Env.Clock.Sleep(time.Second) // off: forwards to the backend clock, no state read (C is nil here)
	if !clk.now.Equal(time.Time{}.Add(time.Second)) || e.Clock != env.Clock(clk) {
		t.Fatalf("clock %v, backend env clock replaced", clk.now)
	}
}
