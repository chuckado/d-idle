// Command stepcat is the step-level Proving Ground catalogue (card stepgrounds): it reads every step of the quest board
// data (data/questboard/act1..5.json) and data/stepgrounds/steps.json and
//
//	gen        writes tools/provingground/scenarios/stepgrounds_gen.json (step scenarios made from the spec's `gen` blocks)
//	catalogue  writes docs/probes/quest-step-catalogue.md (counts, one row per step, verdict from verdicts.json)
//	check      fails when a board step has neither a scenario nor an explicit MISSING reason, or the spec drifts from the data
//
// Adding a step scenario is a data edit in steps.json. Standard library only.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Spec is data/stepgrounds/steps.json.
type Spec struct {
	Schema string `json:"schema"`
	// Kit is the setup every generated scenario starts from (a scenario `pre` and `trigger`, as the scenario format has them).
	Kit   map[string]any `json:"kit"`
	Steps []StepSpec     `json:"steps"`
}

// StepSpec says how one board step is proven: exactly one of Scenario (a hand-written scenario id), Chunk (only chunk-level
// scenarios cover it), Gen (a generated step scenario) or Missing ("F<n>|bot|data: why").
type StepSpec struct {
	ID       string   `json:"id"` // quest/chunk/step
	Scenario string   `json:"scenario,omitempty"`
	Chunk    []string `json:"chunk,omitempty"`
	Also     []string `json:"also,omitempty"` // companion scenarios (a half of the step); shown with their verdicts, not counted
	Gen      *Gen     `json:"gen,omitempty"`
	Missing  string   `json:"missing,omitempty"`
	Note     string   `json:"note,omitempty"`
}

// Gen is one generated scenario: where the hero stands, what is put there, and what must be seen.
type Gen struct {
	Start   string           `json:"start,omitempty"` // area name of areas.json ("" or "Rogue Encampment" style town: start town)
	Quests  []map[string]any `json:"quests,omitempty"`
	Ground  []string         `json:"ground,omitempty"` // item codes given at the hero's feet (debug give <code> 2)
	Pack    []string         `json:"pack,omitempty"`
	Spawn   []map[string]any `json:"spawn,omitempty"`
	Goals   string           `json:"goals,omitempty"`
	Enter   string           `json:"enter,omitempty"`  // regexp of the area name the bot must log "Entered ..." for
	Kill    string           `json:"kill,omitempty"`   // monster name the bot must log "combat kill" for
	Kills   int              `json:"kills,omitempty"`  // at least this many "combat kill" lines of any monster
	Picked  string           `json:"picked,omitempty"` // item display name the bot must click ("loot: clicked")
	Merc    bool             `json:"merc,omitempty"`
	Budget  int              `json:"budget,omitempty"`
	Pending string           `json:"pending,omitempty"` // non-empty: status pending with this reason
}

// Verdicts is data/stepgrounds/verdicts.json: scenario id to its last recorded result.
type Verdicts map[string]struct {
	Verdict string `json:"verdict"` // GREEN RED BLOCKED
	Run     string `json:"run"`
	Tag     string `json:"tag"` // V (named run) or U
	Note    string `json:"note,omitempty"`
}

type step struct {
	Act                int
	Quest, Chunk, Step string
}

func (s step) key() string { return s.Quest + "/" + s.Chunk + "/" + s.Step }

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func boardSteps(root string) ([]step, error) {
	var out []step
	for a := 1; a <= 5; a++ {
		var d struct {
			Quests []struct {
				ID     string `json:"id"`
				Chunks []struct {
					ID    string   `json:"id"`
					Steps []string `json:"steps"`
				} `json:"chunks"`
			} `json:"quests"`
		}
		if err := readJSON(filepath.Join(root, fmt.Sprintf("data/questboard/act%d.json", a)), &d); err != nil {
			return nil, err
		}
		for _, q := range d.Quests {
			for _, c := range q.Chunks {
				for _, s := range c.Steps {
					out = append(out, step{a, q.ID, c.ID, s})
				}
			}
		}
	}
	return out, nil
}

// scenarioIDs lists every scenario id (and quest ref chunk key) of the hand-written scenario files.
func scenarioIDs(root string) (map[string]bool, error) {
	files, _ := filepath.Glob(filepath.Join(root, "tools/provingground/scenarios/*.json"))
	ids := map[string]bool{}
	for _, f := range files {
		var d struct {
			Scenarios []struct {
				ID string `json:"id"`
			} `json:"scenarios"`
		}
		if err := readJSON(f, &d); err != nil {
			return nil, err
		}
		for _, s := range d.Scenarios {
			ids[s.ID] = true
		}
	}
	return ids, nil
}

var missingRe = regexp.MustCompile(`^(F\d+|bot|data): \S`)

// problems is the check: the list of rule breaks (empty is clean).
func problems(root string) ([]string, error) {
	steps, err := boardSteps(root)
	if err != nil {
		return nil, err
	}
	var sp Spec
	if err := readJSON(filepath.Join(root, "data/stepgrounds/steps.json"), &sp); err != nil {
		return nil, err
	}
	ids, err := scenarioIDs(root)
	if err != nil {
		return nil, err
	}
	var ver Verdicts
	if err := readJSON(filepath.Join(root, "data/stepgrounds/verdicts.json"), &ver); err != nil {
		return nil, err
	}
	bySpec := map[string]StepSpec{}
	var bad []string
	for _, s := range sp.Steps {
		if _, dup := bySpec[s.ID]; dup {
			bad = append(bad, "duplicate spec row "+s.ID)
		}
		bySpec[s.ID] = s
	}
	known := map[string]bool{}
	for _, st := range steps {
		known[st.key()] = true
		s, ok := bySpec[st.key()]
		if !ok {
			bad = append(bad, "step "+st.key()+" has neither a scenario nor a MISSING reason")
			continue
		}
		n := 0
		if s.Scenario != "" {
			n++
			if !ids[s.Scenario] {
				bad = append(bad, st.key()+": scenario "+s.Scenario+" does not exist")
			}
		}
		if len(s.Chunk) > 0 {
			n++
			for _, c := range s.Chunk {
				if !ids[c] {
					bad = append(bad, st.key()+": chunk scenario "+c+" does not exist")
				}
			}
		}
		if s.Gen != nil {
			n++
		}
		if s.Missing != "" {
			n++
			if !missingRe.MatchString(s.Missing) {
				bad = append(bad, st.key()+`: MISSING reason must read "F<n>: ...", "bot: ..." or "data: ..."`)
			}
		}
		if n != 1 {
			bad = append(bad, fmt.Sprintf("%s: exactly one of scenario, chunk, gen, missing (has %d)", st.key(), n))
		}
	}
	for k := range bySpec {
		if !known[k] {
			bad = append(bad, "spec row "+k+" is not a step of the quest board data")
		}
	}
	for id, v := range ver {
		if v.Verdict == "GREEN" && (v.Run == "" || v.Tag != "V") {
			bad = append(bad, "verdict "+id+": GREEN needs a run id and tag V")
		}
	}
	sort.Strings(bad)
	return bad, nil
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	var err error
	switch flag.Arg(0) {
	case "gen":
		err = gen(*root)
	case "catalogue":
		err = catalogue(*root)
	case "check":
		var bad []string
		if bad, err = problems(*root); err == nil && len(bad) > 0 {
			err = fmt.Errorf("%d problems:\n  %s", len(bad), strings.Join(bad, "\n  "))
		}
	default:
		err = fmt.Errorf("usage: stepcat [-root DIR] gen|catalogue|check")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "stepcat:", err)
		os.Exit(1)
	}
}
