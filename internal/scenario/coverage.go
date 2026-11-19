package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BehavioursVersion is the behaviour list format.
const BehavioursVersion = "didle.behaviours/1"

// Behaviour is one thing the bot does that a scenario can exercise: a runGoal hook, a chore, a reflex. Quest chunks are
// not listed by hand: ChunkBehaviours derives them from data/questboard.
type Behaviour struct {
	ID    string `json:"id"`
	Group string `json:"group"` // loop-hook, town-chore, resource, recovery, loot, travel, merc, difficulty, quest
	Title string `json:"title"`
	// Where is the code entry (file and function) that implements it: what a change to it must re-prove.
	Where string `json:"where"`
}

// LoadBehaviours reads behaviours.json.
func LoadBehaviours(path string) ([]Behaviour, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f struct {
		Version    string      `json:"version"`
		Behaviours []Behaviour `json:"behaviours"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if f.Version != BehavioursVersion {
		return nil, fmt.Errorf("%s: version %q, want %q", path, f.Version, BehavioursVersion)
	}
	seen := map[string]bool{}
	for _, bh := range f.Behaviours {
		if bh.ID == "" || bh.Where == "" {
			return nil, fmt.Errorf("%s: behaviour %q needs id and where", path, bh.ID)
		}
		if seen[bh.ID] {
			return nil, fmt.Errorf("%s: duplicate behaviour %q", path, bh.ID)
		}
		seen[bh.ID] = true
	}
	return f.Behaviours, nil
}

// ChunkBehaviours derives one behaviour per quest board chunk ("quest.<quest>.<chunk>") from data/questboard/act*.json.
func ChunkBehaviours(boardDir string) ([]Behaviour, error) {
	names, err := filepath.Glob(filepath.Join(boardDir, "act*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var out []Behaviour
	for _, n := range names {
		b, err := os.ReadFile(n)
		if err != nil {
			return nil, err
		}
		var f struct {
			Quests []struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Chunks []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"chunks"`
			} `json:"quests"`
		}
		if err := json.Unmarshal(b, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", n, err)
		}
		for _, q := range f.Quests {
			for _, c := range q.Chunks {
				out = append(out, Behaviour{ID: "quest." + q.ID + "." + c.ID, Group: "quest", Title: q.Title + ": " + c.Title, Where: filepath.Base(n)})
			}
		}
	}
	return out, nil
}

// CoverRow is one behaviour with the scenarios that exercise it.
type CoverRow struct {
	Behaviour
	Scenarios []string `json:"scenarios"`
	Active    int      `json:"active"`  // non-pending scenarios
	Pending   int      `json:"pending"` // skeletons only
}

// Coverage maps every behaviour to the scenarios naming it in Covers (a step scenario also covers its quest chunk) and
// returns an error list for scenarios naming a behaviour that does not exist.
func Coverage(bs []Behaviour, scs []Scenario) ([]CoverRow, []string) {
	rows := make([]CoverRow, len(bs))
	idx := map[string]int{}
	for i, b := range bs {
		rows[i] = CoverRow{Behaviour: b}
		idx[b.ID] = i
	}
	var problems []string
	for _, s := range scs {
		ids := append([]string(nil), s.Covers...)
		if s.Quest != nil {
			ids = append(ids, "quest."+s.Quest.Quest+"."+s.Quest.Chunk)
		}
		for _, id := range ids {
			i, ok := idx[id]
			if !ok {
				problems = append(problems, fmt.Sprintf("scenario %s covers unknown behaviour %q", s.ID, id))
				continue
			}
			rows[i].Scenarios = append(rows[i].Scenarios, s.ID)
			if s.Pending() {
				rows[i].Pending++
			} else {
				rows[i].Active++
			}
		}
	}
	return rows, problems
}

// CoverageReport writes the coverage table: covered behaviours with their scenarios, then the UNCOVERED list (no
// scenario at all) and the PENDING-ONLY list (skeletons only). Quest chunks are summarised per quest to keep the page short.
func CoverageReport(w io.Writer, rows []CoverRow) {
	var covered, pendingOnly, uncovered []CoverRow
	for _, r := range rows {
		switch {
		case r.Active > 0:
			covered = append(covered, r)
		case r.Pending > 0:
			pendingOnly = append(pendingOnly, r)
		default:
			uncovered = append(uncovered, r)
		}
	}
	fmt.Fprintf(w, "behaviours %d: covered %d, pending-only %d, UNCOVERED %d\n\n", len(rows), len(covered), len(pendingOnly), len(uncovered))
	fmt.Fprintln(w, "COVERED")
	for _, r := range covered {
		fmt.Fprintf(w, "  %-44s %s\n", r.ID, strings.Join(r.Scenarios, ", "))
	}
	fmt.Fprintln(w, "\nPENDING-ONLY (a skeleton exists, nothing runs)")
	for _, r := range pendingOnly {
		fmt.Fprintf(w, "  %-44s %s\n", r.ID, strings.Join(r.Scenarios, ", "))
	}
	fmt.Fprintln(w, "\nUNCOVERED")
	byGroup := map[string][]string{}
	var groups []string
	for _, r := range uncovered {
		if _, ok := byGroup[r.Group]; !ok {
			groups = append(groups, r.Group)
		}
		byGroup[r.Group] = append(byGroup[r.Group], r.ID)
	}
	sort.Strings(groups)
	for _, g := range groups {
		fmt.Fprintf(w, "  %s (%d): %s\n", g, len(byGroup[g]), strings.Join(byGroup[g], ", "))
	}
}
