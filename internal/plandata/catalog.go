package plandata

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// CatalogFile is the build text file. It shares the build- prefix with the plan tables but is not one.
const CatalogFile = "build-catalog.json"

// Rating keys. RatingKeys is their fixed order: the detail view order and the tie-break of the small card.
const (
	RatingLeveling     = "leveling"
	RatingFarmingItems = "farmingItems"
	RatingFarmingXP    = "farmingXP"
	RatingBossing      = "bossing"
	RatingDefense      = "defense"
	RatingClearSpeed   = "clearSpeed"
	RatingIdle         = "idle"

	SourceEstimate = "estimate"
	SourceMeasured = "measured"
)

// Text limits so a card fits a phone screen.
const (
	maxIdentity = 80
	maxSummary  = 280
	maxBullet   = 90
	maxNote     = 100
)

var RatingKeys = []string{RatingLeveling, RatingFarmingItems, RatingFarmingXP, RatingBossing, RatingDefense, RatingClearSpeed, RatingIdle}

var ratingLabels = map[string]string{
	RatingLeveling: "Leveling", RatingFarmingItems: "Farming (items)", RatingFarmingXP: "Farming (XP)",
	RatingBossing: "Bossing", RatingDefense: "Defense", RatingClearSpeed: "Clear speed", RatingIdle: "Hands-off",
}

// Rating is one of the seven ratings of a build: 1 to 5 stars, and whether they are an opinion or measured.
type Rating struct {
	Stars  int    `json:"stars"`
	Source string `json:"source"`
	Note   string `json:"note"`
}

// DefenseByDifficulty carries the two defense stars where Normal and Hell differ a lot. Normal equals the plain star.
type DefenseByDifficulty struct {
	Normal int `json:"normal"`
	Hell   int `json:"hell"`
}

// Build is one real build: its text, its ratings and the plan tables that implement it.
type Build struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Class               string               `json:"class"`
	Tables              []string             `json:"tables"`
	Identity            string               `json:"identity"`
	Summary             string               `json:"summary"`
	Pros                []string             `json:"pros"`
	Cons                []string             `json:"cons"`
	Ratings             map[string]Rating    `json:"ratings"`
	DefenseByDifficulty *DefenseByDifficulty `json:"defenseByDifficulty,omitempty"`
	ThresholdsRef       string               `json:"thresholdsRef"`
}

// Uncovered is a plan table that deliberately has no catalogue build, with the reason.
type Uncovered struct {
	Table  string `json:"table"`
	Reason string `json:"reason"`
}

// CatalogDoc is the whole file. Thresholds is data for later measured runs; the code only checks it is present.
type CatalogDoc struct {
	Version    int             `json:"version"`
	Note       string          `json:"note"`
	Thresholds json.RawMessage `json:"thresholds"`
	Builds     []Build         `json:"builds"`
	Uncovered  []Uncovered     `json:"uncovered"`
}

//go:embed build-catalog.json
var catalogRaw []byte

var (
	catalogOnce sync.Once
	catalogDoc  *CatalogDoc
	catalogErr  error
)

// Catalog returns the validated catalogue, parsed once.
func Catalog() (*CatalogDoc, error) {
	catalogOnce.Do(func() { catalogDoc, catalogErr = ParseCatalog(catalogRaw, FS) })
	return catalogDoc, catalogErr
}

// TableIDs lists the plan tables in fsys ("paladin-normal30", ...), sorted; the catalogue file is not one.
func TableIDs(fsys fs.FS) []string {
	ents, _ := fs.ReadDir(fsys, ".")
	var out []string
	for _, e := range ents {
		n := e.Name()
		if n != CatalogFile && strings.HasPrefix(n, "build-") && strings.HasSuffix(n, ".json") {
			out = append(out, strings.TrimSuffix(strings.TrimPrefix(n, "build-"), ".json"))
		}
	}
	sort.Strings(out)
	return out
}

// classIDs are the classes that ship an unlocks table.
func classIDs(fsys fs.FS) map[string]bool {
	ents, _ := fs.ReadDir(fsys, ".")
	out := map[string]bool{}
	for _, e := range ents {
		if n := e.Name(); strings.HasPrefix(n, "unlocks-") && strings.HasSuffix(n, ".json") {
			out[strings.TrimSuffix(strings.TrimPrefix(n, "unlocks-"), ".json")] = true
		}
	}
	return out
}

// ParseCatalog decodes and validates a catalogue against the plan tables and classes in fsys. Every problem is an
// error naming the build, table or class; nothing falls back.
func ParseCatalog(raw []byte, fsys fs.FS) (*CatalogDoc, error) {
	var d CatalogDoc
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}
	tables := map[string]bool{}
	for _, t := range TableIDs(fsys) {
		tables[t] = true
	}
	classes := classIDs(fsys)
	if len(d.Thresholds) == 0 || string(d.Thresholds) == "null" {
		return nil, fmt.Errorf("catalog: no thresholds block")
	}
	owner := map[string]string{}
	ids := map[string]bool{}
	for _, b := range d.Builds {
		if err := b.validate(tables, classes); err != nil {
			return nil, err
		}
		if ids[b.ID] {
			return nil, fmt.Errorf("catalog: duplicate build id %q", b.ID)
		}
		ids[b.ID] = true
		for _, t := range b.Tables {
			if prev := owner[t]; prev != "" {
				return nil, fmt.Errorf("catalog: table %q is in builds %q and %q", t, prev, b.ID)
			}
			owner[t] = b.ID
		}
	}
	for _, u := range d.Uncovered {
		if !tables[u.Table] {
			return nil, fmt.Errorf("catalog: uncovered table %q does not exist", u.Table)
		}
		if strings.TrimSpace(u.Reason) == "" {
			return nil, fmt.Errorf("catalog: uncovered table %q has no reason", u.Table)
		}
		if owner[u.Table] != "" {
			return nil, fmt.Errorf("catalog: table %q is both in build %q and uncovered", u.Table, owner[u.Table])
		}
		owner[u.Table] = "uncovered"
	}
	for t := range tables {
		if owner[t] == "" {
			return nil, fmt.Errorf("catalog: plan table %q has no build text and is not listed as uncovered", t)
		}
	}
	return &d, nil
}

func textOK(what, s string, max int) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s is empty", what)
	}
	if n := utf8.RuneCountInString(s); n > max {
		return fmt.Errorf("%s is %d characters, the limit is %d", what, n, max)
	}
	return nil
}

func (b *Build) validate(tables, classes map[string]bool) error {
	fail := func(err error) error { return fmt.Errorf("catalog: build %q: %w", b.ID, err) }
	if err := textOK("id", b.ID, 40); err != nil {
		return fail(err)
	}
	if err := textOK("name", b.Name, 60); err != nil {
		return fail(err)
	}
	if !classes[b.Class] {
		return fail(fmt.Errorf("unknown class %q", b.Class))
	}
	if len(b.Tables) == 0 {
		return fail(fmt.Errorf("no tables"))
	}
	for _, t := range b.Tables {
		if !tables[t] {
			return fail(fmt.Errorf("unknown plan table %q", t))
		}
	}
	if err := textOK("identity", b.Identity, maxIdentity); err != nil {
		return fail(err)
	}
	if err := textOK("summary", b.Summary, maxSummary); err != nil {
		return fail(err)
	}
	for _, l := range []struct {
		name string
		v    []string
	}{{"pros", b.Pros}, {"cons", b.Cons}} {
		if len(l.v) != 3 {
			return fail(fmt.Errorf("%s has %d entries, want 3", l.name, len(l.v)))
		}
		for _, s := range l.v {
			if err := textOK(l.name+" entry", s, maxBullet); err != nil {
				return fail(err)
			}
		}
	}
	if len(b.Ratings) != len(RatingKeys) {
		return fail(fmt.Errorf("%d ratings, want exactly the seven %v", len(b.Ratings), RatingKeys))
	}
	for _, k := range RatingKeys {
		r, ok := b.Ratings[k]
		if !ok {
			return fail(fmt.Errorf("rating %q missing", k))
		}
		if r.Stars < 1 || r.Stars > 5 {
			return fail(fmt.Errorf("rating %q has %d stars, want 1 to 5", k, r.Stars))
		}
		if r.Source != SourceEstimate && r.Source != SourceMeasured {
			return fail(fmt.Errorf("rating %q source %q is neither %q nor %q", k, r.Source, SourceEstimate, SourceMeasured))
		}
		if err := textOK("rating "+k+" note", r.Note, maxNote); err != nil {
			return fail(err)
		}
	}
	if d := b.DefenseByDifficulty; d != nil {
		if d.Normal < 1 || d.Normal > 5 || d.Hell < 1 || d.Hell > 5 {
			return fail(fmt.Errorf("defenseByDifficulty stars must be 1 to 5"))
		}
		if d.Normal != b.Ratings[RatingDefense].Stars {
			return fail(fmt.Errorf("defense stars %d differ from the Normal value %d", b.Ratings[RatingDefense].Stars, d.Normal))
		}
	}
	if b.ThresholdsRef != "thresholds" {
		return fail(fmt.Errorf("thresholdsRef %q, want %q", b.ThresholdsRef, "thresholds"))
	}
	return nil
}

// RatingView is one rating as the UI shows it. Text is the accessible form: "Defense: 4 of 5 stars, estimate".
type RatingView struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Stars  int    `json:"stars"`
	Source string `json:"source"`
	Note   string `json:"note"`
	Text   string `json:"text"`
}

func ratingView(b *Build, key string) RatingView {
	r := b.Ratings[key]
	text := fmt.Sprintf("%s: %d of 5 stars, %s", ratingLabels[key], r.Stars, r.Source)
	if key == RatingDefense && b.DefenseByDifficulty != nil {
		text += fmt.Sprintf(" (Normal %d, Hell %d)", b.DefenseByDifficulty.Normal, b.DefenseByDifficulty.Hell)
	}
	return RatingView{Key: key, Label: ratingLabels[key], Stars: r.Stars, Source: r.Source, Note: r.Note, Text: text}
}

// CardView is the small card: name, one-line identity and four ratings.
type CardView struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Class    string       `json:"class"`
	Identity string       `json:"identity"`
	Ratings  []RatingView `json:"ratings"`
}

// DetailView is everything for the detail page: all seven ratings in RatingKeys order.
type DetailView struct {
	CardView
	Summary string   `json:"summary"`
	Pros    []string `json:"pros"`
	Cons    []string `json:"cons"`
	Tables  []string `json:"tables"`
	// DefenseByDifficulty is set when Normal and Hell differ a lot (the detail view shows both).
	DefenseByDifficulty *DefenseByDifficulty `json:"defenseByDifficulty,omitempty"`
}

// Card builds the small card: Defense, then the two highest of the other six, then the lowest of the remaining four.
// Ties are broken by RatingKeys order (the earlier key wins among equals, for both the highest and the lowest picks).
func (b *Build) Card() CardView {
	rest := make([]string, 0, len(RatingKeys))
	for _, k := range RatingKeys {
		if k != RatingDefense {
			rest = append(rest, k)
		}
	}
	// stable sort keeps RatingKeys order among equal stars
	sort.SliceStable(rest, func(i, j int) bool { return b.Ratings[rest[i]].Stars > b.Ratings[rest[j]].Stars })
	picked := []string{RatingDefense, rest[0], rest[1]}
	low := rest[2]
	for _, k := range rest[2:] {
		if b.Ratings[k].Stars < b.Ratings[low].Stars {
			low = k
		}
	}
	picked = append(picked, low)
	c := CardView{ID: b.ID, Name: b.Name, Class: b.Class, Identity: b.Identity}
	for _, k := range picked {
		c.Ratings = append(c.Ratings, ratingView(b, k))
	}
	return c
}

// Detail builds the full view with all seven ratings.
func (b *Build) Detail() DetailView {
	d := DetailView{CardView: CardView{ID: b.ID, Name: b.Name, Class: b.Class, Identity: b.Identity},
		Summary: b.Summary, Pros: b.Pros, Cons: b.Cons, Tables: b.Tables, DefenseByDifficulty: b.DefenseByDifficulty}
	for _, k := range RatingKeys {
		d.Ratings = append(d.Ratings, ratingView(b, k))
	}
	return d
}

// Build returns the catalogue build with that id.
func (d *CatalogDoc) Build(id string) (*Build, error) {
	for i := range d.Builds {
		if d.Builds[i].ID == id {
			return &d.Builds[i], nil
		}
	}
	return nil, fmt.Errorf("catalog: no build %q", id)
}

// BuildForTable returns the build that owns a plan table.
func (d *CatalogDoc) BuildForTable(table string) (*Build, error) {
	for i := range d.Builds {
		for _, t := range d.Builds[i].Tables {
			if t == table {
				return &d.Builds[i], nil
			}
		}
	}
	return nil, fmt.Errorf("catalog: no build owns plan table %q", table)
}

// ForClass returns the builds of a class in catalogue order. A class without an unlocks table is an error that names it
// (no fallback to another class); a known class with no catalogue build gives an empty list.
func (d *CatalogDoc) ForClass(class string) ([]*Build, error) {
	c := strings.ToLower(strings.TrimSpace(class))
	if !classIDs(FS)[c] {
		return nil, fmt.Errorf("catalog: unknown class %q", class)
	}
	var out []*Build
	for i := range d.Builds {
		if d.Builds[i].Class == c {
			out = append(out, &d.Builds[i])
		}
	}
	return out, nil
}
