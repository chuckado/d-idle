package points

import (
	"fmt"
	"strings"

	"d2lod-bot/internal/plandata"
)

// The horizon handoff (docs/autoalloc.md): a locked table such as normal30 ends at level 30. The catalogue lists the
// later tables of the same build with `fromLevel`; the allocator follows the next SUPPORTED one once the hero reaches it
// (or the current table has no step left), never an experimental table, and says so in the log, the state file and the
// Points page.

// followed is the table the allocator follows for the hero: the stored handoff target, else the locked table itself.
func followed(s Settings) string {
	if s.Plan != "" {
		return s.Plan
	}
	return s.Build
}

func tierOf(have []BuildInfo, table string) string {
	for _, b := range have {
		if b.Name == table {
			return b.Tier
		}
	}
	return ""
}

func fromLevelOf(cb *plandata.Build, table string) int {
	for _, v := range cb.Variants {
		if v.Table == table {
			return v.FromLevel
		}
	}
	return 0
}

// nextPlan returns the supported later table of the build to follow now ("" when none). A later table is one with a
// higher fromLevel than the table in use. It is due when the hero's level reached its fromLevel, or when exhausted
// (the current table has no step left but points are unspent), then the nearest later one is taken early.
func nextPlan(cb *plandata.Build, current string, level int, have []BuildInfo, exhausted bool) string {
	cur := fromLevelOf(cb, current)
	due, near := "", ""
	for _, v := range cb.Variants {
		if v.FromLevel <= 0 || v.FromLevel <= cur || tierOf(have, v.Table) != BuildSupported {
			continue
		}
		if v.FromLevel <= level && (due == "" || v.FromLevel > fromLevelOf(cb, due)) {
			due = v.Table
		}
		if near == "" || v.FromLevel < fromLevelOf(cb, near) {
			near = v.Table
		}
	}
	if due == "" && exhausted {
		return near
	}
	return due
}

func catalogBuild(table string) *plandata.Build {
	cat, err := plandata.Catalog()
	if err != nil {
		return nil
	}
	cb, err := cat.BuildForTable(table)
	if err != nil {
		return nil
	}
	return cb
}

func shortTable(cb *plandata.Build, table string) string {
	if cb != nil {
		if r := strings.TrimPrefix(table, cb.ID+"_"); r != table {
			return r
		}
	}
	return table
}

// PlanStatus says whether a catalogue build has a plan the allocator can follow, and if not, why: the first table a new
// hero starts on must be supported. The reason names the unsupported skills the capability table lists.
func PlanStatus(cb *plandata.Build, have []BuildInfo) (ok bool, reason string) {
	t := tierOf(have, cb.DefaultVariant)
	if t == BuildSupported {
		return true, ""
	}
	if t == "" {
		return false, fmt.Sprintf("the plan table %s is not loaded by the bot", cb.DefaultVariant)
	}
	var skills []string
	for _, b := range have {
		if b.Name == cb.DefaultVariant {
			for _, n := range b.Needs {
				if n.Tier == "unsupported" {
					skills = append(skills, n.Skill)
				}
			}
		}
	}
	if len(skills) == 0 {
		return false, "its default plan is experimental"
	}
	return false, "its default plan is experimental: the fighter has no rule for " + strings.Join(skills, ", ")
}

// planLine is the sentence the Points page shows about which plan the hero follows.
func planLine(s Settings, level int, have []BuildInfo) string {
	if !s.Locked || s.Build == "" {
		return ""
	}
	cb := catalogBuild(s.Build)
	if cb == nil {
		return ""
	}
	if s.Plan != "" {
		return fmt.Sprintf("Plan: %s, now following the full plan from level %d.", cb.Name, fromLevelOf(cb, s.Plan))
	}
	later := ""
	for _, v := range cb.Variants {
		if v.FromLevel > fromLevelOf(cb, s.Build) && (later == "" || v.FromLevel < fromLevelOf(cb, later)) {
			later = v.Table
		}
	}
	switch {
	case later == "" && level >= 30:
		return fmt.Sprintf("Plan: %s. This build has no plan past level 30, so new points are not spent by the bot.", cb.Name)
	case later == "":
		return fmt.Sprintf("Plan: %s. This build has no plan past its first table; at its end new points stay unspent.", cb.Name)
	case tierOf(have, later) != BuildSupported:
		return fmt.Sprintf("Plan: %s. The plan from level %d is not supported yet, so points stop at the end of this table.", cb.Name, fromLevelOf(cb, later))
	}
	return fmt.Sprintf("Plan: %s. From level %d the bot moves on to the full plan by itself.", cb.Name, fromLevelOf(cb, later))
}
