package combat

import (
	"reflect"
	"slices"
	"testing"

	"d2lod-bot/internal/delivery"
)

// Golden parity (card dlv-rotation, Q11): for every profile of every class and every character level the Rotation
// says what AttackNames, Buffs and AuraLadder said, for a hero that knows everything and for one that knows nothing
// but the requires skills. The Fighter reads slots, so equality here is equality of the decisions.
func TestRotationEqualsTheProfileLists(t *testing.T) {
	for _, pr := range Profiles() {
		pr := pr
		knowAll := func(string) bool { return true }
		knowNone := func(string) bool { return false }
		for _, known := range []func(string) bool{knowAll, knowNone} {
			for clvl := 1; clvl <= 99; clvl++ {
				r := pr.Rotation(clvl, known)
				if got, want := r.Names(delivery.RoleAttack), pr.AttackNames(clvl, known); !slices.Equal(got, want) && !(len(got) == 0 && len(want) == 0) {
					t.Fatalf("%s clvl %d attack: rotation %v, profile %v", pr.ProfileID(), clvl, got, want)
				}
				if got, want := auraPrefs(r), pr.AuraLadder(clvl, known); !reflect.DeepEqual(got, want) && !(len(got) == 0 && len(want) == 0) {
					t.Fatalf("%s clvl %d aura: rotation %+v, profile %+v", pr.ProfileID(), clvl, got, want)
				}
				if got := r.Groups(); !reflect.DeepEqual(got, pr.Buffs) && !(len(got) == 0 && len(pr.Buffs) == 0) {
					t.Fatalf("%s clvl %d buffs: rotation %v, profile %v", pr.ProfileID(), clvl, got, pr.Buffs)
				}
			}
		}
	}
}

// The test must be able to fail: a rotation with one swapped slot differs (two classes).
func TestRotationParityFailsOnAReorder(t *testing.T) {
	for _, id := range []string{"sorceress", "paladin"} {
		var pr *Profile
		for i := range Profiles() {
			if p := Profiles()[i]; p.Class == id {
				pr = &p
				break
			}
		}
		if pr == nil {
			t.Fatalf("no profile for %s", id)
		}
		all := func(string) bool { return true }
		r := pr.Rotation(60, all)
		names := r.Names(delivery.RoleAttack)
		if len(names) < 2 {
			t.Fatalf("%s: need two attack slots, have %v", id, names)
		}
		r.Slots[0], r.Slots[1] = r.Slots[1], r.Slots[0]
		if slices.Equal(r.Names(delivery.RoleAttack), pr.AttackNames(60, all)) {
			t.Errorf("%s: a reordered rotation must differ from the profile", id)
		}
	}
}
