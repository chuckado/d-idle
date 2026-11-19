package bot

import (
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// The seed-149 sequence on a fake epoch clock: a chicken in the zone the legacy rule returns to, the trip's return
// portal asks the model, the model refuses the zone, the alternative is used and held (no flapping).
func TestZoneRecoverReturnSeed149(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			var want, alt int
			g, buf := reaskRig(t, cl, func(area int, cands []ZoneCand) (ZoneVerdict, error) {
				for _, c := range cands {
					if c.Area != want {
						alt = c.Area // the model refuses the burned zone
						return ZoneVerdict{Area: c.Area, Margin: "safe", K: 1.2}, nil
					}
				}
				return ZoneVerdict{}, errors.New("no alternative among the candidates")
			})
			g.w.set(func(w *world) { w.level = 12 })
			g.r.poll()
			s := g.r.latest()
			want = legacy(g, 12).Area
			calls := 0
			dec := g.r.d.ZoneDecide
			g.r.d.ZoneDecide = func(s *d2.Snapshot, a int, d string, p int, dial string, c []ZoneCand) (ZoneVerdict, error) {
				calls++
				return dec(s, a, d, p, dial, c)
			}
			if got := g.r.zoneRecoverReturn(s, want); got != want || calls != 0 {
				t.Fatalf("clean zone: got %d after %d calls, want %d and none", got, calls, want)
			}
			g.r.zoneCostAt("chicken", want)
			got := g.r.zoneRecoverReturn(s, want)
			if got == want || got != alt || alt == 0 {
				t.Fatalf("burned zone %d: got %d, want the alternative %d", want, got, alt)
			}
			if !strings.Contains(buf.String(), "zone recover: reason=chicken via=return") || !strings.Contains(buf.String(), "(alternative)") {
				t.Fatalf("no recover line:\n%s", buf)
			}
			g.clock.Sleep(time.Minute)
			if got := g.r.zoneRecoverReturn(s, want); got != alt || calls != 1 {
				t.Fatalf("hold: got %d after %d calls", got, calls)
			}
			g.clock.Sleep(burnWindow)
			if got := g.r.zoneRecoverReturn(s, want); got != want {
				t.Fatalf("burn expired: got %d, want the legacy zone %d", got, want)
			}
		})
	}
}

func TestZoneRecoverOffErrorAndReturn(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			fail := true
			g, buf := reaskRig(t, cl, func(area int, cands []ZoneCand) (ZoneVerdict, error) {
				if fail {
					return ZoneVerdict{}, errors.New("boom")
				}
				return ZoneVerdict{Area: cands[0].Area, Margin: "safe"}, nil
			})
			g.w.set(func(w *world) { w.level = 12 })
			g.r.poll()
			s := g.r.latest()
			want := legacy(g, 12).Area
			g.r.zoneCostAt("death", want)
			if got := g.r.zoneRecoverReturn(s, want); got != want || !strings.Contains(buf.String(), "legacy choice stands") {
				t.Fatalf("error: got %d\n%s", got, buf)
			}
			fail = false
			if got := g.r.zoneRecoverReturn(s, want); got != want || !strings.Contains(buf.String(), "choice") {
				t.Fatalf("model keeps the zone: got %d\n%s", got, buf)
			}
			g.r.e.Tune.Bot.ZoneModel = 0 // byte-identical: nothing recorded, nothing asked
			g2 := farmLoopRig(t, cl, 10)
			g2.r.d.ZoneDecide = func(*d2.Snapshot, int, string, int, string, []ZoneCand) (ZoneVerdict, error) {
				t.Error("asked with the knob off")
				return ZoneVerdict{}, nil
			}
			g2.r.zoneCostAt("chicken", want)
			if len(g2.r.zst.burn) != 0 || g2.r.zoneRecoverReturn(g2.r.latest(), want) != want {
				t.Fatal("knob 0 recorded or changed something")
			}
		})
	}
}

// The recovery farm area: the zone the legacy rules would farm just cost a chicken, the model picks the other
// area the recovery rules allow; knob 0 keeps the legacy pick.
func TestZoneRecoverFarmArea(t *testing.T) {
	g, s := kitRig(t, 14, true, true, 0)
	legacyArea, ok := g.r.recoveryArea(s)
	if !ok {
		t.Fatal("no legacy recovery area")
	}
	g.r.e.Tune.Bot.ZoneModel = 1
	g.r.d.ZoneDecide = func(_ *d2.Snapshot, _ int, _ string, _ int, _ string, c []ZoneCand) (ZoneVerdict, error) {
		return ZoneVerdict{Area: c[len(c)-1].Area, Margin: "safe"}, nil // the alternative is listed last
	}
	if a, _ := g.r.recoveryArea(s); a != legacyArea {
		t.Fatalf("clean zone: %d, want %d", a, legacyArea)
	}
	g.r.zoneCostAt("chicken", legacyArea)
	a, _ := g.r.recoveryArea(s)
	if a == legacyArea || a == 0 {
		t.Fatalf("burned zone %d still chosen (%d)", legacyArea, a)
	}
	if b, _ := g.r.recoveryArea(s); b != a {
		t.Fatalf("flapped: %d then %d", a, b)
	}
	g.r.e.Tune.Bot.ZoneModel = 0
	if a, _ := g.r.recoveryArea(s); a != legacyArea {
		t.Fatalf("knob 0: %d, want %d", a, legacyArea)
	}
}
