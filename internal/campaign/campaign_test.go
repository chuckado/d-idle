package campaign_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/campaign"
	"d2lod-bot/internal/readiness"
	"d2lod-bot/internal/task"
)

func shipped(t *testing.T) *campaign.Spine {
	t.Helper()
	ct, err := readiness.ShippedChecks()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, c := range ct.Checks {
		ids[c.ID] = true
	}
	s, err := campaign.Shipped(campaign.Validator{
		Quest: func(q string) bool { return task.Handler(q) != nil },
		Check: func(id string) bool { return ids[id] },
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

var t0 = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func TestShippedChapters(t *testing.T) {
	s := shipped(t)
	want := []string{"start", "cold_plains_wp", "den_of_evil", "blood_raven", "rogue", "find_cain", "countess", "tools_of_the_trade", "andariel"}
	if len(s.Chapters) != len(want) {
		t.Fatalf("chapters %d, want %d", len(s.Chapters), len(want))
	}
	for i, id := range want {
		if s.Chapters[i].ID != id {
			t.Errorf("chapter %d = %q, want %q", i, s.Chapters[i].ID, id)
		}
	}
	for _, cl := range []string{"paladin", "sorceress", "amazon", "necromancer", "barbarian", "druid", "assassin"} {
		for i := range s.Chapters {
			if _, err := s.Chapters[i].LevelsFor(cl); err != nil {
				t.Errorf("%s: %v", cl, err)
			}
		}
	}
}

func TestParseErrorsNameTheChapter(t *testing.T) {
	base := `{"schema":"campaign/1","difficulty":"normal","farm":[],"chapters":[{"id":"a","title":"A","act":1,"area":"x","areaId":1,"kind":"quest","done":{"level":2},"reward":"r","why":"w","levels":{"default":{"floor":1,"recommended":2}}}%s]}`
	cases := map[string]string{
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","requires":["zz"],"done":{"level":2},"reward":"r","why":"w","levels":{"default":{"floor":1,"recommended":2}}}`: `chapter "b": requires "zz"`,
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","done":{"quest":"nope"},"reward":"r","why":"w","levels":{"default":{"floor":1,"recommended":2}}}`:              `chapter "b": unknown quest "nope"`,
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","done":{"level":2},"reward":"r","why":"w","levelsFrom":"plandata:mephisto"}`:                                   `chapter "b": unknown levelsFrom`,
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","done":{"level":2},"reward":"r","why":"w","checks":["luck"],"levels":{"default":{"floor":1,"recommended":2}}}`: `chapter "b": unknown readiness check "luck"`,
		`,{"id":"a","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","done":{"level":2},"reward":"r","why":"w","levels":{"default":{"floor":1,"recommended":2}}}`:                   `chapter "a": duplicate id`,
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"walk","done":{"level":2},"reward":"r","why":"w","levels":{"default":{"floor":1,"recommended":2}}}`:                    `chapter "b": unknown kind`,
		`,{"id":"b","title":"B","act":1,"area":"x","areaId":1,"kind":"quest","done":{"level":2},"reward":"r","why":"w","levels":{"paladin":{"floor":1,"recommended":2}}}`:                   `chapter "b": levels has no default row`,
	}
	v := campaign.Validator{Quest: func(q string) bool { return task.Handler(q) != nil }, Check: func(id string) bool { return id == "level" }}
	if _, err := campaign.Parse([]byte(strings.Replace(base, "%s", "", 1)), v); err != nil {
		t.Fatalf("base: %v", err)
	}
	for extra, want := range cases {
		_, err := campaign.Parse([]byte(strings.Replace(base, "%s", extra, 1)), v)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("err %v, want %q", err, want)
		}
	}
}

func TestBands(t *testing.T) {
	s := shipped(t)
	andy := s.Chapter("andariel")
	cases := []struct {
		class string
		lvl   int
		merc  bool
		want  campaign.Band
	}{
		{"paladin", 16, false, campaign.TooEarly}, // floor 17 without a merc (ADR 0011)
		{"paladin", 17, false, campaign.Ready},
		{"paladin", 20, false, campaign.Comfortable}, // recommended 20 alone, 18 with the Rogue (boss1-trace, I)
		{"paladin", 15, true, campaign.Ready},
		{"paladin", 18, true, campaign.Comfortable},
		{"sorceress", 16, false, campaign.TooEarly},
		{"sorceress", 17, false, campaign.Ready},
		{"sorceress", 16, true, campaign.Comfortable},
	}
	for _, c := range cases {
		r, err := campaign.Rate(andy, campaign.Hero{Class: c.class, Level: c.lvl, Merc: c.merc}, nil)
		if err != nil || r.Band != c.want {
			t.Errorf("%s clvl %d merc %v: %v %v, want %v", c.class, c.lvl, c.merc, r.Band, err, c.want)
		}
	}
	// A failing listed check caps Comfortable at Ready; an unlisted one is ignored.
	r, _ := campaign.Rate(andy, campaign.Hero{Class: "paladin", Level: 20}, []campaign.Check{{ID: "potions", State: "fail", Reason: "2 < 8"}, {ID: "damage", State: "fail"}})
	if r.Band != campaign.Ready || len(r.Reasons) != 1 {
		t.Errorf("capped: %+v", r)
	}
	if _, err := campaign.Rate(andy, campaign.Hero{Class: "monk", Level: 18}, nil); err == nil || !strings.Contains(err.Error(), "monk") {
		t.Errorf("unknown class: %v", err)
	}
}

// The ladder of chapters a hero walks when it levels and finishes each chapter it is sent to (both test classes).
func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// The live hero of 2026-11-11 (Paladin clvl ~12, Den and Blood Raven done, no living merc) is sent to Cain, not to a
// "level up to 16" grind for Andariel.
func TestStorePersists(t *testing.T) {
	s := shipped(t)
	path := filepath.Join(t.TempDir(), "state", "campaign.json")
	st, err := campaign.LoadStore(path)
	if err != nil {
		t.Fatal(err)
	}
	h := st.Hero("Palli")
	_ = h.SetIntent(s, "countess", campaign.Skip, t0)
	_ = h.SetGoAt(campaign.GoFloor)
	if err := st.Save(path); err != nil {
		t.Fatal(err)
	}
	back, err := campaign.LoadStore(path)
	if err != nil {
		t.Fatal(err)
	}
	b := back.Hero("Palli")
	if b.Intents["countess"] != campaign.Skip || b.GoAt != campaign.GoFloor || b.Mode != campaign.Guided || b.Skipped["countess"].IsZero() {
		t.Errorf("%+v", b)
	}
	if other := back.Hero("Sorci"); other.Mode != campaign.Guided || len(other.Intents) != 0 {
		t.Errorf("new hero %+v", other)
	}
}

// Blood Raven done and no merc ever claimed: the Rogue chapter comes first; a claimed Rogue that died does not repeat it.
