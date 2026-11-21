// Package supply is the town planner's pure core (docs/design/fable-review-2026-11-21/town-planner.md): one supply
// counter, one supply need per excursion and a Solve that turns the two into a town plan. It reads no snapshot, clock or
// input, so a guard, the viewer and a unit test get the same answer, and it names no class or skill: the build's needs
// arrive as data (Inputs), the numbers as profile rows (profiles.json).
package supply

// Kind is a potion family.
type Kind uint8

const (
	HP Kind = iota
	MP
	RV
	kinds
)

// Potion is one carried potion as the caller's item catalogue classifies it.
type Potion struct {
	Kind   Kind
	Tier   int
	InBelt bool // false: in the backpack
}

// Floors is the lowest tier worth counting per kind (0 = every tier counts, which is what the old counters did).
type Floors struct{ HP, MP int }

// Count is what the hero carries and owns, by the one counter. Healing and mana count belt PLUS pack (town-planner
// section 3.3, user answer Q1); a potion under the tier floor is counted in BelowFloor only.
type Count struct {
	Belt, Pack, BelowFloor [kinds]int
	BeltTier               [kinds]int // best tier in the belt per kind
	Capacity               [kinds]int // belt capacity per kind, set by the caller from the belt plan (0 = unknown)

	PortalCharges int  // tome charges plus loose scrolls, set by the caller (a tome read needs the game)
	HasTome       bool //
	Purse, Stash  int
	StashKnown    bool // false: the stash gold could not be read and counts 0 (Q4)
	FreeCells     int
}

// CountOf counts potions. Gold, scrolls and capacity are filled by the caller (they need reads this package must not do).
func CountOf(ps []Potion, f Floors) Count {
	var c Count
	for _, p := range ps {
		if p.Kind >= kinds {
			continue
		}
		if p.Tier > 0 && p.Tier < floorOf(f, p.Kind) {
			c.BelowFloor[p.Kind]++
			continue
		}
		if p.InBelt {
			c.Belt[p.Kind]++
			c.BeltTier[p.Kind] = max(c.BeltTier[p.Kind], p.Tier)
		} else {
			c.Pack[p.Kind]++
		}
	}
	return c
}

func floorOf(f Floors, k Kind) int {
	switch k {
	case HP:
		return f.HP
	case MP:
		return f.MP
	}
	return 0
}

// Healing is the healing and rejuvenation potions in belt and pack: the set the route guard, the kit gate and the
// chore planner already counted, so a guard and the shop finally agree.
func (c Count) Healing() int { return c.Belt[HP] + c.Belt[RV] + c.Pack[HP] + c.Pack[RV] }

// Mana is the mana potions in belt and pack (the route guard counted the belt only: rule conflict I8).
func (c Count) Mana() int { return c.Belt[MP] + c.Pack[MP] }

// Wealth is purse plus stash, the number "can the hero afford X" reads (Q4). An unreadable stash counts 0.
func (c Count) Wealth() int {
	if !c.StashKnown {
		return c.Purse
	}
	return c.Purse + c.Stash
}

// Have is the carried count of a planned line.
func (c Count) Have(l Line) int {
	switch l {
	case Healing:
		return c.Healing()
	case Mana:
		return c.Mana()
	case Rejuvenation:
		return c.Belt[RV] + c.Pack[RV]
	case Portal:
		return c.PortalCharges
	}
	return 0
}
