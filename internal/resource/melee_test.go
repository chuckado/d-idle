package resource

import (
	"math"
	"testing"
)

// Expected values come from tools/emu/sk_melee.py's reading functions (hit_chance, dmg_reading, muldiv), run offline
// without the emulator on the same inputs (docs/skmeleebot.md); the rule is FUN_0057d9b0 / FUN_0057b420 (V run).
func TestHitChance(t *testing.T) {
	for _, c := range []struct {
		name                           string
		ar, pct, def, alvl, dlvl, want int
	}{
		{"high AR, low defense: capped at 95", 95, 0, 15, 8, 3, 95},
		{"equal levels", 95, 0, 15, 8, 8, 86},
		{"defense bites", 95, 0, 100, 8, 8, 48},
		{"percent and levels", 200, 30, 50, 20, 10, 95},
		{"no AR, no defense", 0, 0, 0, 5, 5, 95},
		{"negative AR: floor 5 (D5)", -40, 0, 10, 10, 10, 5},
		{"negative percent", 100, -50, 40, 10, 10, 55},
		{"outleveled: floor 5", 95, 0, 15, 1, 30, 5},
	} {
		if got := HitChance(c.ar, c.pct, c.def, c.alvl, c.dlvl); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
}

func TestMeanDamage(t *testing.T) {
	for _, c := range []struct {
		name string
		b    Blow
		f    Foe
		want float64
	}{
		// dmg_reading: 1-5 weapon, p 10: [281, 1408) 256ths
		{"staff p 10", Blow{Weapon: true, Min: 1, Max: 5, DmgPct: 10}, Foe{}, 3.296875},
		{"staff p 60", Blow{Weapon: true, Min: 1, Max: 5, DmgPct: 60}, Foe{}, 4.796875},
		{"flat 111 and p 35", Blow{Weapon: true, Min: 2, Max: 9, Flat: 1, DmgPct: 35}, Foe{}, 8.771484375},
		{"p floored at -90", Blow{Weapon: true, Min: 1, Max: 2, DmgPct: -95}, Foe{}, 0.150390625},
		{"fixed damage gets +1 point of range (max <= min)", Blow{Weapon: true, Min: 5, Max: 5}, Foe{}, 5.498046875},
		{"stat 17 / 18 not used under reading A", Blow{Weapon: true, Min: 6, Max: 20, Flat: 2, DmgPct: 110}, Foe{}, 31.49609375},
		// D4: barehanded, min 1 / max 2, Strength as a percent
		{"barehanded str 30", Blow{Min: 0, Max: 0, Str: 30}, Foe{}, 1.9453125},
		// D2: Blunt +50 points against undead (reading: p 10 + 50 = the p 60 row)
		{"blunt vs undead", Blow{Weapon: true, Blunt: true, Min: 1, Max: 5, DmgPct: 10}, Foe{Undead: true}, 4.796875},
		{"blunt vs living: no bonus", Blow{Weapon: true, Blunt: true, Min: 1, Max: 5, DmgPct: 10}, Foe{}, 3.296875},
		{"edged vs undead: no bonus", Blow{Weapon: true, Min: 1, Max: 5, DmgPct: 10}, Foe{Undead: true}, 3.296875},
	} {
		if got := c.b.MeanDamage(c.f); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: %.9f, want %.9f", c.name, got, c.want)
		}
	}
}

func TestStatsAndBlunt(t *testing.T) {
	if got := AttackRating(0, 25, 5); got != 95 { // stat 19 + (dex - 7) x 5 + ToHitFactor (FUN_00622560)
		t.Errorf("AR %d", got)
	}
	if got := StatDmgPct(25, 10, 20, 100, 50, 10); got != 55 {
		t.Errorf("pct %d", got)
	}
	isA := func(typ, anc string) bool { return typ == "staf" && anc == BluntType || typ == anc }
	if !IsBlunt(isA, "staf") || IsBlunt(isA, "swor") || IsBlunt(nil, "staf") {
		t.Error("blunt chain")
	}
}

// Swing is a Bernoulli(HitPct) times the rolled range: the mean over many swings matches Expected.
func TestSwingMean(t *testing.T) {
	b := Blow{AR: 95, Level: 8, Weapon: true, Min: 1, Max: 5, DmgPct: 10}
	f := Foe{Defense: 100, Level: 8} // 48 % to hit
	r := rng(11)
	rnd := func(n int) int { return int(r.next() % uint64(n)) }
	sum := 0.0
	const n = 200000
	for i := 0; i < n; i++ {
		sum += b.Swing(f, rnd)
	}
	if got, want := sum/n, b.Expected(f); math.Abs(got-want) > 0.02 {
		t.Errorf("mean %.3f, want %.3f", got, want)
	}
}

// B2: the AR term. Same level pack, but a defense that the character's AR cannot reach is no longer weak.
// B3: a Blunt weapon raises the melee rung's pack limit against undead only. Two builds: a staff caster and a
// sword fighter with different tables (no skill or weapon name in the code).
func TestWeakPackAndUndeadRung(t *testing.T) {
	tu := DefaultTuning()
	caster := Blow{AR: 95, Level: 8, Weapon: true, Min: 1, Max: 5, DmgPct: 10, Blunt: true}
	fighter := Blow{AR: 400, Level: 8, Weapon: true, Min: 8, Max: 20, DmgPct: 60}
	staff := Skill{Name: "x", Melee: true, Ready: true}
	mk := func(b *Blow, f Foe, hostiles int, weak bool) Input {
		return Input{Mana: 0, MaxMana: 40, Life: 90, MaxLife: 100, Hostiles: hostiles, WeakPack: weak, Blow: b, Foe: &f, Skills: []Skill{staff}}
	}
	cases := []struct {
		name string
		in   Input
		want AttackKind
	}{
		// to-hit 86 % (see TestHitChance): weak stays weak
		{"B2 weak, AR holds", mk(&caster, Foe{Defense: 15, Level: 8}, 5, true), Melee},
		// defense 100: 48 % < WeakMinHitPct: not weak, 5 hostiles > pack limit 3: no standing (retreat)
		{"B2 weak by level, AR does not", mk(&caster, Foe{Defense: 100, Level: 8}, 5, true), Retreat},
		{"B2 no model: level only", mk(nil, Foe{Defense: 100, Level: 8}, 5, true), Melee},
		{"B2 fighter reaches the same defense", mk(&fighter, Foe{Defense: 100, Level: 8}, 5, true), Melee},
		// B3: limit 3 -> 3 x 1.45 = 4 vs undead for a blunt caster; living foes keep 3
		{"B3 blunt vs undead, 4 hostiles", mk(&caster, Foe{Defense: 15, Level: 8, Undead: true}, 4, false), Melee},
		{"B3 blunt vs living, 4 hostiles", mk(&caster, Foe{Defense: 15, Level: 8}, 4, false), Retreat},
		{"B3 edged vs undead, 4 hostiles", mk(&fighter, Foe{Defense: 15, Level: 8, Undead: true}, 4, false), Retreat},
	}
	for _, c := range cases {
		if got := AttackChoice(c.in, tu).Kind; got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	// gain: 1-5 staff p 10 vs p 60 (python rows above): 4.796875 / 3.296875
	if g := caster.undeadGain(Foe{Undead: true}); math.Abs(g-4.796875/3.296875) > 1e-9 {
		t.Errorf("gain %.6f", g)
	}
}

// B1 before / after: the stock Blood Moor scenario with the flat 3 and with the blow model. The flat model lands 3
// of 3 a swing; the model lands 95 % (HitChance) of a mean 3.297 (TestMeanDamage) = 3.13 a swing in expectation.
func TestStockBlowVersusFlat(t *testing.T) {
	sc := BloodMoorSorc()
	if got := sc.Blow.Expected(sc.Foe); math.Abs(got-0.95*3.296875) > 1e-9 {
		t.Fatalf("expected %.4f per swing", got)
	}
	flat := sc
	flat.Blow = nil
	tu := DefaultTuning()
	a, b := Replay(flat, ResourcePolicy(), tu), Replay(sc, ResourcePolicy(), tu)
	t.Logf("flat 3: kills %d melee %d; blow model: kills %d melee %d", a.Kills, a.MeleeHits, b.Kills, b.MeleeHits)
	if b.Deaths > 0 || b.Kills == 0 {
		t.Fatalf("blow model replay: %+v", b)
	}
	// an undead pack: the same staff hits harder (3.297 -> 4.797 a landed blow)
	und := sc
	und.Foe.Undead = true
	if und.Blow.Expected(und.Foe) <= sc.Blow.Expected(sc.Foe) {
		t.Error("undead must raise a blunt weapon's expected damage")
	}
}
