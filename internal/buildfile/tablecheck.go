package buildfile

import (
	"fmt"
	"sort"

	"d2lod-bot/internal/leveler"
)

// Problem codes of CheckTable (a per-level plan table, not a build file).
const (
	PTableCap    = "table_over_cap"       // a skill gets more hard points than its maximum
	PTableLevel  = "table_level_early"    // a skill is first taken below its required level
	PTablePrereq = "table_prereq_missing" // a skill is taken before every prerequisite owns a point
	PTableSkill  = "table_skill_unknown"  // the rules do not know the skill (never guessed)
)

// CheckTable verifies a shipped plan table row by row against the game rules: no skill above its cap, no first point
// below the required level, no point before each prerequisite owns one (earlier row, or earlier in the same row).
// With a respec the table's rows describe the pre-respec path up to the respec level and the post-respec build after
// it (rows 2..25 are refunded, docs/points.md); the pool is checked as a set at the respec level, and the rows after
// it start from the pool. U: the rule "clvl >= reqlevel for the first point" is what Expand uses; the kolbot rule
// asks one more clvl per extra point and is stricter (docs/agency.md).
func CheckTable(lb *leveler.Build, rules Rules) []Problem {
	var ps []Problem
	add := func(code, where, f string, a ...any) { ps = append(ps, Problem{code, where, fmt.Sprintf(f, a...)}) }
	rows := append([]leveler.Row(nil), lb.Rows...)
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Clvl < rows[j].Clvl })
	cum := map[string]int{}
	spend := func(where string, clvl int, s leveler.SkillPts) {
		in, ok := rules.Skill(s.Skill)
		if !ok {
			add(PTableSkill, where, "unknown skill %q", s.Skill)
			return
		}
		if cum[in.Name] == 0 {
			if clvl < in.ReqLevel {
				add(PTableLevel, where, "%q taken at clvl %d, required level %d", in.Name, clvl, in.ReqLevel)
			}
			for _, p := range in.ReqSkills {
				if pi, ok := rules.Skill(p); ok && cum[pi.Name] == 0 {
					add(PTablePrereq, where, "%q taken before its prerequisite %q has a point", in.Name, pi.Name)
				}
			}
		}
		cum[in.Name] += s.Points
		if cum[in.Name] > in.maxLevel() {
			add(PTableCap, where, "%q reaches %d hard points, cap %d", in.Name, cum[in.Name], in.maxLevel())
		}
	}
	respecDone := lb.Respec == 0
	for _, r := range rows {
		if !respecDone && r.Clvl >= lb.Respec {
			respecDone = true
			cum = map[string]int{}
			w := fmt.Sprintf("respec_pool@%d", lb.Respec)
			for _, s := range lb.RespecPool { // a set: register every pool skill first, then check each against the others
				cum[s.Skill] += s.Points
			}
			held := cum
			cum = map[string]int{}
			for _, s := range lb.RespecPool {
				in, ok := rules.Skill(s.Skill)
				if !ok {
					add(PTableSkill, w, "unknown skill %q", s.Skill)
					continue
				}
				if lb.Respec < in.ReqLevel {
					add(PTableLevel, w, "%q needs clvl %d, respec at %d", in.Name, in.ReqLevel, lb.Respec)
				}
				for _, p := range in.ReqSkills {
					if pi, ok := rules.Skill(p); ok && held[pi.Name] == 0 && held[p] == 0 {
						add(PTablePrereq, w, "%q in the pool without its prerequisite %q", in.Name, pi.Name)
					}
				}
				if held[s.Skill] > in.maxLevel() {
					add(PTableCap, w, "%q pool of %d, cap %d", in.Name, held[s.Skill], in.maxLevel())
				}
				cum[in.Name] = held[s.Skill]
			}
		}
		for _, s := range r.Skills {
			spend(fmt.Sprintf("clvl %d", r.Clvl), r.Clvl, s)
		}
	}
	return ps
}
