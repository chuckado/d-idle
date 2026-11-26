package bot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/d2"
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

	// Monitor incidents (review P8): a hit of the invariant monitor writes the pre-state too, at most one file per
	// invariant name per incidentMonGapSec, at most incidentMonPerHour in any hour, and at most incidentMonKeep files
	// of this kind are kept (so a storm cannot push the death files out). Tools only; nothing here acts.
	incidentMonGapSec  = 600
	incidentMonPerHour = 6
	incidentMonKeep    = 20
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
	// Invariant is the monitor row that fired (kind "monitor"): the replay row is red when it fires again.
	Invariant string `json:"invariant,omitempty"`
	Hard      bool   `json:"hard,omitempty"`
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

func trimIncidents(dir string, keep int) { trimGlob(dir, "*.json", keep) }

func trimGlob(dir, pattern string, keep int) {
	names, _ := filepath.Glob(filepath.Join(dir, pattern))
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

// incidentGate is the rate limit of monitor incidents (guarded by the invMon mutex of its owner).
type incidentGate struct {
	last  map[string]time.Time
	times []time.Time
}

func (g *incidentGate) allow(name string, now time.Time) bool {
	if t, ok := g.last[name]; ok && now.Sub(t) < incidentMonGapSec*time.Second {
		return false
	}
	keep := g.times[:0]
	for _, t := range g.times {
		if now.Sub(t) < time.Hour {
			keep = append(keep, t)
		}
	}
	g.times = keep
	if len(g.times) >= incidentMonPerHour {
		return false
	}
	if g.last == nil {
		g.last = map[string]time.Time{}
	}
	g.last[name] = now
	g.times = append(g.times, now)
	return true
}

// monitorIncident builds the incident of an invariant hit from the live snapshot; nil without a player.
func (r *Runner) monitorIncident(snap *d2.Snapshot, h invHit) *Incident {
	if snap == nil || snap.Player == nil {
		return nil
	}
	p := snap.Player
	k := r.kitOf(snap)
	start := "field"
	if p.InTown() {
		start = "town"
	}
	r.mu.Lock()
	state, step := string(r.st), r.activity
	r.mu.Unlock()
	row := invRowOf(h.Name)
	in := &Incident{Version: IncidentVersion, Kind: "monitor", Time: r.e.Clock.Now(), AreaName: r.areaName(p.Area), Class: p.Class.String(), Tags: []string{"SETUP"},
		Pre: IncidentPre{Start: start, Area: p.Area, Level: p.Level, Gold: p.Gold, Stash: p.StashGold, Difficulty: difficultyName(p.Difficulty), LifePct: p.LifePct()},
		Facts: IncidentFacts{BeltHP: k.BeltHP, BeltMP: k.BeltMP, PackHP: k.PackHP, TownPortals: k.TownPortals, State: state, Activity: step,
			Invariant: h.Name, Hard: row.Hard, Summary: h.Name + ": " + h.Detail}}
	if start == "field" {
		in.Facts.HostilesNear = r.hostileNear(snap, r.e.Tune.Combat.FightRadius)
	}
	return in
}

// noteMonitorIncident writes the pre-state of an invariant hit (hard or soft), rate limited. Failures are only logged.
func (r *Runner) noteMonitorIncident(snap *d2.Snapshot, h invHit) {
	if r.cfg.DeathLogPath == "" {
		return
	}
	now := r.e.Clock.Now()
	r.inv.mu.Lock()
	ok := r.inv.incidents.allow(h.Name, now)
	r.inv.mu.Unlock()
	if !ok {
		return
	}
	in := r.monitorIncident(snap, h)
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
	name := strconv.FormatInt(now.Unix(), 10) + "-inv-" + h.Name + ".json"
	if err := os.WriteFile(filepath.Join(dir, name), append(b, '\n'), 0o644); err != nil {
		r.e.Log.Warn("incident not saved", "err", err)
		return
	}
	trimGlob(dir, "*-inv-*.json", incidentMonKeep)
	trimIncidents(dir, incidentKeep)
}
