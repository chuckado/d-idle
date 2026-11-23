package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The repository data must pass the check (every board step has a scenario or a MISSING reason).
func TestRepoCatalogueIsClean(t *testing.T) {
	bad, err := problems("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) > 0 {
		t.Fatalf("catalogue problems:\n%s", strings.Join(bad, "\n"))
	}
}

// Must-fail control: a copy of the data with one spec row removed, one scenario id invented and one empty MISSING reason is refused.
func TestCheckFailsOnGaps(t *testing.T) {
	root := t.TempDir()
	copyTree := func(rel string) {
		b, err := os.ReadFile(filepath.Join("../..", rel))
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(root, rel)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for a := '1'; a <= '5'; a++ {
		copyTree("data/questboard/act" + string(a) + ".json")
	}
	copyTree("data/stepgrounds/verdicts.json")
	copyTree("tools/provingground/scenarios/stepgrounds.json")
	spec := `{"schema":"stepgrounds/1","steps":[
 {"id":"den_of_evil/clear_den/talk_akara","scenario":"no.such.scenario"},
 {"id":"den_of_evil/clear_den/clear_den","missing":"just broken"}]}`
	_ = os.MkdirAll(filepath.Join(root, "data/stepgrounds"), 0o755)
	if err := os.WriteFile(filepath.Join(root, "data/stepgrounds/steps.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	bad, err := problems(root)
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(bad, "\n")
	for _, want := range []string{"does not exist", "MISSING reason must read", "neither a scenario nor a MISSING reason"} {
		if !strings.Contains(all, want) {
			t.Errorf("want a problem containing %q, got:\n%s", want, all)
		}
	}
}
