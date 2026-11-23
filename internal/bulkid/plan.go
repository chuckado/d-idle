package bulkid

import (
	"sort"

	"d2lod-bot/internal/env"
)

// Ways an item is identified.
const (
	ViaCain   = "cain"
	ViaScroll = "scroll"
)

// Reasons a plan is empty or an item is skipped (shown in the viewer and the log; stable words).
const (
	WhyManual    = "identify is manual"
	WhyNever     = "identify policy is never"
	WhyNoCain    = "Cain is not rescued yet: identify is manual"
	WhyFilter    = "the filter does not select it"
	WhyValue     = "below identify.minValue"
	WhyCap       = "over the per-visit cap"
	WhyGoldFloor = "buying a scroll would leave less than identify.goldFloor"
	WhyIdentd    = "already identified"
)

// Hero is the state the planner reads; nothing else about the hero matters (no class, no build).
type Hero struct {
	CainOpen    bool // Cain is rescued and stands in this town
	Charges     int  // identify scrolls plus tome charges held; -1 unknown counts as none
	Gold        int
	ScrollPrice int // gold per scroll at the vendor (0: unknown, nothing is bought)
}

// Pick is one item to identify and how.
type Pick struct {
	Item Item   `json:"item"`
	Via  string `json:"via"`
	Rule string `json:"rule"`
}

// Skip is a candidate that was left and why.
type Skip struct {
	ID     uint32 `json:"id"`
	Reason string `json:"reason"`
}

// Plan is the planner's answer for one visit.
type Plan struct {
	Reason  string `json:"reason,omitempty"` // why the plan is empty or limited as a whole
	Picks   []Pick `json:"picks"`
	Skips   []Skip `json:"skips,omitempty"`
	NeedBuy int    `json:"needBuy"` // scrolls to buy first (already within the floor)
}

// Stash returns the picks that lie in the stash (they must be moved to the pack first).
func (p Plan) Stash() []Pick { return p.where(PlaceStash) }

func (p Plan) where(place string) []Pick {
	var out []Pick
	for _, k := range p.Picks {
		if k.Item.Place == place {
			out = append(out, k)
		}
	}
	return out
}

// Make plans one visit. Unidentified items only: the caller lists what the game shows as unidentified. The order is
// priority, then value, then id, so the same input gives the same plan.
func Make(f *Filter, tu env.IdentifyTuning, h Hero, items []Item) Plan {
	switch tu.Policy {
	case env.IdentifyNever:
		return Plan{Reason: WhyNever}
	case env.IdentifyManual:
		return Plan{Reason: WhyManual}
	case env.IdentifyCainFree:
		if !h.CainOpen {
			return Plan{Reason: WhyNoCain}
		}
	}
	type cand struct {
		it   Item
		rule string
		prio int
	}
	var cs []cand
	var p Plan
	for _, it := range items {
		rule, prio, ok := f.Match(it)
		if !ok {
			p.Skips = append(p.Skips, Skip{it.ID, WhyFilter})
			continue
		}
		cs = append(cs, cand{it, rule, prio})
	}
	sort.SliceStable(cs, func(i, j int) bool {
		a, b := cs[i], cs[j]
		if a.prio != b.prio {
			return a.prio > b.prio
		}
		if a.it.Value != b.it.Value {
			return a.it.Value > b.it.Value
		}
		return a.it.ID < b.it.ID
	})
	if h.CainOpen {
		for _, c := range cs {
			if len(p.Picks) >= tu.TripCap {
				p.Skips = append(p.Skips, Skip{c.it.ID, WhyCap})
				continue
			}
			p.Picks = append(p.Picks, Pick{c.it, ViaCain, c.rule})
		}
		return p
	}
	// Scrolls (policy 2 only: policy 1 returned above without Cain). Held charges first, then purchases within the floor.
	have := max(h.Charges, 0)
	canBuy := 0
	if h.ScrollPrice > 0 && h.Gold-tu.GoldFloor >= h.ScrollPrice {
		canBuy = (h.Gold - tu.GoldFloor) / h.ScrollPrice
	}
	left := tu.MaxScrolls
	for _, c := range cs {
		switch {
		case c.it.Value < tu.MinValue:
			p.Skips = append(p.Skips, Skip{c.it.ID, WhyValue})
		case left <= 0:
			p.Skips = append(p.Skips, Skip{c.it.ID, WhyCap})
		case have > 0:
			have--
			left--
			p.Picks = append(p.Picks, Pick{c.it, ViaScroll, c.rule})
		case canBuy > 0:
			canBuy--
			left--
			p.NeedBuy++
			p.Picks = append(p.Picks, Pick{c.it, ViaScroll, c.rule})
		default:
			p.Skips = append(p.Skips, Skip{c.it.ID, WhyGoldFloor})
		}
	}
	return p
}
