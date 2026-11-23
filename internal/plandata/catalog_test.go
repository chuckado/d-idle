package plandata

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestBuildCatalog is the contract: every plan table has build text (or an explicit reason), and the text is valid.
func TestBuildCatalog(t *testing.T) {
	d, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	covered := map[string]int{}
	for _, b := range d.Builds {
		for _, tb := range b.Tables {
			covered[tb]++
		}
	}
	for _, u := range d.Uncovered {
		covered[u.Table]++
	}
	for _, tb := range TableIDs(FS) {
		if covered[tb] != 1 {
			t.Errorf("plan table %q is covered %d times by the catalogue; add build text or an uncovered entry with a reason", tb, covered[tb])
		}
	}
	if len(d.Builds) < 15 {
		t.Errorf("only %d builds", len(d.Builds))
	}
}

func TestCatalogRejectsBadData(t *testing.T) {
	fsys := fstest.MapFS{
		"build-a.json":         {Data: []byte("{}")},
		"unlocks-paladin.json": {Data: []byte("{}")},
	}
	good := func() string {
		r := `{"stars":3,"source":"estimate","note":"n"}`
		return `{"version":1,"thresholds":{},"builds":[{"id":"b","name":"B","class":"paladin","tables":["a"],"variants":[{"table":"a","label":"Leveling to level 30","why":"The first thirty levels."}],"defaultVariant":"a","identity":"i","summary":"s",
"pros":["a","b","c"],"cons":["a","b","c"],"thresholdsRef":"thresholds","ratings":{"leveling":` + r + `,"farmingItems":` + r +
			`,"farmingXP":` + r + `,"bossing":` + r + `,"defense":` + r + `,"clearSpeed":` + r + `,"idle":` + r + `}}]}`
	}
	if _, err := ParseCatalog([]byte(good()), fsys); err != nil {
		t.Fatalf("good catalogue rejected: %v", err)
	}
	for name, c := range map[string]struct{ from, to, want string }{
		"unknown class":    {`"paladin"`, `"monk"`, "monk"},
		"unknown table":    {`["a"]`, `["zz"]`, "zz"},
		"stars out of 1-5": {`"stars":3,"source":"estimate","note":"n"}`, `"stars":6,"source":"estimate","note":"n"}`, "6 stars"},
		"bad source":       {`"source":"estimate"`, `"source":"guess"`, "guess"},
		"empty note":       {`"note":"n"`, `"note":" "`, "empty"},
		"two pros":         {`"pros":["a","b","c"]`, `"pros":["a","b"]`, "want 3"},
		"missing rating":   {`,"idle":{"stars":3,"source":"estimate","note":"n"}`, ``, "seven"},
		"too long":         {`"identity":"i"`, `"identity":"` + strings.Repeat("x", 81) + `"`, "limit"},
	} {
		_, err := ParseCatalog([]byte(strings.Replace(good(), c.from, c.to, 1)), fsys)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want an error containing %q", name, err, c.want)
		}
	}
	// a table without text fails, and an uncovered entry fixes it
	fsys["build-new.json"] = &fstest.MapFile{Data: []byte("{}")}
	if _, err := ParseCatalog([]byte(good()), fsys); err == nil || !strings.Contains(err.Error(), `"new"`) {
		t.Errorf("a new table without text must fail naming it, got %v", err)
	}
	withUncovered := strings.Replace(good(), `"builds"`, `"uncovered":[{"table":"new","reason":"scratch"}],"builds"`, 1)
	if _, err := ParseCatalog([]byte(withUncovered), fsys); err != nil {
		t.Errorf("uncovered with a reason must pass: %v", err)
	}
}

func cardKeys(c CardView) []string {
	var k []string
	for _, r := range c.Ratings {
		k = append(k, r.Key)
	}
	return k
}

func TestSmallCard(t *testing.T) {
	d, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		// Defense, two highest (ties: earlier key), lowest of the remaining
		"paladin_hammerdin":    {"defense", "farmingItems", "farmingXP", "leveling"},
		"sorceress_blizzard":   {"defense", "farmingItems", "leveling", "idle"},
		"necromancer_summoner": {"defense", "leveling", "farmingItems", "bossing"},
	}
	for id, w := range want {
		b, err := d.Build(id)
		if err != nil {
			t.Fatal(err)
		}
		c := b.Card()
		if got := cardKeys(c); strings.Join(got, ",") != strings.Join(w, ",") {
			t.Errorf("%s card = %v, want %v", id, got, w)
		}
		if len(c.Ratings) != 4 || len(b.Detail().Ratings) != 7 {
			t.Errorf("%s: card has %d ratings, detail %d", id, len(c.Ratings), len(b.Detail().Ratings))
		}
	}
	b, _ := d.Build("paladin_hammerdin")
	if got := b.Card().Ratings[0].Text; got != "Defense: 5 of 5 stars, estimate (Normal 5, Hell 4)" {
		t.Errorf("defense text = %q", got)
	}
	s, _ := d.Build("sorceress_fire")
	if got := s.Card().Ratings[0].Text; got != "Defense: 2 of 5 stars, estimate" {
		t.Errorf("defense text = %q", got)
	}
}

func TestSMEDifferences(t *testing.T) {
	d, _ := Catalog()
	if b, _ := d.Build("necromancer_summoner"); b.DefenseByDifficulty == nil || b.DefenseByDifficulty.Normal != 4 || b.DefenseByDifficulty.Hell != 3 {
		t.Errorf("summoner defense must be N4/H3 (SME)")
	}
	if b, _ := d.Build("sorceress_blizzard"); b.Ratings[RatingIdle].Stars != 3 {
		t.Errorf("blizzard idle must be 3 (SME)")
	}
}

// TestCatalogVariants is the contract of the plain-language variants: every build has exactly one default among its own
// tables, non-empty player text that never shows a table id, and an unknown build or table is an error naming it.
func TestCatalogVariants(t *testing.T) {
	d, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range d.Builds {
		if len(b.Variants) != len(b.Tables) {
			t.Errorf("%s: %d variants for %d tables", b.ID, len(b.Variants), len(b.Tables))
		}
		defaults := 0
		for _, v := range b.Variants {
			if v.Table == b.DefaultVariant {
				defaults++
			}
			if strings.TrimSpace(v.Label) == "" || strings.TrimSpace(v.Why) == "" || strings.Contains(v.Label+v.Why, v.Table) {
				t.Errorf("%s/%s: label %q or why %q is empty or shows the table id", b.ID, v.Table, v.Label, v.Why)
			}
		}
		if defaults != 1 {
			t.Errorf("%s: default %q matches %d variants, want 1", b.ID, b.DefaultVariant, defaults)
		}
		if got := b.RecommendedVariant(1, nil); got != b.DefaultVariant {
			t.Errorf("%s: level 1 recommends %q, want the default %q", b.ID, got, b.DefaultVariant)
		}
	}
	if _, err := d.Build("no_such_build"); err == nil || !strings.Contains(err.Error(), "no_such_build") {
		t.Errorf("unknown build: %v", err)
	}
	b, _ := d.Build("necromancer_summoner")
	if _, err := b.Variant("zz"); err == nil || !strings.Contains(err.Error(), "zz") {
		t.Errorf("unknown table: %v", err)
	}
	// level horizon: the full plan from level 30, only when it is usable
	if got := b.RecommendedVariant(30, nil); got != "necromancer_summoner_endgame90" {
		t.Errorf("level 30 recommends %q", got)
	}
	if got := b.RecommendedVariant(50, func(string) bool { return false }); got != b.DefaultVariant {
		t.Errorf("unusable full plan still recommended: %q", got)
	}
}
