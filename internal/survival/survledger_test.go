package survival

import (
	"d2lod-bot/internal/power"
	"strings"
	"testing"
	"time"
)

// TestMercDeathAndFightsReachTheLedger replays the harness sequence on two builds with a lockstep-style epoch clock (the
// clock bug of survfix put events 56 years off): fights, a merc death (the merc flips dead, so the hero's power input
// of the next decision is read with a dead merc), and a decision. The row must keep the death and the fights, k must
// move, and every decision line must print k.
func TestMercDeathAndFightsReachTheLedger(t *testing.T) {
	epoch := time.Unix(0, 0).UTC()
	for _, tc := range []struct {
		name, arch string
		hero       func(int) power.Hero
	}{{"paladin", "bruiser", paladin}, {"sorceress", "caster", sorceress}} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t, tc.arch)
			l := NewLedger(e.p)
			h := tc.hero(14)
			k := ZoneKey{e.cands(h)[0].Key.Area, e.cands(h)[0].Key.Diff}
			now := epoch.Add(20 * time.Minute)
			in := e.input(h, k, now)
			d, err := l.Choose(in)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(d.Line(), "  k ") {
				t.Fatalf("the zone choice line must print k: %q", d.Line())
			}
			before := l.KZone(k, now)
			// fights without a prediction still accumulate kills and seconds, and leave the offence pool alone
			l.Observe(k, now, Event{Kind: EvFight, Kills: 12, Seconds: 90, Taken: 300}, in.Power)
			l.Observe(k, now, Event{Kind: EvFight, Kills: 8, Seconds: 60, Taken: 100, PredTaken: 100, PredKillSec: 50}, in.Power)
			s := l.peek(k)
			if s.Kills < 19 || s.Seconds < 140 {
				t.Fatalf("fights did not accumulate: kills %.1f seconds %.1f", s.Kills, s.Seconds)
			}
			if l.HeroKill < 7.9 || l.HeroKill > 8.1 {
				t.Fatalf("the unpredicted fight leaked into the offence pool: %.1f", l.HeroKill)
			}
			// the merc dies; the next decision sees a hero whose merc is dead
			l.Observe(k, now.Add(time.Minute), Event{Kind: EvMercDeath}, in.Power)
			dead := h
			dead.Merc.Alive = false
			in2 := e.input(dead, k, now.Add(2*time.Minute))
			if in2.Power != in.Power {
				t.Fatalf("a merc death must not move the power scalar: %v vs %v", in.Power, in2.Power)
			}
			if _, err := l.Choose(in2); err != nil {
				t.Fatal(err)
			}
			s = l.peek(k)
			if s.MercDeaths < 1 || s.Hard <= 1 || s.Kills < 19 {
				t.Fatalf("the death or the fights were wiped by the next decision: %+v", s)
			}
			if after := l.KZone(k, now.Add(2*time.Minute)); after <= before {
				t.Fatalf("k did not move after a merc death: %.3f -> %.3f", before, after)
			}
		})
	}
}
