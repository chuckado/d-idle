package main

import (
	"os"
	"path/filepath"
	"testing"

	"d2lod-bot/internal/campaign"
	"d2lod-bot/internal/questboard"
)

// importRig copies a fixture of an old state/campaign.json into a state directory and opens the hero's plan.
func importRig(t *testing.T, fixture, hero string) (stateDir string, st *questboard.Store) {
	t.Helper()
	stateDir = t.TempDir()
	b, err := os.ReadFile(filepath.Join("testdata", "journey", fixture))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, campaign.StateFile), b, 0o644); err != nil {
		t.Fatal(err)
	}
	board, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err = questboard.Open(board, filepath.Join(stateDir, "quests"), hero)
	if err != nil {
		t.Fatal(err)
	}
	return stateDir, st
}

func noLog(string, ...any) {}

// A Paladin mid-Journey: the Countess and Blood Raven skips carry over as quest skips; the done chapters, the hold, the question and
// Go now are not imported (the board derives progress from the game); the old file is untouched and backed up.
func TestImportJourneyPaladinMidJourney(t *testing.T) {
	dir, st := importRig(t, "campaign-paladin-midjourney.json", "Palli")
	before, _ := os.ReadFile(filepath.Join(dir, campaign.StateFile))
	for i := 0; i < 2; i++ { // twice: a restart changes nothing
		importJourney(dir, "Palli", st, noLog)
	}
	p, _, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Skipped) != 2 || p.Skipped[0].Quest != "forgotten_tower" || p.Skipped[1].Quest != "sisters_burial_grounds" || len(p.Accepted) != 0 || p.Mode != "" {
		t.Errorf("plan: %+v", p)
	}
	after, _ := os.ReadFile(filepath.Join(dir, campaign.StateFile))
	bak, err := os.ReadFile(filepath.Join(dir, campaign.StateFile+".pre-board"))
	if err != nil || string(after) != string(before) || string(bak) != string(before) {
		t.Errorf("the Journey file changed or has no backup: %v", err)
	}
}

// A Sorceress in orders mode farms between quests; a hero the file does not know and a missing or broken file leave the
// plan clean without an error.
func TestImportJourneySorceressAndEdges(t *testing.T) {
	dir, st := importRig(t, "campaign-sorceress-orders.json", "Sorci")
	importJourney(dir, "Sorci", st, noLog)
	if p, _, _ := st.Load(); p.Mode != "farm" || len(p.Skipped) != 0 {
		t.Errorf("orders: %+v", p)
	}

	dir2, st2 := importRig(t, "campaign-sorceress-orders.json", "Other")
	importJourney(dir2, "Other", st2, noLog)
	if p, _, _ := st2.Load(); p.Mode != "" || p.Legacy != "" {
		t.Errorf("unknown hero: %+v", p)
	}

	dir3, st3 := importRig(t, "campaign-sorceress-orders.json", "Sorci")
	if err := os.WriteFile(filepath.Join(dir3, campaign.StateFile), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	importJourney(dir3, "Sorci", st3, noLog)
	if p, _, _ := st3.Load(); p.Legacy != "" {
		t.Errorf("broken file imported: %+v", p)
	}
	importJourney(t.TempDir(), "Sorci", st3, noLog) // no file at all
}
