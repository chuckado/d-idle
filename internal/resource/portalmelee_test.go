package resource

import "testing"

func TestAbortFightCutCast(t *testing.T) {
	tu := DefaultTuning()
	base := func(adj, lifePct int) Input {
		return Input{Life: lifePct, MaxLife: 100, Mana: 10, MaxMana: 10, HasTP: true, Hostiles: 5, Adjacent: adj, LastTripGold: -1,
			Skills: []Skill{staff}}
	}
	chicken := tu.Chicken(0)
	for _, c := range []struct {
		name string
		in   Input
		want AbortKind
	}{
		{"chicken line, scroll, melee: leave", base(3, chicken), LeaveGame},
		{"chicken line, scroll, nobody adjacent: portal", base(0, chicken), PortalOut},
		{"chicken line, scroll, 2 adjacent: portal", base(2, chicken), PortalOut},
	} {
		if got := AbortFight(c.in, tu).Kind; got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
	thin := base(0, 40)
	thin.Belt = Belt{HP: 1}
	if got := AbortFight(thin, tu); got.Kind != PortalOut {
		t.Skipf("thin-belt setup does not reach escape here: %v", got)
	}
	thin.Adjacent = 3
	if got := AbortFight(thin, tu).Kind; got != LeaveGame {
		t.Errorf("thin belt, 3 adjacent: got %v want LeaveGame", got)
	}
}
