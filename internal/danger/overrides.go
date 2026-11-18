package danger

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed overrides.json
var overridesJSON []byte

// MaxPerClass is the override budget per class (user decision 2026-11-14, docs/survival/decisions-2026-11-14.md).
const MaxPerClass = 5

// OverrideRow replaces one number of the generated table where the tables cannot know it. Paths: density.units,
// density.perScreen, uniquesPerClear, encounters.{typical,bad,elite}.{n,melee,ranged}, or for a boss row (Boss set)
// life, level, minions. Every row names its evidence file and an expiry date (YYYY-MM-DD).
type OverrideRow struct {
	ID       string  `json:"id"`
	Area     int     `json:"area,omitempty"`
	Boss     string  `json:"boss,omitempty"`
	Diff     string  `json:"diff"`
	Class    string  `json:"class,omitempty"` // "" applies to every class
	Path     string  `json:"path"`
	Value    float64 `json:"value"`
	Evidence string  `json:"evidence"` // repo-relative file
	Tag      string  `json:"tag"`
	Expires  string  `json:"expires"`
}

// Overrides is the parsed file.
type Overrides struct {
	Schema string        `json:"schema"`
	Rows   []OverrideRow `json:"rows"`
	// Now is the clock for expiry checks; nil means time.Now.
	Now func() time.Time `json:"-"`
}

// ShippedOverrides parses the embedded override file.
func ShippedOverrides() (*Overrides, error) { return ParseOverrides(overridesJSON) }

// ParseOverrides reads and checks a file: known fields only, unique ids, evidence, tag, valid expiry, the per-class cap.
func ParseOverrides(b []byte) (*Overrides, error) {
	var o Overrides
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&o); err != nil {
		return nil, fmt.Errorf("danger overrides: %w", err)
	}
	if o.Schema != "danger.overrides/1" {
		return nil, fmt.Errorf("danger overrides: schema %q", o.Schema)
	}
	ids := map[string]bool{}
	perClass := map[string]int{}
	for _, r := range o.Rows {
		switch {
		case r.ID == "" || ids[r.ID]:
			return nil, fmt.Errorf("danger overrides: row id %q is empty or repeated", r.ID)
		case r.Evidence == "":
			return nil, fmt.Errorf("danger overrides: row %s has no evidence", r.ID)
		case r.Tag != TagV && r.Tag != TagI && r.Tag != TagU:
			return nil, fmt.Errorf("danger overrides: row %s: tag %q is not V, I or U", r.ID, r.Tag)
		case r.Boss == "" && r.Area == 0:
			return nil, fmt.Errorf("danger overrides: row %s names neither area nor boss", r.ID)
		case !validDiff(r.Diff):
			return nil, fmt.Errorf("danger overrides: row %s: unknown difficulty %q", r.ID, r.Diff)
		}
		if _, err := time.Parse("2006-01-02", r.Expires); err != nil {
			return nil, fmt.Errorf("danger overrides: row %s: expires %q: %w", r.ID, r.Expires, err)
		}
		ids[r.ID] = true
		perClass[r.Class]++
		if perClass[r.Class] > MaxPerClass {
			return nil, fmt.Errorf("danger overrides: class %q has more than %d rows", r.Class, MaxPerClass)
		}
	}
	return &o, nil
}

func validDiff(d string) bool {
	for _, n := range DiffNames {
		if n == d {
			return true
		}
	}
	return false
}

// CheckEvidence fails for a row whose evidence file does not exist under root (the repo root).
func (o *Overrides) CheckEvidence(root string) error {
	for _, r := range o.Rows {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(r.Evidence))); err != nil {
			return fmt.Errorf("danger overrides: row %s: evidence %s: %w", r.ID, r.Evidence, err)
		}
	}
	return nil
}

func (o *Overrides) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now()
}

// apply writes every unexpired row into t. An expired row is skipped; a row that matches nothing is an error.
func (o *Overrides) apply(t *Table) error {
	for _, r := range o.Rows {
		exp, _ := time.Parse("2006-01-02", r.Expires)
		if o.now().After(exp.Add(24 * time.Hour)) {
			continue
		}
		if err := r.applyTo(t); err != nil {
			return err
		}
		t.Overrides = append(t.Overrides, r.ID)
	}
	return nil
}

// Expired lists the rows past their date, so a test can force a review.
func (o *Overrides) Expired() []string {
	var out []string
	for _, r := range o.Rows {
		exp, _ := time.Parse("2006-01-02", r.Expires)
		if o.now().After(exp.Add(24 * time.Hour)) {
			out = append(out, r.ID)
		}
	}
	return out
}

func (r OverrideRow) applyTo(t *Table) error {
	if r.Boss != "" {
		b, err := t.FindBoss(r.Boss, r.Diff)
		if err != nil {
			return fmt.Errorf("danger override %s: %w", r.ID, err)
		}
		switch r.Path {
		case "life":
			b.Life = int(r.Value)
		case "level":
			b.Level = int(r.Value)
		case "minions":
			b.Minions = int(r.Value)
		default:
			return fmt.Errorf("danger override %s: boss path %q is not life, level or minions", r.ID, r.Path)
		}
		return nil
	}
	z, err := t.Find(r.Area, r.Diff)
	if err != nil {
		return fmt.Errorf("danger override %s: %w", r.ID, err)
	}
	enc := map[string]*Encounter{"typical": &z.Encounters.Typical, "bad": &z.Encounters.Bad, "elite": &z.Encounters.Elite}
	switch parts := strings.Split(r.Path, "."); {
	case r.Path == "density.units":
		z.Density.Units = r.Value
	case r.Path == "density.perScreen":
		z.Density.PerScreen = r.Value
	case r.Path == "uniquesPerClear":
		z.Uniques = r.Value
	case len(parts) == 3 && parts[0] == "encounters" && enc[parts[1]] != nil:
		e := enc[parts[1]]
		switch parts[2] {
		case "n":
			e.N = int(r.Value)
		case "melee":
			e.Melee = int(r.Value)
		case "ranged":
			e.Ranged = int(r.Value)
		default:
			return fmt.Errorf("danger override %s: unknown path %q", r.ID, r.Path)
		}
	default:
		return fmt.Errorf("danger override %s: unknown path %q", r.ID, r.Path)
	}
	if z.Source == nil {
		z.Source = map[string]string{}
	}
	z.Source["override:"+r.ID] = r.Tag
	return nil
}
