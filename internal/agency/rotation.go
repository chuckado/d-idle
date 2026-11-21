package agency

import "d2lod-bot/internal/delivery"

// Rotation lays the plan out as a delivery.Rotation (card dlv-rotation): the plan's attack, then the others the planner ranked as usable
// (best first, unsupported ones left out), then the buffs, the summons and the movement skills as upkeep slots (one
// group each, in plan order) and the aura. It is the plan side of the shadow comparison; the Fighter does not read it.
func (p Plan) Rotation() delivery.Rotation {
	var r delivery.Rotation
	if p.Attack != nil {
		r.Slots = append(r.Slots, delivery.Slot{Skill: p.Attack.Skill, Role: delivery.RoleAttack})
	}
	for _, c := range p.Candidates {
		if c.Tier != Unsupported && (p.Attack == nil || c.Skill != p.Attack.Skill) {
			r.Slots = append(r.Slots, delivery.Slot{Skill: c.Skill, Role: delivery.RoleAttack})
		}
	}
	g := 0
	for _, list := range [][]Choice{p.Buffs, p.Summons, p.Movement} {
		for _, c := range list {
			r.Slots = append(r.Slots, delivery.Slot{Skill: c.Skill, Role: delivery.RoleUpkeep, Group: g})
			g++
		}
	}
	if p.Aura != nil {
		r.Slots = append(r.Slots, delivery.Slot{Skill: p.Aura.Skill, Role: delivery.RoleAura})
	}
	return r
}

// ShadowDiff compares the plan's picks with the combat profile's, slot by slot: the head of the attack role and the
// head of the aura role (the profile's picks are Prior and PriorAura). Nil when they agree.
func ShadowDiff(p Plan) []string {
	plan := p.Rotation()
	var prior delivery.Rotation
	if p.Prior != "" {
		prior.Slots = append(prior.Slots, delivery.Slot{Skill: p.Prior, Role: delivery.RoleAttack})
	}
	if p.PriorAura != "" {
		prior.Slots = append(prior.Slots, delivery.Slot{Skill: p.PriorAura, Role: delivery.RoleAura})
	}
	head := func(r delivery.Rotation) delivery.Rotation {
		var out delivery.Rotation
		seen := map[string]bool{}
		for _, s := range r.Slots {
			if !seen[s.Role] && (s.Role == delivery.RoleAttack || s.Role == delivery.RoleAura) {
				seen[s.Role] = true
				out.Slots = append(out.Slots, s)
			}
		}
		return out
	}
	return delivery.Diff(head(plan), head(prior), delivery.RoleAttack, delivery.RoleAura)
}
