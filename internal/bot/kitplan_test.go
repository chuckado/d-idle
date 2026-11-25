package bot

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/supply"
)

type manaCatalog struct{ fakeCatalog }

func (c manaCatalog) Item(class int) (ItemMeta, bool) {
	if class == 6 {
		return ItemMeta{Name: "Mana Potion", W: 1, H: 1, Potion: PotionMana}, true
	}
	return c.fakeCatalog.Item(class)
}

// fixedPlanner answers every guard question with one plan, to see what the bot hands the guards.
type fixedPlanner struct {
	Townsman
	pl supply.Plan
}

func (f fixedPlanner) SupplyPlan(*d2.Snapshot, string, int) (supply.Plan, error) { return f.pl, nil }

// Mana counts belt plus pack at every stage (the quest gate used to see the belt only: a Necromancer with its potions in the pack
// was refused "0 mana potions" for 29 minutes while the shop saw nothing to buy); stage 2 also hands the guards the plan.
func TestFillKitStages(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin, d2.Necromancer} {
		for _, stage := range []int{0, 1, 2, 10} { // 10: stage 0 with quests.boardSM (the guard reads the plan then)
			g := newRig(t)
			g.r.d.Catalog = manaCatalog{}
			g.r.e.Tune.Town.Plan = stage
			if stage == 10 {
				g.r.e.Tune.Town.Plan, g.r.e.Tune.Quests.BoardSM = 0, 1
			}
			g.w.set(func(w *world) {
				w.class = class
				w.items = []d2.Unit{
					{ID: 10, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContBelt}},
					{ID: 11, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}},
					{ID: 12, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}},
				}
			})
			pl := supply.Plan{Lines: []supply.PlanLine{{Line: supply.Healing, Min: 3}, {Line: supply.Mana, Min: 2}, {Line: supply.Portal, Min: 2}}, Verdict: supply.ShortGold, Open: "healing", Short: 60}
			g.r.d.Town = fixedPlanner{Townsman: g.r.d.Town, pl: pl}
			g.r.poll()
			w := g.r.guardWorld(g.r.latest())
			wantMana, wantPlan := 3, false
			if stage >= 2 { // stage 10 attaches the plan like stage 2
				wantPlan = true
			}
			if w.ManaPotions != wantMana || (w.KitPlan != nil) != wantPlan {
				t.Fatalf("%v stage %d: mana %d plan %v, want %d and %v", class, stage, w.ManaPotions, w.KitPlan, wantMana, wantPlan)
			}
			if wantPlan && (w.KitPlan.Healing != 3 || w.KitPlan.Short != "short of gold: needs 60 more for healing") {
				t.Fatalf("%v: plan %+v", class, *w.KitPlan)
			}
		}
	}
}

// The failing case of the Necromancer drill: no mana potion in the belt, all of them in the pack, the quest board's world
// (taskWorld) at the default stage 0.
func TestTaskWorldCountsPackMana(t *testing.T) {
	for _, class := range []d2.Class{d2.Necromancer, d2.Sorceress, d2.Paladin} {
		g := newRig(t)
		g.r.d.Catalog = manaCatalog{}
		g.w.set(func(w *world) {
			w.class = class
			w.items = []d2.Unit{
				{ID: 11, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}},
				{ID: 12, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}},
				{ID: 13, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}},
			}
		})
		g.r.poll()
		w, err := taskEnv{g.r}.World(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if w.ManaPotions != 3 {
			t.Errorf("%v: quest gate world counts %d mana potions, want 3 (pack only)", class, w.ManaPotions)
		}
		if gw := g.r.guardWorld(g.r.latest()); gw.ManaPotions != 3 {
			t.Errorf("%v: route guard world counts %d, want 3", class, gw.ManaPotions)
		}
	}
}

type profilePlanner struct {
	Townsman
	got *string
}

func (f profilePlanner) SupplyPlan(_ *d2.Snapshot, profile string, _ int) (supply.Plan, error) {
	*f.got = profile
	return supply.Plan{Lines: []supply.PlanLine{{Line: supply.Healing, Min: 8}, {Line: supply.GoldCarry, Min: 3000}}}, nil
}

// V live 2026-11-24: the boss chunk kept the road's kit (4 healing) and the purse fell to 188. Within two levels of Andariel's floor
// in Normal, before she is killed, the guards and the town read the "andariel" profile (8 healing, the boss purse); an earlier
// hero keeps the road's. Two classes.
func TestKitPlanUsesTheBossProfileNearAndariel(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Necromancer} {
		for _, tc := range []struct {
			level int
			want  string
		}{{20, "andariel"}, {5, kitProfile}} {
			g := newRig(t)
			g.r.e.Tune.Quests.BoardSM = 1
			g.r.e.Tune.Boss.AndarielHold = 0
			var got string
			g.r.d.Town = profilePlanner{Townsman: g.r.d.Town, got: &got}
			g.w.set(func(w *world) { w.class = class; w.level = tc.level })
			g.r.poll()
			k := g.r.kitPlanOf(g.r.latest())
			if got != tc.want {
				t.Errorf("%v level %d: profile %q, want %q", class, tc.level, got, tc.want)
			}
			if tc.want == "andariel" && (k == nil || k.GoldCarry != 3000 || g.r.bossPurse() != 3000) {
				t.Errorf("%v: boss purse not handed on: %+v", class, k)
			}
			// the boss floor is what belt and pack carry: no buffer above it, the shop could never close it (V live 2026-11-24)
			if h, _ := g.r.routeKit(); tc.want == "andariel" && h != 8 {
				t.Errorf("%v: the trip buys up to %d healing, want the plan's 8 without a buffer", class, h)
			}
		}
	}
}

// While boss.andarielHold holds Andariel back (and no owner try is on), the boss kit floor is not asked of a hero that only levels
// (V live 2026-11-25: 815 refusals "route guard (4 healing potions < 5)"). Two classes.
func TestBossKitIsNotAskedWhileAndarielIsHeld(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Necromancer} {
		g := newRig(t)
		g.r.e.Tune.Quests.BoardSM = 1
		g.r.e.Tune.Boss.AndarielHold = 1
		var got string
		g.r.d.Town = profilePlanner{Townsman: g.r.d.Town, got: &got}
		g.w.set(func(w *world) { w.class = class; w.level = 20 })
		g.r.poll()
		g.r.kitPlanOf(g.r.latest())
		if got != kitProfile {
			t.Errorf("%v: profile %q while Andariel is held, want %q", class, got, kitProfile)
		}
		if g.r.bossPurse() != 0 {
			t.Errorf("%v: boss purse while held", class)
		}
	}
}

// Questy bed 2026-11-25: a clvl 4 Paladin with no potions walked Stony Field, the Underground Passage and toward the Dark Wood and
// died. The early road (earlyroutes.json) is closed below the class family's level and short of 2 healing potions and a portal
// scroll; the owner's try anyway opens it. Two classes (melee 7, caster 8).
func TestEarlyRouteClosesBelowTheLevelAndTheKit(t *testing.T) {
	for _, tc := range []struct {
		class   d2.Class
		level   int
		potions int
		maxLife int
		want    string // "" open, else a part of the reason
	}{
		{d2.Paladin, 4, 3, 150, "level 7 needed"},
		{d2.Paladin, 7, 0, 150, "healing potions"},
		{d2.Paladin, 7, 2, 150, ""},
		{d2.Paladin, 8, 3, 69, "life 110 needed"}, // the Questy hero: clvl 8, nothing in vitality
		{d2.Sorceress, 7, 3, 90, "level 8 needed"},
		{d2.Sorceress, 8, 2, 90, ""},
		{d2.Sorceress, 8, 3, 50, "life 70 needed"},
	} {
		g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) { tu.Route.EarlyRoute = 1 })
		g.w.set(func(w *world) {
			w.class, w.level, w.area, w.maxL, w.life = tc.class, tc.level, 3, tc.maxLife, tc.maxLife
			var keep []d2.Unit
			for _, u := range w.items {
				if u.Class != 1 && u.Class != 5 { // the rig's own potions and scrolls go: the case sets them
					keep = append(keep, u)
				}
			}
			w.items = keep
		})
		for i := 0; i < tc.potions; i++ {
			g.w.set(func(w *world) {
				w.items = append(w.items, d2.Unit{ID: uint32(700 + i), Type: d2.TypeItem, Class: 1, Item: &d2.Item{Where: d2.ContBelt}})
			})
		}
		g.w.set(func(w *world) {
			w.items = append(w.items, d2.Unit{ID: 799, Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}}) // a portal scroll
		})
		g.r.poll()
		got := g.r.earlyRouteWhy(g.r.latest(), 10)
		if tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%v clvl %d potions %d: %q, want %q", tc.class, tc.level, tc.potions, got, tc.want)
		}
	}
	if (&Runner{}).earlyRouteWhy(nil, 10) != "" {
		t.Error("no snapshot must not close the road")
	}
}

// The walk itself is refused: a quest travel from Stony Field through the Underground Passage to the Dark Wood returns the gate
// error for a clvl 4 hero (the board shows the reason and the hero levels), and a walk to town is still fine. Two classes.
func TestEarlyRoadRefusesTheQuestWalk(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) { tu.Route.EarlyRoute = 1 })
		g.w.set(func(w *world) { w.class, w.level, w.area = class, 4, 4 })
		g.r.poll()
		s := g.r.latest()
		if err := g.r.routeAllowed(s, 4, 5, []int{10, 5}); !errors.Is(err, errAreaGated) {
			t.Errorf("%v: walk to the Dark Wood: %v, want the gate error", class, err)
		}
		if err := g.r.routeAllowed(s, 4, 1, []int{3, 2, 1}); err != nil {
			t.Errorf("%v: the walk home was refused: %v", class, err)
		}
	}
}
