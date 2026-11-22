package scenario

import (
	"fmt"
	"strings"
)

// classSkillBase is the first skills.txt id of each class (S: internal/d2dsrc/tables.go classSkillBase, Skills.txt: Amazon
// 6, Sorceress 36, Necromancer 66, Paladin 96, Barbarian 126, Druid 221, Assassin 251). The dump's skills.base_levels lists
// the class's 30 skills in ascending id order (U, same source), so skill id = base + index.
var classSkillBase = map[string]int{"amazon": 6, "sorceress": 36, "necromancer": 66, "paladin": 96, "barbarian": 126, "druid": 221, "assassin": 251}

// SkillPointsOf returns the hard points in skill id for class from the dump's base_levels, and false when the class is
// unknown, the dump had no levels, or the id is not one of the class's skills.
func SkillPointsOf(class string, levels []int, id int) (int, bool) {
	base, ok := classSkillBase[strings.ToLower(class)]
	if !ok || levels == nil || id < base || id-base >= len(levels) {
		return 0, false
	}
	return levels[id-base], true
}

func evalSkillPoints(a Assert, st State) AssertResult {
	want := fmt.Sprintf("skill %d >= %d", a.Skill, a.Min)
	if a.Max != nil {
		want += fmt.Sprintf(" and <= %d", *a.Max)
	}
	if a.AtLevel > 0 {
		want += fmt.Sprintf(" at clvl %d", a.AtLevel)
	}
	if st.Level == nil && a.AtLevel > 0 {
		return unobs(want, "level")
	}
	if a.AtLevel > 0 && *st.Level != a.AtLevel {
		return unobs(want, fmt.Sprintf("a state at clvl %d (observed clvl %d)", a.AtLevel, *st.Level))
	}
	v, ok := SkillPointsOf(st.Class, st.SkillLevels, a.Skill)
	if !ok {
		return unobs(want, fmt.Sprintf("skill %d of class %q", a.Skill, st.Class))
	}
	return AssertResult{OK: v >= a.Min && (a.Max == nil || v <= *a.Max), Want: want, Got: fmt.Sprintf("skill %d has %d points", a.Skill, v)}
}
