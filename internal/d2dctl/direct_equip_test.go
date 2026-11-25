package d2dctl

import (
	"reflect"
	"testing"
)

func TestParseDirectShapes(t *testing.T) {
	one, err := parseDirect(&Reply{Rest: "equipped 1 id=7 removed=aa11,bb22 removed_norm=aa,bb"})
	if err != nil || len(one) != 1 || one[0].Unit != 7 || !reflect.DeepEqual(one[0].Removed, []string{"aa11", "bb22"}) {
		t.Fatalf("equip %+v %v", one, err)
	}
	none, _ := parseDirect(&Reply{Rest: "equipped 1 id=8 removed= removed_norm="})
	if len(none[0].Removed) != 0 {
		t.Fatalf("empty removed: %+v", none)
	}
	un, _ := parseDirect(&Reply{Rest: "unequipped 3 id=9 hex=cc33 norm=cc"})
	if un[0].Removed[0] != "cc33" {
		t.Fatalf("unequip %+v", un)
	}
	set, err := parseDirect(&Reply{Lines: []string{"piece 0 slot=3 ok id=11 removed=dd", "piece 1 slot=4 err level needs level 30 ok really"}, Rest: "wearset applied=1/2 removed=dd"})
	if err != nil || len(set) != 2 || !set[0].OK || set[1].OK || set[1].Code != "level" || set[1].Why != "needs level 30 ok really" {
		t.Fatalf("set %+v %v", set, err)
	}
}

func TestDirectFakeSwapAndTwoHander(t *testing.T) {
	f := &DirectFake{Worn: map[string]string{"head": "aa", "lhand": "bb"}, Pushes: func(hex, slot string) []string {
		if hex == "2h" {
			return []string{"lhand"}
		}
		return nil
	}}
	r, err := f.EquipDirect(nil, "2h", "rhand", "o1")
	if err != nil || len(r.Removed) != 1 || r.Removed[0] != "bb" {
		t.Fatalf("%+v %v", r, err)
	}
	f.Measured = true
	if _, err := f.EquipDirect(nil, "cc", "head", "o2"); err != ErrDirectMeasured {
		t.Fatal(err)
	}
}
