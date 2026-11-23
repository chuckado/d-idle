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

// Every town visit buys healing up to the route guard's floor (belt and pack) while the guard is on in Normal from the kit
// gate level: the loop's own trips used to fill only the belt plan and leave a four box belt one potion short. Must fail
// without routeKit in fullChores. Both classes; guard off, Nightmare and a low level ask for nothing.
func TestFullChoresBuyUpToTheRouteKit(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, tc := range []struct {
			name  string
			guard int
			diff  d2.Difficulty
			lvl   int
			want  bool
		}{{"guard on", 20, d2.Normal, 19, true}, {"guard off", 0, d2.Normal, 19, false}, {"nightmare", 20, d2.Nightmare, 19, false}, {"below the kit gate", 20, d2.Normal, 3, false}} {
			rg := newRig(t)
			rg.r.e.Tune.Route.MonasteryMinLevel = tc.guard
			rg.r.obMu.Lock()
			rg.r.ob.lastSnap = &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 1, Level: tc.lvl, Class: class, Difficulty: tc.diff, Life: 100, MaxLife: 100}}
			rg.r.obMu.Unlock()
			p := rg.r.fullChores("loop")
			if got := p.KitHealing > 0; got != tc.want || tc.want && p.KitHealing != rg.r.e.Tune.Resource.KitMinHP+kitHealingBuffer {
				t.Errorf("%v %s: KitHealing %d, want set %v", class, tc.name, p.KitHealing, tc.want)
			}
		}
	}
}

// A potion trip built by the resource policy (a plan that buys potions and names no kit) gets the route guard kit before
// it reaches the shop; a plan that names its own kit, or buys no potions, is left alone. Must fail without withRouteKit.
func TestWithRouteKit(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		rg := newRig(t)
		rg.r.e.Tune.Route.MonasteryMinLevel = 20
		rg.r.obMu.Lock()
		rg.r.ob.lastSnap = &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 1, Level: 19, Class: class, Difficulty: d2.Normal, Life: 100, MaxLife: 100}}
		rg.r.obMu.Unlock()
		want := rg.r.e.Tune.Resource.KitMinHP + kitHealingBuffer
		if got := rg.r.withRouteKit(ChorePlan{Potions: true}); got.KitHealing != want {
			t.Errorf("%v: potion trip KitHealing %d, want %d", class, got.KitHealing, want)
		}
		if got := rg.r.withRouteKit(ChorePlan{Potions: true, KitHealing: 9}); got.KitHealing != 9 {
			t.Errorf("%v: own kit overwritten: %d", class, got.KitHealing)
		}
		if got := rg.r.withRouteKit(ChorePlan{Heal: true}); got.KitHealing != 0 {
			t.Errorf("%v: a heal trip got a kit: %d", class, got.KitHealing)
		}
	}
}

// Below the class floor of the Barracks (route.barracks*) the picker never offers the Barracks or what lies past it
// (Jail to Catacombs: the farm walk crosses the arrival), and a walk through them is refused with the reason. At the
// floors both are open. Paladin and Sorceress; fails without barracksGuardWhy.
func TestBarracksFloorKeepsTheFarmWalkOut(t *testing.T) {
	for _, tc := range []struct {
		class       d2.Class
		level, life int
		closed      bool
	}{
		{d2.Paladin, 20, 279, true}, {d2.Paladin, 24, 420, false}, {d2.Paladin, 24, 300, true},
		{d2.Sorceress, 21, 300, true}, {d2.Sorceress, 22, 300, false},
	} {
		rg := newRig(t)
		s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Area: 5, Level: tc.level, Class: tc.class, Difficulty: d2.Normal, Life: tc.life, MaxLife: tc.life}}
		why := rg.r.barracksGuardWhy(s, 30)
		ex := map[int]bool{}
		rg.r.guardedAreasNoFarm(s, ex)
		if (why != "") != tc.closed {
			t.Errorf("%v clvl %d life %d: why %q, closed want %v", tc.class, tc.level, tc.life, why, tc.closed)
		}
		for _, a := range []int{28, 29, 33, 35, 37} {
			// at the floors the route guard may still close them (the rig has no kit): only the closed case is asserted
			if tc.closed && !ex[a] {
				t.Errorf("%v clvl %d life %d: area %d not excluded below the floor", tc.class, tc.level, tc.life, a)
			}
		}
	}
}
