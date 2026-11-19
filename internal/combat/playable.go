package combat

// Playable says whether the fighter has a rule for a skill today: some profile names it (attack band, aura ladder,
// buffs or layout) and, for an attack skill, no profile that names it lists it in Avoid on its own. It is the interim
// capability test of the points sheet (docs/points.md) until the capability table is generated from Proving Ground
// results (design review 2026-11-19, 04 section 3.1). Names are skills.txt skill names.
func Playable(skill string) bool {
	db := loadClassData()
	if !db.names[skill] {
		return false
	}
	named, avoided := 0, 0
	for _, p := range db.profiles {
		if profileNames(p, skill) {
			named++
			for _, a := range p.Avoid {
				if a == skill {
					avoided++
					break
				}
			}
		}
	}
	return named > avoided
}

func profileNames(p Profile, skill string) bool {
	for _, b := range p.Bands {
		for _, n := range b.Attack {
			if n == skill {
				return true
			}
		}
		for _, a := range b.Aura {
			if a.Name == skill {
				return true
			}
		}
	}
	for _, a := range p.Aura {
		if a.Name == skill {
			return true
		}
	}
	for _, g := range p.Buffs {
		for _, n := range g {
			if n == skill {
				return true
			}
		}
	}
	_, inLayout := p.Layout[skill]
	return inLayout
}
