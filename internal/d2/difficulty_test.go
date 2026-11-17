package d2

import (
	"strings"
	"testing"
)

func TestDoctorLine(t *testing.T) {
	p := &Player{Difficulty: Hell, Level: 70, Area: 107, Resist: Resists{Fire: 90, Cold: 30, Lightning: 75, Poison: 0}}
	l := DoctorLine(p, 85)
	for _, want := range []string{"Hell (byte 2)", "penalty=-100", "res(fire/cold/ltng/pois)=-10/-70/-25/-100", "mlvl=85 gap=15"} {
		if !strings.Contains(l, want) {
			t.Fatalf("%q missing: %s", want, l)
		}
	}
	if strings.Contains(DoctorLine(p, 0), "gap=") || DoctorLine(nil, 0) == "" {
		t.Fatal("optional parts")
	}
	if Normal.ResistPenalty() != 0 || Nightmare.ResistPenalty() != -40 {
		t.Fatal("penalty")
	}
}

func TestResistRulesOrder(t *testing.T) {
	r := ResistTable
	for _, c := range []struct {
		raw, bonus int
		d          Difficulty
		want       int
	}{{90, 0, Nightmare, 50}, {120, 0, Hell, 20}, {190, 20, Hell, 90}, {500, 50, Normal, 95}, {-20, 0, Hell, -100}, {75, 0, Normal, 75}} {
		if got := r.Effective(c.raw, c.bonus, c.d); got != c.want {
			t.Errorf("Effective(%d,%d,%v)=%d want %d", c.raw, c.bonus, c.d, got, c.want)
		}
	}
	// decomp37 Q5: fire 50 percent resist, 20 percent absorb leaves 40 percent of the hit (8000/10000 -> res 60)
	if got := r.Intake(50, 20); got != 60 {
		t.Errorf("Intake(50,20)=%d", got)
	}
	if r.Intake(50, 0) != 50 || r.Intake(-30, 40) != 22 {
		t.Errorf("intake %d %d", r.Intake(50, 0), r.Intake(-30, 40))
	}
}
