package combat

import "d2lod-bot/internal/d2"

// MinionRow is one summon as the engine read reports it (fork `debug minions`; the d2dctl row, reduced to what the census
// needs). Skill is the Skills.txt id that made it.
type MinionRow struct {
	Skill int
	Name  string
}

// census is the minion count the Necromancer rules use. With the combat.minionRead knob on and a reader wired it comes from
// the engine read, mapped to kinds by the profile's own summon and golem names (data, no skill literal); a failed read, an
// unknown skill or the knob off falls back to the snapshot census.
func (f *Fighter) census(snap *d2.Snapshot) Minions {
	fallback := CountMinions(snap, f.cat)
	if f.e.Tune.Combat.MinionRead == 0 || f.MinionRead == nil || snap == nil || snap.Player == nil {
		return fallback
	}
	nd := f.profileFor(snap.Player).Necro
	if nd == nil {
		return fallback
	}
	rows, err := f.MinionRead()
	if err != nil {
		return fallback
	}
	kind := map[int]string{}
	for _, g := range nd.Golems {
		if s := f.cat.SkillByName(g); s != nil {
			kind[s.ID] = "golem"
		}
	}
	for _, r := range nd.Summons {
		if s := f.cat.SkillByName(r.Name); s != nil {
			kind[s.ID] = r.Kind
		}
	}
	var m Minions
	for _, r := range rows {
		switch kind[r.Skill] {
		case "golem":
			m.Golems++
		case "skeleton":
			m.Skeletons++
		case "mage":
			m.Mages++
		default:
			m.Other++
		}
	}
	return m
}
