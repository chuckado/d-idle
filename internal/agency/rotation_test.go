package agency

import (
	"testing"
)

// The plan's rotation and the shadow diff agree with the existing Agrees flag on the corpus the fixtures use
// (two classes); a made-up plan that parts from its prior must be reported.
func TestShadowDiffMatchesAgrees(t *testing.T) {
	env := testEnv(t)
	for _, cl := range []struct {
		class  string
		skills map[string]int
	}{
		{"sorceress", map[string]int{"Fire Bolt": 10, "Frozen Armor": 1}},
		{"paladin", map[string]int{"Might": 5, "Zeal": 5}},
	} {
		a := alloc(cl.class, 20, cl.skills)
		pl, err := Derive(a, env)
		if err != nil {
			t.Fatal(err)
		}
		rot := pl.Rotation()
		if pl.Attack != nil && len(rot.Names("attack")) == 0 {
			t.Errorf("%s: plan has an attack %s but the rotation has none", cl.class, pl.Attack.Skill)
		}
		same := pl.Attack != nil && pl.Attack.Skill == pl.Prior
		if auraSame := (pl.Aura == nil && pl.PriorAura == "") || (pl.Aura != nil && pl.Aura.Skill == pl.PriorAura); same && auraSame && len(ShadowDiff(pl)) != 0 {
			t.Errorf("%s: plan agrees with its prior but the diff says %v", cl.class, ShadowDiff(pl))
		}
	}
	pl := Plan{Prior: "Fire Bolt", Attack: &Choice{Skill: "Ice Bolt"}, Candidates: []Choice{{Skill: "Ice Bolt", Tier: Experimental}}}
	if d := ShadowDiff(pl); len(d) != 1 {
		t.Errorf("a plan that parts from its prior must be reported once: %v", d)
	}
}
