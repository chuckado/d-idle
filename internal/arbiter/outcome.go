package arbiter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// The outcome harness (docs/arbiter.md, "exit rule by outcome"). A family leaves the chain when the arbiter-led run has
// fewer invariant hits than the chain-led run and no new kind of hit. This file scores ONE decision point: a case holds
// the facts of an incident pre-state (or a recorded divergence), the family the chain recorded, and the arbiter decides
// from the same facts. Both choices are scored by the same data rules: a rule names an invariant that the choice would
// hit under some facts. The rules are a model of the monitor, not the monitor (U until a lockstep run confirms a row).

const OutcomeVersion = "didle.arbiter-outcomes/1"

// OutcomeRule says: choosing one of Families (hook families, see Family) while all When conditions hold hits Hit.
type OutcomeRule struct {
	Name     string   `json:"name"`
	Hit      string   `json:"hit"`      // an invariant monitor row (cmd/d2scenario/invariants.json)
	Families []string `json:"families"` // the choice's hook family; "*" for any
	When     []string `json:"when"`     // "fact op fact-or-number", op one of < <= > >= == !=; all must hold
	Why      string   `json:"why"`
}

// OutcomeCase is one decision point to score.
type OutcomeCase struct {
	ID         string         `json:"id"`
	Incident   string         `json:"incident,omitempty"` // the scenario id in tools/provingground/incidents it mirrors
	Family     string         `json:"family"`             // the exit-rule family the case counts for (docs/arbiter.md)
	Area       string         `json:"area"`
	Facts      map[string]int `json:"facts"`
	Objective  string         `json:"objectiveArea,omitempty"`
	Chain      string         `json:"chain"` // the hook family the chain acted for (recorded)
	Provenance string         `json:"provenance"`
}

type OutcomeFile struct {
	Version string        `json:"version"`
	Rules   []OutcomeRule `json:"rules"`
	Cases   []OutcomeCase `json:"cases"`
}

// LoadOutcomes reads and validates an outcome file.
func LoadOutcomes(path string) (OutcomeFile, error) {
	var f OutcomeFile
	b, err := os.ReadFile(path)
	if err != nil {
		return f, err
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return f, fmt.Errorf("%s: %w", path, err)
	}
	if f.Version != OutcomeVersion {
		return f, fmt.Errorf("%s: version %q, want %q", path, f.Version, OutcomeVersion)
	}
	for _, r := range f.Rules {
		for _, c := range r.When {
			if _, _, _, err := parseCond(c); err != nil {
				return f, fmt.Errorf("%s: rule %s: %w", path, r.Name, err)
			}
		}
	}
	for _, c := range f.Cases {
		if c.ID == "" || c.Chain == "" || c.Provenance == "" {
			return f, fmt.Errorf("%s: case %q needs id, chain and provenance", path, c.ID)
		}
	}
	return f, nil
}

var condOps = []string{"<=", ">=", "==", "!=", "<", ">"}

func parseCond(s string) (l, op, r string, err error) {
	for _, o := range condOps {
		if i := strings.Index(s, o); i > 0 {
			return strings.TrimSpace(s[:i]), o, strings.TrimSpace(s[i+len(o):]), nil
		}
	}
	return "", "", "", fmt.Errorf("condition %q has no operator", s)
}

func operand(s string, facts map[string]int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return facts[s]
}

func holds(c string, facts map[string]int) bool {
	l, op, r, err := parseCond(c)
	if err != nil {
		return false
	}
	a, b := operand(l, facts), operand(r, facts)
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "==":
		return a == b
	}
	return a != b
}

// Hits returns the invariant names the rules say a choice of this family would hit under these facts, sorted.
func (f OutcomeFile) Hits(family string, facts map[string]int) []string {
	set := map[string]bool{}
	for _, r := range f.Rules {
		if !inList(r.Families, family) {
			continue
		}
		ok := true
		for _, c := range r.When {
			if !holds(c, facts) {
				ok = false
				break
			}
		}
		if ok {
			set[r.Hit] = true
		}
	}
	var out []string
	for h := range set {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

func inList(l []string, s string) bool {
	for _, x := range l {
		if x == "*" || x == s {
			return true
		}
	}
	return false
}

// CaseScore is the result of one case.
type CaseScore struct {
	ID, Incident, Family string
	Chain, Arbiter       string   // the families chosen ("nothing" when the arbiter chose nothing)
	ArbiterIntent        string   // Compact form of the arbiter's choice
	ChainHits            []string // invariant names the chain's choice hits
	ArbiterHits          []string
	NewKinds             []string // hits of the arbiter that the chain did not have
}

// Score decides the case with the standard sources and scores both choices.
func (f OutcomeFile) Score(c OutcomeCase, cfg Config) CaseScore {
	facts := c.Facts
	if facts == nil {
		facts = map[string]int{}
	}
	d := New(cfg, StdSources()...).Decide(View{Area: c.Area, Facts: facts, ObjectiveArea: c.Objective})
	mine := "nothing"
	if d.Chosen != nil {
		mine = Family(d.Chosen.Source)
	}
	s := CaseScore{ID: c.ID, Incident: c.Incident, Family: c.Family, Chain: c.Chain, Arbiter: mine, ArbiterIntent: d.Compact(),
		ChainHits: f.Hits(c.Chain, facts), ArbiterHits: f.Hits(mine, facts)}
	for _, h := range s.ArbiterHits {
		if !inList(s.ChainHits, h) {
			s.NewKinds = append(s.NewKinds, h)
		}
	}
	return s
}

// Verdict is the exit rule for one exit-rule family: fewer hard hits for the arbiter (hard decides who is right), no
// new kind of hit of any tier, and nothing worse in the soft hits. A family with no case has no verdict.
type Verdict struct {
	Family             string
	Cases              int
	ChainHard, ArbHard int
	ChainSoft, ArbSoft int
	NewKinds           []string
	Pass               bool
	Note               string
}

// Judge folds case scores into one verdict per family. hard says whether an invariant name is a hard row.
func Judge(scores []CaseScore, hard func(string) bool) []Verdict {
	byFam := map[string]*Verdict{}
	var order []string
	for _, s := range scores {
		v := byFam[s.Family]
		if v == nil {
			v = &Verdict{Family: s.Family}
			byFam[s.Family] = v
			order = append(order, s.Family)
		}
		v.Cases++
		for _, h := range s.ChainHits {
			if hard(h) {
				v.ChainHard++
			} else {
				v.ChainSoft++
			}
		}
		for _, h := range s.ArbiterHits {
			if hard(h) {
				v.ArbHard++
			} else {
				v.ArbSoft++
			}
		}
		v.NewKinds = append(v.NewKinds, s.NewKinds...)
	}
	sort.Strings(order)
	var out []Verdict
	for _, n := range order {
		v := byFam[n]
		fewer := v.ArbHard < v.ChainHard || v.ArbHard == v.ChainHard && v.ArbSoft < v.ChainSoft
		worse := v.ArbHard > v.ChainHard || v.ArbSoft > v.ChainSoft
		v.Pass = fewer && !worse && len(v.NewKinds) == 0
		switch {
		case v.Pass:
			v.Note = "arbiter has fewer hits, no new kind"
		case len(v.NewKinds) > 0:
			v.Note = "new kind of hit: " + strings.Join(v.NewKinds, ",")
		case worse:
			v.Note = "arbiter has more hits"
		default:
			v.Note = "no fewer hits than the chain (the arbiter is no better here)"
		}
		out = append(out, *v)
	}
	return out
}
