// Package agency derives how the hero would play from the skills the player actually bought, and says so when it
// cannot (ADR 0020, path D; docs/agency.md). It reads internal/power (damage estimate, archetype rule, survival
// margin), internal/combat (the profile bands, the bot's current choice) and the capability table in
// data/capabilities. P1 runs it in SHADOW: it never changes what the hero does.
package agency

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	capdata "d2lod-bot/data/capabilities"
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/power"
)

const capsSchema = "didle.cap/1"

// Tiers: what the combat engine can do with a skill today (not a judgement of the build).
const (
	Supported    = "supported"    // engine code and a live-verified loop
	Experimental = "experimental" // code exists, unverified
	Unsupported  = "unsupported"  // no engine support
)

// Roles and deliveries of a skill.
const (
	RoleAttack   = "damage"
	RoleAura     = "aura"
	RoleBuff     = "buff"
	RoleSummon   = "summon"
	RoleCurse    = "curse" // Necromancer, Assassin: a buff of the monsters' side
	RoleMovement = "movement"
	RolePassive  = "passive"
	RoleUtility  = "utility"
)

// ManaSrc says where the mana cost per skill level comes from.
type ManaSrc struct {
	Source string `json:"source"` // skilldamage.json:mana | skills.txt:... | none
}

// Cap is one row of the capability table.
type Cap struct {
	Class    string `json:"class"`
	Role     string `json:"kind"`     // damage | aura | buff | curse | summon | utility | movement | passive
	Delivery string `json:"delivery"` // melee | projectile | nova | ground | trap | pet | self
	Tier     string `json:"tier"`
	Element  string `json:"element"` // fire cold ltng pois mag phys, "" none or not an attack
	// Note is why the tier is what it is; required for every row that is not supported.
	Note string `json:"reason"`
	// ReqLevel and ReqSkills are the skills.txt requirements (0 and nil: not in the table, read the rules).
	ReqLevel  int      `json:"reqLevel,omitempty"`
	ReqSkills []string `json:"reqSkills,omitempty"`
	Mana      ManaSrc  `json:"mana"`
	// Synergies names the skills that add damage to this one; nil means NOT KNOWN (never "none"), [] means none.
	Synergies []string `json:"synergies"`
	// Derived marks a row that was not in the table and came from power's skillKinds.
	Derived bool `json:"-"`
	// FromTables marks a row whose requirements and synergies were replaced by skills.txt (Overlay).
	FromTables bool `json:"-"`
}

// ManaCfg is the simple sustain budget (all U).
type ManaCfg struct {
	PerEnergy       map[string]float64 `json:"perEnergy"`
	PerLevel        map[string]float64 `json:"perLevel"`
	Base            float64            `json:"base"`
	RegenFracPerSec float64            `json:"regenFracPerSec"`
	FightSeconds    float64            `json:"fightSeconds"`
}

// Caps is the loaded capability table.
type Caps struct {
	Schema string         `json:"schema"`
	Skills map[string]Cap `json:"skills"`
	Mana   ManaCfg        `json:"mana"`
	// Planner holds the planner's own constants (all U until calibrated on tracks known to be good).
	Planner PlannerCfg `json:"planner"`
}

// PlannerCfg are the planner's constants.
type PlannerCfg struct {
	// PriorMargin: the estimate must beat the combat profile's pick by this factor to overrule it; 0 = the estimate alone.
	PriorMargin float64 `json:"priorMargin"`
}

// LoadCaps parses the embedded table and checks every row.
func LoadCaps() (*Caps, error) {
	var c Caps
	if err := json.Unmarshal(capdata.JSON, &c); err != nil {
		return nil, fmt.Errorf("agency: capabilities.json: %w", err)
	}
	if c.Schema != capsSchema {
		return nil, fmt.Errorf("agency: capabilities.json schema %q, want %q", c.Schema, capsSchema)
	}
	roles := map[string]bool{RoleAttack: true, RoleAura: true, RoleBuff: true, RoleSummon: true, RoleCurse: true, RoleMovement: true, RolePassive: true, RoleUtility: true}
	dels := map[string]bool{"melee": true, "projectile": true, "nova": true, "ground": true, "trap": true, "pet": true, "self": true}
	tiers := map[string]bool{Supported: true, Experimental: true, Unsupported: true}
	for n, s := range c.Skills {
		if !roles[s.Role] || !dels[s.Delivery] || !tiers[s.Tier] || !validClass(s.Class) {
			return nil, fmt.Errorf("agency: capabilities.json: skill %q has an invalid kind, delivery, tier or class", n)
		}
		if s.Tier != Supported && strings.TrimSpace(s.Note) == "" {
			return nil, fmt.Errorf("agency: capabilities.json: skill %q is %s and gives no reason", n, s.Tier)
		}
	}
	return &c, nil
}

// Classes is the class list of the game (lower case, as the plan tables spell it).
var Classes = []string{"amazon", "sorceress", "necromancer", "paladin", "barbarian", "druid", "assassin"}

func validClass(c string) bool {
	for _, k := range Classes {
		if k == c {
			return true
		}
	}
	return false
}

// Lookup finds the row of a skill. A skill missing from the table gets a row derived from power's skillKinds,
// marked Experimental and Derived; a skill in neither is not found (the caller treats it as Unsupported).
func (c *Caps) Lookup(name string, kinds map[string]string) (Cap, bool) {
	if s, ok := c.Skills[name]; ok {
		return s, true
	}
	d := Cap{Tier: Experimental, Derived: true, Note: "not in capabilities.json; derived from power skillKinds"}
	switch kinds[name] {
	case "melee":
		d.Role, d.Delivery = RoleAttack, "melee"
	case "missile":
		d.Role, d.Delivery = RoleAttack, "projectile"
	case "spell":
		d.Role, d.Delivery = RoleAttack, "ground"
	case "summon":
		d.Role, d.Delivery = RoleSummon, "pet"
	default:
		return Cap{}, false
	}
	return d, true
}

// Counts returns tier counts per class (rows of the table only).
func (c *Caps) Counts() map[string]map[string]int {
	out := map[string]map[string]int{}
	for _, s := range c.Skills {
		if out[s.Class] == nil {
			out[s.Class] = map[string]int{}
		}
		out[s.Class][s.Tier]++
	}
	return out
}

// Rules are the skills.txt rows the validator needs (reqlevel, prerequisites, max level). No game tables are on
// every box, so the caller supplies them.
type Rules struct {
	Skills map[string]data.Skill
	// MaxQuest is the most quest skill points a character can have (U: 4 per difficulty, 12 in all).
	MaxQuest int
}

// known reports whether a skill name is known to any source; an unknown name is an error for the caller.
func known(name string, c *Caps, p *power.Profiles, sd *power.SkillData, r *Rules) bool {
	if _, ok := c.Skills[name]; ok {
		return true
	}
	if _, ok := p.SkillKinds[name]; ok {
		return true
	}
	if sd != nil {
		if _, ok := sd.Skills[name]; ok {
			return true
		}
	}
	if r != nil {
		if _, ok := r.Skills[name]; ok {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	k := make([]string, 0, len(m))
	for n := range m {
		k = append(k, n)
	}
	sort.Strings(k)
	return k
}

func lower(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// Overlay replaces the requirements and the synergy lists of every table row that has a skills.txt row by what the
// tables say (the table's own values are typed; the game's are the source). A skill with an expression-free damage
// calc gets [] (known: none), not nil. It returns the names of rows it could not find in the tables.
func (c *Caps) Overlay(skills func(name string) *data.Skill) (missing []string) {
	for _, n := range sortedKeys(c.Skills) {
		s := skills(n)
		if s == nil {
			missing = append(missing, n)
			continue
		}
		row := c.Skills[n]
		row.ReqLevel = s.ReqLevel
		row.ReqSkills = nil
		for _, r := range s.ReqSkills {
			if r != "" {
				row.ReqSkills = append(row.ReqSkills, r)
			}
		}
		row.Synergies = append([]string{}, s.Synergies()...)
		row.FromTables = true
		c.Skills[n] = row
	}
	return missing
}

// combatBuild is the combat package's name for the combat line of a plan table.
func combatBuild(s string) combat.Build { return combat.Build(s) }
