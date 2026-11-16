package merc

import (
	"reflect"
	"testing"

	"d2lod-bot/internal/data"
)

// Item type chains as data.DB.IsA walks them (the real tree: swor < mele < weap, bow < mis? < weap, tors < armo).
var (
	chainBow    = []string{"bow", "bowq", "mis?", "weap"}
	chainXbow   = []string{"xbow", "bowq", "weap"}
	chainSpear  = []string{"spea", "mele", "weap"}
	chainSword  = []string{"swor", "mele", "weap"}
	chainShield = []string{"shld", "armo"}
	chainArmor  = []string{"tors", "armo"}
	chainPelt   = []string{"pelt", "helm", "armo"}
	chainCirc   = []string{"circ", "helm", "armo"}
	chainBoots  = []string{"boot", "armo"}
	chainJave   = []string{"jave", "mele", "weap"}
)

func TestWearTypesTwoHirelings(t *testing.T) {
	for _, tc := range []struct {
		label     string
		chain     []string
		twoHanded bool
		want      []string
	}{
		{"bow: the Rogue only", chainBow, false, []string{"rogue"}},
		{"crossbow: nobody", chainXbow, true, nil},
		{"spear: the Desert Guard only", chainSpear, true, []string{"desert"}},
		{"one-handed sword: Iron Wolf and Barbarian", chainSword, false, []string{"barbarian", "ironwolf"}},
		{"two-handed sword: nobody", chainSword, true, nil},
		{"shield: the Iron Wolf only", chainShield, false, []string{"ironwolf"}},
		{"body armor: every kind", chainArmor, false, []string{"barbarian", "desert", "ironwolf", "rogue"}},
		{"circlet counts as a helm", chainCirc, false, []string{"barbarian", "desert", "ironwolf", "rogue"}},
		{"druid pelt: nobody", chainPelt, false, nil},
		{"boots: nobody", chainBoots, false, nil},
		{"javelin: nobody (U, refused)", chainJave, false, nil},
	} {
		if got := WearTypes(tc.chain, tc.twoHanded); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %v want %v", tc.label, got, tc.want)
		}
	}
}

func TestCanWearUnknownKindNamesIt(t *testing.T) {
	if _, err := CanWear("necromancer", chainBow, false); err == nil || !contains(err.Error(), "necromancer") {
		t.Fatalf("err %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// The table agrees with Allowed for every kind, slot and category (the old merc policy stays the reference).
func TestWearTableMatchesAllowed(t *testing.T) {
	cases := []struct {
		kind  Kind
		slot  Slot
		chain []string
		two   bool
	}{
		{KindHelm, SlotHelm, chainCirc, false}, {KindArmor, SlotArmor, chainArmor, false},
		{KindBow, SlotWeapon, chainBow, false}, {KindSpear, SlotWeapon, chainSpear, true},
		{KindSword1H, SlotWeapon, chainSword, false}, {KindShield, SlotShield, chainShield, false},
	}
	for class, name := range map[int]string{ClassRogue: "rogue", ClassDesert: "desert", ClassIronWolf: "ironwolf", ClassBarb: "barbarian"} {
		for _, c := range cases {
			want := Allowed(class, c.slot, Item{Kind: c.kind})
			got, err := CanWear(name, c.chain, c.two)
			if err != nil || got != want {
				t.Errorf("%s %v: table %v (%v), Allowed %v", name, c.chain[0], got, err, want)
			}
		}
	}
	if !reflect.DeepEqual(WearKinds(), []string{"barbarian", "desert", "ironwolf", "rogue"}) {
		t.Errorf("kinds %v", WearKinds())
	}
}

func TestWTypeGaps(t *testing.T) {
	g, err := WTypeGaps(&data.Hireling{Name: "Rogue Scout", Class: ClassRogue, WType: []string{"bow", ""}})
	if err != nil || len(g) != 0 {
		t.Fatalf("%v %v", g, err)
	}
	g, _ = WTypeGaps(&data.Hireling{Name: "Desert Guard", Class: ClassDesert, WType: []string{"spea", "pole", "axe"}})
	if !reflect.DeepEqual(g, []string{"axe"}) {
		t.Errorf("gaps %v", g)
	}
	if _, err := WTypeGaps(&data.Hireling{Name: "Mystery", Class: 1}); err == nil || !contains(err.Error(), "Mystery") {
		t.Errorf("err %v", err)
	}
}
