package buildfile

import (
	"fmt"
	"strings"
)

// Problem codes.
const (
	PClass        = "class_unknown"
	PSkillUnknown = "skill_unknown"
	PWrongClass   = "skill_wrong_class"
	PTarget       = "target_range"
	PCycle        = "prereq_cycle"
	PUnreachable  = "skill_unreachable"
	PSkillBudget  = "skills_over_budget"
	PStat         = "stat_unknown"
	PStatTarget   = "stat_target"
	PStatBudget   = "stats_over_budget"
	PRest         = "rest_not_last"
	PWhen         = "when_invalid"
	PPhase        = "phase_invalid"
	PRespec       = "respec_invalid"
	PQuest        = "quest_invalid"
	PNoGoals      = "no_goals"
	PStatAfter    = "stat_after_unsupported"
)

// Max extra points a character can get from quests (U: 12 skill points = Den, Radament and Izual in three
// difficulties, matching the 101 skill points of the repo's endgame tables over 89 levels; 5 stat points Lam Esen).
const (
	MaxQuestSkill = 12
	MaxQuestStat  = 5
)

// Problem is one thing wrong with a file. Where names the field ("skills[2]").
type Problem struct {
	Code  string
	Where string
	Msg   string
}

func (p Problem) String() string { return fmt.Sprintf("%s at %s: %s", p.Code, p.Where, p.Msg) }

// Validate checks a parsed build against game data. An empty result means Expand will place every goal.
func Validate(b *Build, rules Rules) []Problem {
	var ps []Problem
	add := func(code, where, f string, a ...any) { ps = append(ps, Problem{code, where, fmt.Sprintf(f, a...)}) }

	start, classOK := rules.StartStats(b.Class)
	if !classOK {
		add(PClass, "class", "unknown class %q (want one of %s)", b.Class, strings.Join(ClassNames(), ", "))
	}
	if len(b.Skills) == 0 && len(b.Phases) == 0 {
		add(PNoGoals, "skills", "a build needs at least one skill goal")
	}

	// quest points
	qs, qt := 0, 0
	for i, q := range b.Quest {
		w := fmt.Sprintf("quest_points[%d]", i)
		if q.Clvl < 2 || q.Clvl > MaxClvl || q.Skill < 0 || q.Stat < 0 {
			add(PQuest, w, "clvl must be 2..%d and points not negative", MaxClvl)
		}
		qs += q.Skill
		qt += q.Stat
	}
	if qs > MaxQuestSkill || qt > MaxQuestStat {
		add(PQuest, "quest_points", "%d skill / %d stat quest points, at most %d / %d exist", qs, qt, MaxQuestSkill, MaxQuestStat)
	}

	// goals
	check := func(g Goal, where string, from int) {
		in, ok := rules.Skill(g.Skill)
		if !ok {
			add(PSkillUnknown, where, "unknown skill %q", g.Skill)
			return
		}
		if classOK && in.Class != b.Class {
			add(PWrongClass, where, "%q belongs to %q, build is %q", in.Name, in.Class, b.Class)
		}
		if g.To < 1 || g.To > in.maxLevel() {
			add(PTarget, where, "%q to %d: want 1..%d hard points", in.Name, g.To, in.maxLevel())
		}
		if g.When != nil {
			if g.When.Clvl < 0 || g.When.Clvl > MaxClvl {
				add(PWhen, where, "when.clvl %d out of range", g.When.Clvl)
			}
			if a := g.When.After; a != nil {
				if ai, ok := rules.Skill(a.Skill); !ok {
					add(PSkillUnknown, where+".when.after", "unknown skill %q", a.Skill)
				} else if a.Level < 1 || a.Level > ai.maxLevel() {
					add(PWhen, where+".when.after", "level %d out of range", a.Level)
				}
			}
		}
		if from > MaxClvl {
			add(PWhen, where, "phase starts after clvl %d", MaxClvl)
		}
	}
	for i, g := range b.Skills {
		check(g, fmt.Sprintf("skills[%d]", i), 0)
	}
	prevUntil := 0
	for i, p := range b.Phases {
		w := fmt.Sprintf("phases[%d]", i)
		if p.FromClvl < 2 || p.FromClvl > MaxClvl || (p.UntilClvl != 0 && p.UntilClvl < p.FromClvl) || p.FromClvl < prevUntil {
			add(PPhase, w, "from_clvl %d until_clvl %d: want 2 <= from <= until, phases in order and not overlapping", p.FromClvl, p.UntilClvl)
		}
		if p.UntilClvl == 0 {
			prevUntil = MaxClvl + 1
		} else {
			prevUntil = p.UntilClvl
		}
		if len(p.Skills) == 0 {
			add(PPhase, w, "phase %q has no skills", p.Name)
		}
		for j, g := range p.Skills {
			check(g, fmt.Sprintf("%s.skills[%d]", w, j), p.FromClvl)
		}
	}
	if b.RespecAt != nil {
		if b.RespecAt.Clvl < 2 || b.RespecAt.Clvl >= MaxClvl {
			add(PRespec, "respec_at", "clvl %d out of range 2..%d", b.RespecAt.Clvl, MaxClvl-1)
		}
		if len(b.Phases) == 0 {
			add(PRespec, "respec_at", "a respec needs phases (what is played before it)")
		}
	}

	// prerequisite cycles in the data, for every skill the file touches
	for _, name := range touched(b) {
		if in, ok := rules.Skill(name); ok {
			if cyc := findCycle(in.Name, rules); cyc != "" {
				add(PCycle, "skills", "prerequisite cycle: %s", cyc)
				break
			}
		}
	}

	// stats
	need := [4]int{}
	restSeen := false
	for i, g := range b.Stats {
		w := fmt.Sprintf("stats[%d]", i)
		si := statIndex(g.Stat)
		if si < 0 {
			add(PStat, w, "unknown stat %q (str, dex, vit, ene)", g.Stat)
			continue
		}
		if restSeen {
			add(PRest, w, "goal after a \"rest\" goal can never get a point")
		}
		if g.To.Rest {
			restSeen = true
		} else {
			if classOK && g.To.N < start[si] {
				add(PStatTarget, w, "%s to %d is below the class start value %d", g.Stat, g.To.N, start[si])
			}
			need[si] = max(need[si], g.To.N-start[si])
		}
		if g.Per < 0 || g.Per > StatPerLevel+MaxQuestStat {
			add(PStatTarget, w, "per %d out of range", g.Per)
		}
		if g.When != nil && (g.When.After != nil || g.When.Clvl < 0 || g.When.Clvl > MaxClvl) {
			add(PStatAfter, w, "stat goals only take when.clvl")
		}
	}
	total := 0
	for _, n := range need {
		total += max(n, 0)
	}
	if budget := StatPerLevel*(MaxClvl-1) + qt; classOK && total > budget {
		add(PStatBudget, "stats", "targets need %d points, a character earns %d by clvl %d (5 per level plus %d quest)", total, budget, MaxClvl, qt)
	}

	if len(ps) > 0 {
		return ps
	}
	// reachability by dry run
	exp, err := Expand(b, rules)
	if err != nil {
		add(PCycle, "skills", "%v", err)
		return ps
	}
	if len(exp.Unplaced) > 0 {
		steps := 0
		if e, err := newEngine(allGoals(b), rules, b.Quest); err == nil {
			steps = len(e.steps)
		}
		if steps > earnedSkill(b.Quest, MaxClvl) {
			add(PSkillBudget, "skills", "%d points (with prerequisites) but only %d exist by clvl %d: first unplaced %q", steps, earnedSkill(b.Quest, MaxClvl), MaxClvl, exp.Unplaced[0])
		} else {
			add(PUnreachable, "skills", "never placeable by clvl %d (skill level, when or after never met): %s", MaxClvl, strings.Join(exp.Unplaced, ", "))
		}
	}
	return ps
}

func allGoals(b *Build) []goalCtx {
	var out []goalCtx
	for _, p := range b.Phases {
		for _, g := range p.Skills {
			out = append(out, goalCtx{Goal: g, min: max(p.FromClvl, whenClvl(g)), until: p.UntilClvl})
		}
	}
	for _, g := range b.Skills {
		out = append(out, goalCtx{Goal: g, min: whenClvl(g)})
	}
	return out
}

func touched(b *Build) []string {
	var out []string
	for _, c := range allGoals(b) {
		out = append(out, c.Skill)
	}
	return out
}

func findCycle(name string, rules Rules) string {
	var path []string
	onPath := map[string]bool{}
	var dfs func(n string) string
	dfs = func(n string) string {
		in, ok := rules.Skill(n)
		if !ok {
			return ""
		}
		if onPath[in.Name] {
			return strings.Join(append(path, in.Name), " -> ")
		}
		onPath[in.Name] = true
		path = append(path, in.Name)
		for _, r := range in.ReqSkills {
			if s := dfs(r); s != "" {
				return s
			}
		}
		path = path[:len(path)-1]
		onPath[in.Name] = false
		return ""
	}
	return dfs(name)
}
