package combat

import (
	"bytes"
	"context"
	"d2lod-bot/internal/d2"
	"log/slog"
	"strings"
	"testing"
)

func TestFightEmitsTelemetryLine(t *testing.T) {
	h := newHarness(t)
	var buf bytes.Buffer
	h.f.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
	h.s.stock(itHP1, 0, 2)
	h.s.addMon(10, clsFallen, 120, 100, 30, 5, 5)
	if _, err := h.f.Clear(context.Background(), ClearOpts{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if n := strings.Count(out, "fight area="); n != 1 || !strings.Contains(out, "kills=1") {
		t.Fatalf("want one fight line with kills=1:\n%s", out)
	}
}

func TestSafetyCurseRetreatDrinks(t *testing.T) {
	h := newHarness(t)
	h.s.stock(itHP1, 0, 2)
	h.s.addMon(10, clsFallen, 110, 100, 30, 0, 0)
	snap, _ := h.s.snapshot()
	p := snap.Player
	p.MaxLife, p.Life = 100, 55 // above the plain drink line, below the curse line
	id := curseIDs.Amplify
	p.States[id>>5] |= 1 << (uint(id) & 31)
	if a := h.f.Safety(snap); a != DrinkLife {
		t.Fatalf("cursed at 55%%: %v", a)
	}
	p.States = [6]uint32{}
	if a := h.f.Safety(snap); a != SafetyNone {
		t.Fatalf("uncursed at 55%%: %v", a)
	}
}

// A living character with life 0 (a reload after a chicken, V live Questy 2026-11-26) drinks from the belt, in town as in the
// field; a dying one does not; with no potion nothing is asked. Must fail while Safety returns nothing for life 0. Two classes.
func TestSafetyZeroLifeDrinks(t *testing.T) {
	for _, cl := range []d2.Class{d2.Paladin, d2.Sorceress} {
		t.Run(cl.String(), func(t *testing.T) {
			h := newHarness(t)
			h.s.stock(itHP1, 0, 2)
			snap, _ := h.s.snapshot()
			p := snap.Player
			p.Class = cl
			p.MaxLife, p.Life, p.Mode = 176, 0, 5 // standing, no life
			if a := h.f.Safety(snap); a != DrinkLife {
				t.Fatalf("life 0, standing: %v", a)
			}
			p.Mode = ModeDying
			if a := h.f.Safety(snap); a == DrinkLife {
				t.Fatalf("a dying character was asked to drink")
			}
		})
	}
}
