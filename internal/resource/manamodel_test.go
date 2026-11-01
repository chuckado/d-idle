package resource

import "testing"

// V-d2d DRILLS15 M4: max mana 45, 750 ticks (30 s) = +2250/256, Warmth 1 (+30 percent) the same.
func TestManaRegenMeasured(t *testing.T) {
	for _, tc := range []struct {
		max, bonus, frame int
	}{
		{45, 0, 3}, {45, 30, 3}, // V-d2d M4
		{35, 0, 2}, {100, 0, 8}, {250, 0, 21}, // trace section 4 (arithmetic on V read)
		{1, 0, 1}, {0, 0, 0}, {250, 100, 42},
	} {
		if got := ManaPerFrame256(tc.max, tc.bonus); got != tc.frame {
			t.Errorf("max %d bonus %d: %d a frame, want %d", tc.max, tc.bonus, got, tc.frame)
		}
	}
	got := ManaPerSec(45, 0) * 30 * 256
	if got < 2249 || got > 2251 {
		t.Fatalf("750 ticks: %.1f/256, want 2250", got)
	}
	if s := (Input{MaxMana: 45}).RestSecs(45); s < 152 || s > 155 {
		t.Fatalf("fill 45: %.1f s, want about 153", s)
	}
}

func TestManaBookBackends(t *testing.T) {
	for _, tc := range []struct {
		name     string
		m        ManaBooking
		at       float64 // time of the first Take
		cut      bool
		wantEarl int
		wantLate int
	}{
		{"d2d one tick", BookD2D, 0.05, false, 10, 0},
		{"d2d not yet", BookD2D, 0.02, false, 0, 10},
		{"d2d cut cast still pays", BookD2D, 0.05, true, 10, 0},
		{"game.exe not before action frame", BookGameExe, 0.2, false, 0, 10},
		{"game.exe at action frame", BookGameExe, 0.33, false, 10, 0},
		{"game.exe cut cast never pays", BookGameExe, 0.2, true, 0, 0},
	} {
		b := ManaBook{Model: tc.m}
		b.Add(0, 10)
		if tc.cut {
			b.Cut()
		}
		if got := b.Take(tc.at); got != tc.wantEarl {
			t.Errorf("%s: first take %d, want %d", tc.name, got, tc.wantEarl)
		}
		if got := b.Take(5); got != tc.wantLate {
			t.Errorf("%s: late take %d, want %d", tc.name, got, tc.wantLate)
		}
	}
}

// A cast booked at the command, game.exe: the foreign-loss check must not see a gain before the mana drops.
func TestManaBurnTrackerBookedAtPayment(t *testing.T) {
	var m ManaBurnTracker
	b := ManaBook{Model: BookGameExe}
	m.Observe(0, 100, 100, 0)
	b.Add(0.1, 30)
	// the command is out, the pool is still full: nothing may be taken off the samples yet
	if got := m.Observe(0.2, 100, 100, b.Take(0.2)); got != ManaOK {
		t.Fatal(got)
	}
	// the action frame: the pool dropped by our cast, which is not foreign loss
	if got := m.Observe(0.5, 70, 100, b.Take(0.5)); got != ManaOK {
		t.Fatal(got)
	}
}

// V read (B3): a mana potion arrives over 128 frames, not at once.
func TestSimManaPotionOverTime(t *testing.T) {
	w := &world{sc: Scenario{MaxMana: 100, MaxLife: 100}, life: 100, mana: 10}
	w.addManaPot(40)
	w.advance(1, 0)
	if w.mana < 15 || w.mana > 19 {
		t.Fatalf("after 1 s: %.1f, want about 18 (40 / 5.12 s)", w.mana)
	}
	w.advance(5, 0)
	if w.mana < 49.9 || w.mana > 50.1 {
		t.Fatalf("after 6 s: %.1f, want 50", w.mana)
	}
}
