package d2

import "testing"

func TestSkillBonus(t *testing.T) {
	const sor = Sorceress
	own := func(tab int) SkillRef { return SkillRef{Own: true, Tab: tab} }
	other := SkillRef{}
	shrine := StateSet{}
	shrine[StateShrineSkill>>5] |= 1 << (StateShrineSkill & 31)
	cases := []struct {
		name   string
		st     Stats
		states StateSet
		id     int
		ref    SkillRef
		want   int
	}{
		{"none", Stats{}, StateSet{}, 38, own(2), 0},
		{"all skills", Stats{subKey(127, 0): 1}, StateSet{}, 38, own(2), 1},
		{"all skills other class", Stats{subKey(127, 0): 1}, StateSet{}, 70, other, 1},
		{"class own", Stats{subKey(83, 1): 2}, StateSet{}, 38, own(2), 2},
		{"class not other", Stats{subKey(83, 1): 2}, StateSet{}, 70, other, 0},
		{"class other id ignored", Stats{subKey(83, 2): 2}, StateSet{}, 38, own(2), 0},
		{"tab matches", Stats{subKey(188, 1*8+2-1): 3}, StateSet{}, 38, own(2), 3},
		{"tab other tab", Stats{subKey(188, 1*8+2-1): 3}, StateSet{}, 36, own(3), 0},
		{"tab other class", Stats{subKey(188, 1*8+2-1): 3}, StateSet{}, 38, other, 0},
		{"no tab skill", Stats{subKey(188, 1*8-1): 3}, StateSet{}, 38, own(0), 0},
		{"oskill own capped", Stats{subKey(97, 38): 5}, StateSet{}, 38, own(2), 3},
		{"oskill own under cap", Stats{subKey(97, 38): 2}, StateSet{}, 38, own(2), 2},
		{"oskill other full", Stats{subKey(97, 70): 5}, StateSet{}, 70, other, 5},
		{"oskill other skill", Stats{subKey(97, 70): 5}, StateSet{}, 71, other, 0},
		{"elem fire", Stats{subKey(126, 1): 3}, StateSet{}, 36, SkillRef{Own: true, Tab: 1, Elem: ElemFire}, 3},
		{"elem not other", Stats{subKey(126, 1): 3}, StateSet{}, 38, SkillRef{Own: true, Tab: 2, Elem: ElemLight}, 0},
		{"elem skill without element", Stats{subKey(126, 1): 3}, StateSet{}, 61, own(1), 0},
		{"single", Stats{subKey(107, 36): 1}, StateSet{}, 36, own(1), 1},
		{"single other skill", Stats{subKey(107, 36): 1}, StateSet{}, 37, own(1), 0},
		{"shrine", Stats{}, shrine, 38, own(2), 2},
		{"shrine other class", Stats{}, shrine, 70, other, 2},
		{"all sources", Stats{subKey(127, 0): 1, subKey(83, 1): 2, subKey(188, 9): 3, subKey(97, 38): 7,
			subKey(126, 2): 4, subKey(107, 38): 1}, shrine, 38, SkillRef{Own: true, Tab: 2, Elem: ElemLight}, 1 + 2 + 2 + 3 + 3 + 4 + 1},
	}
	for _, c := range cases {
		if got := SkillBonus(c.st, c.states, sor, c.id, c.ref); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
}

// Without bonus stats Effective is the base points, exactly like Skill.Effective of a skill with no 107.
func TestPlayerEffectiveNoBonus(t *testing.T) {
	p := &Player{Class: Sorceress, Stats: Stats{}, Skills: []Skill{{ID: 36, Level: 4}, {ID: 38, Level: 0}}}
	for _, sk := range p.Skills {
		if got := p.Effective(sk.ID, SkillRef{Own: true, Tab: 1, Elem: ElemFire}); got != sk.Effective() {
			t.Errorf("skill %d: %d, want %d", sk.ID, got, sk.Effective())
		}
	}
}

// Synergies read base points: Base ignores every bonus source.
func TestBaseIgnoresBonus(t *testing.T) {
	p := &Player{Class: Sorceress, Stats: Stats{subKey(127, 0): 2, subKey(107, 38): 1}, Skills: []Skill{{ID: 38, Level: 5, Bonus: 1}}}
	if p.Base(38) != 5 || p.Base(99) != 0 {
		t.Errorf("Base = %d / %d", p.Base(38), p.Base(99))
	}
	if got := p.Effective(38, SkillRef{Own: true, Tab: 2}); got != 8 {
		t.Errorf("Effective = %d, want 8", got)
	}
}
