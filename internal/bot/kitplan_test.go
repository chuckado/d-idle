package bot

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
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
