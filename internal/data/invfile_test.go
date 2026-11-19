package data

import (
	"os"
	"path/filepath"
	"testing"
)

// Fixture rows are made up: a weapon with a unique and a set icon, an armor with none, a misc rune and gem,
// uniques with and without an override.
func invDB() *DB {
	return &DB{
		Items: []*ItemBase{
			{Code: "ssd", InvFile: "invSsd", UniqueInvFile: "invSsdU", SetInvFile: "invSsdS"},
			{Code: "cap", InvFile: "invCap"},
			{Code: "r01", InvFile: "invrEl"},
			{Code: "gpv", InvFile: "invgsve"},
			{Code: "nofile"},
		},
		Uniques: []*UniqueItem{
			{Row: 0, Code: "ssd", InvFile: "invSsdX"},
			{Row: 1, Code: "ssd"},
			{Row: 2, Code: "cap", InvFile: "invCapX"},
			{Row: 4, Code: "cap", InvFile: "invCapY"}, // row 3 is the "Expansion" separator, not loaded
		},
		SetItems: []*SetItem{{Row: 0, Code: "ssd", InvFile: "invSetOverride"}, {Row: 1, Code: "ssd"}},
	}
}

func TestInvFile(t *testing.T) {
	db := invDB()
	for _, tc := range []struct {
		name  string
		class int
		q     InvQuality
		idx   int
		want  string
		ok    bool
	}{
		{"weapon plain", 0, InvPlain, 0, "invssd", true},
		{"weapon plain ignores index", 0, InvPlain, 2, "invssd", true},
		{"unique with row override", 0, InvUnique, 0, "invssdx", true},
		{"unique without override uses base uniqueinvfile", 0, InvUnique, 1, "invssdu", true},
		{"unique row of another base is ignored", 0, InvUnique, 2, "invssdu", true},
		{"unique, unknown row", 0, InvUnique, 99, "invssdu", true},
		{"armor unique falls back to invfile", 1, InvUnique, 3, "invcap", true},
		{"armor unique override after separator", 1, InvUnique, 4, "invcapy", true},
		{"set with row override", 0, InvSet, 0, "invsetoverride", true},
		{"set base setinvfile", 0, InvSet, 1, "invssds", true},
		{"armor set falls back", 1, InvSet, 0, "invcap", true},
		{"rune lowercased", 2, InvPlain, 0, "invrel", true},
		{"gem", 3, InvPlain, 0, "invgsve", true},
		{"no icon name", 4, InvPlain, 0, "", false},
		{"unknown class", 99, InvPlain, 0, "", false},
		{"negative class", -1, InvPlain, 0, "", false},
	} {
		got, ok := db.InvFile(tc.class, tc.q, tc.idx)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: %q %v, want %q %v", tc.name, got, ok, tc.want, tc.ok)
		}
	}
	var nilDB *DB
	if _, ok := nilDB.InvFile(0, InvPlain, 0); ok {
		t.Error("nil db")
	}
}

// The columns reach the structs through a real extraction, including the schema bump that invalidates old
// caches (an old cache has no invfile column).
func TestInvFileColumns(t *testing.T) {
	inst, cache := t.TempDir(), filepath.Join(t.TempDir(), "cache")
	writeInstall(t, inst, func(patch, exp, base map[string][]byte) {
		exp[excelDir+"weapons.txt"] = []byte("name\ttype\tcode\tnamestr\tinvwidth\tinvheight\tinvfile\tuniqueinvfile\tsetinvfile\n" +
			"Short Sword\tswor\tssd\tssd\t1\t3\tinvssd\tinvssdu\tinvssds\n")
		exp[excelDir+"misc.txt"] = []byte("name\ttype\tcode\tnamestr\tinvwidth\tinvheight\tinvfile\tuniqueinvfile\n" +
			"El Rune\trune\tr01\tr01\t1\t1\tinvrEl\t\n")
		exp[excelDir+"uniqueitems.txt"] = []byte("index\tenabled\tlvl\tlvl req\tcode\tinvfile\nThe Gnasher\t1\t1\t1\tssd\tinvgnash\nExpansion\t\t\t\t\t\nHarlequin\t1\t62\t62\tssd\t\n")
	})
	db, err := Open(inst, cache)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := db.InvFile(0, InvUnique, 0); got != "invgnash" {
		t.Errorf("unique 0 icon %q", got)
	}
	if got, _ := db.InvFile(0, InvUnique, 2); got != "invssdu" { // the separator row counts: Harlequin is row 2
		t.Errorf("unique 2 icon %q", got)
	}
	if got, ok := db.InvFile(3, InvPlain, 0); !ok || got != "invrel" {
		t.Errorf("rune %q %v", got, ok)
	}
	if _, err := os.Stat(filepath.Join(cache, "weapons.tsv")); err != nil {
		t.Fatal(err)
	}
}

// Different gem, rune and charm bases must each keep their own icon: a lookup that collapses many items to one image
// (the stash showed one egg for everything) fails here. Jewels legitimately share one invfile; they differ by colour, not art.
func TestInvFileDistinctPerBase(t *testing.T) {
	db := &DB{Items: []*ItemBase{
		{Code: "gpv", InvFile: "invgsv"}, {Code: "gpb", InvFile: "invgsb"}, {Code: "r01", InvFile: "invr01"}, {Code: "r02", InvFile: "invr02"},
		{Code: "cm1", InvFile: "invch1"}, {Code: "cm2", InvFile: "invch2"}, {Code: "jew", InvFile: "invjw"},
	}}
	seen := map[string]string{}
	for i, b := range db.Items {
		for _, q := range []InvQuality{InvPlain, InvUnique, InvSet} {
			n, ok := db.InvFile(i, q, 0)
			if !ok || n != b.InvFile {
				t.Fatalf("%s q=%d: got %q ok=%v want %q", b.Code, q, n, ok, b.InvFile)
			}
		}
		if o, dup := seen[b.InvFile]; dup {
			t.Fatalf("%s and %s share icon %s", o, b.Code, b.InvFile)
		}
		seen[b.InvFile] = b.Code
	}
}

// Made-up types with the real column shape: jewels 6 pictures, rings 5, charms 3 (shared by two sizes), a type with none.
func picDB() *DB {
	return &DB{
		ItemTypes: []*ItemType{
			{Code: "jewl", InvGfx: [6]string{"invjw1", "invjw2", "invjw3", "invjw4", "invjw5", "invjw6"}},
			{Code: "ring", InvGfx: [6]string{"invrin1", "invrin2", "invrin3", "invrin4", "invrin5"}},
			{Code: "scha", InvGfx: [6]string{"invch1", "invch4", "invch7"}},
			{Code: "swor"},
		},
		Items: []*ItemBase{
			{Code: "jew", Type: "jewl", InvFile: "invgswe"},
			{Code: "rin", Type: "ring", InvFile: "invrin"},
			{Code: "cm1", Type: "scha", InvFile: "invchm"},
			{Code: "ssd", Type: "swor", InvFile: "invSsd"},
		},
		Uniques: []*UniqueItem{{Row: 0, Code: "rin", InvFile: "invsoj"}},
	}
}

func TestInvFilePic(t *testing.T) {
	db := picDB()
	seen := map[string]int{}
	for pic := 0; pic < 6; pic++ {
		n, ok := db.InvFilePic(0, InvPlain, 0, pic)
		if !ok || n != "invjw"+string(rune('1'+pic)) {
			t.Errorf("jewel pic %d: %q %v", pic, n, ok)
		}
		seen[n]++
	}
	if len(seen) != 6 {
		t.Errorf("jewel pictures not distinct: %v", seen)
	}
	for _, tc := range []struct {
		name  string
		class int
		q     InvQuality
		idx   int
		pic   int
		want  string
	}{
		{"ring pic 2", 1, InvPlain, 0, 2, "invrin3"},
		{"charm pic 1", 2, InvPlain, 0, 1, "invch4"},
		{"no pic: base icon", 0, InvPlain, 0, -1, "invgswe"},
		{"past the type's list", 1, InvPlain, 0, 5, "invrin"},
		{"past 6", 0, InvPlain, 0, 9, "invgswe"},
		{"type without pictures", 3, InvPlain, 0, 2, "invssd"},
		{"unique with its own icon keeps it", 1, InvUnique, 0, 3, "invsoj"},
		{"unique without an icon row uses the picture", 0, InvUnique, 7, 1, "invjw2"},
	} {
		if got, _ := db.InvFilePic(tc.class, tc.q, tc.idx, tc.pic); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
	if got, _ := db.InvFile(0, InvPlain, 0); got != "invgswe" {
		t.Errorf("InvFile changed: %q", got)
	}
	files := map[string]bool{}
	for _, f := range db.InvFiles() {
		files[f] = true
	}
	for _, f := range []string{"invjw6", "invrin5", "invch7"} {
		if !files[f] {
			t.Errorf("InvFiles lacks %s (never pre-extracted)", f)
		}
	}
}
