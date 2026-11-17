package survival

import (
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/power"
)

// both builds: the Paladin (bruiser archetype) and the Sorceress (caster archetype), the same rule on each.
var twoBuilds = []struct {
	name, arch string
	hero       func(int) power.Hero
}{{"paladin", "bruiser", paladin}, {"sorceress", "caster", sorceress}}

func activeIn(e env, h power.Hero, arch string, now time.Time) Input {
	in := e.input(h, ZoneKey{}, now)
	in.Active, in.Archetype = true, arch
	return in
}

func TestRealizedYieldOverridesPrior(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			l := NewLedger(e.p)
			h := b.hero(14)
			in := activeIn(e, h, b.arch, t0)
			first, err := l.Choose(in)
			if err != nil {
				t.Fatal(err)
			}
			k := ZoneKey{first.Area, "normal"}
			prior := 0.0
			for _, c := range in.Cands {
				if c.Key == k {
					prior = c.Eco.XPPerMin
				}
			}
			// one bad minute is below the trust floor: the prior stands
			l.Observe(k, t0, Event{Kind: EvXP, XP: 5, Seconds: 60}, in.Power)
			if got, real := l.BlendYield(k, prior, b.arch); got != prior || real != 0 {
				t.Fatalf("one minute must not move the value: %v %v", got, real)
			}
			// an hour at a fiftieth of the prior: the realized value takes over, but never fully
			l.Observe(k, t0.Add(time.Minute), Event{Kind: EvXP, XP: prior / 50 * 60, Seconds: 3600}, in.Power)
			got, real := l.BlendYield(k, prior, b.arch)
			if !(real > 0 && got < 0.5*prior && got > real) {
				t.Fatalf("blend %.1f realized %.1f prior %.1f", got, real, prior)
			}
			again, err := l.Choose(activeIn(e, h, b.arch, t0.Add(62*time.Minute)))
			if err != nil {
				t.Fatal(err)
			}
			if again.Area == first.Area {
				t.Fatalf("a zone that pays a fraction of its prior stays the pick: %s", again.Line())
			}
			if !strings.Contains(again.Line(), "xp/min") {
				t.Fatalf("line: %s", again.Line())
			}
			// the shadow (Active false) never uses it
			sh := activeIn(e, h, b.arch, t0.Add(62*time.Minute))
			sh.Active = false
			if s, _ := l.Choose(sh); s.Area != first.Area {
				t.Fatalf("shadow must keep the model's pick: %s", s.Line())
			}
		})
	}
}

func TestStarvedZoneLosesValue(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			e.p.StarveValueMul = 0.2 // a hard cut: the synthetic zones differ by a factor of 5 in pay
			l := NewLedger(e.p)
			h := b.hero(14)
			first, _ := l.Choose(activeIn(e, h, b.arch, t0))
			k := ZoneKey{first.Area, "normal"}
			limit := e.p.starveLimit(b.arch)
			l.Observe(k, t0, Event{Kind: EvFight, Seconds: 1, Starved: limit - 1}, PowerOf(h))
			if l.Starving(k, b.arch) {
				t.Fatal("below the limit")
			}
			if d, _ := l.Choose(activeIn(e, h, b.arch, t0)); d.Area != first.Area {
				t.Fatalf("below the limit nothing changes: %s", d.Line())
			}
			l.Observe(k, t0, Event{Kind: EvFight, Seconds: 1, Starved: 1.5}, PowerOf(h))
			if !l.Starving(k, b.arch) {
				t.Fatal("at the limit")
			}
			d, _ := l.Choose(activeIn(e, h, b.arch, t0))
			if d.Area == first.Area {
				t.Fatalf("a starving zone stays the pick: %s", d.Line())
			}
			sh := activeIn(e, h, b.arch, t0)
			sh.Active = false
			if s, _ := l.Choose(sh); s.Area != first.Area {
				t.Fatal("shadow ignores the mana floor")
			}
		})
	}
	// the limit is data per archetype
	p := params(t)
	if p.starveLimit("caster") >= p.starveLimit("bruiser") || p.starveLimit("") != p.StarveCasts {
		t.Fatalf("archetype table: caster %v bruiser %v global %v", p.starveLimit("caster"), p.starveLimit("bruiser"), p.StarveCasts)
	}
}

func TestFreshGameNeedsKnownOffenceAndMerc(t *testing.T) {
	for _, b := range twoBuilds {
		t.Run(b.name, func(t *testing.T) {
			e := newEnv(t, b.arch)
			in := activeIn(e, b.hero(14), b.arch, t0)
			in.FreshAllowed = true
			for i := range in.Cands {
				in.Cands[i].Depleted = true
			}
			if d, _ := NewLedger(e.p).Choose(in); !d.Fresh || d.FreshRefused != "" {
				t.Fatalf("known offence, merc alive: want fresh: %s", d.Line())
			}
			in.OffenceUnread = true
			d, _ := NewLedger(e.p).Choose(in)
			if d.Fresh || d.FreshRefused != "offence unread" || !strings.Contains(d.Line(), "fresh game refused: offence unread") {
				t.Fatalf("unread offence: %s", d.Line())
			}
			in.OffenceUnread, in.MercDead = false, true
			if d, _ = NewLedger(e.p).Choose(in); d.FreshRefused != "" {
				t.Fatalf("merc dead is no reason to refuse by default (freshNeedsMerc false): %s", d.Line())
			}
			strict := *e.p
			strict.FreshNeedsMerc = true
			if d, _ = NewLedger(&strict).Choose(in); d.Fresh || d.FreshRefused != "merc dead" {
				t.Fatalf("merc dead: %s", d.Line())
			}
			in.Active = false // the shadow keeps its old behaviour
			if d, _ = NewLedger(e.p).Choose(in); !d.Fresh {
				t.Fatalf("shadow: %s", d.Line())
			}
		})
	}
}

func TestReviveLoopTripsAndResets(t *testing.T) {
	k := ZoneKey{3, "normal"}
	other := ZoneKey{4, "normal"}
	l := &ReviveLoop{N: 3, Within: 10 * time.Minute, Hold: time.Hour}
	if _, known, _, _ := l.Died(k, t0); known {
		t.Fatal("a death with no revive is not a loop datum")
	}
	now := t0
	for i := 1; i <= 3; i++ {
		l.Revived(now)
		now = now.Add(time.Minute)
		life, known, fast, tripped := l.Died(k, now)
		if !known || !fast || life != time.Minute || tripped != (i == 3) {
			t.Fatalf("death %d: %v %v %v %v", i, life, known, fast, tripped)
		}
	}
	if !l.Blocked(k, now) || !l.BlockedArea(3, now) || !l.BlockedArea(0, now) || l.Blocked(other, now) || l.BlockedArea(4, now) {
		t.Fatal("only the zone that killed him is closed")
	}
	if l.Blocked(k, now.Add(61*time.Minute)) {
		t.Fatal("the closure expires")
	}
	// a long life resets the count
	l2 := &ReviveLoop{N: 2, Within: 10 * time.Minute, Hold: time.Hour}
	l2.Revived(t0)
	l2.Died(k, t0.Add(time.Minute))
	l2.Revived(t0.Add(2 * time.Minute))
	if _, _, fast, _ := l2.Died(k, t0.Add(40*time.Minute)); fast {
		t.Fatal("38 minutes is not quick")
	}
	l2.Revived(t0.Add(41 * time.Minute))
	if _, _, _, tripped := l2.Died(k, t0.Add(42*time.Minute)); tripped || l2.Blocked(k, t0.Add(42*time.Minute)) {
		t.Fatal("the long life must have reset the count")
	}
	// explicit reset and N 0 (off)
	l.Reset(k)
	if l.Blocked(k, now) {
		t.Fatal("reset")
	}
	off := &ReviveLoop{Within: time.Hour, Hold: time.Hour}
	for i := 0; i < 5; i++ {
		off.Revived(t0)
		off.Died(k, t0.Add(time.Minute))
	}
	if off.Blocked(k, t0.Add(2*time.Minute)) {
		t.Fatal("N 0 is off")
	}
}
