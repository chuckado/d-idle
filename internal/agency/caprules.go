package agency

import (
	"strings"

	"d2lod-bot/internal/buildfile"
)

// BuildRules adapts the capability table to buildfile.Rules, so a build file for a class the table covers validates and
// expands without the game tables (a row without requirements is a skill the table does not know the rules of: it is
// reported as unknown by Validate, never guessed). After Overlay the rules are the game's own.
func (c *Caps) BuildRules() buildfile.Rules {
	idx := map[string]string{}
	for n := range c.Skills {
		idx[strings.ToLower(n)] = n
	}
	return capRules{c, idx}
}

type capRules struct {
	c   *Caps
	idx map[string]string
}

func (r capRules) Skill(name string) (buildfile.SkillInfo, bool) {
	n, ok := r.idx[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return buildfile.SkillInfo{}, false
	}
	row := r.c.Skills[n]
	if row.ReqLevel == 0 { // a real required level is at least 1: 0 means the table has no requirements for this row
		return buildfile.SkillInfo{}, false
	}
	return buildfile.SkillInfo{Name: n, Class: row.Class, ReqLevel: row.ReqLevel, MaxLevel: maxSkillLevel, ReqSkills: row.ReqSkills}, true
}

func (r capRules) StartStats(class string) ([4]int, bool) {
	s, ok := r.c.ClassStart[strings.ToLower(class)]
	return s, ok
}
