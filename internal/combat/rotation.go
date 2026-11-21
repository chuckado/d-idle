package combat

import "d2lod-bot/internal/delivery"

// Rotation is the profile written as a delivery.Rotation for a character level: the attack names of the active band
// (Avoid applied), the buff groups and the aura ladder. It is the adapter of card dlv-rotation: the Fighter reads
// slots, and the result equals what AttackNames, Buffs and AuraLadder said (golden test in rotation_test.go).
func (p *Profile) Rotation(clvl int, known func(string) bool) delivery.Rotation {
	var r delivery.Rotation
	for _, n := range p.AttackNames(clvl, known) {
		r.Slots = append(r.Slots, delivery.Slot{Skill: n, Role: delivery.RoleAttack})
	}
	for g, names := range p.Buffs {
		for _, n := range names {
			r.Slots = append(r.Slots, delivery.Slot{Skill: n, Role: delivery.RoleUpkeep, Group: g})
		}
	}
	for _, a := range p.AuraLadder(clvl, known) {
		r.Slots = append(r.Slots, delivery.Slot{Skill: a.Name, Role: delivery.RoleAura,
			When: delivery.When{HeroLifeBelowPct: a.RestLifePct, CrowdKnob: a.NeedsCrowd, TargetNotImmuneTo: a.Elem}})
	}
	return r
}

// auraPrefs turns the aura slots of a rotation back into the ladder entries the aura code reads.
func auraPrefs(r delivery.Rotation) []AuraPref {
	var out []AuraPref
	for _, s := range r.Of(delivery.RoleAura) {
		out = append(out, AuraPref{Name: s.Skill, Elem: s.When.TargetNotImmuneTo, RestLifePct: s.When.HeroLifeBelowPct, NeedsCrowd: s.When.CrowdKnob})
	}
	return out
}
