package bot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Incident files (Fable logic review 2026-11-25, T2/P8, ADR 0020 amended 2026-11-25): at each death the bot writes the
// declared state of that moment as a scenario `pre`, so `d2scenario incident` can turn it into a regression row that the
// lockstep Proving Ground replays. The state is a RECORDED state, tagged SETUP under the test-character provenance rule:
// a row built from it may test mechanics, it is never evidence about game behaviour. Tier: report-only (tools only, not
// the merge gate). Format: docs/probes/incident-scenarios.md.

const (
	// IncidentVersion is the version string of an incident file.
	IncidentVersion = "didle.incident/1"
	incidentDir     = "incidents"
	incidentKeep    = 50
)

// IncidentPre mirrors the fields of the scenario `pre` (internal/scenario.Pre) that a death records. The belt is kept as
// counts in Facts: the tier of the potions is not recorded (U), the converter fills the lowest tier.
type IncidentPre struct {
	Start      string `json:"start"`
	Area       int    `json:"area"`
	Level      int    `json:"level"`
	Gold       int    `json:"gold,omitempty"`
	Stash      int    `json:"stash,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
	LifePct    int    `json:"life_pct,omitempty"`
}

// IncidentFacts are the recorded facts the pre cannot hold; the converter and the reader use them.
type IncidentFacts struct {
	BeltHP       int    `json:"beltHP"`
	BeltMP       int    `json:"beltMP"`
	PackHP       int    `json:"packHP"`
	TownPortals  int    `json:"townPortals"`
	HostilesNear int    `json:"hostilesNear"`
	Killer       string `json:"killer,omitempty"`
	State        string `json:"state,omitempty"`
	Activity     string `json:"activity,omitempty"`
	LastSafety   string `json:"lastSafety,omitempty"`
	Bug          bool   `json:"bug"`
	Summary      string `json:"summary,omitempty"`
}

// Incident is one file of the incidents directory.
type Incident struct {
	Version  string        `json:"version"`
	Kind     string        `json:"kind"` // death, or death-bug when the autopsy hit its potion bug test
	Time     time.Time     `json:"time"`
	AreaName string        `json:"areaName"`
	Class    string        `json:"class"`
	Build    string        `json:"build,omitempty"`
	Tags     []string      `json:"tags"` // always SETUP: a recorded state, not a played one
	Pre      IncidentPre   `json:"pre"`
	Facts    IncidentFacts `json:"facts"`
}

// incidentOf builds the incident of a death record; nil without an autopsy.
func incidentOf(rec DeathRecord) *Incident {
	au := rec.Autopsy
	if au == nil {
		return nil
	}
	in := &Incident{Version: IncidentVersion, Kind: "death", Time: rec.Time, AreaName: rec.AreaName, Class: au.Class, Build: au.Build,
		Tags: []string{"SETUP"},
		Pre:  IncidentPre{Start: "field", Area: rec.Area, Level: rec.Level, Gold: au.GoldCarried, Stash: au.StashGold, Difficulty: rec.Difficulty},
		Facts: IncidentFacts{BeltHP: au.BeltHP, BeltMP: au.BeltMP, PackHP: au.Kit.PackHP, TownPortals: au.Kit.TownPortals,
			HostilesNear: au.HostilesNear, Killer: rec.Killer, State: au.State, Activity: au.Activity, LastSafety: au.LastSafety, Bug: au.Bug, Summary: au.Summary}}
	if au.Bug {
		in.Kind = "death-bug"
	}
	if n := len(rec.LifeCurve); n > 0 { // the oldest sample of the window: the hero as the trouble began
		in.Pre.LifePct = rec.LifeCurve[0].Pct
	}
	return in
}

// writeIncident saves the incident next to the death log (<dir of the death log>/incidents/<unix>-<kind>.json) and keeps
// the newest incidentKeep files. A failure is only logged; nothing here changes what the hero does.
func (r *Runner) writeIncident(rec DeathRecord) {
	if r.cfg.DeathLogPath == "" {
		return
	}
	in := incidentOf(rec)
	if in == nil {
		return
	}
	dir := filepath.Join(filepath.Dir(r.cfg.DeathLogPath), incidentDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		r.e.Log.Warn("incident not saved", "err", err)
		return
	}
	b, err := json.MarshalIndent(in, "", " ")
	if err != nil {
		return
	}
	name := strconv.FormatInt(rec.Time.Unix(), 10) + "-" + in.Kind + ".json"
	if err := os.WriteFile(filepath.Join(dir, name), append(b, '\n'), 0o644); err != nil {
		r.e.Log.Warn("incident not saved", "err", err)
		return
	}
	trimIncidents(dir, incidentKeep)
}

func trimIncidents(dir string, keep int) {
	names, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(names) <= keep {
		return
	}
	sort.Strings(names) // unix seconds first: the oldest sort first
	for _, n := range names[:len(names)-keep] {
		if strings.HasSuffix(n, ".json") {
			os.Remove(n)
		}
	}
}
