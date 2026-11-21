package bot

import (
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

// Stage 2 counts mana in belt plus pack and hands the guards the plan; stage 0 and 1 keep the belt-only count and no plan.
func TestFillKitStages(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
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
			wantMana, wantPlan := 1, false
			if stage >= 2 { // stage 10 counts like stage 2
				wantMana, wantPlan = 3, true
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
