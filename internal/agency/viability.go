package agency

import (
	"fmt"
	"sort"
	"strings"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
)

// Viability answers "can this hero clear Normal Act N at this level?" with the survival model's own margin and nothing
// new: every zone of the act at the difficulty is judged through power.Model.Zone with the plan's archetype and attack,
// under the default dial. The words are the ones the quest board shows (safe, risky, too hard).
type Viability struct {
	Diff      string   `json:"diff"`
	Act       int      `json:"act"`
	Zones     int      `json:"zones"`
	Safe      int      `json:"safe"`
	Risky     int      `json:"risky"`
	TooHard   int      `json:"tooHard"`
	Word      string   `json:"word"`      // the worst zone's word: the act is only as clear as its hardest zone
	Hardest   string   `json:"hardest"`   // name of that zone
	Margin    float64  `json:"margin"`    // the bad-pack margin of that zone, Best
	Weakest   string   `json:"weakest"`   // the least certain term of that zone ("defence_default"), "" when all are measured
	Archetype string   `json:"archetype"` // the plan's archetype, the one the margin used
	Line      string   `json:"line"`
	Notes     []string `json:"notes,omitempty"`
}

var wordRank = map[string]int{"safe": 0, "risky": 1, "too hard": 2}

// Judge runs the margin of every zone of act at diff for the hero. hero is the read hero (power.HeroFromPlayer); the
// plan's attack is put on it the way the survival wiring does. An act the table has no zone for is an error naming it.
func Judge(m *power.Model, tab *danger.Table, hero power.Hero, pl Plan, diff string, act int) (Viability, error) {
	v := Viability{Diff: diff, Act: act}
	if m == nil || tab == nil {
		return v, fmt.Errorf("agency: viability needs the survival model and the danger table")
	}
	if pl.Arch == nil {
		return v, fmt.Errorf("agency: the plan has no archetype (%d warnings), so no margin", len(pl.Warnings))
	}
	if err := m.Attach(&hero, pl.Spec); err != nil {
		v.Notes = append(v.Notes, err.Error())
	}
	v.Archetype = pl.Arch.ID
	var worst power.ZoneMargin
	for _, z := range tab.Zones {
		if z.Act != act || z.Diff != diff {
			continue
		}
		zm, err := m.Zone(hero, pl.Arch, z, m.P.DefaultDial)
		if err != nil {
			return v, fmt.Errorf("agency: zone %q: %w", z.Name, err)
		}
		v.Zones++
		switch zm.Word() {
		case "safe":
			v.Safe++
		case "risky":
			v.Risky++
		default:
			v.TooHard++
		}
		if v.Zones == 1 || wordRank[zm.Word()] > wordRank[worst.Word()] || (wordRank[zm.Word()] == wordRank[worst.Word()] && zm.Bad.Best < worst.Bad.Best) {
			worst = zm
		}
	}
	if v.Zones == 0 {
		return v, fmt.Errorf("agency: no zone of act %d at %q in the danger table", act, diff)
	}
	v.Word, v.Hardest, v.Margin, v.Weakest = worst.Word(), worst.Name, worst.Bad.Best, worst.Weakest()
	est := "measured"
	if v.Weakest != "" {
		est = "estimate, " + strings.ReplaceAll(v.Weakest, "_", " ")
	}
	v.Line = fmt.Sprintf("%s act %d: %s (%d safe, %d risky, %d too hard of %d zones; hardest %s; %s)",
		strings.ToUpper(diff[:1])+diff[1:], act, v.Word, v.Safe, v.Risky, v.TooHard, v.Zones, v.Hardest, est)
	return v, nil
}

// Acts lists the acts the table has zones for at diff, ascending.
func Acts(tab *danger.Table, diff string) []int {
	seen := map[int]bool{}
	for _, z := range tab.Zones {
		if z.Diff == diff {
			seen[z.Act] = true
		}
	}
	var out []int
	for a := range seen {
		out = append(out, a)
	}
	sort.Ints(out)
	return out
}
