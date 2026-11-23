package charsel

import (
	"os"
	"path/filepath"
	"testing"
)

// A parked folder with no .d2s is a hero the game has not saved yet: it is reported (never silently dropped), the playing hero and a
// saved hero are not.
func TestUnsavedParkedHeroes(t *testing.T) {
	root := t.TempDir()
	st := &Store{SaveDir: filepath.Join(root, "save"), StateRoot: filepath.Join(root, "state")}
	os.MkdirAll(st.SaveDir, 0o755)
	for _, n := range []string{"Ghost", "Playing", "Saved", "zed"} {
		os.MkdirAll(filepath.Join(st.StateRoot, n), 0o755)
	}
	os.WriteFile(filepath.Join(st.StateRoot, "stray.txt"), []byte("x"), 0o600) // a file is not a hero
	got := st.Unsaved("playing", []Entry{{Name: "SAVED"}})
	if len(got) != 2 || got[0] != "Ghost" || got[1] != "zed" {
		t.Errorf("unsaved = %v, want [Ghost zed]", got)
	}
	if g := (&Store{}).Unsaved("", nil); len(g) != 0 {
		t.Errorf("no state root: %v", g)
	}
}
