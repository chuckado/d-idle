package bot

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

var reaskClasses = []d2.Class{d2.Sorceress, d2.Paladin}

func reaskRig(t *testing.T, cl d2.Class, dec func(area int, cands []ZoneCand) (ZoneVerdict, error)) (*rig, *bytes.Buffer) {
	g, _ := modelRig(t, cl, dec)
	buf := new(bytes.Buffer)
	g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
	return g, buf
}

// A chicken makes the next pick ask the model at once, the model's answer is used, and the line says why.
func TestZoneReaskAfterChicken(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			area, k := 8, 1.0
			g, buf := reaskRig(t, cl, func(int, []ZoneCand) (ZoneVerdict, error) { return ZoneVerdict{Area: area, K: k}, nil })
			g.modelPick(12, nil)
			if g.r.zoneRecheckDue() {
				t.Fatal("due without a cost event")
			}
			g.clock.Sleep(10 * time.Second)
			area, k = 2, 1.5
			g.r.zoneCost("chicken")
			if !g.r.zoneRecheckDue() {
				t.Fatal("a chicken did not make the pick due")
			}
			st, ok := g.modelPick(12, nil)
			if !ok || st.Area != 2 {
				t.Fatalf("pick %d (%v), want the model's zone 2", st.Area, ok)
			}
			if !strings.Contains(buf.String(), "zone reask: reason=chicken k 1.00 -> 1.50 choice 2") {
				t.Fatalf("no reask line:\n%s", buf)
			}
			if g.r.zoneRecheckDue() {
				t.Fatal("still due after the answer")
			}
		})
	}
}

// Rate limit: no re-ask closer than reaskGap, at most reaskMax per window.
func TestZoneReaskNoFlapping(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			calls := 0
			g, _ := reaskRig(t, cl, func(int, []ZoneCand) (ZoneVerdict, error) { calls++; return ZoneVerdict{Area: 2}, nil })
			g.modelPick(12, nil)
			base := calls
			g.r.zoneCost("chicken")
			g.modelPick(12, nil)
			g.clock.Sleep(time.Second)
			g.r.zoneCost("death")
			if g.r.zoneRecheckDue() {
				t.Fatal("re-ask inside the gap")
			}
			done := 0
			for i := 0; i < 20; i++ {
				g.clock.Sleep(reaskGap + time.Second)
				g.r.zoneCost("chicken")
				if g.r.zoneRecheckDue() {
					g.modelPick(12, nil)
					done++
				}
			}
			if done+1 > reaskMax {
				t.Fatalf("%d re-asks inside the window, max %d", done+1, reaskMax)
			}
			if calls-base != done+1 {
				t.Fatalf("calls %d", calls-base)
			}
			g.clock.Sleep(reaskWindow)
			g.r.zoneCost("chicken")
			if !g.r.zoneRecheckDue() {
				t.Fatal("window did not reopen")
			}
		})
	}
}

// Seed 149: six chickens in 35 s. A model that closes the zone on the second chicken (the step-down row) is heard
// after at most two of them.
func TestZoneReaskSeed149Pattern(t *testing.T) {
	for _, cl := range reaskClasses {
		t.Run(cl.String(), func(t *testing.T) {
			chickens := 0
			g, buf := reaskRig(t, cl, func(int, []ZoneCand) (ZoneVerdict, error) {
				if chickens >= 2 {
					return ZoneVerdict{Area: 2, K: 3}, nil
				}
				return ZoneVerdict{Area: 8, K: 1 + float64(chickens)}, nil
			})
			g.modelPick(12, nil)
			left := 0
			for i := 1; i <= 6 && left == 0; i++ {
				g.clock.Sleep(7 * time.Second)
				chickens = i
				g.r.zoneCost("chicken")
				if g.r.zoneRecheckDue() {
					if st, ok := g.modelPick(12, nil); ok && st.Area == 2 {
						left = i
					}
				}
			}
			if left == 0 || left > 2 {
				t.Fatalf("left after %d chickens, want at most 2\n%s", left, buf)
			}
		})
	}
}

// Two emergency drinks within a minute are a cost event; one is not.
func TestZoneReaskPotionCrisis(t *testing.T) {
	g, _ := reaskRig(t, d2.Paladin, func(int, []ZoneCand) (ZoneVerdict, error) { return ZoneVerdict{Area: 2}, nil })
	g.modelPick(12, nil)
	g.clock.Sleep(10 * time.Second)
	g.r.zonePotion()
	if g.r.zoneRecheckDue() {
		t.Fatal("one drink is not a crisis")
	}
	g.clock.Sleep(10 * time.Second)
	g.r.zonePotion()
	if !g.r.zoneRecheckDue() {
		t.Fatal("two drinks did not trigger")
	}
}

// Knob off: nothing is recorded and nothing is due.
func TestZoneReaskKnobOff(t *testing.T) {
	for _, cl := range reaskClasses {
		g, _ := reaskRig(t, cl, func(int, []ZoneCand) (ZoneVerdict, error) { return ZoneVerdict{Area: 2}, nil })
		g.modelPick(12, nil)
		g.r.e.Tune.Bot.ZoneModel = 0
		g.r.zoneCost("chicken")
		g.r.zonePotion()
		g.r.zonePotion()
		g.r.e.Tune.Bot.ZoneModel = 1
		if g.r.zst.reask.pending != "" || g.r.zoneRecheckDue() {
			t.Fatal("cost events leaked into a run with the knob off")
		}
	}
}
