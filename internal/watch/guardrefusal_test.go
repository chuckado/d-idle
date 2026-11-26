package watch

import "testing"

func TestGuardRefusalReadsTheTypedAndTheLegacyWording(t *testing.T) {
	for _, c := range []struct {
		note, want string
		ok         bool
	}{
		{"area not allowed now: Forgotten Tower (20): route guard (no merc, clvl 6 < 16)", "Forgotten Tower (20): route guard (no merc, clvl 6 < 16)", true},
		{"refused (route_closed): area not allowed now: Forgotten Tower (20): route guard (no merc, clvl 6 < 16)", "Forgotten Tower (20): route guard (no merc, clvl 6 < 16) (need: route_closed)", true},
		{"refused (too_weak): not ready: clvl 8 < 12", "not ready: clvl 8 < 12 (need: too_weak)", true},
		{"walk failed: no route", "", false},
		{"the word refused (here) is not a typed refusal", "", false},
	} {
		got, ok := guardRefusal(c.note)
		if ok != c.ok || got != c.want {
			t.Errorf("%q: got %q %v, want %q %v", c.note, got, ok, c.want, c.ok)
		}
	}
}
