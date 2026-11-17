package questboard

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestLoadEmbedded(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if n := len(b.Quests()); n != 14 {
		t.Fatalf("quests = %d, want 14 (7 per act)", n)
	}
	a1, err := b.Act(1)
	if err != nil || len(a1) != 7 {
		t.Fatalf("act 1: %d %v", len(a1), err)
	}
	staff, _ := b.Quest("horadric_staff")
	if len(staff.Chunks) != 3 {
		t.Fatalf("staff chunks = %d", len(staff.Chunks))
	}
}

// Every quest id must be a save id (internal/saves questTable), so Derive can read its word (V: read from the
// source file; the table has no exported list).
func TestQuestIDsAreSaveIDs(t *testing.T) {
	src, err := os.ReadFile("../saves/progress.go")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, m := range regexp.MustCompile(`\{"([a-z0-9_]+)", "[^"]+", \d, \d+\}`).FindAllStringSubmatch(string(src), -1) {
		ids[m[1]] = true
	}
	if len(ids) < 30 {
		t.Fatalf("parsed %d save ids, the pattern is stale", len(ids))
	}
	b, _ := Load()
	for _, q := range b.Quests() {
		if !ids[q.ID] {
			t.Errorf("quest %q is not a save quest id", q.ID)
		}
	}
}

func TestEveryQuestHasTagsAndMeta(t *testing.T) {
	b, _ := Load()
	for _, q := range b.Quests() {
		if len(q.Tags) == 0 {
			t.Errorf("%s: no tags", q.ID)
		}
		if q.Kind != "unlock" && len(q.Meta) == 0 {
			t.Errorf("%s: no meta column", q.ID)
		}
		for k, v := range q.Tags {
			if !strings.ContainsAny(v[:1], "VSUEI") {
				t.Errorf("%s tag %s = %q does not start with a tag", q.ID, k, v)
			}
		}
	}
}

func TestFloorByClassOnTwoBuilds(t *testing.T) {
	b, _ := Load()
	q, _ := b.Quest("sisters_burial_grounds")
	for class, want := range map[string]int{"paladin": 7, "sorceress": 8} {
		got, ok, err := b.Floor(q, class)
		if err != nil || !ok || got != want {
			t.Errorf("%s floor = %d %v %v, want %d", class, got, ok, err, want)
		}
	}
	if _, _, err := b.Floor(q, "bard"); err == nil || !strings.Contains(err.Error(), `"bard"`) {
		t.Errorf("unknown class error = %v", err)
	}
}
