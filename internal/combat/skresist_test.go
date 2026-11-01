package combat

import (
	"math"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// every damage type, taken from the type table so a new type is covered without a literal list.
func allTypes() []data.DamageType {
	var out []data.DamageType
	for d := data.Physical; d <= data.Poison; d++ {
		out = append(out, d)
	}
	return out
}

// Q8: the floor is -100 for all six damage types, so the factor never exceeds 2.
func TestExpFactorFloor(t *testing.T) {
	for _, tc := range []struct {
		res  int
		want float64
	}{{-30, 1.3}, {-100, 2}, {-101, 2}, {-250, 2}, {0, 1}, {99, 0.01}, {100, 0}, {140, 0}} {
		if got := expFactor(tc.res); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("expFactor(%d) = %v want %v", tc.res, got, tc.want)
		}
	}
	for _, d := range allTypes() {
		var tg Target
		tg.HasRes = true
		tg.Res[d] = -180
		if got := DamageFactor(SkillSpec{Elem: d}, &tg); got != 2 {
			t.Errorf("type %d: factor %v want 2", d, got)
		}
	}
}

// Q4: pierce applies only while the total is under 100; Q2/Q5: immune (>= 100) does 0; Q8: floor after pierce.
func TestEffResPierce(t *testing.T) {
	for _, tc := range []struct {
		name        string
		res, pierce int
		wantRes     int
		wantFactor  float64
	}{
		{"below 100 pierced", 75, 20, 55, 0.45},
		{"just below 100", 99, 20, 79, 0.21},
		{"exactly 100 stays immune", 100, 20, 100, 0},
		{"above 100 stays immune", 110, 50, 110, 0},
		{"no pierce", 75, 0, 75, 0.25},
		{"pierce into the floor", -90, 50, -100, 2},
		{"already below the floor", -150, 0, -100, 2},
	} {
		for _, d := range allTypes() {
			var tg Target
			tg.HasRes = true
			tg.Res[d], tg.Pierce[d] = tc.res, tc.pierce
			if got := tg.EffRes(d); got != tc.wantRes {
				t.Errorf("%s type %d: EffRes %d want %d", tc.name, d, got, tc.wantRes)
			}
			got := DamageFactor(SkillSpec{Elem: d}, &tg)
			if d == data.Physical {
				continue // Stone Skin path is unrelated; physical has no pierce stat but EffRes is generic
			}
			if math.Abs(got-tc.wantFactor) > 1e-9 {
				t.Errorf("%s type %d: factor %v want %v", tc.name, d, got, tc.wantFactor)
			}
			if (tc.wantFactor == 0) != tg.Immune(d) {
				t.Errorf("%s type %d: Immune %v", tc.name, d, tg.Immune(d))
			}
		}
	}
}

// Q4: pierce comes from the player's stat list; only the types that have a pierce stat are read.
func TestPierceFromStats(t *testing.T) {
	st := d2.Stats{uint32(pierceStat[data.Cold]): 20, uint32(pierceStat[data.Fire]): 5}
	p := PierceFromStats(st)
	if p[data.Cold] != 20 || p[data.Fire] != 5 || p[data.Lightning] != 0 || p[data.Physical] != 0 || p[data.Magic] != 0 {
		t.Fatalf("%v", p)
	}
	if PierceFromStats(nil) != [6]int{} {
		t.Fatal("empty list")
	}
}

// Q4 + Q2: a caster whose preferred element has pierce keeps it against a 75 target (45 % vs fire 60 % is within
// resistTolerance), while without pierce she switches (25 % < 0.6 * 60 %); an immune element is never picked even
// when its pierce is large. Elements are taken from the table, any two elemental types give the same outcome.
func TestPickByResistPierce(t *testing.T) {
	elems := []data.DamageType{data.Fire, data.Lightning, data.Cold}
	for _, a := range elems {
		for _, b := range elems {
			if a == b {
				continue
			}
			order := []SkillSpec{{Name: "pref", Elem: a, Cost: 1}, {Name: "alt", Elem: b, Cost: 1}}
			mk := func(pierce int) *Target {
				tg := &Target{HasRes: true}
				tg.Res[a], tg.Res[b] = 75, 40
				tg.Pierce[a] = pierce
				return tg
			}
			if got, _ := PickByResist(order, mk(20), 50, nil); got.Name != "pref" {
				t.Errorf("%d over %d with pierce: got %s", a, b, got.Name)
			}
			if got, _ := PickByResist(order, mk(0), 50, nil); got.Name != "alt" {
				t.Errorf("%d over %d without pierce: got %s", a, b, got.Name)
			}
			imm := &Target{HasRes: true}
			imm.Res[a], imm.Pierce[a] = 110, 60
			imm.Res[b] = 30
			if got, why := PickByResist(order, imm, 50, nil); why != PickOK || got.Name != "alt" {
				t.Errorf("%d immune with pierce: got %s %v", a, got.Name, why)
			}
		}
	}
}

// Q8 through ImmuneWorkaround: a type under the floor is ranked with factor 2, not more.
func TestFloorRanking(t *testing.T) {
	tg := &Target{HasRes: true}
	tg.Res[data.Fire], tg.Res[data.Cold], tg.Res[data.Lightning] = -400, -100, 100
	if expFactor(tg.EffRes(data.Fire)) != expFactor(tg.EffRes(data.Cold)) {
		t.Fatal("-400 and -100 must rank the same")
	}
}

// B3 is left alone (sk-resist-trace.md section 6 tags it S): the fallback keeps its current rule.
func TestAffixResistsFallbackUnchanged(t *testing.T) {
	var r [6]int
	r[data.Fire], r[data.Cold], r[data.Lightning] = 100, 100, 30
	got := AffixResists(r, AffixLightningEnchanted)
	if got[data.Lightning] != 30 {
		t.Fatalf("third immunity must be skipped: %v", got)
	}
	r = [6]int{}
	if got := AffixResists(r, AffixColdEnchanted); got[data.Cold] != enchantedBonus {
		t.Fatalf("%v", got)
	}
}
