package loot

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
)

// A synthetic item table (no game files): the shape of the real item type tree for the bases the loot rule tests need.
// Columns follow data/items.go. The rows are test data, not game values; only the structure (types, equivalents, class
// restriction, base tier codes, two-handed flag) matters.

func tsv(rows ...string) string { return strings.Join(rows, "\n") + "\n" }

const itemHdr = "name\tcode\ttype\ttype2\t2handed\treqstr\treqdex\tlevelreq\tcost\tinvwidth\tinvheight\tstackable\tgemsockets\tnormcode\tubercode\tultracode\tquest"

func synthDB(t testing.TB) *data.DB {
	t.Helper()
	dir := t.TempDir()
	put := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name+".tsv"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"monstats", "levels", "experience"} {
		put(n, "x\n")
	}
	put("skills", "auralencalc\n")
	put("charstats", tsv("class\tstr\tdex", "Amazon\t20\t25", "Sorceress\t10\t25"))
	put("itemtypes", tsv("ItemType\tCode\tEquiv1\tEquiv2\tClass\tBody\tBodyLoc1",
		"Weapon\tweap\t\t\t\t\t", "Melee\tmele\tweap\t\t\t\t", "Missile\tmiss\tweap\t\t\t\t", "Sword\tswor\tmele\t\t\t\t", "Spear\tspea\tmele\t\t\t\t",
		"Bow\tbow\tmiss\t\t\t\t", "Crossbow\txbow\tmiss\t\t\t\t", "Javelin\tjave\tmele\t\t\t\t", "Amazon Javelin\tajav\tjave\t\tama\t\t",
		"Armor\tarmo\t\t\t\t\t", "Body Armor\ttors\tarmo\t\t\t1\ttors", "Helm\thelm\tarmo\t\t\t\t", "Druid Pelt\tpelt\thelm\t\tdru\t\t",
		"Shield\tshld\tarmo\t\t\t\t", "Ring\tring\t\t\t\t\t", "Rune\trune\t\t\t\t\t", "Gem\tgem\t\t\t\t\t", "Skull\tskul\tgem\t\t\t\t",
		"Book\tbook\t\t\t\t\t", "Potion\tpota\t\t\t\t\t", "Healing Potion\thpot\tpota\t\t\t\t", "Gold\tgold\t\t\t\t\t"))
	put("weapons", tsv(itemHdr,
		"Short Sword\tsst\tswor\t\t0\t0\t0\t0\t100\t1\t3\t0\t2\tsst\tcrs\tgsd\t0",
		"Crystal Sword\tcrs\tswor\t\t0\t43\t0\t25\t1000\t2\t3\t0\t5\tsst\tcrs\tgsd\t0",
		"Claymore\tclm\tswor\t\t1\t47\t0\t0\t300\t2\t3\t0\t5\tclm\t\t\t0",
		"Short Bow\tsbw\tbow\t\t1\t0\t0\t0\t150\t2\t3\t0\t4\tsbw\t\t\t0",
		"Light Crossbow\tlxb\txbow\t\t1\t21\t0\t0\t150\t2\t3\t0\t4\tlxb\t\t\t0",
		"Spear\tspr\tspea\t\t1\t20\t0\t0\t100\t2\t4\t0\t5\tspr\t\t\t0",
		"Javelin\tjav\tjave\t\t0\t0\t0\t0\t60\t1\t3\t1\t0\tjav\t\t\t0",
		"Maiden Javelin\tams\tajav\t\t0\t0\t0\t0\t60\t1\t3\t1\t0\tams\t\t\t0"))
	put("armor", tsv(itemHdr,
		"Quilted Armor\tqui\ttors\t\t0\t12\t0\t0\t64\t2\t3\t0\t4\tqui\tdbp\t\t0",
		"Buckler\tbuc\tshld\t\t0\t12\t0\t0\t30\t2\t2\t0\t1\tbuc\t\t\t0",
		"Cap\tcap\thelm\t\t0\t0\t0\t0\t20\t2\t2\t0\t2\tcap\t\t\t0",
		"Wolf Head\tdr1\tpelt\t\t0\t10\t0\t0\t20\t2\t2\t0\t2\tdr1\t\t\t0"))
	put("misc", tsv(itemHdr,
		"Ring\trin\tring\t\t0\t0\t0\t0\t100\t1\t1\t0\t0\t\t\t\t0",
		"El Rune\tr01\trune\t\t0\t0\t0\t11\t10\t1\t1\t0\t0\t\t\t\t0",
		"Zod Rune\tr33\trune\t\t0\t0\t0\t69\t10\t1\t1\t0\t0\t\t\t\t0",
		"Chipped Skull\tskc\tskul\t\t0\t0\t0\t0\t20\t1\t1\t0\t0\t\t\t\t0",
		"Perfect Skull\tskz\tskul\t\t0\t0\t0\t0\t20\t1\t1\t0\t0\t\t\t\t0",
		"Tome of Town Portal\ttbk\tbook\t\t0\t0\t0\t0\t450\t1\t2\t0\t0\t\t\t\t0",
		"Minor Healing Potion\thp1\thpot\t\t0\t0\t0\t0\t30\t1\t1\t0\t0\t\t\t\t0",
		"Gold\tgld\tgold\t\t0\t0\t0\t0\t0\t1\t1\t1\t0\t\t\t\t0"))
	put("hireling", tsv("Hireling\tSubType\tId\tClass\tAct\tDifficulty\tLevel\tStr\tDex\tWType1\tWType2",
		"Rogue Scout\tFire\t1\t271\t1\t1\t1\t30\t30\tbow\t",
		"Desert Guard\tPrayer\t2\t338\t2\t1\t1\t40\t25\tspea\tpole"))
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

func synthLooter(t testing.TB) (*Looter, *data.DB) {
	t.Helper()
	db := synthDB(t)
	tune := env.DefaultTuning()
	e := &env.Env{ID: "test", Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Data: db, Tune: &tune}
	return New(e, nil), db
}

// unit builds an item unit of the base code.
func synthUnit(t testing.TB, db *data.DB, id uint32, code string, quality int, flags uint32, where d2.Container) d2.Unit {
	t.Helper()
	for i, b := range db.Items {
		if b.Code == code {
			return d2.Unit{ID: id, Addr: 0x1000 + id, Type: d2.TypeItem, Class: i, Item: &d2.Item{Quality: quality, Flags: flags, Where: where, ILvl: 20}}
		}
	}
	t.Fatalf("no base %s", code)
	return d2.Unit{}
}
