package bot

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

type beltFillFake struct {
	*fakeTown
	needs bool
	calls int
}

func (b *beltFillFake) NeedsBeltFill(*d2.Snapshot) bool { return b.needs }
func (b *beltFillFake) FillBelt(context.Context) (int, error) {
	b.calls++
	return 2, nil
}

func TestBeltFillStep(t *testing.T) {
	rg := newRig(t)
	bf := &beltFillFake{fakeTown: rg.town, needs: true}
	rg.r.d.Town = bf
	r := rg.r
	s := townSnap(2, 100)
	s.Inventory = &d2.Inventory{}
	if !r.beltFillStep(context.Background(), s) || bf.calls != 1 {
		t.Fatalf("no fill: calls %d", bf.calls)
	}
	if r.beltFillStep(context.Background(), s) || bf.calls != 1 {
		t.Error("filled again inside the backoff")
	}
	r.e.Clock.Sleep(beltFillBackoff + time.Second)
	bf.needs = false
	if r.beltFillStep(context.Background(), s) || bf.calls != 1 {
		t.Error("filled without a need")
	}
}

// A belt with no healing potion and healing in the pack fills with a hostile 20 subtiles away (outside the urgent
// radius, inside the sweep radius); with healing in the belt, or the hostile at 8, it does not. Must fail without the
// urgent radius. The decision reads belt and pack counts only, no class: both classes run it.
func TestBeltFillStepUrgentRadius(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, tc := range []struct {
			name      string
			beltHP    bool
			hostileAt int
			want      bool
		}{{"empty belt, hostile at 20", false, 20, true}, {"empty belt, hostile at 8", false, 8, false}, {"healing in the belt, hostile at 20", true, 20, false}} {
			rg := newRig(t)
			bf := &beltFillFake{fakeTown: rg.town, needs: true}
			rg.r.d.Town = bf
			p := &d2.Player{Area: 7, Life: 200, MaxLife: 250, Class: class, X: 100, Y: 100}
			inv := &d2.Inventory{}
			inv.Items = append(inv.Items, d2.Unit{ID: 50, Type: d2.TypeItem, Class: 1, Item: &d2.Item{Where: d2.ContInventory}})
			if tc.beltHP {
				inv.Items = append(inv.Items, d2.Unit{ID: 51, Type: d2.TypeItem, Class: 1, Item: &d2.Item{Where: d2.ContBelt}})
			}
			m := d2.Unit{ID: 9, Type: d2.TypeMonster, Class: 10, Mode: d2.MonModeStand, X: 100 + tc.hostileAt, Y: 100}
			s := &d2.Snapshot{State: d2.StateInGame, Player: p, Inventory: inv, Monsters: []d2.Unit{m}}
			if got := rg.r.beltFillStep(context.Background(), s); got != tc.want {
				t.Errorf("%v %s: filled %v, want %v", class, tc.name, got, tc.want)
			}
		}
	}
}
