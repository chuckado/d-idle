package town

import (
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/resource"
	"d2lod-bot/internal/supply"
	"d2lod-bot/internal/town/belt"
)

// Stage 1 of the town planner (docs/design/fable-review-2026-11-21/town-planner.md, knob town.plan): the plan is solved
// beside every chores visit and logged, and nothing acts on it. PlanStage is the knob's value.
func (t *Town) PlanStage() int { return t.tune().Plan }

// PlanShadow solves the supply plan of the current snapshot for the excursion profile of the chore plan (the `field`
// profile when the caller names none). It reads and changes nothing.
func (t *Town) PlanShadow(plan ChorePlan) (supply.Plan, error) {
	s, err := t.current()
	if err != nil {
		return supply.Plan{}, err
	}
	return t.planOf(s, plan)
}

func (t *Town) planOf(s *d2.Snapshot, plan ChorePlan) (supply.Plan, error) {
	p := s.Player
	c := t.supplyCount(s)
	c.Purse, c.Stash, c.StashKnown = p.Gold, p.StashGold, p.StashGold > 0
	c.PortalCharges, c.HasTome = t.scrollCount(s, "tsc", "tbk")
	return solveShadow(p.Class, t.plan(p), belt.Slots(t.e.Data, s.Inventory), c, plan, &t.e.Tune.Loot, &t.e.Tune.Resource)
}

// solveShadow is planOf without the snapshot reads, so it runs on any class and band without the game tables.
func solveShadow(class d2.Class, band belt.Band, slots int, c supply.Count, plan ChorePlan, lt *env.LootTuning, r *env.ResourceTuning) (supply.Plan, error) {
	capHP, capMP := band.Capacity(belt.HP, slots), band.Capacity(belt.MP, slots)
	need := resource.ManaNeed(class)
	if capMP == 0 {
		need = 0 // the belt plan has no mana column: the kit asks for none
	}
	in := supply.Inputs{
		ManaNeed:   need,
		Capacity:   map[supply.Line]int{supply.Healing: capHP + lt.PotionBuffer, supply.Mana: capMP + lt.PotionBufferMana},
		BaseTarget: map[supply.Line]int{supply.Healing: capHP + lt.PotionBuffer, supply.Mana: capMP + lt.PotionBufferMana},
	}
	profile := plan.Profile
	if profile == "" {
		profile = "field"
	}
	n, err := supply.NeedFor("town", profile, in)
	if err != nil {
		return supply.Plan{}, err
	}
	prices := supply.Prices{supply.Healing: r.KitPotionGold, supply.Mana: r.KitPotionGold, supply.Portal: r.KitTPGold, supply.GoldCarry: 1}
	return supply.Solve(n, c, prices, supply.Policy{Reserve: plan.ReserveGold}, 0), nil
}

// logPlanShadow writes the stage 1 trace line when the knob is on (docs/design town-planner section 8).
func (t *Town) logPlanShadow(s *d2.Snapshot, plan ChorePlan) {
	if t.PlanStage() < 1 {
		return
	}
	pl, err := t.planOf(s, plan)
	if err != nil {
		t.e.Log.Warn("town plan (shadow): " + err.Error())
		return
	}
	t.e.Log.Info("town plan (shadow): " + pl.Trace())
}
