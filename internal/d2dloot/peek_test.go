package d2dloot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/data"
)

const hdr = "name\tcode\ttype\ttype2\t2handed\treqstr\treqdex\tlevelreq\tcost\tinvwidth\tinvheight\tstackable\tgemsockets\tnormcode\tubercode\tultracode\tquest"

// peekDB is a small itemtypes/items table so type chains (helm is armo) are real.
func peekDB(t *testing.T) (*data.DB, map[string]int) {
	t.Helper()
	dir := t.TempDir()
	put := func(name string, rows ...string) {
		if err := os.WriteFile(filepath.Join(dir, name+".tsv"), []byte(strings.Join(rows, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"monstats", "levels", "experience"} {
		put(n, "x")
	}
	put("skills", "auralencalc")
	put("charstats", "class\tstr\tdex", "Sorceress\t10\t25")
	put("itemtypes", "ItemType\tCode\tEquiv1\tEquiv2\tClass\tBody\tBodyLoc1",
		"Armor\tarmo\t\t\t\t\t", "Body Armor\ttors\tarmo\t\t\t1\ttors", "Helm\thelm\tarmo\t\t\t\t",
		"Weapon\tweap\t\t\t\t\t", "Sword\tswor\tweap\t\t\t\t",
		"Ring\tring\t\t\t\t\t", "Amulet\tamul\t\t\t\t\t", "Jewel\tjewl\t\t\t\t\t", "Charm\tchar\t\t\t\t\t", "Small Charm\tscha\tchar\t\t\t\t",
		"Potion\tpota\t\t\t\t\t", "Healing Potion\thpot\tpota\t\t\t\t", "Gold\tgold\t\t\t\t\t")
	put("weapons", hdr, "Short Sword\tssd\tswor\t\t0\t0\t0\t0\t100\t1\t3\t0\t2\tssd\t\t\t0")
	put("armor", hdr, "Quilted Armor\tqui\ttors\t\t0\t12\t0\t0\t64\t2\t3\t0\t4\tqui\t\t\t0", "Cap\tcap\thelm\t\t0\t0\t0\t0\t20\t2\t2\t0\t2\tcap\t\t\t0")
	put("misc", hdr,
		"Ring\trin\tring\t\t0\t0\t0\t0\t100\t1\t1\t0\t0\t\t\t\t0", "Amulet\tamu\tamul\t\t0\t0\t0\t0\t100\t1\t1\t0\t0\t\t\t\t0",
		"Jewel\tjew\tjewl\t\t0\t0\t0\t0\t100\t1\t1\t0\t0\t\t\t\t0", "Small Charm\tcm1\tscha\t\t0\t0\t0\t0\t100\t1\t1\t0\t0\t\t\t\t0",
		"Minor Healing Potion\thp1\thpot\t\t0\t0\t0\t0\t30\t1\t1\t0\t0\t\t\t\t0", "Gold\tgld\tgold\t\t0\t0\t0\t0\t0\t1\t1\t1\t0\t\t\t\t0")
	put("hireling", "Hireling\tSubType\tId\tClass\tAct\tDifficulty\tLevel\tStr\tDex\tWType1\tWType2", "Rogue Scout\tFire\t1\t271\t1\t1\t1\t30\t30\tbow\t")
	for n, c := range map[string]string{"stamp.json": "{}", "strings.tsv": ""} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	db, err := data.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	cls := map[string]int{}
	for i, b := range db.Items {
		if b != nil {
			cls[b.Code] = i
		}
	}
	return db, cls
}

// countPeek counts calls and serves a fixed floor through the real wire parser.
type countPeek struct {
	lines   []string
	caps    d2dctl.VaultCaps
	capsErr error
	err     error
	ground  atomic.Int32
	capsN   atomic.Int32
}

func (c *countPeek) VaultCaps(context.Context) (d2dctl.VaultCaps, error) {
	c.capsN.Add(1)
	return c.caps, c.capsErr
}

func (c *countPeek) GroundFull(context.Context, int) ([]d2dctl.VaultRecord, error) {
	c.ground.Add(1)
	if c.err != nil {
		return nil, c.err
	}
	var out []d2dctl.VaultRecord
	for _, l := range c.lines {
		r, err := d2dctl.ParseVaultLine(l)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func line(id int, code string, q int, ident bool, ilvl int) string {
	r := d2dctl.FakeRecord(id, code, q, ilvl, "")
	r.Identified, r.HasPos, r.X, r.Y = ident, true, 10, 20
	return d2dctl.FormatVaultLine(r)
}

type fl struct {
	id    uint32
	code  string
	label string
}

func floor(cls map[string]int, in ...fl) (*d2dsrc.Result, []uint32) {
	snap := &d2.Snapshot{Player: &d2.Player{X: 1, Y: 1}}
	r := &d2dsrc.Result{Snap: snap}
	r.Extras.Ground = map[uint32]d2dsrc.ItemExtra{}
	var ids []uint32
	for _, f := range in {
		snap.Ground = append(snap.Ground, d2.Unit{Type: d2.TypeItem, ID: f.id, Class: cls[f.code], Mode: d2.ItemModeGround, X: 5, Y: 5, Item: &d2.Item{Where: d2.ContGround}})
		r.Extras.Ground[f.id] = d2dsrc.ItemExtra{Code: f.code, Label: f.label}
		ids = append(ids, f.id)
	}
	return r, ids
}

func qOf(r *d2dsrc.Result, id uint32) *d2.Item {
	for _, u := range r.Snap.Ground {
		if u.ID == id {
			return u.Item
		}
	}
	return nil
}

func TestPeekRealQuality(t *testing.T) {
	db, cls := peekDB(t)
	cp := &countPeek{caps: d2dctl.VaultCaps{GroundFull: true}, lines: []string{
		line(1, "rin", d2.QualityRare, false, 61),  // unidentified rare ring: label shows only "Ring"
		line(2, "cap", d2.QualityMagic, false, 30), // unidentified magic helm
		line(3, "qui", d2.QualityNormal, true, 12), // white armor
		line(4, "amu", d2.QualityMagic, true, 40),
		line(5, "cm1", d2.QualityMagic, true, 20),
	}}
	s := (&Source{DB: db}).UsePeek(cp)
	r, _ := floor(cls, fl{1, "rin", "Ring"}, fl{2, "cap", "Cap"}, fl{3, "qui", "Quilted Armor"}, fl{4, "amu", "Amulet"}, fl{5, "cm1", "Small Charm"},
		fl{6, "hp1", "Minor Healing Potion"})
	s.Enrich(r)
	for id, want := range map[uint32]int{1: d2.QualityRare, 2: d2.QualityMagic, 3: d2.QualityNormal, 4: d2.QualityMagic, 5: d2.QualityMagic} {
		if got := qOf(r, id).Quality; got != want {
			t.Errorf("unit %d quality %d, want %d", id, got, want)
		}
	}
	if it := qOf(r, 1); it.Identified() || it.ILvl != 61 {
		t.Errorf("ring: identified %v ilvl %d, want false 61", it.Identified(), it.ILvl)
	}
	if !qOf(r, 4).Identified() {
		t.Error("identified flag not carried")
	}
	if got := qOf(r, 6).Quality; got != d2.QualityNormal {
		t.Errorf("potion quality %d", got)
	}
	// Second pass, same floor: cached, no new call. Bounded: one ground call and one caps call so far.
	r2, _ := floor(cls, fl{1, "rin", "Ring"}, fl{2, "cap", "Cap"})
	s.Enrich(r2)
	if cp.ground.Load() != 1 || cp.capsN.Load() != 1 {
		t.Errorf("calls ground=%d caps=%d, want 1 and 1", cp.ground.Load(), cp.capsN.Load())
	}
	if qOf(r2, 1).Quality != d2.QualityRare {
		t.Error("cache lost the ring")
	}
}

func TestPeekNoCallForPotionsAndGold(t *testing.T) {
	db, cls := peekDB(t)
	cp := &countPeek{caps: d2dctl.VaultCaps{GroundFull: true}}
	s := (&Source{DB: db}).UsePeek(cp)
	r, _ := floor(cls, fl{1, "hp1", "Minor Healing Potion"}, fl{2, "gld", "12 Gold"})
	s.Enrich(r)
	if cp.ground.Load() != 0 || cp.capsN.Load() != 0 {
		t.Errorf("calls ground=%d caps=%d for potions and gold", cp.ground.Load(), cp.capsN.Load())
	}
}

// Must-fail guard: without the verb a ring, amulet, jewel or charm is never "normal"; plain gear keeps the label guess.
func TestNoVerbNeverNormal(t *testing.T) {
	db, cls := peekDB(t)
	cases := map[string]*countPeek{
		"no peeker":      nil,
		"caps lack verb": {caps: d2dctl.VaultCaps{}},
		"unsupported":    {caps: d2dctl.VaultCaps{GroundFull: true}, err: d2dctl.ErrVaultUnsupported},
		"off":            {caps: d2dctl.VaultCaps{GroundFull: true}, err: d2dctl.ErrVaultOff},
		"transient":      {caps: d2dctl.VaultCaps{GroundFull: true}, err: errors.New("boom")},
		"caps error":     {capsErr: errors.New("boom")},
	}
	for name, cp := range cases {
		s := &Source{DB: db}
		if cp != nil {
			s.UsePeek(cp)
		}
		r, _ := floor(cls, fl{1, "rin", "Ring"}, fl{2, "amu", "Amulet"}, fl{3, "jew", "Jewel"}, fl{4, "cm1", "Small Charm"},
			fl{5, "qui", "Quilted Armor"}, fl{6, "cap", "Cap"})
		s.Enrich(r)
		for _, id := range []uint32{1, 2, 3, 4} {
			if q := qOf(r, id).Quality; q == d2.QualityNormal {
				t.Errorf("%s: unit %d is normal", name, id)
			}
		}
		if q := qOf(r, 5).Quality; q != d2.QualityNormal {
			t.Errorf("%s: white armor quality %d, want normal", name, q)
		}
		if q := qOf(r, 6).Quality; q != d2.QualityNormal {
			t.Errorf("%s: cap quality %d, want the label guess (normal)", name, q)
		}
	}
}

func TestPeekVerbMissingIsFinal(t *testing.T) {
	db, cls := peekDB(t)
	cp := &countPeek{caps: d2dctl.VaultCaps{}}
	s := (&Source{DB: db}).UsePeek(cp)
	for i := 0; i < 3; i++ {
		r, _ := floor(cls, fl{1, "rin", "Ring"})
		s.Enrich(r)
	}
	if cp.capsN.Load() != 1 || cp.ground.Load() != 0 {
		t.Errorf("caps=%d ground=%d: a missing verb must be asked once", cp.capsN.Load(), cp.ground.Load())
	}
}

func TestPeekTransientRetriesAndMissingUnitCached(t *testing.T) {
	db, cls := peekDB(t)
	cp := &countPeek{caps: d2dctl.VaultCaps{GroundFull: true}, err: errors.New("boom")}
	s := (&Source{DB: db}).UsePeek(cp)
	r, _ := floor(cls, fl{1, "rin", "Ring"})
	s.Enrich(r)
	cp.err = nil
	cp.lines = []string{line(1, "rin", d2.QualityMagic, false, 5)}
	r, _ = floor(cls, fl{1, "rin", "Ring"}, fl{2, "cap", "Cap"}) // unit 2 is not in the reply
	s.Enrich(r)
	if qOf(r, 1).Quality != d2.QualityMagic {
		t.Errorf("ring after the retry: %d", qOf(r, 1).Quality)
	}
	n := cp.ground.Load()
	r, _ = floor(cls, fl{1, "rin", "Ring"}, fl{2, "cap", "Cap"})
	s.Enrich(r)
	if cp.ground.Load() != n {
		t.Error("a unit the verb did not return was asked for again")
	}
}
