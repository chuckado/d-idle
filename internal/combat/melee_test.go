package combat

import (
	"os"
	"path/filepath"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/resource"
)

// meleeTables is a cache with two classes and a few weapons. The numbers are synthetic (U): the test checks the wiring,
// not the game's tables.
func meleeTables(t *testing.T) *data.DB {
	t.Helper()
	dir := t.TempDir()
	tab := map[string]string{
		"charstats": "class\tToHitFactor\nSorceress\t5\nPaladin\t30\nExpansion\t\n",
		"itemtypes": "ItemType\tCode\tEquiv1\tEquiv2\nStaff\tstaf\trod\t\nRod\trod\tblun\t\nBlunt\tblun\tmele\t\n" +
			"Sword\tswor\tmele\t\nMelee\tmele\tweap\t\nBow\tbow\tmiss\t\nMissile\tmiss\tweap\t\nAny\tweap\t\t\n",
		"weapons": "name\ttype\tcode\tmindam\tmaxdam\tStrBonus\tDexBonus\tthrowable\n" +
			"Staff\tstaf\tstf\t1\t5\t100\t0\t0\nLong Sword\tswor\tlsd\t3\t10\t100\t0\t0\nShort Bow\tbow\tsbw\t1\t4\t0\t0\t0\n" +
			"Throwing Knife\tswor\ttkf\t2\t3\t0\t0\t1\n",
	}
	for _, n := range []string{"monstats", "levels", "skills", "experience", "armor", "misc"} {
		tab[n] = "x\n"
	}
	tab["skills"] = "auralencalc\n"
	for n, c := range tab {
		if err := os.WriteFile(filepath.Join(dir, n+".tsv"), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for n, c := range map[string]string{"stamp.json": "{}", "strings.tsv": ""} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	db, err := data.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func meleePlayer(class d2.Class, lvl, str, dex int, mind, maxd int32) *d2.Player {
	return &d2.Player{Class: class, Level: lvl, Strength: str, Dexterity: dex,
		Stats: d2.Stats{statMinDamage: mind, statMaxDamage: maxd}}
}

func wielding(db *data.DB, code string) *d2.Inventory {
	b := db.ItemByCode(code)
	return &d2.Inventory{Items: []d2.Unit{{Class: b.Index, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyRHand}}}}
}

func foeTarget(ac, lvl int, undead bool) Target {
	m := &data.Monster{LUndead: undead}
	m.AC[d2.Normal], m.Level[d2.Normal] = ac, lvl
	return Target{Mon: m, Dist: 5}
}

func TestMeleeBlowStaffCasterAndSwordFighter(t *testing.T) {
	db := meleeTables(t)
	// weapons.txt order is the Items index: staff 0, sword 1, bow 2, knife 3 (the fixture has no other weapon rows).
	sor := meleeBlow(meleePlayer(d2.Sorceress, 8, 10, 25, 1, 5), wielding(db, "stf"), db)
	if sor == nil {
		t.Fatal("staff caster: no model")
	}
	want := resource.Blow{AR: resource.AttackRating(0, 25, 5), Level: 8, Weapon: true, Min: 1, Max: 5, Str: 10,
		DmgPct: resource.StatDmgPct(0, 10, 25, 100, 0, 0), Blunt: true}
	if *sor != want {
		t.Errorf("staff caster blow %+v, want %+v", *sor, want)
	}
	pal := meleePlayer(d2.Paladin, 12, 30, 20, 3, 10)
	pal.Stats[statToHit], pal.Stats[statDmgPct], pal.Stats[statToHitPct], pal.Stats[statDmgFlat] = 40, 15, 20, 2
	sw := meleeBlow(pal, wielding(db, "lsd"), db)
	want = resource.Blow{AR: 40 + (20-7)*5 + 30, ARPct: 20, Level: 12, Weapon: true, Min: 3, Max: 10, Flat: 2, Str: 30,
		DmgPct: 15 + 30, Blunt: false}
	if sw == nil || *sw != want {
		t.Errorf("sword fighter blow %+v, want %+v", sw, want)
	}
	// Same living pack, same level: the sword fighter's higher to-hit factor and gear show up as a better hit chance;
	// against an undead pack only the staff gains the Blunt term.
	if sw.HitPct(resource.Foe{Defense: 150, Level: 12}) <= sor.HitPct(resource.Foe{Defense: 150, Level: 12}) {
		t.Error("the better attack rating must hit more often")
	}
}

func TestMeleeBlowNilWhenAnInputIsMissing(t *testing.T) {
	db := meleeTables(t)
	p := func() *d2.Player { return meleePlayer(d2.Sorceress, 8, 10, 25, 1, 5) }
	noStats := p()
	noStats.Stats = nil
	noDmg := p()
	noDmg.Stats = d2.Stats{statToHit: 5}
	unknownClass := p()
	unknownClass.Class = d2.Assassin // not in the fixture
	bad := &d2.Inventory{Items: []d2.Unit{{Class: 999, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyRHand}}}}
	cases := map[string]*resource.Blow{
		"no player":         meleeBlow(nil, wielding(db, "stf"), db),
		"no inventory":      meleeBlow(p(), nil, db),
		"no tables":         meleeBlow(p(), wielding(db, "stf"), nil),
		"no stat vector":    meleeBlow(noStats, wielding(db, "stf"), db),
		"no weapon dmg":     meleeBlow(noDmg, wielding(db, "stf"), db),
		"unknown class":     meleeBlow(unknownClass, wielding(db, "stf"), db),
		"unknown weapon":    meleeBlow(p(), bad, db),
		"bow":               meleeBlow(p(), wielding(db, "sbw"), db),
		"thrown weapon":     meleeBlow(p(), wielding(db, "tkf"), db),
		"armor row as hand": meleeBlow(p(), &d2.Inventory{Items: []d2.Unit{{Class: len(db.Items), Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyRHand}}}}, db),
	}
	for name, b := range cases {
		if b != nil {
			t.Errorf("%s: want nil, got %+v", name, *b)
		}
	}
	// Barehanded with a readable inventory is a model (trace D4), not an unknown.
	if b := meleeBlow(p(), &d2.Inventory{}, db); b == nil || b.Weapon {
		t.Errorf("barehanded: %+v", b)
	}
}

func TestMeleeFoeUndeadAndHardest(t *testing.T) {
	db := meleeTables(t)
	p := meleePlayer(d2.Sorceress, 8, 10, 25, 1, 5)
	p.Difficulty = d2.Normal
	b := meleeBlow(p, wielding(db, "stf"), db)
	living := []Target{foeTarget(15, 3, false), foeTarget(60, 6, false)}
	if f := meleeFoe(p, living, 30, b); f == nil || f.Undead || f.Defense != 60 || f.Level != 6 {
		t.Errorf("living pack: %+v", f)
	}
	dead := []Target{foeTarget(100, 3, true), foeTarget(200, 4, true)}
	if f := meleeFoe(p, dead, 30, b); f == nil || !f.Undead {
		t.Errorf("undead pack: %+v", f)
	}
	mixed := append(append([]Target{}, dead...), foeTarget(10, 3, false))
	if f := meleeFoe(p, mixed, 30, b); f == nil || f.Undead {
		t.Errorf("a living member makes the pack not undead: %+v", f)
	}
	far := foeTarget(500, 99, false)
	far.Dist = 99
	if f := meleeFoe(p, append([]Target{far}, dead...), 30, b); f == nil || f.Defense != 200 {
		t.Errorf("a far hostile is outside the radius: %+v", f)
	}
	if meleeFoe(p, nil, 30, b) != nil || meleeFoe(p, []Target{far}, 30, b) != nil {
		t.Error("nobody in range: nil")
	}
	if meleeFoe(p, []Target{{Dist: 5}}, 30, b) != nil {
		t.Error("an unknown class: nil")
	}
	if meleeFoe(p, dead, 30, nil) != nil {
		t.Error("no blow: nil")
	}
}

// resInput fills Blow and Foe only when the tables, the wielded weapon and a known pack are all there; otherwise the
// input is exactly what it was before the wiring.
func TestResInputFeedsTheMeleeModel(t *testing.T) {
	db := meleeTables(t)
	h := newHarness(t)
	p := meleePlayer(d2.Sorceress, 8, 10, 25, 1, 5)
	snap := &d2.Snapshot{Player: p, Inventory: wielding(db, "stf")}
	hs := []Target{foeTarget(15, 3, true), foeTarget(15, 3, true)}
	if in := h.f.resInput(snap, hs, 2, nil); in.Blow != nil || in.Foe != nil {
		t.Fatalf("no tables: model fed %+v %+v", in.Blow, in.Foe)
	}
	h.f.e.Data = db
	in := h.f.resInput(snap, hs, 2, nil)
	if in.Blow == nil || in.Foe == nil || !in.Blow.Blunt || !in.Foe.Undead {
		t.Fatalf("staff vs undead pack: %+v %+v", in.Blow, in.Foe)
	}
	if in := h.f.resInput(snap, nil, 0, nil); in.Blow != nil || in.Foe != nil {
		t.Error("no hostile in range: model not fed")
	}
	if in := h.f.resInput(snap, []Target{{Dist: 5}}, 1, nil); in.Blow != nil || in.Foe != nil {
		t.Error("unknown monster class: model not fed, never a half-fed pair")
	}
	snap.Inventory = nil
	if in := h.f.resInput(snap, hs, 2, nil); in.Blow != nil || in.Foe != nil {
		t.Error("no inventory: model not fed")
	}
}
