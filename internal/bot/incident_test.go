package bot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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

// An invariant hit (hard or soft) writes the pre-state with the invariant name, tagged SETUP; the same name inside the gap
// and more than incidentMonPerHour per hour write nothing; the monitor files are trimmed to incidentMonKeep. Two classes.
func TestMonitorHitWritesIncident(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		g := newRig(t)
		dir := t.TempDir()
		g.r.cfg.DeathLogPath = filepath.Join(dir, "deaths.json")
		g.w.set(func(w *world) { w.class = class })
		s, _ := g.w.snapshot()
		files := func() []string { n, _ := filepath.Glob(filepath.Join(dir, incidentDir, "*-inv-*.json")); return n }
		g.r.noteMonitorIncident(s, invHit{"stuck-in-pack", "hostiles=6"})
		if len(files()) != 1 {
			t.Fatalf("%v: %d files after a hit", class, len(files()))
		}
		b, _ := os.ReadFile(files()[0])
		var in Incident
		if err := json.Unmarshal(b, &in); err != nil {
			t.Fatal(err)
		}
		if in.Kind != "monitor" || in.Facts.Invariant != "stuck-in-pack" || in.Class != class.String() || len(in.Tags) != 1 || in.Tags[0] != "SETUP" || in.Facts.Hard {
			t.Fatalf("%v: incident %+v", class, in)
		}
		g.r.noteMonitorIncident(s, invHit{"stuck-in-pack", "again"})
		if len(files()) != 1 {
			t.Fatalf("%v: the same name inside the gap wrote a second file", class)
		}
		for _, name := range []string{"potion-not-drunk", "life-zero-alive", "xp-frozen", "mode-no-exit", "refusal-storm", "corpse-unresolved", "town-thrash"} {
			g.clock.Sleep(time.Second)
			g.r.noteMonitorIncident(s, invHit{name, "x"})
		}
		if n := len(files()); n != incidentMonPerHour {
			t.Fatalf("%v: %d files, want the hourly cap %d", class, n, incidentMonPerHour)
		}
	}
}

func TestTrimGlobKeepsDeathFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "1-death.json"), []byte("{}"), 0o644)
	for i := 0; i < 25; i++ {
		os.WriteFile(filepath.Join(dir, strconv.Itoa(100+i)+"-inv-x.json"), []byte("{}"), 0o644)
	}
	trimGlob(dir, "*-inv-*.json", incidentMonKeep)
	all, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(all) != incidentMonKeep+1 {
		t.Fatalf("%d files left", len(all))
	}
}
