package data

import "testing"

func TestTeleportRule(t *testing.T) {
	cases := []struct {
		name string
		l    Level
		rule TeleportRule
		can  bool
	}{
		{"refused", Level{ID: 2, Teleport: 0}, TeleportRefused, false},
		{"allowed", Level{ID: 2, Teleport: 1}, TeleportAllowed, true},
		{"clear line (Duriel's Lair)", Level{ID: 73, Teleport: 2}, TeleportClearLine, true},
		{"unknown value is refused", Level{ID: 2, Teleport: 7}, TeleportRefused, false},
		{"town refuses whatever the column says", Level{ID: 40, Teleport: 1}, TeleportAllowed, false},
	}
	for _, c := range cases {
		if got := c.l.TeleportRule(); got != c.rule {
			t.Errorf("%s: rule %v want %v", c.name, got, c.rule)
		}
		if got := CanTeleportIn(&c.l); got != c.can {
			t.Errorf("%s: CanTeleportIn %v want %v", c.name, got, c.can)
		}
	}
	if CanTeleportIn(nil) {
		t.Error("nil level")
	}
	for _, a := range []int{1, 40, 75, 103, 109} {
		if CanTeleportIn(&Level{ID: a, Teleport: 1}) {
			t.Errorf("town %d", a)
		}
	}
}

// With a real install the lair must be the only row with value 2 and no town row may matter (SKTELE: all other rows are 1).
func TestRealLairTeleport(t *testing.T) {
	db := realDB(t)
	for _, l := range db.Levels {
		if (l.Teleport == 2) != (l.ID == 73) {
			t.Errorf("area %d (%s): Teleport %d", l.ID, l.Key, l.Teleport)
		}
	}
	if l := db.AreaByID(73); l == nil || l.TeleportRule() != TeleportClearLine || !CanTeleportIn(l) {
		t.Errorf("lair %+v", l)
	}
}
