package scenario

import (
	"fmt"
	"sort"
	"strings"
)

// Run is one executable case: a scenario at one class and one seed.
type Run struct {
	ID    string `json:"id"`
	Class string `json:"class"`
	Seed  int64  `json:"seed"`
}

// Key is the stable name of a run (result file, determinism pairing).
func (r Run) Key() string { return fmt.Sprintf("%s/%s/%d", r.ID, r.Class, r.Seed) }

// Selector picks scenarios. Empty fields match everything.
type Selector struct {
	IDs   []string // exact ids or id prefixes ending in "."
	Tier  string   // fast: tier fast only; full: fast and full; nightly: everything
	Kind  string
	Class string // restrict the matrix to this class
	Seed  *int64 // restrict the matrix to this seed
	// Changed are the files a merge changed: only scenarios whose Touches match one (or declare none) are kept.
	Changed []string
}

// tierRank orders the tiers: a run of tier T includes every lower tier.
var tierRank = map[string]int{TierFast: 0, TierFull: 1, TierNightly: 2}

// Select returns the scenarios matching sel (pending ones included: they are listed and reported, never run).
func Select(all []Scenario, sel Selector) []Scenario {
	var out []Scenario
	for _, s := range all {
		if sel.Tier != "" && tierRank[s.Tier] > tierRank[sel.Tier] {
			continue
		}
		if sel.Kind != "" && s.Kind != sel.Kind {
			continue
		}
		if len(sel.IDs) > 0 && !matchID(sel.IDs, s.ID) {
			continue
		}
		if len(sel.Changed) > 0 && !touched(s.Touches, sel.Changed) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func touched(prefixes, changed []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	for _, p := range prefixes {
		for _, c := range changed {
			if strings.HasPrefix(c, p) {
				return true
			}
		}
	}
	return false
}

func matchID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id || strings.HasSuffix(x, ".") && strings.HasPrefix(id, x) {
			return true
		}
	}
	return false
}

// Expand lists the runs of the selected scenarios. The fast tier runs the first class and first seed of each scenario
// (one cell, to stay in the minutes a merge gate has); the full and nightly tiers run the whole matrix. A Class or Seed in
// the selector overrides that for a one-scenario dev loop. Pending scenarios give no runs. An error names a class the
// scenario's matrix lacks.
func Expand(scs []Scenario, sel Selector) ([]Run, error) {
	var runs []Run
	for _, s := range scs {
		if s.Pending() {
			continue
		}
		classes, seeds := s.Matrix.Classes, s.Matrix.Seeds
		if sel.Tier == TierFast && sel.Class == "" && sel.Seed == nil {
			classes, seeds = classes[:1], seeds[:1]
		}
		if sel.Class != "" {
			if _, err := ClassIndex(sel.Class); err != nil {
				return nil, err
			}
			classes = []string{sel.Class}
		}
		if sel.Seed != nil {
			seeds = []int64{*sel.Seed}
		}
		for _, c := range classes {
			for _, sd := range seeds {
				runs = append(runs, Run{ID: s.ID, Class: c, Seed: sd})
			}
		}
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].Key() < runs[j].Key() })
	return runs, nil
}

// MaxLabSlots is the most scenarios run at once on the builder (memory guard of the long lane: about 1 GB per d2d and
// bot pair, docs/builder.md). One lab slot per run.
const MaxLabSlots = 4

// Estimate is the wall-clock estimate of a plan in minutes at the given parallelism, from each scenario's wall budget
// (an upper bound, not a prediction).
func Estimate(scs []Scenario, runs []Run, parallel int) int {
	if parallel < 1 {
		parallel = 1
	}
	if parallel > MaxLabSlots {
		parallel = MaxLabSlots
	}
	by := map[string]Scenario{}
	for _, s := range scs {
		by[s.ID] = s
	}
	total := 0
	for _, r := range runs {
		total += by[r.ID].Budget.WallMin
	}
	return (total + parallel - 1) / parallel
}
