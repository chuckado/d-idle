package agency

import (
	"fmt"
	"sort"
)

// Problem is one rule an allocation breaks.
type Problem struct{ Code, Skill, Text string }

const maxSkillLevel = 20 // hard points per skill

// Validate checks an allocation against skills.txt style rules: required level, prerequisites, the 20 point cap and
// the point budget (clvl - 1 plus quest points). Rules rows are the caller's (no game tables on every box).
func Validate(a Allocation, r Rules) []Problem {
	var out []Problem
	add := func(code, skill, f string, args ...any) {
		out = append(out, Problem{code, skill, fmt.Sprintf(f, args...)})
	}
	names := make([]string, 0, len(a.Skills))
	for n, p := range a.Skills {
		if p > 0 {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, n := range names {
		p := a.Skills[n]
		row, ok := r.Skills[n]
		if !ok {
			add("no_rules", n, "no skills.txt row for %q in the rules supplied", n)
			continue
		}
		if p > maxSkillLevel {
			add("max_level", n, "%s has %d points, the cap is %d", n, p, maxSkillLevel)
		}
		if a.Clvl < row.ReqLevel {
			add("req_level", n, "%s needs clvl %d, the character is %d", n, row.ReqLevel, a.Clvl)
		}
		for _, q := range row.ReqSkills {
			if q != "" && a.Skills[q] <= 0 {
				add("prereq", n, "%s needs a point in %s", n, q)
			}
		}
	}
	if a.QuestSkillPts < 0 || (r.MaxQuest > 0 && a.QuestSkillPts > r.MaxQuest) {
		add("quest_points", "", "quest skill points %d outside 0..%d", a.QuestSkillPts, r.MaxQuest)
	}
	if budget := a.Clvl - 1 + a.QuestSkillPts; a.spent() > budget {
		add("budget", "", "%d points spent, only %d available at clvl %d with %d quest points", a.spent(), budget, a.Clvl, a.QuestSkillPts)
	}
	return out
}

// CanAdd says whether one more hard point in skill is legal now, and if not, why.
func CanAdd(a Allocation, r Rules, skill string) (bool, string) {
	row, ok := r.Skills[skill]
	if !ok {
		return false, fmt.Sprintf("no skills.txt row for %q in the rules supplied", skill)
	}
	switch {
	case a.Skills[skill] >= maxSkillLevel:
		return false, fmt.Sprintf("%s is at the cap of %d points", skill, maxSkillLevel)
	case a.spent() >= a.Clvl-1+a.QuestSkillPts:
		return false, "no unspent skill point"
	case a.Clvl < row.ReqLevel:
		return false, fmt.Sprintf("%s needs clvl %d", skill, row.ReqLevel)
	}
	for _, q := range row.ReqSkills {
		if q != "" && a.Skills[q] <= 0 {
			return false, fmt.Sprintf("%s needs a point in %s first", skill, q)
		}
	}
	return true, ""
}

// Delta is what one more point changes.
type Delta struct {
	DPS          float64 // change in the main attack's estimated dps
	MainChanged  bool
	NewWarnings  []Warning
	GoneWarnings []Warning
}

// Preview derives the plan before and after one more point in skill (legality is CanAdd's job).
func Preview(a Allocation, env Env, skill string) (before, after Plan, err error) {
	if before, err = Derive(a, env); err != nil {
		return
	}
	b := a.clone()
	if b.Skills == nil {
		b.Skills = map[string]int{}
	}
	b.Skills[skill]++
	after, err = Derive(b, env)
	return
}

// Diff compares two plans.
func Diff(before, after Plan) Delta {
	var d Delta
	dps := func(p Plan) float64 {
		if p.Attack != nil {
			return p.Attack.DPS
		}
		return 0
	}
	d.DPS = dps(after) - dps(before)
	d.MainChanged = (before.Main == nil) != (after.Main == nil) || (before.Main != nil && before.Main.Skill != after.Main.Skill)
	for _, w := range after.Warnings {
		if !before.Has(w.Code) {
			d.NewWarnings = append(d.NewWarnings, w)
		}
	}
	for _, w := range before.Warnings {
		if !after.Has(w.Code) {
			d.GoneWarnings = append(d.GoneWarnings, w)
		}
	}
	return d
}
