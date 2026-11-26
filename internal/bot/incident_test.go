package bot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// A death writes one incident file with the recorded state tagged SETUP; the keep limit drops the oldest files.
func TestDeathWritesIncident(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g := newRig(t)
		dir := t.TempDir()
		g.r.cfg.DeathLogPath = filepath.Join(dir, "deaths.json")
		now := g.r.e.Clock.Now()
		for i := 0; i < 5; i++ {
			p := &d2.Player{Level: 9, Class: class, Life: 90 - 10*i, MaxLife: 100, Mana: 20, MaxMana: 50, Area: 3, Gold: 35, StashGold: 7}
			g.r.obMu.Lock()
			g.r.ob.lastGood.area = 3
			g.r.trackAlive(&d2.Snapshot{State: d2.StateInGame, Player: p}, now.Add(time.Duration(i)*time.Second))
			g.r.obMu.Unlock()
		}
		g.r.obMu.Lock()
		g.r.recordDeath(&d2.Player{Level: 9, Class: class, Area: 3}, now.Add(6*time.Second))
		g.r.obMu.Unlock()
		names, _ := filepath.Glob(filepath.Join(dir, incidentDir, "*.json"))
		if len(names) != 1 {
			t.Fatalf("%v: %d incident files", class, len(names))
		}
		b, _ := os.ReadFile(names[0])
		var in Incident
		if err := json.Unmarshal(b, &in); err != nil {
			t.Fatal(err)
		}
		if in.Version != IncidentVersion || in.Class != class.String() || in.Pre.Start != "field" || in.Pre.Area != 3 || in.Pre.Level != 9 ||
			in.Pre.Gold != 35 || in.Pre.Stash != 7 || len(in.Tags) != 1 || in.Tags[0] != "SETUP" || in.Pre.LifePct == 0 || in.Kind != "death" {
			t.Fatalf("%v: incident %+v", class, in)
		}
	}
}

func TestTrimIncidentsKeepsTheNewest(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"100-death.json", "200-death.json", "300-death-bug.json"} {
		os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644)
	}
	trimIncidents(dir, 2)
	left, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(left) != 2 || filepath.Base(left[0]) != "200-death.json" {
		t.Fatalf("%v", left)
	}
}
