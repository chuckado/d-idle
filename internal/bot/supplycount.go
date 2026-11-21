package bot

import (
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/supply"
)

// supplyCount is the bot's side of the one supply counter (internal/supply, town-planner section 3.3): belt and pack
// potions by kind. known is false before the catalogue and the inventory can be read. No tier floor here (the old
// counters had none); the planner stages pass one.
func (r *Runner) supplyCount(s *d2.Snapshot) (c supply.Count, known bool) {
	if r.d.Catalog == nil || s == nil || s.Inventory == nil {
		return c, false
	}
	var ps []supply.Potion
	for i := range s.Inventory.Items {
		u := &s.Inventory.Items[i]
		if u.Item == nil || u.Item.Where != d2.ContBelt && u.Item.Where != d2.ContInventory {
			continue
		}
		m, ok := r.d.Catalog.Item(u.Class)
		if !ok {
			continue
		}
		var k supply.Kind
		switch m.Potion {
		case PotionHealing:
			k = supply.HP
		case PotionMana:
			k = supply.MP
		case PotionRejuv:
			k = supply.RV
		default:
			continue
		}
		ps = append(ps, supply.Potion{Kind: k, Tier: m.Tier, InBelt: u.Item.Where == d2.ContBelt})
	}
	return supply.CountOf(ps, supply.Floors{}), true
}
