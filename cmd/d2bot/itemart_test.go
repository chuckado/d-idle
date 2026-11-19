package main

import (
	"encoding/json"
	"strings"
	"testing"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/status"
)

// Made-up rows: a staff (Sorceress gear) and a scepter (Paladin gear), a unique override on the scepter.
func artCat() bot.Catalog {
	return bot.NewDataCatalog(&data.DB{
		Items: []*data.ItemBase{
			{Code: "sst", InvFile: "invSst", UniqueInvFile: "invSstU"},
			{Code: "scp", InvFile: "invScp"},
		},
		Uniques: []*data.UniqueItem{{Row: 5, Code: "scp", InvFile: "invScpX"}},
	})
}

func TestItemArt(t *testing.T) {
	cat := artCat()
	unit := func(class, quality, idx int) d2.Unit {
		return d2.Unit{Class: class, Item: &d2.Item{Quality: quality, FileIndex: idx}}
	}
	for _, tc := range []struct {
		name string
		u    d2.Unit
		want string
	}{
		{"sorceress staff, magic", unit(0, 4, 0), "invsst"},
		{"sorceress staff, unique falls back to uniqueinvfile", unit(0, d2.QualityUnique, 5), "invsstu"},
		{"paladin scepter, rare", unit(1, 6, 0), "invscp"},
		{"paladin scepter, unique override", unit(1, d2.QualityUnique, 5), "invscpx"},
		{"paladin scepter, unique wrong row", unit(1, d2.QualityUnique, 9), "invscp"},
		{"set without setinvfile", unit(1, d2.QualitySet, 5), "invscp"},
		{"unknown class", unit(77, 4, 0), ""},
		{"no item data", d2.Unit{Class: 0}, ""},
	} {
		if got := itemArt(cat, tc.u); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

// A catalog without icon names (the test fakes, a bot without tables) leaves art empty, so the page draws cells.
func TestItemArtNoIcons(t *testing.T) {
	if got := itemArt(fakeCat{}, d2.Unit{Item: &d2.Item{}}); got != "" {
		t.Errorf("fakeCat: %q", got)
	}
}

// art is additive: absent from the JSON when empty, so existing consumers and goldens see the same bytes.
func TestStatusItemArtOmitEmpty(t *testing.T) {
	b, _ := json.Marshal(status.Item{Code: "1"})
	if strings.Contains(string(b), "art") {
		t.Errorf("empty art serialised: %s", b)
	}
	b, _ = json.Marshal(status.Item{Code: "1", Art: "invsst"})
	if !strings.Contains(string(b), `"art":"invsst"`) {
		t.Errorf("art missing: %s", b)
	}
}

// A jewel carries its kind word (the page tints the untinted jewel art and names the type when the icon is missing); the
// base code is the key, not the name or the quality, so a magic or unidentified jewel is a jewel too.
func TestInvItemKind(t *testing.T) {
	cat := bot.NewDataCatalog(&data.DB{
		Items: []*data.ItemBase{{Code: "jew", Type: "jewl", InvFile: "invjwl"}, {Code: "rin", Type: "ring", InvFile: "invrin"}, {Code: "zzz", Type: "weird"}},
	})
	for class, want := range map[int]string{0: "jewel", 1: "ring", 2: ""} {
		m, _ := cat.Item(class)
		if m.GoalType != want {
			t.Errorf("class %d: kind %q, want %q", class, m.GoalType, want)
		}
	}
	for _, q := range []int{d2.QualityUnique, 4, 6} { // unique, magic, rare: same icon, no name needed
		if got := itemArt(cat, d2.Unit{Class: 0, Item: &d2.Item{Quality: q, FileIndex: 3}}); got != "invjwl" {
			t.Errorf("jewel quality %d: art %q", q, got)
		}
	}
}

// itemcolour: the tint rides along with the art (a jewel and a coloured armor), never without an icon, and is absent for
// older d2d builds; omitted from the JSON when empty.
func TestStatusItemTint(t *testing.T) {
	b, _ := json.Marshal(status.Item{Code: "1", Art: "invsst"})
	if strings.Contains(string(b), "tint") {
		t.Errorf("empty tint serialised: %s", b)
	}
	b, _ = json.Marshal(status.Item{Code: "1", Art: "invjwl", Tint: "t1c3"})
	if !strings.Contains(string(b), `"tint":"t1c3"`) {
		t.Errorf("tint missing: %s", b)
	}
}

// Item pictures: the six jewel looks are six different art keys, a ring is a second kind, items without variants and
// items with an unknown or absent picture keep the base icon.
func TestItemArtPictures(t *testing.T) {
	jw := [6]string{"invjw1", "invjw2", "invjw3", "invjw4", "invjw5", "invjw6"}
	cat := bot.NewDataCatalog(&data.DB{
		ItemTypes: []*data.ItemType{{Code: "jewl", InvGfx: jw}, {Code: "ring", InvGfx: [6]string{"invrin1", "invrin2", "invrin3", "invrin4", "invrin5"}}, {Code: "swor"}},
		Items:     []*data.ItemBase{{Code: "jew", Type: "jewl", InvFile: "invgswe"}, {Code: "rin", Type: "ring", InvFile: "invrin"}, {Code: "ssd", Type: "swor", InvFile: "invssd"}},
	})
	unit := func(class, q, pic int) d2.Unit { return d2.Unit{Class: class, Item: &d2.Item{Quality: q, Pic: pic}} }
	keys := map[string]bool{}
	for pic := 0; pic < 6; pic++ {
		for _, q := range []int{d2.QualityMagic, 6} { // magic and rare; identified or not does not change the picture
			got := itemArt(cat, unit(0, q, pic+1))
			if got != jw[pic] {
				t.Errorf("jewel pic %d quality %d: %q, want %q", pic, q, got, jw[pic])
			}
			keys[got] = true
		}
	}
	if len(keys) != 6 {
		t.Errorf("six jewel pictures gave %d art keys: %v", len(keys), keys)
	}
	for _, tc := range []struct {
		name string
		u    d2.Unit
		want string
	}{
		{"ring pic 3", unit(1, 4, 4), "invrin4"},
		{"jewel without a picture (older d2d)", unit(0, 4, 0), "invgswe"},
		{"ring pic past the list", unit(1, 4, 6), "invrin"},
		{"sword ignores a picture", unit(2, 4, 3), "invssd"},
	} {
		if got := itemArt(cat, tc.u); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}
