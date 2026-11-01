package d2

// Skill level bonuses. game.exe's FUN_00644180 builds the level a skill casts at (damage, mana cost, duration,
// passive stats) as base points plus the sources below (V read: docs/research/sk-passives-trace.md Q8 on branch
// agent/skpass). Synergies read base points only (flag 0 of FUN_006442a0), so Base and Effective stay separate.
// None of the stat ids or sub indexes below was read live from the real stat vector yet (docs/fixskilllvl.md);
// 107 with the skill id as sub index is V live (starting staff, Fire Bolt).
const (
	StatAllSkills    = 127 // item_allskills, sub 0
	StatClassSkills  = 83  // item_addclassskills, sub = class id (own class only)
	StatSkillTab     = 188 // item_addskill_tab, sub = class*8 + tab - 1
	StatOSkill       = 97  // item_nonclassskill (oskill), sub = skill id
	StatElemSkill    = 126 // item_elemskill, sub = the row's element byte
	StatSingleSkill  = 107 // item_singleskill, sub = skill id
	StateShrineSkill = 134 // states.txt shrine_skill: +2 to every skill
	shrineSkillBonus = 2
	ownOSkillCap     = 3
)

// Element bytes (Skills.txt row +0x1dc) of the elemental skills: Fire Bolt 1, Charged Bolt 2, Ice Bolt 4 are V run in
// the trace. Poison and magic have no known byte and get 0 (U).
const (
	ElemNone  = 0
	ElemFire  = 1
	ElemLight = 2
	ElemCold  = 4
)

// SkillRef is what the bonus rules need to know about one skill from the data tables.
type SkillRef struct {
	Own  bool // the skill belongs to the player's class (Skills.txt charclass)
	Tab  int  // skilldesc SkillPage, 1 to 3, 0 when none
	Elem int  // element byte, ElemNone when not elemental
}

func subKey(id, sub int) uint32 { return uint32(sub)<<16 | uint32(id) }

// SkillBonus sums the +skills that game.exe adds to skill id's base points. Each term:
//   - 127 all skills, shrine +2 (state 134): every skill (V read)
//   - own class only: 83 (sub class), 188 (sub class*8+tab-1, only when the skill has a tab), min(97, 3) (V read)
//   - other class: 97 in full (V read)
//   - 126 by element byte (V read; the Elem mapping is partly U) and 107 for this skill (V read, V live)
func SkillBonus(st Stats, states StateSet, class Class, id int, ref SkillRef) int {
	return skillBonusExcl107(st, states, class, id, ref) + int(st[subKey(StatSingleSkill, id)])
}

func skillBonusExcl107(st Stats, states StateSet, class Class, id int, ref SkillRef) int {
	n := int(st[subKey(StatAllSkills, 0)])
	if states.Has(StateShrineSkill) {
		n += shrineSkillBonus
	}
	oskill := int(st[subKey(StatOSkill, id)])
	if ref.Own {
		n += int(st[subKey(StatClassSkills, int(class))])
		if ref.Tab > 0 {
			n += int(st[subKey(StatSkillTab, int(class)*8+ref.Tab-1)])
		}
		n += min(oskill, ownOSkillCap)
	} else {
		n += oskill
	}
	if ref.Elem != ElemNone {
		n += int(st[subKey(StatElemSkill, ref.Elem)])
	}
	return n
}

// Base is the skill's base points (what synergies and the leveler read), 0 when the skill is not in the list.
func (p *Player) Base(id int) int {
	for _, sk := range p.Skills {
		if sk.ID == id {
			return sk.Level
		}
	}
	return 0
}

// Effective is the level the skill casts at: base points plus SkillBonus (107 from Skill.Bonus). Only as complete as p.Stats: the d2d
// backend has no effective stat vector (its bonus stays 0).
func (p *Player) Effective(id int, ref SkillRef) int {
	for _, sk := range p.Skills {
		if sk.ID == id {
			// Skill.Bonus is stat 107 as the reader (or a test) filled it; the other sources come from the vector.
			return sk.Level + sk.Bonus + skillBonusExcl107(p.Stats, p.States, p.Class, id, ref)
		}
	}
	return SkillBonus(p.Stats, p.States, p.Class, id, ref)
}
