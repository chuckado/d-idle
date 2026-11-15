// Package danger builds the danger profile of survival slice S-A (docs/survival/design-2026-11-14.md section 3):
// per area and difficulty, what the zone's monsters are, how they come (typical, bad and elite packs), what the named
// bosses are, and what a kill pays. It is generated from the install tables at run time (data.DB) plus a small
// evidence-checked override file; no game data is committed. Nothing here decides or changes behaviour.
package danger

import "fmt"

// Schema tags a generated table.
const Schema = "danger.profile/1"

// Difficulties in data.Difficulty order.
var DiffNames = [3]string{"normal", "nightmare", "hell"}

// Tags say how far a number can be trusted: V verified (tables read), I inferred (a model on top), U unverified.
const (
	TagV = "V"
	TagI = "I"
	TagU = "U"
)

// Monster is one spawn of a zone. Damage terms stay raw (Dmg, ToHit, PeriodSec) so the margin can redo the hit chance
// against the real hero; Dps is the convenience value against the reference defence of Params.
type Monster struct {
	Key     string   `json:"key"`
	Share   float64  `json:"share"` // share of the zone's units (I: pool entries weighted by mean group size)
	Life    int      `json:"life"`
	Def     int      `json:"def"`
	ToHit   int      `json:"toHit"`
	Dmg     float64  `json:"dmg"`   // mean damage of one attack, level scaling applied
	Dps     float64  `json:"dps"`   // Dmg x hit chance against the reference hero / period
	Reach   string   `json:"reach"` // melee or ranged
	Speed   int      `json:"speed"` // run velocity from monstats
	Group   [2]int   `json:"group"` // min, max pack members
	Hazards []string `json:"hazards"`
	// Unblockable: a ranged attacker whose missile has no physical damage, so a shield cannot block it (V, unblockable.json).
	Unblockable bool `json:"unblockable,omitempty"`
}

// Encounter is a reference pack: n monsters of which melee and ranged reach the hero.
type Encounter struct {
	N      int      `json:"n"`
	Melee  int      `json:"melee"`
	Ranged int      `json:"ranged"`
	Leader string   `json:"leader,omitempty"`
	Mods   []string `json:"mods,omitempty"`
	Rate   float64  `json:"rate,omitempty"` // elite packs per clear
}

// Encounters are the typical pack (median that reaches the hero together), the bad pack (several packs at once) and the elite.
type Encounters struct {
	Typical Encounter `json:"typical"`
	Bad     Encounter `json:"bad"`
	Elite   Encounter `json:"elite"`
}

// Density is how many units the zone holds.
type Density struct {
	Units     float64 `json:"units"`     // estimate per full clear
	PerScreen float64 `json:"perScreen"` // units on one screen
}

// Pay is what a kill is worth.
type Pay struct {
	XPPerKill float64            `json:"xpPerKill"` // full award (no level-gap penalty), mean over the pool
	XPByClvl  map[string]float64 `json:"xpByClvl"`  // after the level-gap penalty, keys "c1" "c3" ...
}

// Zone is the profile of one area at one difficulty.
type Zone struct {
	Area        int               `json:"area"`
	Name        string            `json:"name"`
	Act         int               `json:"act"`
	Diff        string            `json:"diff"`
	MLvl        int               `json:"mlvl"`
	Density     Density           `json:"density"`
	Monsters    []Monster         `json:"monsters"`
	Encounters  Encounters        `json:"encounters"`
	Uniques     float64           `json:"uniquesPerClear"`
	Bosses      []string          `json:"bosses,omitempty"` // keys into Table.Bosses
	Immune      []string          `json:"immune,omitempty"`
	EliteImmune []string          `json:"eliteImmune,omitempty"`
	Pay         Pay               `json:"pay"`
	Source      map[string]string `json:"source"`
}

// Boss is a quest boss or superunique with its mods.
type Boss struct {
	Key      string   `json:"key"` // bosses.json key
	Name     string   `json:"name"`
	Class    string   `json:"class"` // monstats key
	Area     int      `json:"area"`
	Diff     string   `json:"diff"`
	Kind     string   `json:"kind"` // act or superunique
	Level    int      `json:"level"`
	Life     int      `json:"life"`
	Def      int      `json:"def"`
	Dmg      float64  `json:"dmg"`
	ToHit    int      `json:"toHit"`
	XP       int      `json:"xp"`
	Mods     []string `json:"mods,omitempty"`
	Minions  int      `json:"minions"`
	Immune   []string `json:"immune,omitempty"`
	Evidence string   `json:"evidence"`
	Source   string   `json:"source"`
}

// Table is the generated result.
type Table struct {
	Schema    string   `json:"schema"`
	Params    Params   `json:"params"`
	Zones     []*Zone  `json:"zones"`
	Bosses    []*Boss  `json:"bosses"`
	Overrides []string `json:"overrides,omitempty"` // ids of the rows that were applied
}

// Find returns the zone for an area id and difficulty name; an unknown pair is an error naming it.
func (t *Table) Find(area int, diff string) (*Zone, error) {
	for _, z := range t.Zones {
		if z.Area == area && z.Diff == diff {
			return z, nil
		}
	}
	return nil, fmt.Errorf("danger: no profile for area %d at %q", area, diff)
}

// FindBoss returns a boss row by bosses.json key and difficulty.
func (t *Table) FindBoss(key, diff string) (*Boss, error) {
	for _, b := range t.Bosses {
		if b.Key == key && b.Diff == diff {
			return b, nil
		}
	}
	return nil, fmt.Errorf("danger: no boss %q at %q", key, diff)
}
