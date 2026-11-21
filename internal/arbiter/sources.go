package arbiter

import "fmt"

// World fact names the standard sources read. The bot adapter and the replay tool fill them; a missing fact is 0.
const (
	FactInTown       = "in_town"
	FactLifePct      = "life_pct"
	FactHostile      = "hostile"      // hostiles near the hero
	FactPlayerOrder  = "player_order" // 1 while a player order runs
	FactTrip         = "trip"         // 1 while a town trip is running or chores are due
	FactTripEmerg    = "trip_emergency"
	FactClaim        = "claim_pending" // quest rewards to claim in town
	FactPotions      = "potions"
	FactPotionsMin   = "potions_min" // belt refill line; 0 = no resupply source
	FactPackGold     = "pack_gold"
	FactStashGold    = "stash_gold"
	FactNeedGold     = "need_gold" // cost of the refill; 0 = unknown
	FactRecovery     = "recovery"
	FactGoldFarm     = "gold_farm"
	FactObjective    = "objective"
	FactSurviveBelow = "survive_below" // life percent under which a hostile makes the hero flee or drink
)

// Source names and the hook family each belongs to (the counters of the minute line are per family).
const (
	SrcSurvive  = "survive"
	SrcPlayer   = "player"
	SrcTrip     = "trip"
	SrcClaim    = "claim"
	SrcResupply = "resupply"
	SrcRecovery = "recovery"
	SrcGoldFarm = "goldfarm"
	SrcQuest    = "quest"
	SrcLevel    = "level"
)

// Family maps a source name to its hook family: the trip and the belt refill are both "chores".
func Family(source string) string {
	if source == SrcTrip || source == SrcResupply {
		return "chores"
	}
	return source
}

func fn(name string, f func(View) *Intent) Source { return FuncSource{N: name, F: f} }

// StdSources are the sources the shadow runs. Each reads only View (world facts), never another source. Build data
// (belt line, flee line, gold need) arrives as facts, so no class or build literal lives here.
func StdSources() []Source {
	return []Source{
		fn(SrcSurvive, func(v View) *Intent {
			f := v.Facts
			if f[FactInTown] == 0 && f[FactHostile] > 0 && f[FactSurviveBelow] > 0 && f[FactLifePct] < f[FactSurviveBelow] {
				return &Intent{Kind: "drink", Class: Reflex, Urgency: 1, NeedsArea: Any, EstCost: 2,
					Reason: fmt.Sprintf("life %d%% under %d%% with hostiles near", f[FactLifePct], f[FactSurviveBelow])}
			}
			return nil
		}),
		fn(SrcPlayer, func(v View) *Intent {
			if v.Facts[FactPlayerOrder] > 0 {
				return &Intent{Kind: "order", Class: Player, Urgency: 1, NeedsArea: Any, Reason: "player order", Resumable: true}
			}
			return nil
		}),
		fn(SrcTrip, func(v View) *Intent {
			f := v.Facts
			if f[FactTrip] == 0 {
				return nil
			}
			u, why := 0.5, "town trip running or chores due"
			if f[FactTripEmerg] > 0 {
				u, why = 0.9, "emergency chores"
			}
			return &Intent{Kind: "chores", Class: Upkeep, Urgency: u, NeedsArea: Town, EstCost: 60, Reason: why}
		}),
		fn(SrcClaim, func(v View) *Intent {
			if v.Facts[FactClaim] == 0 {
				return nil
			}
			return &Intent{Kind: "claim", Class: Upkeep, Urgency: 0.4, NeedsArea: Town, EstCost: 30, Reason: "quest reward to claim"}
		}),
		FuncSource{N: SrcResupply, F: func(v View) *Intent {
			f := v.Facts
			if f[FactPotionsMin] == 0 || f[FactPotions] >= f[FactPotionsMin] {
				return nil
			}
			in := Intent{Kind: "refill", Class: Upkeep, Urgency: 0.5, NeedsArea: Town, EstCost: 60,
				Reason: fmt.Sprintf("belt %d under %d", f[FactPotions], f[FactPotionsMin])}
			need := f[FactNeedGold]
			if need > 0 && f[FactPackGold] < need {
				if f[FactPackGold]+f[FactStashGold] < need {
					return nil // cannot pay: recovery's business
				}
				in.Steps = []string{fmt.Sprintf("withdraw_gold %d", need-f[FactPackGold]), "buy_potions"}
				in.Reason += fmt.Sprintf(", pack %d + stash %d covers %d", f[FactPackGold], f[FactStashGold], need)
			}
			return &in
		}, Town: func(v View) *Intent { return nil }},
		fn(SrcRecovery, func(v View) *Intent {
			if v.Facts[FactRecovery] == 0 {
				return nil
			}
			return &Intent{Kind: "farm_gold", Class: Objective, Urgency: 0.8, NeedsArea: Any, EstCost: 600, Reason: "broke: recovery farm"}
		}),
		fn(SrcGoldFarm, func(v View) *Intent {
			if v.Facts[FactGoldFarm] == 0 {
				return nil
			}
			return &Intent{Kind: "farm", Class: Objective, Urgency: 0.7, NeedsArea: Any, EstCost: 600, Reason: "player's gold farm order"}
		}),
		fn(SrcQuest, func(v View) *Intent {
			if v.Facts[FactObjective] == 0 {
				return nil
			}
			need := v.ObjectiveArea
			if need == "" {
				need = Any
			}
			return &Intent{Kind: "step", Class: Objective, Urgency: 0.5, NeedsArea: need, EstCost: 300, Reason: "quest step in " + need}
		}),
		fn(SrcLevel, func(View) *Intent {
			return &Intent{Kind: "sweep", Class: Filler, Urgency: 0.1, NeedsArea: Any, EstCost: 300, Reason: "nothing else wanted: level the hero"}
		}),
	}
}
