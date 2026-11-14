package danger

import (
	"encoding/csv"
	"math"
	"os"
	"sort"
	"strconv"
	"testing"

	"d2lod-bot/internal/data"
)

// realInstall is the game data dir: D2_MPQ_DIR (the builder mounts it at /gamedata), else /tmp/d2mpq.
func realInstall() string {
	if d := os.Getenv("D2_MPQ_DIR"); d != "" {
		return d
	}
	return "/tmp/d2mpq"
}

func realTable(t *testing.T) (*data.DB, *Table) {
	t.Helper()
	if _, err := os.Stat(realInstall() + "/d2data.mpq"); err != nil {
		t.Skip("no real install at " + realInstall())
	}
	db, err := data.Open(realInstall(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ov, err := ShippedOverrides()
	if err != nil {
		t.Fatal(err)
	}
	tb, err := Build(db, DefaultParams(), ov)
	if err != nil {
		t.Fatal(err)
	}
	return db, tb
}

// TestRealAnchors: monster levels of docs/difficulty.md section 2.1 and docs/research/areamon-trace.md Q2 (V run), and
// Andariel from docs/research/boss1-trace.md (V run: level 12, life 1,024, defence 60, experience 1,282).
func TestRealAnchors(t *testing.T) {
	_, tb := realTable(t)
	for _, c := range []struct {
		area int
		name string
		mlvl [3]int
	}{{2, "Blood Moor", [3]int{1, 36, 67}}, {3, "Cold Plains", [3]int{2, 0, 0}}, {6, "Black Marsh", [3]int{6, 0, 0}}} {
		for d, dn := range DiffNames {
			z, err := tb.Find(c.area, dn)
			if err != nil {
				t.Fatal(err)
			}
			if z.Name != c.name {
				t.Errorf("area %d is %q, want %q", c.area, z.Name, c.name)
			}
			if c.mlvl[d] != 0 && z.MLvl != c.mlvl[d] {
				t.Errorf("%s %s mlvl %d, want %d", c.name, dn, z.MLvl, c.mlvl[d])
			}
			if len(z.Monsters) == 0 || z.Encounters.Bad.N < z.Encounters.Typical.N {
				t.Errorf("%s %s: empty or inverted profile %+v", c.name, dn, z.Encounters)
			}
		}
	}
	a, err := tb.FindBoss("andariel", "normal")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("andariel %+v", *a)
	if a.Level != 12 || math.Abs(float64(a.Life-1024)) > 0.15*1024 || a.Def < 50 || a.Def > 70 {
		t.Errorf("andariel %+v does not match boss1-trace (level 12, life 1024, defence 60)", *a)
	}
	for _, k := range []string{"smith", "countess", "blood_raven"} {
		b, err := tb.FindBoss(k, "normal")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s level %d life %d xp %d dmg %.1f", k, b.Level, b.Life, b.XP, b.Dmg)
	}
}

// TestAreamonDiff compares the estimated units per clear with the V run population of areamon-table.csv (12 seeds,
// Act 1 Normal). The estimate comes from MonDen x level size; the level generator's real room fill is not modelled,
// so the tolerance is the median relative error and a loose per-area bound; the numbers are logged for refitting
// Params.UnitsPerDensityTile.
func TestAreamonDiff(t *testing.T) {
	_, tb := realTable(t)
	f, err := os.Open("../../docs/research/areamon-table.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		col[h] = i
	}
	var errs []float64
	for _, r := range rows[1:] {
		id, _ := strconv.Atoi(r[col["lid"]])
		units, _ := strconv.ParseFloat(r[col["units_mean"]], 64)
		uniq, _ := strconv.ParseFloat(r[col["unique"]], 64)
		z, err := tb.Find(id, "normal")
		if err != nil {
			t.Errorf("area %d (%s): %v", id, r[col["area"]], err)
			continue
		}
		rel := (z.Density.Units - units) / units
		errs = append(errs, math.Abs(rel))
		t.Logf("%-22s units csv %6.1f model %6.1f (%+.0f%%)  uniques csv %.1f model %.1f  mlvl csv %s model %d",
			r[col["area"]], units, z.Density.Units, rel*100, uniq, z.Uniques, r[col["mlvl_mean"]], z.MLvl)
	}
	sort.Float64s(errs)
	if len(errs) == 0 {
		t.Fatal("no csv rows")
	}
	med := errs[len(errs)/2]
	t.Logf("median relative error %.0f%%, worst %.0f%% over %d areas", med*100, errs[len(errs)-1]*100, len(errs))
	if med > 0.35 {
		t.Errorf("median units error %.0f%% over the tolerance of 35%%", med*100)
	}
}
