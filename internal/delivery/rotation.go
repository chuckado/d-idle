package delivery

import (
	"fmt"
	"slices"
)

// Roles of a rotation slot. The order inside a role is the preference order (grammar appendix B section 4).
const (
	RoleAttack = "attack" // offered by the attack driver, first usable wins
	RoleUpkeep = "upkeep" // a buff or summon the hero keeps up; slots of one Group are alternatives, one cast per group
	RoleAura   = "aura"   // an aura of the ladder, first wanted and known wins
)

// When is the conjunction of closed predicates of a slot (grammar appendix B section 4). The profile adapter fills the
// three that profiles already carry; the others join when their rule leaves Go.
type When struct {
	// HeroLifeBelowPct > 0 makes the slot a resting one: wanted only while resting below this life percent (Prayer).
	HeroLifeBelowPct int `json:"heroLifeBelowPct,omitempty"`
	// CrowdKnob marks a slot that pays only against a pack; the count is a knob until crowdAtLeast carries it.
	CrowdKnob bool `json:"crowdKnob,omitempty"`
	// TargetNotImmuneTo is an element the target must not be immune to ("" none); today the aura's own element.
	TargetNotImmuneTo string `json:"targetNotImmuneTo,omitempty"`
}

// Slot is one entry of a rotation.
type Slot struct {
	Skill string `json:"skill"`
	Role  string `json:"role"`
	// Group numbers the alternatives of an upkeep slot (the profile's buff groups); 0 for the other roles.
	Group int  `json:"group,omitempty"`
	When  When `json:"when,omitzero"`
}

// Rotation is what the Fighter reads: the ordered list of slots for one hero at one level. It carries no class and
// no skill knowledge; the adapters (combat.Profile.Rotation, agency.Plan.Rotation) write it.
type Rotation struct {
	Slots []Slot `json:"slots"`
}

// Names returns the skill names of a role in slot order.
func (r Rotation) Names(role string) []string {
	var out []string
	for _, s := range r.Slots {
		if s.Role == role {
			out = append(out, s.Skill)
		}
	}
	return out
}

// Groups returns the upkeep slots by group, groups in first-seen order, names in slot order.
func (r Rotation) Groups() [][]string {
	var out [][]string
	idx := map[int]int{}
	for _, s := range r.Slots {
		if s.Role != RoleUpkeep {
			continue
		}
		i, ok := idx[s.Group]
		if !ok {
			i = len(out)
			idx[s.Group] = i
			out = append(out, nil)
		}
		out[i] = append(out[i], s.Skill)
	}
	return out
}

// Slots of one role as a slice, for callers that need the When of each.
func (r Rotation) Of(role string) []Slot {
	var out []Slot
	for _, s := range r.Slots {
		if s.Role == role {
			out = append(out, s)
		}
	}
	return out
}

// Diff compares two rotations role by role and says where they part, one line per difference; nil when equal. It is
// the slot-by-slot comparison of the shadow planner (the plan against the profile).
func Diff(a, b Rotation, roles ...string) []string {
	var out []string
	for _, role := range roles {
		x, y := a.Names(role), b.Names(role)
		if slices.Equal(x, y) {
			continue
		}
		n := min(len(x), len(y))
		i := 0
		for i < n && x[i] == y[i] {
			i++
		}
		out = append(out, fmt.Sprintf("%s: slot %d is %s against %s", role, i, nameAt(x, i), nameAt(y, i)))
	}
	return out
}

func nameAt(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return "(none)"
}
